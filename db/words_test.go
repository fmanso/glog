package db

import (
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestWordIndex_Search(t *testing.T) {
	store, err := NewDocumentStore("./testwordindexsearch.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		err = os.Remove("./testwordindexsearch.db")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document Title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Test Content",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.Search("test content")
	if err != nil {
		t.Fatalf("Failed to search for word: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected to find at least one document, found none")
	}

	doc.Blocks[0].Content = "New Content"
	err = store.Save(doc)
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	results, err = store.Search("test content")
	if err != nil {
		t.Fatalf("Failed to search for word after update: %v", err)
	}

	if len(results) != 0 {
		t.Fatal("Expected to find none after update, found some")
	}

	results, err = store.Search("Document Title")
	if err != nil {
		t.Fatalf("Failed to search for title words: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected to find at least one document by title, found none")
	}
}

func TestWordIndex_Search_NotFound(t *testing.T) {
	store, err := NewDocumentStore("./testnotfound.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		err = os.Remove("./testnotfound.db")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document Title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Test Content",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.Search("test notfound")
	if err != nil {
		t.Fatalf("Failed to search for word: %v", err)
	}

	if len(results) != 0 {
		t.Fatal("Expected to find none, found some")
	}
}
