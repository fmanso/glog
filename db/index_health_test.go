package db

import (
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestIndexHealthTracking(t *testing.T) {
	store, err := NewDocumentStore("./testindexhealth.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}
		_ = os.Remove("./testindexhealth.db")
		_ = os.RemoveAll("./testindexhealth.db.bleve")
	}()

	// Check initial health
	health := store.GetIndexHealth()
	if !health.IsHealthy {
		t.Errorf("Index should be healthy initially")
	}
	if health.FailedDocuments != 0 {
		t.Errorf("Should have 0 failed documents initially, got %d", health.FailedDocuments)
	}

	// Save a document successfully
	title := "Test Document " + uuid.NewString()
	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: title,
		Date:  time.Now(),
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

	// Check health after successful save
	health = store.GetIndexHealth()
	if !health.IsHealthy {
		t.Errorf("Index should still be healthy after successful save")
	}
	if health.FailedDocuments != 0 {
		t.Errorf("Should have 0 failed documents after successful save, got %d", health.FailedDocuments)
	}
}

func TestReindexSearchHealth(t *testing.T) {
	store, err := NewDocumentStore("./testreindex.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}
		_ = os.Remove("./testreindex.db")
		_ = os.RemoveAll("./testreindex.db.bleve")
	}()

	// Save some documents
	for i := 0; i < 5; i++ {
		doc := &domain.Document{
			ID:    domain.DocumentID(uuid.New()),
			Title: "Test Document " + uuid.NewString(),
			Date:  time.Now(),
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
	}

	// Reindex
	err = store.ReindexSearch()
	if err != nil {
		t.Fatalf("Failed to reindex: %v", err)
	}

	// Check health after reindex
	health := store.GetIndexHealth()
	if !health.IsHealthy {
		t.Errorf("Index should be healthy after reindex")
	}
}

func TestRetryFailedIndexing(t *testing.T) {
	store, err := NewDocumentStore("./testretry.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}
		_ = os.Remove("./testretry.db")
		_ = os.RemoveAll("./testretry.db.bleve")
	}()

	// Save a document
	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document " + uuid.NewString(),
		Date:  time.Now(),
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

	// Try to retry failed indexing (should be none)
	count, err := store.RetryFailedIndexing()
	if err != nil {
		t.Fatalf("Failed to retry indexing: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 retried documents, got %d", count)
	}

	// Check health
	health := store.GetIndexHealth()
	if !health.IsHealthy {
		t.Errorf("Index should be healthy")
	}
}
