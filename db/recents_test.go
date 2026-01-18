package db

import (
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRecents_Get(t *testing.T) {
	store, err := NewDocumentStore("./testrecents.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testrecents.db")
		_ = os.RemoveAll("./testrecents.db.bleve")
	}()

	doc1 := &domain.Document{
		ID:     domain.DocumentID(uuid.New()),
		Title:  "Test Document Title",
		Date:   time.Now().UTC(),
		Blocks: []*domain.Block{},
	}

	err = store.Save(doc1)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	_, err = store.LoadDocumentByTitle(doc1.Title)
	if err != nil {
		t.Fatalf("Failed to load document by title: %v", err)
	}

	ids, err := store.GetRecents()
	if err != nil {
		t.Fatalf("Failed to get recents: %v", err)
	}

	assert.NotNil(t, ids)
	assert.Equal(t, 1, len(ids))
	assert.Equal(t, doc1.ID, ids[0])
}

func TestRecents_Delete(t *testing.T) {
	store, err := NewDocumentStore("./testrecents.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testrecents.db")
		_ = os.RemoveAll("./testrecents.db.bleve")
	}()

	doc1 := &domain.Document{
		ID:     domain.DocumentID(uuid.New()),
		Title:  "Test Document Title",
		Date:   time.Now().UTC(),
		Blocks: []*domain.Block{},
	}

	err = store.Save(doc1)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	_, err = store.LoadDocumentByTitle(doc1.Title)
	if err != nil {
		t.Fatalf("Failed to load document by title: %v", err)
	}

	err = store.Delete(uuid.UUID(doc1.ID))
	if err != nil {
		t.Fatalf("Failed deleting: %v", err)
	}

	ids, err := store.GetRecents()
	if err != nil {
		t.Fatalf("Failed to get recents: %v", err)
	}

	assert.NotNil(t, ids)
	assert.Equal(t, 0, len(ids))
}

func TestRecents_Update(t *testing.T) {
	store, err := NewDocumentStore("./testrecents.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testrecents.db")
		_ = os.RemoveAll("./testrecents.db.bleve")
	}()

	doc1 := &domain.Document{
		ID:     domain.DocumentID(uuid.New()),
		Title:  "Test Document Title",
		Date:   time.Now().UTC(),
		Blocks: []*domain.Block{},
	}

	doc2 := &domain.Document{
		ID:     domain.DocumentID(uuid.New()),
		Title:  "Test Document Title 2",
		Date:   time.Now().UTC(),
		Blocks: []*domain.Block{},
	}

	err = store.Save(doc1)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	err = store.Save(doc2)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	_, err = store.LoadDocumentByTitle(doc1.Title)
	if err != nil {
		t.Fatalf("Failed to load document by title: %v", err)
	}

	_, err = store.LoadDocumentByTitle(doc2.Title)
	if err != nil {
		t.Fatalf("Failed to load document by title: %v", err)
	}

	ids, err := store.GetRecents()
	if err != nil {
		t.Fatalf("Failed to get recents: %v", err)
	}

	assert.NotNil(t, ids)
	assert.Equal(t, 2, len(ids))
	assert.Equal(t, doc2.ID, ids[0])
	assert.Equal(t, doc1.ID, ids[1])
}
