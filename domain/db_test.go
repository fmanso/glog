package domain

import (
	"encoding/gob"
	"testing"
	"time"

	"github.com/google/uuid"
)

func init() {
	gob.Register(Document{})
	gob.Register(Paragraph{})
}

func TestNewDocumentStore(t *testing.T) {
	db, err := NewDocumentStore("./testdb.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer db.Close()
}

func TestSave(t *testing.T) {
	store, err := NewDocumentStore("./testdb.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer store.Close()

	pid := ParagraphID(uuid.New())
	var paragraphs []Paragraph
	paragraphs = append(paragraphs, Paragraph{
		ID:         pid,
		DocumentID: DocumentID(uuid.New()),
		Content:    "This is a test paragraph.",
	})

	doc := NewDocument("Test Document", time.Now(), paragraphs)

	err = store.Save(doc, paragraphs)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
}

func TestLoad(t *testing.T) {
	store, err := NewDocumentStore("./testdb.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer store.Close()

	pid := ParagraphID(uuid.New())
	var paragraphs []Paragraph
	paragraphs = append(paragraphs, Paragraph{
		ID:         pid,
		DocumentID: DocumentID(uuid.New()),
		Content:    "This is a test paragraph.",
	})

	doc := NewDocument("Test Document", time.Now(), paragraphs)

	err = store.Save(doc, paragraphs)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	loadedDoc, loadedParagraphs, err := store.LoadDocument(doc.ID)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}

	if loadedDoc.Title != doc.Title {
		t.Fatalf("Loaded document title mismatch: got %s, want %s", loadedDoc.Title, doc.Title)
	}

	if len(loadedParagraphs) != len(paragraphs) {
		t.Fatalf("Loaded paragraphs count mismatch: got %d, want %d", len(loadedParagraphs), len(paragraphs))
	}

	if loadedParagraphs[doc.Body[0]].Content != paragraphs[0].Content {
		t.Fatalf("Loaded paragraph content mismatch: got %s, want %s", loadedParagraphs[doc.Body[0]].Content, paragraphs[0].Content)
	}

	from := time.Now().Add(-1 * time.Hour)
	to := time.Now().Add(1 * time.Hour)
	docs, err := store.Load(ToDateTime(from), ToDateTime(to))
	if err != nil {
		t.Fatalf("Failed to load documents by date range: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected one document in date range, got %d", len(docs))
	}

	to = time.Now().Add(-20 * time.Minute)
	docs, err = store.Load(ToDateTime(from), ToDateTime(to))
	if err != nil {
		t.Fatalf("Failed to load documents by date range: %v", err)
	}

	if len(docs) != 0 {
		t.Fatalf("Expected no documents in date range, got %d", len(docs))
	}
}
