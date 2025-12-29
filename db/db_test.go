package db

import (
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewDocumentStore(t *testing.T) {
	store, err := NewDocumentStore("./testnewdocumentstore.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		err = os.Remove("./testnewdocumentstore.db")
	}()

	if store == nil {
		t.Fatal("DocumentStore is nil")
	}
}

func TestDocumentStore_Save(t *testing.T) {
	store, err := NewDocumentStore("./testsave.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		err = os.Remove("./testsave.db")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document",
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

	got, err := store.LoadDocument(doc.ID)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}

	if got.Title != doc.Title {
		t.Errorf("Loaded document title mismatch: got %v, want %v", got.Title, doc.Title)
	}

	if len(got.Blocks) != len(doc.Blocks) {
		t.Fatalf("Loaded document blocks length mismatch: got %v, want %v", len(got.Blocks), len(doc.Blocks))
	}

	if got.Blocks[0].Content != doc.Blocks[0].Content {
		t.Errorf("Loaded document block content mismatch: got %v, want %v", got.Blocks[0].Content, doc.Blocks[0].Content)
	}
}
