package db

import (
	"errors"
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

		_ = os.Remove("./testnewdocumentstore.db")
		_ = os.RemoveAll("./testnewdocumentstore.db.bleve")
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

		_ = os.Remove("./testsave.db")
		_ = os.RemoveAll("./testsave.db.bleve")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: uuid.NewString() + " Test Document",
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

func TestDocumentStore_ListDocuments(t *testing.T) {
	store, err := NewDocumentStore("./testlistdocuments.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testlistdocuments.db")
		_ = os.RemoveAll("./testlistdocuments.db.bleve")
	}()

	doc1 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document 1",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Test Content 1",
				Indent:  0,
			},
		},
	}

	doc2 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document 2",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Test Content 2",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc1)
	if err != nil {
		t.Fatalf("Failed to save document 1: %v", err)
	}

	err = store.Save(doc2)
	if err != nil {
		t.Fatalf("Failed to save document 2: %v", err)
	}

	docs, err := store.ListDocuments()
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}

	if len(docs) != 2 {
		t.Errorf("Listed documents length mismatch: got %v, want %v", len(docs), 2)
	}
}

func TestLoadDocumentByTime(t *testing.T) {
	store, err := NewDocumentStore("./testloadbytime.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testloadbytime.db")
		_ = os.RemoveAll("./testloadbytime.db.bleve")
	}()

	doc1 := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     "Test Document 1",
		Date:      time.Date(2024, 1, 1, 6, 0, 0, 0, time.UTC),
		IsJournal: true,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Test Content 1",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc1)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	doc2 := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     "Test Document 2",
		Date:      time.Date(2024, 1, 2, 6, 0, 0, 0, time.UTC),
		IsJournal: true,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Test Content 2",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc2)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)

	docs, err := store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load document by time: %v", err)
	}

	if len(docs) != 2 {
		t.Errorf("Loaded documents length mismatch: got %v, want %v", len(docs), 2)
	}
}

func TestLoadDocumentByTitle(t *testing.T) {
	store, err := NewDocumentStore("./testloadbytitle.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testloadbytitle.db")
		_ = os.RemoveAll("./testloadbytitle.db.bleve")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Unique Test Document Title",
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

	got, err := store.LoadDocumentByTitle("Unique Test Document Title")
	if err != nil {
		t.Fatalf("Failed to load document by title: %v", err)
	}

	if got.ID != doc.ID {
		t.Errorf("Loaded document ID mismatch: got %v, want %v", got.ID, doc.ID)
	}
}

func TestDocumentStore_SaveRejectsDuplicateTitle(t *testing.T) {
	store, err := NewDocumentStore("./testsaveduplicatetitle.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testsaveduplicatetitle.db")
		_ = os.RemoveAll("./testsaveduplicatetitle.db.bleve")
	}()

	doc1 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Dupe Title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{{
			ID:      domain.BlockID(uuid.New()),
			Content: "Doc 1",
			Indent:  0,
		}},
	}
	if err := store.Save(doc1); err != nil {
		t.Fatalf("Failed to save first document: %v", err)
	}

	doc2 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Dupe Title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{{
			ID:      domain.BlockID(uuid.New()),
			Content: "Doc 2",
			Indent:  0,
		}},
	}
	if err := store.Save(doc2); !errors.Is(err, ErrDuplicateTitle) {
		t.Fatalf("Expected ErrDuplicateTitle, got %v", err)
	}
}

func TestScheduledTasks(t *testing.T) {
	store, err := NewDocumentStore("./testscheduledtasks.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testscheduledtasks.db")
		_ = os.RemoveAll("./testscheduledtasks.db.bleve")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Unique Test Document Title",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Test Content /scheduled 2024-12-31",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	got, err := store.GetScheduledTasks(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 5)
	if err != nil {
		t.Fatalf("Failed to load scheduled tasks: %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("Expected 1 scheduled task, got %v", len(got))
	}

	if got[0].DocID != doc.ID {
		t.Errorf("Scheduled task DocID mismatch: got %v, want %v", got[0].DocID, doc.ID)
	}

	if got[0].BlockID != doc.Blocks[0].ID {
		t.Errorf("Scheduled task BlockID mismatch: got %v, want %v", got[0].BlockID, doc.Blocks[0].ID)
	}
}

func TestReindexSearch(t *testing.T) {
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

	// Create documents with searchable content
	doc1 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Programming Languages",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Go is a statically typed compiled language",
				Indent:  0,
			},
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Python is a dynamically typed interpreted language",
				Indent:  0,
			},
		},
	}

	doc2 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Database Systems",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "BoltDB is an embedded key-value database",
				Indent:  0,
			},
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "PostgreSQL is a powerful relational database",
				Indent:  0,
			},
		},
	}

	// Save documents before reindexing
	err = store.Save(doc1)
	if err != nil {
		t.Fatalf("Failed to save document 1: %v", err)
	}

	err = store.Save(doc2)
	if err != nil {
		t.Fatalf("Failed to save document 2: %v", err)
	}

	// Verify search works before reindex
	results, err := store.Search("language")
	if err != nil {
		t.Fatalf("Failed to search before reindex: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result before reindex, got %v", len(results))
	}

	// Perform reindexing
	err = store.ReindexSearch()
	if err != nil {
		t.Fatalf("Failed to reindex search: %v", err)
	}

	// Verify search still works after reindex
	results, err = store.Search("language")
	if err != nil {
		t.Fatalf("Failed to search after reindex: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result after reindex, got %v", len(results))
	}

	if results[0] != doc1.ID {
		t.Errorf("Expected document ID %v, got %v", doc1.ID, results[0])
	}

	// Verify both documents are searchable
	results, err = store.Search("database")
	if err != nil {
		t.Fatalf("Failed to search for 'database': %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'database', got %v", len(results))
	}

	if results[0] != doc2.ID {
		t.Errorf("Expected document ID %v, got %v", doc2.ID, results[0])
	}

	// Search by title
	results, err = store.Search("Programming")
	if err != nil {
		t.Fatalf("Failed to search by title: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for title search, got %v", len(results))
	}

	if results[0] != doc1.ID {
		t.Errorf("Expected document ID %v, got %v", doc1.ID, results[0])
	}
}

func TestReindexSearchEmpty(t *testing.T) {
	store, err := NewDocumentStore("./testreindexempty.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testreindexempty.db")
		_ = os.RemoveAll("./testreindexempty.db.bleve")
	}()

	// Reindex on empty database should not fail
	err = store.ReindexSearch()
	if err != nil {
		t.Fatalf("Failed to reindex empty database: %v", err)
	}

	// Search should return no results
	results, err := store.Search("anything")
	if err != nil {
		t.Fatalf("Failed to search after reindex: %v", err)
	}

	if len(results) != 0 {
		t.Errorf("Expected 0 results, got %v", len(results))
	}
}

func TestReindexSearchWithNewDocuments(t *testing.T) {
	store, err := NewDocumentStore("./testreindexnew.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testreindexnew.db")
		_ = os.RemoveAll("./testreindexnew.db.bleve")
	}()

	// Add first document
	doc1 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "First Document",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Original content before reindex",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc1)
	if err != nil {
		t.Fatalf("Failed to save document 1: %v", err)
	}

	// Reindex
	err = store.ReindexSearch()
	if err != nil {
		t.Fatalf("Failed to reindex: %v", err)
	}

	// Add second document after reindex
	doc2 := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Second Document",
		Date:  time.Now().UTC(),
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "New content after reindex",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc2)
	if err != nil {
		t.Fatalf("Failed to save document 2: %v", err)
	}

	// Search for first document
	results, err := store.Search("Original")
	if err != nil {
		t.Fatalf("Failed to search for first document: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'Original', got %v", len(results))
	}

	if len(results) > 0 && results[0] != doc1.ID {
		t.Errorf("Expected document ID %v, got %v", doc1.ID, results[0])
	}

	// Search for second document
	results, err = store.Search("New")
	if err != nil {
		t.Fatalf("Failed to search for second document: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'New', got %v", len(results))
	}

	if len(results) > 0 && results[0] != doc2.ID {
		t.Errorf("Expected document ID %v, got %v", doc2.ID, results[0])
	}

	// Verify both documents are indexed
	results, err = store.Search("content")
	if err != nil {
		t.Fatalf("Failed to search for 'content': %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'content', got %v", len(results))
	}
}

// TestReindexSearchConcurrentSave verifies that Save operations can run
// concurrently with ReindexSearch without causing race conditions or panics.
func TestReindexSearchConcurrentSave(t *testing.T) {
	store, err := NewDocumentStore("./testreindexconcurrentsave.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testreindexconcurrentsave.db")
		_ = os.RemoveAll("./testreindexconcurrentsave.db.bleve")
	}()

	// Add initial documents
	for i := 0; i < 5; i++ {
		doc := &domain.Document{
			ID:    domain.DocumentID(uuid.New()),
			Title: uuid.NewString() + " Initial Document",
			Date:  time.Now().UTC(),

			Blocks: []*domain.Block{
				{
					ID:      domain.BlockID(uuid.New()),
					Content: "Initial content for search",
					Indent:  0,
				},
			},
		}
		if err := store.Save(doc); err != nil {
			t.Fatalf("Failed to save initial document: %v", err)
		}
	}

	// Start reindexing in a goroutine
	reindexDone := make(chan error, 1)
	go func() {
		reindexDone <- store.ReindexSearch()
	}()

	// Concurrently save documents while reindexing
	saveDone := make(chan error, 1)
	go func() {
		for i := 0; i < 5; i++ {
			doc := &domain.Document{
				ID:    domain.DocumentID(uuid.New()),
				Title: uuid.NewString() + " Concurrent Document",
				Date:  time.Now().UTC(),

				Blocks: []*domain.Block{
					{
						ID:      domain.BlockID(uuid.New()),
						Content: "Concurrent content for search",
						Indent:  0,
					},
				},
			}
			if err := store.Save(doc); err != nil {
				saveDone <- err
				return
			}
			time.Sleep(10 * time.Millisecond) // Small delay to increase chance of interleaving
		}
		saveDone <- nil
	}()

	// Wait for both operations to complete
	if err := <-reindexDone; err != nil {
		t.Fatalf("ReindexSearch failed: %v", err)
	}
	if err := <-saveDone; err != nil {
		t.Fatalf("Concurrent Save failed: %v", err)
	}

	// Verify all documents are searchable
	results, err := store.Search("content")
	if err != nil {
		t.Fatalf("Failed to search after concurrent operations: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 documents after concurrent operations, got %v", len(results))
	}
}

// TestReindexSearchConcurrentSearch verifies that Search operations can run
// concurrently with ReindexSearch without causing race conditions or panics.
func TestReindexSearchConcurrentSearch(t *testing.T) {
	store, err := NewDocumentStore("./testreindexconcurrentsearch.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testreindexconcurrentsearch.db")
		_ = os.RemoveAll("./testreindexconcurrentsearch.db.bleve")
	}()

	// Add documents to search
	for i := 0; i < 10; i++ {
		doc := &domain.Document{
			ID:    domain.DocumentID(uuid.New()),
			Title: uuid.NewString() + " Test Document",
			Date:  time.Now().UTC(),

			Blocks: []*domain.Block{
				{
					ID:      domain.BlockID(uuid.New()),
					Content: "Searchable content",
					Indent:  0,
				},
			},
		}
		if err := store.Save(doc); err != nil {
			t.Fatalf("Failed to save document: %v", err)
		}
	}

	// Start reindexing in a goroutine
	reindexDone := make(chan error, 1)
	go func() {
		reindexDone <- store.ReindexSearch()
	}()

	// Concurrently perform searches while reindexing
	searchDone := make(chan error, 1)
	go func() {
		for i := 0; i < 20; i++ {
			_, err := store.Search("Searchable")
			if err != nil {
				searchDone <- err
				return
			}
			time.Sleep(5 * time.Millisecond) // Small delay to increase chance of interleaving
		}
		searchDone <- nil
	}()

	// Wait for both operations to complete
	if err := <-reindexDone; err != nil {
		t.Fatalf("ReindexSearch failed: %v", err)
	}
	if err := <-searchDone; err != nil {
		t.Fatalf("Concurrent Search failed: %v", err)
	}

	// Verify search still works after reindex
	results, err := store.Search("Searchable")
	if err != nil {
		t.Fatalf("Failed to search after concurrent operations: %v", err)
	}

	if len(results) != 10 {
		t.Errorf("Expected 10 documents after concurrent operations, got %v", len(results))
	}
}

// TestConcurrentSaveAndSearch verifies that Save and Search operations
// can run concurrently without race conditions.
func TestConcurrentSaveAndSearch(t *testing.T) {
	store, err := NewDocumentStore("./testconcurrentsavesearch.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testconcurrentsavesearch.db")
		_ = os.RemoveAll("./testconcurrentsavesearch.db.bleve")
	}()

	// Number of concurrent operations
	numSavers := 5
	numSearchers := 5
	docsPerSaver := 10

	// Launch concurrent savers
	saveDone := make(chan error, numSavers)
	for s := 0; s < numSavers; s++ {
		go func(saverID int) {
			for i := 0; i < docsPerSaver; i++ {
				doc := &domain.Document{
					ID:    domain.DocumentID(uuid.New()),
					Title: "Concurrent Test " + uuid.New().String(),
					Date:  time.Now().UTC(),
					Blocks: []*domain.Block{
						{
							ID:      domain.BlockID(uuid.New()),
							Content: "Concurrent content",
							Indent:  0,
						},
					},
				}
				if err := store.Save(doc); err != nil {
					saveDone <- err
					return
				}
			}
			saveDone <- nil
		}(s)
	}

	// Launch concurrent searchers
	searchDone := make(chan error, numSearchers)
	for s := 0; s < numSearchers; s++ {
		go func() {
			for i := 0; i < 20; i++ {
				_, err := store.Search("Concurrent")
				if err != nil {
					searchDone <- err
					return
				}
				time.Sleep(5 * time.Millisecond)
			}
			searchDone <- nil
		}()
	}

	// Wait for all savers
	for i := 0; i < numSavers; i++ {
		if err := <-saveDone; err != nil {
			t.Fatalf("Saver %d failed: %v", i, err)
		}
	}

	// Wait for all searchers
	for i := 0; i < numSearchers; i++ {
		if err := <-searchDone; err != nil {
			t.Fatalf("Searcher %d failed: %v", i, err)
		}
	}

	// Give Bleve a moment to flush its internal buffers
	time.Sleep(100 * time.Millisecond)

	// Verify final state
	results, err := store.Search("Concurrent")
	if err != nil {
		t.Fatalf("Final search failed: %v", err)
	}

	expectedDocs := numSavers * docsPerSaver
	if len(results) != expectedDocs {
		t.Errorf("Expected %d documents, got %v", expectedDocs, len(results))
	}
}

// TestJournalIndex tests that journals are properly indexed and retrieved
func TestJournalIndex(t *testing.T) {
	store, err := NewDocumentStore("./testjournalindex.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testjournalindex.db")
		_ = os.RemoveAll("./testjournalindex.db.bleve")
	}()

	// Create a journal document
	journal := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     "Monday, January 1, 2024",
		Date:      time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		IsJournal: true,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Journal entry for Monday",
				Indent:  0,
			},
		},
	}

	err = store.Save(journal)
	if err != nil {
		t.Fatalf("Failed to save journal: %v", err)
	}

	// Create a regular document (not a journal)
	regularDoc := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     "Regular Document",
		Date:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		IsJournal: false,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "This is not a journal",
				Indent:  0,
			},
		},
	}

	err = store.Save(regularDoc)
	if err != nil {
		t.Fatalf("Failed to save regular document: %v", err)
	}

	// Load journals for that date range - should only return the journal
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	journals, err := store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load journals: %v", err)
	}

	if len(journals) != 1 {
		t.Errorf("Expected 1 journal, got %v", len(journals))
	}

	if len(journals) > 0 && journals[0].ID != journal.ID {
		t.Errorf("Expected journal ID %v, got %v", journal.ID, journals[0].ID)
	}

	if len(journals) > 0 && !journals[0].IsJournal {
		t.Error("Expected loaded document to have IsJournal=true")
	}
}

// TestJournalIndexMultipleDays tests journal retrieval across multiple days
func TestJournalIndexMultipleDays(t *testing.T) {
	store, err := NewDocumentStore("./testjournalmultipledays.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testjournalmultipledays.db")
		_ = os.RemoveAll("./testjournalmultipledays.db.bleve")
	}()

	// Create journals for 5 consecutive days
	journalIDs := make([]domain.DocumentID, 5)
	for i := 0; i < 5; i++ {
		journal := &domain.Document{
			ID:        domain.DocumentID(uuid.New()),
			Title:     time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC).Format("Monday, January 2, 2006"),
			Date:      time.Date(2024, 1, i+1, 0, 0, 0, 0, time.UTC),
			IsJournal: true,
			Blocks: []*domain.Block{
				{
					ID:      domain.BlockID(uuid.New()),
					Content: "Journal entry",
					Indent:  0,
				},
			},
		}
		journalIDs[i] = journal.ID

		err = store.Save(journal)
		if err != nil {
			t.Fatalf("Failed to save journal %d: %v", i, err)
		}
	}

	// Load all 5 journals
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 5, 0, 0, 0, 0, time.UTC)

	journals, err := store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load journals: %v", err)
	}

	if len(journals) != 5 {
		t.Errorf("Expected 5 journals, got %v", len(journals))
	}

	// Load subset (days 2-4)
	from = time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
	to = time.Date(2024, 1, 4, 0, 0, 0, 0, time.UTC)

	journals, err = store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load journals: %v", err)
	}

	if len(journals) != 3 {
		t.Errorf("Expected 3 journals for subset, got %v", len(journals))
	}
}

// TestJournalNotIndexedWhenIsJournalFalse ensures non-journals are not in the journal index
func TestJournalNotIndexedWhenIsJournalFalse(t *testing.T) {
	store, err := NewDocumentStore("./testjournalnotindexed.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testjournalnotindexed.db")
		_ = os.RemoveAll("./testjournalnotindexed.db.bleve")
	}()

	// Create multiple regular documents
	for i := 0; i < 3; i++ {
		doc := &domain.Document{
			ID:        domain.DocumentID(uuid.New()),
			Title:     "Regular Document " + uuid.NewString(),
			Date:      time.Date(2024, 1, i+1, 12, 30, 0, 0, time.UTC),
			IsJournal: false,
			Blocks: []*domain.Block{
				{
					ID:      domain.BlockID(uuid.New()),
					Content: "Not a journal",
					Indent:  0,
				},
			},
		}

		err = store.Save(doc)
		if err != nil {
			t.Fatalf("Failed to save document %d: %v", i, err)
		}
	}

	// Load journals - should be empty
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)

	journals, err := store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load journals: %v", err)
	}

	if len(journals) != 0 {
		t.Errorf("Expected 0 journals (non-journals should not be indexed), got %v", len(journals))
	}
}

// TestJournalIsJournalFieldPersistence ensures IsJournal field is properly saved and loaded
func TestJournalIsJournalFieldPersistence(t *testing.T) {
	store, err := NewDocumentStore("./testjournalpersistence.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testjournalpersistence.db")
		_ = os.RemoveAll("./testjournalpersistence.db.bleve")
	}()

	// Create and save a journal
	journalID := domain.DocumentID(uuid.New())
	journal := &domain.Document{
		ID:        journalID,
		Title:     "Persisted Journal",
		Date:      time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		IsJournal: true,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Content",
				Indent:  0,
			},
		},
	}

	err = store.Save(journal)
	if err != nil {
		t.Fatalf("Failed to save journal: %v", err)
	}

	// Load by ID and verify IsJournal is true
	loadedDoc, err := store.LoadDocument(journalID)
	if err != nil {
		t.Fatalf("Failed to load document: %v", err)
	}

	if !loadedDoc.IsJournal {
		t.Error("Expected loaded document to have IsJournal=true, got false")
	}

	// Create and save a non-journal
	regularID := domain.DocumentID(uuid.New())
	regularDoc := &domain.Document{
		ID:        regularID,
		Title:     "Regular Document",
		Date:      time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC),
		IsJournal: false,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Content",
				Indent:  0,
			},
		},
	}

	err = store.Save(regularDoc)
	if err != nil {
		t.Fatalf("Failed to save regular document: %v", err)
	}

	// Load by ID and verify IsJournal is false
	loadedRegular, err := store.LoadDocument(regularID)
	if err != nil {
		t.Fatalf("Failed to load regular document: %v", err)
	}

	if loadedRegular.IsJournal {
		t.Error("Expected loaded document to have IsJournal=false, got true")
	}
}

// TestJournalSameDayDifferentTimes tests that journals are indexed by date, not exact time
func TestJournalSameDayDifferentTimes(t *testing.T) {
	store, err := NewDocumentStore("./testjournalsamedaytimes.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testjournalsamedaytimes.db")
		_ = os.RemoveAll("./testjournalsamedaytimes.db.bleve")
	}()

	// Create a journal with a specific time (e.g., noon)
	journal := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     "Journal at noon",
		Date:      time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		IsJournal: true,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Created at noon",
				Indent:  0,
			},
		},
	}

	err = store.Save(journal)
	if err != nil {
		t.Fatalf("Failed to save journal: %v", err)
	}

	// Should be retrievable by querying the day regardless of time
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	journals, err := store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load journals: %v", err)
	}

	if len(journals) != 1 {
		t.Errorf("Expected 1 journal, got %v", len(journals))
	}
}

func TestDocumentStore_Delete(t *testing.T) {
	store, err := NewDocumentStore("./testdelete.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testdelete.db")
		_ = os.RemoveAll("./testdelete.db.bleve")
	}()

	// Create a document with references and scheduled tasks
	doc := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     "Document To Delete",
		Date:      time.Date(2024, 6, 15, 10, 0, 0, 0, time.UTC),
		IsJournal: false,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Some content with [[Reference]]",
				Indent:  0,
			},
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "A scheduled task /scheduled 2024-06-20",
				Indent:  0,
			},
		},
	}

	err = store.Save(doc)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	// Verify document exists
	_, err = store.LoadDocument(doc.ID)
	if err != nil {
		t.Fatalf("Document should exist before deletion: %v", err)
	}

	// Verify it's searchable
	searchResults, err := store.Search("Delete")
	if err != nil {
		t.Fatalf("Search should work: %v", err)
	}
	if len(searchResults) != 1 {
		t.Errorf("Expected 1 search result, got %v", len(searchResults))
	}

	// Delete the document
	err = store.Delete(uuid.UUID(doc.ID))
	if err != nil {
		t.Fatalf("Failed to delete document: %v", err)
	}

	// Verify document no longer exists
	_, err = store.LoadDocument(doc.ID)
	if err != ErrDocumentNotFound {
		t.Errorf("Document should not exist after deletion, got error: %v", err)
	}

	// Verify it's no longer searchable
	searchResults, err = store.Search("Delete")
	if err != nil {
		t.Fatalf("Search should work: %v", err)
	}
	if len(searchResults) != 0 {
		t.Errorf("Expected 0 search results after deletion, got %v", len(searchResults))
	}

	// Verify it's not in the document list
	docs, err := store.ListDocuments()
	if err != nil {
		t.Fatalf("Failed to list documents: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("Expected 0 documents in list after deletion, got %v", len(docs))
	}

	// Verify it can't be loaded by title
	_, err = store.LoadDocumentByTitle("Document To Delete")
	if err != ErrDocumentNotFound {
		t.Errorf("Document should not be found by title after deletion, got error: %v", err)
	}
}

func TestDocumentStore_DeleteNonExistent(t *testing.T) {
	store, err := NewDocumentStore("./testdeletenonexistent.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testdeletenonexistent.db")
		_ = os.RemoveAll("./testdeletenonexistent.db.bleve")
	}()

	// Try to delete a document that doesn't exist
	err = store.Delete(uuid.New())
	if err != ErrDocumentNotFound {
		t.Errorf("Expected ErrDocumentNotFound when deleting non-existent document, got: %v", err)
	}
}

func TestDocumentStore_DeleteJournal(t *testing.T) {
	store, err := NewDocumentStore("./testdeletejournal.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		_ = os.Remove("./testdeletejournal.db")
		_ = os.RemoveAll("./testdeletejournal.db.bleve")
	}()

	// Create a journal entry
	journal := &domain.Document{
		ID:        domain.DocumentID(uuid.New()),
		Title:     "Monday, June 15, 2024",
		Date:      time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC),
		IsJournal: true,
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "Journal entry content",
				Indent:  0,
			},
		},
	}

	err = store.Save(journal)
	if err != nil {
		t.Fatalf("Failed to save journal: %v", err)
	}

	// Verify journal can be loaded by date
	from := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)

	journals, err := store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load journals: %v", err)
	}
	if len(journals) != 1 {
		t.Errorf("Expected 1 journal before deletion, got %v", len(journals))
	}

	// Delete the journal
	err = store.Delete(uuid.UUID(journal.ID))
	if err != nil {
		t.Fatalf("Failed to delete journal: %v", err)
	}

	// Verify journal can no longer be loaded by date
	journals, err = store.LoadJournals(from, to)
	if err != nil {
		t.Fatalf("Failed to load journals: %v", err)
	}
	if len(journals) != 0 {
		t.Errorf("Expected 0 journals after deletion, got %v", len(journals))
	}
}
