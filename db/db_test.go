package db

import (
	"encoding/gob"
	"errors"
	"fmt"
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

func TestDbHandleReferences(t *testing.T) {
	store, err := NewDocumentStore("./testreferences.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}

	defer func() {
		defer store.Close()
		defer os.Remove("./testreferences.db")
	}()

	docId := uuid.New()
	doc := &domain.Document{
		ID:    domain.DocumentID(docId),
		Title: "Test Document",
		Date:  domain.ToDateTime(time.Now().Truncate(24 * time.Hour)),
		Body:  []*domain.Paragraph{},
	}
	_ = store.Save(doc)

	refDocId := uuid.New()
	refDoc := &domain.Document{
		ID:    domain.DocumentID(refDocId),
		Title: "Referencing Document",
		Date:  domain.ToDateTime(time.Now().UTC()),
		Body:  []*domain.Paragraph{},
	}

	_ = refDoc.InsertParagraphAt(0, "", 0)

	_ = store.Save(refDoc)

	para := refDoc.Body[0]
	err = store.SetParagraphContent(para.ID, domain.Content(fmt.Sprintf("This is a reference to [[%s:%s]].", docId.String(), "Test Document")))
	if err != nil {
		t.Fatalf("Failed to set paragraph content: %v", err)
	}

	references, err := store.GetReferences(domain.DocumentID(docId))
	if err != nil {
		t.Fatalf("Failed to get references: %v", err)
	}

	if len(references) != 1 {
		t.Fatalf("Expected 1 reference, got %d", len(references))
	}
}
