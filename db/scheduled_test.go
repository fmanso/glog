package db

import (
	"glog/domain"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestScheduled_GetTasks(t *testing.T) {
	store, err := NewDocumentStore("./testscheduled.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		err = os.Remove("./testscheduled.db")
	}()

	docID := domain.DocumentID(uuid.New())
	blockID := domain.BlockID(uuid.New())
	err = store.ScheduleTask(time.Now().UTC(), docID, blockID)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.GetScheduledTasks(time.Now().UTC(), 5)
	if err != nil {
		t.Fatalf("Failed to search for word: %v", err)
	}

	if len(results) != 1 {
		t.Fatal("Expected to find one task, found none")
	}

	if results[0].DocID != docID {
		t.Fatalf("Expected DocID %v, got %v", docID, results[0].DocID)
	}

	if results[0].BlockID != blockID {
		t.Fatalf("Expected BlockID %v, got %v", blockID, results[0].BlockID)
	}
}

func TestScheduled_RemoveObsoleteTasks(t *testing.T) {
	store, err := NewDocumentStore("./testscheduled.db")
	if err != nil {
		t.Fatalf("Failed to create DocumentStore: %v", err)
	}
	defer func() {
		err := store.Close()
		if err != nil {
			t.Errorf("Failed to close DocumentStore: %v", err)
		}

		err = os.Remove("./testscheduled.db")
	}()

	doc := &domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: "Test Document",
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

	tasks, err := store.GetScheduledTasks(time.Date(2024, 12, 29, 0, 0, 0, 0, time.UTC), 5)
	if err != nil {
		t.Fatalf("Failed to load scheduled tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 scheduled tasks, got %v", len(tasks))
	}

	doc.Blocks[0].Content = "Test Content /scheduled 2025-12-31"
	err = store.Save(doc)
	if err != nil {
		t.Fatalf("Failed to update document: %v", err)
	}

	tasks, err = store.GetScheduledTasks(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC), 5)
	if err != nil {
		t.Fatalf("Failed to load scheduled tasks: %v", err)
	}

	if len(tasks) != 0 {
		t.Fatalf("Expected 0 scheduled tasks for obsolete date, got %v", len(tasks))
	}

	tasks, err = store.GetScheduledTasks(time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), 5)
	if err != nil {
		t.Fatalf("Failed to load scheduled tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 scheduled task for new date, got %v", len(tasks))
	}
}
