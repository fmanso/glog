package db

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/analysis/analyzer/standard"
	"github.com/blevesearch/bleve/v2/mapping"
	"github.com/google/uuid"
)

const (
	bleveFieldTitle   = "title"
	bleveFieldContent = "content"

	titleBoost = 2.0
)

type bleveDoc struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Date    string `json:"date,omitempty"`
}

type bleveSearch struct {
	path  string
	index bleve.Index
}

func openBleveSearch(path string) (*bleveSearch, error) {
	idx, err := bleve.Open(path)
	if err == nil {
		return &bleveSearch{path: path, index: idx}, nil
	}
	if !errors.Is(err, bleve.ErrorIndexPathDoesNotExist) {
		return nil, err
	}

	indexMapping := mapping.NewIndexMapping()
	indexMapping.DefaultAnalyzer = standard.Name

	// Ensure fields exist and are indexed as text.
	docMapping := mapping.NewDocumentMapping()
	textMapping := mapping.NewTextFieldMapping()
	textMapping.Store = false
	textMapping.Index = true
	docMapping.AddFieldMappingsAt(bleveFieldTitle, textMapping)
	docMapping.AddFieldMappingsAt(bleveFieldContent, textMapping)

	indexMapping.DefaultMapping = docMapping

	idx, err = bleve.New(path, indexMapping)
	if err != nil {
		return nil, err
	}
	return &bleveSearch{path: path, index: idx}, nil
}

func (s *bleveSearch) Close() error {
	if s == nil || s.index == nil {
		return nil
	}
	return s.index.Close()
}

func (s *bleveSearch) DeleteIndexDir() error {
	if s == nil {
		return nil
	}
	return os.RemoveAll(s.path)
}

func (s *bleveSearch) IndexDoc(doc *DocDb) error {
	if s == nil || s.index == nil {
		return nil
	}

	blocks := make([]string, 0, len(doc.Blocks))
	for _, block := range doc.Blocks {
		if block == nil {
			continue
		}
		blocks = append(blocks, block.Content)
	}

	bdoc := bleveDoc{
		Title:   doc.Title,
		Content: strings.Join(blocks, "\n"),
		Date:    doc.Date,
	}

	return s.index.Index(doc.ID.String(), bdoc)
}

func (s *bleveSearch) Search(query string) ([]uuid.UUID, error) {
	if s == nil || s.index == nil {
		return nil, nil
	}

	phrases, tokens := parseSearchQuery(query)

	conj := bleve.NewConjunctionQuery()

	for _, phrase := range phrases {
		phrase = strings.TrimSpace(phrase)
		if phrase == "" {
			continue
		}

		contentQ := bleve.NewPhraseQuery(strings.Fields(phrase), bleveFieldContent)
		titleQ := bleve.NewPhraseQuery(strings.Fields(phrase), bleveFieldTitle)
		titleQ.SetBoost(titleBoost)
		conj.AddQuery(bleve.NewDisjunctionQuery(titleQ, contentQ))
	}

	for _, token := range tokens {
		fuzziness := 1
		if len(token) >= 5 {
			fuzziness = 2
		}

		contentQ := bleve.NewFuzzyQuery(token)
		contentQ.SetField(bleveFieldContent)
		contentQ.SetFuzziness(fuzziness)

		titleQ := bleve.NewFuzzyQuery(token)
		titleQ.SetField(bleveFieldTitle)
		titleQ.SetFuzziness(fuzziness)
		titleQ.SetBoost(titleBoost)

		conj.AddQuery(bleve.NewDisjunctionQuery(titleQ, contentQ))
	}

	searchRequest := bleve.NewSearchRequest(conj)
	searchRequest.Fields = []string{}
	searchRequest.Size = 10000 // Set a high limit for search results
	searchResult, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	ids := make([]uuid.UUID, 0, len(searchResult.Hits))
	for _, hit := range searchResult.Hits {
		id, err := uuid.Parse(hit.ID)
		if err != nil {
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}

var quoteRe = regexp.MustCompile(`"([^"]+)"`)

func parseSearchQuery(query string) (phrases []string, tokens []string) {
	matches := quoteRe.FindAllStringSubmatch(query, -1)
	for _, match := range matches {
		if len(match) > 1 {
			phrases = append(phrases, strings.ToLower(match[1]))
		}
	}

	stripped := quoteRe.ReplaceAllString(query, " ")
	for _, token := range strings.Fields(strings.ToLower(stripped)) {
		if len(token) < 2 {
			continue
		}
		tokens = append(tokens, token)
	}

	return phrases, tokens
}

func bleveIndexPath(boltPath string) string {
	return filepath.Clean(boltPath + ".bleve")
}
