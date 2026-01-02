package db

import (
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestReferences_Search(t *testing.T) {
	store, err := NewDocumentStore("./testreferences.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		err = os.Remove("./testreferences.db")
	}()

	referencedDoc := &domain.Document{
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

	err = store.Save(referencedDoc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	referencingDoc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Referencing Document",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "This document references [[Test Document Title]].",
				Indent:  0,
			},
		},
	}

	err = store.Save(referencingDoc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.GetReferences("Test Document Title")
	if err != nil {
		t.Fatalf("Failed to search for word: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("Expected to find at least one referencing document, found none")
	}
}
