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
	err = store.ScheduleTask(time.Now().UTC().Add(24*time.Hour), docID, blockID)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	results, err := store.GetScheduledTasks(time.Now().UTC())
	if err != nil {
		t.Fatalf("Failed to search for word: %v", err)
	}

	if len(results) == 1 {
		t.Fatal("Expected to find at least one task, found none")
	}

	if results[0].DocID != docID {
		t.Fatalf("Expected DocID %v, got %v", docID, results[0].DocID)
	}

	if results[0].BlockID != blockID {
		t.Fatalf("Expected BlockID %v, got %v", blockID, results[0].BlockID)
	}
}
