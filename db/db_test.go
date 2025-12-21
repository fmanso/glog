package db

import (
	"encoding/gob"
	"errors"
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func init() {
	gob.Register(domain.Document{})
	gob.Register(domain.Paragraph{})
}

func TestNewDocumentStore(t *testing.T) {
	db, err := NewDocumentStore("./testnew.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer os.Remove("./testnew.db")
	defer db.Close()
}

func TestLoadForEmpty(t *testing.T) {
	store, err := NewDocumentStore("./testloadfor.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer os.Remove("./testloadfor.db")
	defer store.Close()

	date := time.Now().Truncate(24 * time.Hour)
	_, err = store.GetDocumentFor(date)
	if errors.Is(err, ErrDocumentNotFound) {
		return
	}

	t.Fatalf("Expected ErrDocumentNotFound, got: %v", err)
}

func TestSaveAndLoad(t *testing.T) {
	store, err := NewDocumentStore("./testloadfor.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer os.Remove("./testloadfor.db")
	defer store.Close()

	date := time.Now().Truncate(24 * time.Hour)
	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document",
		Date:  domain.ToDateTime(date),
		Body:  []*domain.Paragraph{},
	}
	err = store.Save(doc)

	_, err = store.GetDocumentFor(date)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}
}
