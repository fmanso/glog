package db

import (
	"bytes"
	"encoding/gob"
	"glog/domain"
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

	paragraph := &domain.Paragraph{
		ID:      domain.ParagraphID(uuid.New()),
		Content: domain.Content("This is a reference to [[123e4567-e89b-12d3-a456-426614174000:Sample Document]]."),
	}

	paragraph2 := &domain.Paragraph{
		ID:      domain.ParagraphID(uuid.New()),
		Content: domain.Content("This is a reference to [[123e4567-e89b-12d3-a456-426614174000:Sample Document]]."),
	}

	err = referencesDb.handleReferences(paragraph)
	if err != nil {
		t.Fatalf("Failed to handle references: %v", err)
	}

	err = referencesDb.handleReferences(paragraph2)
	if err != nil {
		t.Fatalf("Failed to handle references for second paragraph: %v", err)
	}

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(referencesDb.bucketReferences)
		docKey := []byte(uuid.MustParse("123e4567-e89b-12d3-a456-426614174000").String())
		data := bucket.Get(docKey)
		if data == nil {
			t.Fatalf("No references found for document ID")
		}

		var docReferences map[domain.ParagraphID]struct{}
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err = dec.Decode(&docReferences)
		if err != nil {
			t.Fatalf("Failed to decode references: %v", err)
		}

		if _, exists := docReferences[paragraph.ID]; !exists {
			t.Fatalf("Paragraph ID 1 not found in references")
		}

		if _, exists := docReferences[paragraph2.ID]; !exists {
			t.Fatalf("Paragraph ID 2 not found in references")
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
