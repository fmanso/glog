package db

import (
	"bytes"
	"encoding/gob"
	"log"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

type wordIndex struct {
	db               *bolt.DB
	bucketWordsIndex []byte
	bucketDocsIndex  []byte
}

func newWordIndex(db *bolt.DB) (*wordIndex, error) {
	wordsIndexKey := []byte("word_index")
	docsIndexKey := []byte("doc_index")

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(wordsIndexKey)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(docsIndexKey)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &wordIndex{
		db:               db,
		bucketWordsIndex: wordsIndexKey,
		bucketDocsIndex:  docsIndexKey,
	}, nil
}

func (wi *wordIndex) save(tx *bolt.Tx, doc *DocDb) error {
	err := wi.saveWordIndex(tx, doc)
	if err != nil {
		return err
	}

	err = wi.saveDocIndex(tx, doc)
	if err != nil {
		return err
	}

	return nil
}

func (wi *wordIndex) saveDocIndex(tx *bolt.Tx, doc *DocDb) error {
	words := getWords(doc)
	docsBucket := tx.Bucket(wi.bucketDocsIndex)
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(words)
	if err != nil {
		return err
	}

	err = docsBucket.Put([]byte(doc.ID.String()), buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (wi *wordIndex) saveWordIndex(tx *bolt.Tx, doc *DocDb) error {
	words := getWords(doc)
	wordsBucket := tx.Bucket(wi.bucketWordsIndex)
	for _, word := range words {
		existing := wordsBucket.Get([]byte(word))
		docIds := decodeUUIDSet(existing)
		docIds[doc.ID] = struct{}{}

		log.Println("Indexing word:", word, "for document ID:", doc.ID)
		data, err := encodeUUIDSet(docIds)
		if err != nil {
			return err
		}

		err = wordsBucket.Put([]byte(word), data)
		if err != nil {
			return err
		}
	}
	return nil
}

func decodeUUIDSet(data []byte) map[uuid.UUID]struct{} {
	var docIDs map[uuid.UUID]struct{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&docIDs)
	if err != nil {
		return make(map[uuid.UUID]struct{})
	}
	return docIDs
}

func encodeUUIDSet(docIDs map[uuid.UUID]struct{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(docIDs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// getWords extracts unique words from the document's blocks.
func getWords(doc *DocDb) []string {
	wordMap := map[string]struct{}{}

	for _, block := range doc.Blocks {
		words := strings.Fields(block.Content)
		for _, word := range words {
			cleanedWord := strings.ToLower(strings.Trim(word, ".,!?\"'()[]{}<>:;"))
			if cleanedWord != "" {
				wordMap[cleanedWord] = struct{}{}
			}
		}
	}

	words := make([]string, 0, len(wordMap))
	for word := range wordMap {
		words = append(words, word)
	}
	return words
}

func (wi *wordIndex) Search(tx *bolt.Tx, query string) ([]uuid.UUID, error) {
	wordsBucket := tx.Bucket(wi.bucketWordsIndex)
	queryWords := strings.Fields(strings.ToLower(query))
	resultSet := make(map[uuid.UUID]struct{})

	for i, word := range queryWords {
		cleanedWord := strings.Trim(word, ".,!?\"'()[]{}<>:;")
		if cleanedWord == "" {
			continue
		}

		data := wordsBucket.Get([]byte(cleanedWord))
		docIDs := decodeUUIDSet(data)

		log.Printf("Searching for word: %s, found %d documents\n", cleanedWord, len(docIDs))
		if i == 0 {
			for id := range docIDs {
				resultSet[id] = struct{}{}
			}
		} else {
			for id := range resultSet {
				if _, exists := docIDs[id]; !exists {
					delete(resultSet, id)
				}
			}
		}
	}

	results := make([]uuid.UUID, 0, len(resultSet))
	for id := range resultSet {
		results = append(results, id)
	}

	return results, nil
}
