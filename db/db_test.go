package db

import (
	"encoding/gob"
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
	_, err = store.GetDocumentForToday(date)
	if errors.Is(err, ErrDocumentNotFound) {
		// Expected error for no document found
	} else if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	} else {
		t.Fatalf("Expected ErrDocumentNotFound, but document was found")
	}
}

func TestSave(t *testing.T) {
	store, err := NewDocumentStore("./testsave.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer os.Remove("./testsave.db")
	defer store.Close()

	pid := domain.ParagraphID(uuid.New())
	var paragraphs []domain.Paragraph
	paragraphs = append(paragraphs, domain.Paragraph{
		ID:         pid,
		DocumentID: domain.DocumentID(uuid.New()),
		Content:    "This is a test paragraph.",
	})

	doc := domain.NewDocument("Test Document", time.Now(), paragraphs)

	err = store.Save(doc, paragraphs)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
}

func TestLoad(t *testing.T) {
	store, err := NewDocumentStore("./testloaddb.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer os.Remove("./testloaddb.db")
	defer store.Close()

	pid := domain.ParagraphID(uuid.New())
	var paragraphs []domain.Paragraph
	paragraphs = append(paragraphs, domain.Paragraph{
		ID:         pid,
		DocumentID: domain.DocumentID(uuid.New()),
		Content:    "This is a test paragraph.",
	})

	doc := domain.NewDocument("Test Document", time.Now(), paragraphs)

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
	docs, err := store.Load(domain.ToDateTime(from), domain.ToDateTime(to))
	if err != nil {
		t.Fatalf("Failed to load documents by date range: %v", err)
	}

	if len(docs) != 1 {
		t.Fatalf("Expected one document in date range, got %d", len(docs))
	}

	to = time.Now().Add(-20 * time.Minute)
	docs, err = store.Load(domain.ToDateTime(from), domain.ToDateTime(to))
	if err != nil {
		t.Fatalf("Failed to load documents by date range: %v", err)
	}

	if len(docs) != 0 {
		t.Fatalf("Expected no documents in date range, got %d", len(docs))
	}
}

func TestSearch(t *testing.T) {
	store, err := NewDocumentStore("./testsearch.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer os.Remove("./testsearch.db")
	defer store.Close()

	pid := domain.ParagraphID(uuid.New())
	var paragraphs []domain.Paragraph
	paragraphs = append(paragraphs, domain.Paragraph{
		ID:         pid,
		DocumentID: domain.DocumentID(uuid.New()),
		Content:    "This is a test paragraph.",
	})

	doc := domain.NewDocument("Test Document", time.Now(), paragraphs)

	err = store.Save(doc, paragraphs)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}
	results, err := store.Search([]string{"test"})
	if err != nil {
		t.Fatalf("Failed to search documents: %v", err)
	}

	if len(results) == 0 {
		t.Fatalf("Expected at least one search result, got %d", len(results))
	}

	results, err = store.Search([]string{"missingterm"})
	if err != nil {
		t.Fatalf("Failed to search documents: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("Expected no search results, got %d", len(results))
	}
}
