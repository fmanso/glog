package db

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

var referenceRegex = regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)

type referencesIndex struct {
	db             *bolt.DB
	referenceIndex []byte
}

func newReferencesIndex(db *bolt.DB) (*referencesIndex, error) {
	referencesKey := []byte("references_index")
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(referencesKey)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &referencesIndex{
		db:             db,
		referenceIndex: referencesKey,
	}, nil
}

func (ri *referencesIndex) save(tx *bolt.Tx, doc *DocDb) error {
	log.Printf("Updating references index for document ID: %s, Title: %s", doc.ID, doc.Title)
	referencedTitles := getReferencedTitles(doc)

	log.Printf("Document ID: %s references titles: %v", doc.ID, referencedTitles)
	bucket := tx.Bucket(ri.referenceIndex)
	if bucket == nil {
		return fmt.Errorf("references index not found for document ID: %s", doc.ID)
	}

	for _, title := range referencedTitles {
		key := []byte(strings.ToLower(title))
		existing := bucket.Get(key)
		ids := decodeUUIDSet(existing)
		ids[doc.ID] = struct{}{}
		encoded, err := encodeUUIDSet(ids)
		if err != nil {
			return err
		}

		log.Println("Saving reference index:", title, "->", doc.ID)
		err = bucket.Put(key, encoded)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ri *referencesIndex) getReferences(title string) ([]uuid.UUID, error) {
	var result []uuid.UUID
	err := ri.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(ri.referenceIndex)
		if bucket == nil {
			return nil
		}

		data := bucket.Get([]byte(strings.ToLower(title)))
		if data == nil {
			return nil
		}

		ids := decodeUUIDSet(data)
		for id := range ids {
			result = append(result, id)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// getReferencedTitles extracts referenced titles from the document
// Referenced titles are enclosed by double square brackets [[Title]]
func getReferencedTitles(doc *DocDb) []string {
	if doc == nil {
		return nil
	}

	seen := make(map[string]struct{})
	var refs []string

	scan := func(text string) {
		matches := referenceRegex.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			title := strings.TrimSpace(match[1])
			if title == "" {
				continue
			}
			if _, exists := seen[title]; exists {
				continue
			}
			seen[title] = struct{}{}
			refs = append(refs, title)
		}
	}

	scan(doc.Title)
	for _, block := range doc.Blocks {
		if block == nil {
			continue
		}
		scan(block.Content)
	}

	return refs
}
