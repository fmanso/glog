package db

import (
	"fmt"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

func TestHandleReferences(t *testing.T) {
	db, err := bolt.Open("./testreferences.db", 0600, nil)
	if err != nil {
		t.Fatalf("Failed to open BoltDB: %v", err)
	}
	defer func() {
		db.Close()
		os.Remove("./testreferences.db")
	}()

	referencesDb, err := newReferencesDb(db)
	if err != nil {
		t.Fatalf("Failed to create referencesDb: %v", err)
	}

	paragraph := &ParagraphDb{
		ID:      uuid.New(),
		Content: "This is a reference to [[123e4567-e89b-12d3-a456-426614174000:Sample Document]].",
	}

	paragraph2 := &ParagraphDb{
		ID:      uuid.New(),
		Content: "This is a reference to [[123e4567-e89b-12d3-a456-426614174000:Sample Document]].",
	}

	err = db.Update(func(tx *bolt.Tx) error {
		err = referencesDb.handleReferences(tx, paragraph)
		if err != nil {
			return fmt.Errorf("failed to handle references for paragraph 1: %v", err)
		}

		err = referencesDb.handleReferences(tx, paragraph2)
		if err != nil {
			return fmt.Errorf("failed to handle references for paragraph 2: %v", err)
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Failed to update DB with references: %v", err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		ids, err := referencesDb.getParagraphIds(tx, uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"))
		if err != nil {
			return fmt.Errorf("failed to get paragraph IDs: %v", err)
		}

		if ids[0] != paragraph.ID {
			t.Fatalf("Expected paragraph ID %s, got %s", paragraph.ID, ids[0])
		}

		if ids[1] != paragraph2.ID {
			t.Fatalf("Expected paragraph ID %s, got %s", paragraph2.ID, ids[1])
		}

		return nil
	})
	
	if err != nil {
		t.Fatalf("Failed to verify references in DB: %v", err)
	}
}

func TestRegExpReferences(t *testing.T) {
	text := "This is a reference to [[123e4567-e89b-12d3-a456-426614174000:Sample Document]] and another to [[987e6543-e21b-12d3-a456-426614174999:Another Document]]."
	refs := getReferences(text)
	if len(refs) != 2 {
		t.Fatalf("Expected 2 references, got %d", len(refs))
	}

	expectedIDs := []uuid.UUID{
		uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		uuid.MustParse("987e6543-e21b-12d3-a456-426614174999"),
	}

	for i, ref := range refs {
		if uuid.UUID(ref) != expectedIDs[i] {
			t.Errorf("Expected reference ID %s, got %s", expectedIDs[i], ref)
		}
	}
}
