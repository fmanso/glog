package db

import (
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestSearchBleve_ContentTermMatch(t *testing.T) {
	store, err := NewDocumentStore("./test_bleve_content_term.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		_ = store.Close()
		_ = os.Remove("./test_bleve_content_term.db")
		_ = os.RemoveAll("./test_bleve_content_term.db.bleve")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Unrelated title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{{
			ID:      domain.BlockID(uuid.New()),
			Content: "hello world",
			Indent:  0,
		}},
	}

	if err := store.Save(doc); err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.Search("world")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0] != doc.ID {
		t.Fatalf("Expected %v, got %v", doc.ID, results[0])
	}
}

func TestSearchBleve_TitleMatch(t *testing.T) {
	store, err := NewDocumentStore("./test_bleve_title_term.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		_ = store.Close()
		_ = os.Remove("./test_bleve_title_term.db")
		_ = os.RemoveAll("./test_bleve_title_term.db.bleve")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Golang Bleve",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{{
			ID:      domain.BlockID(uuid.New()),
			Content: "nothing to see here",
			Indent:  0,
		}},
	}

	if err := store.Save(doc); err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.Search("bleve")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
}

func TestSearchBleve_PhraseMatch(t *testing.T) {
	store, err := NewDocumentStore("./test_bleve_phrase.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		_ = store.Close()
		_ = os.Remove("./test_bleve_phrase.db")
		_ = os.RemoveAll("./test_bleve_phrase.db.bleve")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Some title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{{
			ID:      domain.BlockID(uuid.New()),
			Content: "the quick brown fox",
			Indent:  0,
		}},
	}

	if err := store.Save(doc); err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.Search("\"quick brown\"")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	results, err = store.Search("\"brown quick\"")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("Expected 0 results, got %d", len(results))
	}
}

func TestSearchBleve_FuzzyMatch(t *testing.T) {
	store, err := NewDocumentStore("./test_bleve_fuzzy.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		_ = store.Close()
		_ = os.Remove("./test_bleve_fuzzy.db")
		_ = os.RemoveAll("./test_bleve_fuzzy.db.bleve")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Some title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{{
			ID:      domain.BlockID(uuid.New()),
			Content: "elephant",
			Indent:  0,
		}},
	}

	if err := store.Save(doc); err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.Search("elephnt")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
}
