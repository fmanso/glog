package db

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

var referenceRegex = regexp.MustCompile(`\[\[([^\[\]]+)\]\]`)

type referencesIndex struct {
	db                *bolt.DB
	referenceIndex    []byte
	docReferenceIndex []byte
}

func newReferencesIndex(db *bolt.DB) (*referencesIndex, error) {
	referencesKey := []byte("references_index")
	docReferenceKey := []byte("doc_reference_index")
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(referencesKey)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(docReferenceKey)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &referencesIndex{
		db:                db,
		referenceIndex:    referencesKey,
		docReferenceIndex: docReferenceKey,
	}, nil
}

func (ri *referencesIndex) updateInvertedIndex(tx *bolt.Tx, doc *DocDb, newReferences []string) error {
	log.Printf("Deleting old references for document ID: %s, Title: %s, newReferences: %v", doc.ID, doc.Title, newReferences)
	bucket := tx.Bucket(ri.docReferenceIndex)
	if bucket == nil {
		return fmt.Errorf("document reference index not found for document ID: %s", doc.ID)
	}

	data := bucket.Get([]byte(doc.ID.String()))
	var oldTitles map[string]struct{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&oldTitles)
	if err != nil {
		oldTitles = map[string]struct{}{}
	}

	newTitlesSet := make(map[string]struct{})
	for _, title := range newReferences {
		newTitlesSet[strings.ToLower(title)] = struct{}{}
	}

	log.Printf("newTitlesSet: %v", newTitlesSet)
	for oldTitle := range oldTitles {
		log.Printf("Checking old title: %s for document ID: %s", oldTitle, doc.ID)
		if _, exists := newTitlesSet[oldTitle]; exists {
			continue
		}

		err = ri.deleteDocFromReferences(tx, oldTitle, doc.ID)
		if err != nil {
			return err
		}
	}

	// Save the new set of referenced titles for the document
	var bufNew bytes.Buffer
	enc := gob.NewEncoder(&bufNew)
	err = enc.Encode(newTitlesSet)
	if err != nil {
		return err
	}

	log.Printf("Saving new referenced titles for document ID: %s, Titles: %v", doc.ID, newReferences)
	err = bucket.Put([]byte(doc.ID.String()), bufNew.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (ri *referencesIndex) deleteDocFromReferences(tx *bolt.Tx, title string, docID uuid.UUID) error {
	log.Printf("Deleting old references for document ID: %s, Title: %s", docID, title)
	bucket := tx.Bucket(ri.referenceIndex)

	data := bucket.Get([]byte(title))
	if data == nil {
		return nil
	}

	ids := decodeUUIDSet(data)
	delete(ids, docID)
	encoded, err := encodeUUIDSet(ids)
	if err != nil {
		return err
	}

	log.Println("Deleting document ID:", docID, "from reference index for title:", title)
	return bucket.Put([]byte(title), encoded)
}

func (ri *referencesIndex) save(tx *bolt.Tx, doc *DocDb) error {
	log.Printf("Updating references index for document ID: %s, Title: %s", doc.ID, doc.Title)
	referencedTitles := getReferencedTitles(doc)

	err := ri.updateInvertedIndex(tx, doc, referencedTitles)
	if err != nil {
		return fmt.Errorf("error deleting references index for document ID: %s, Title: %s", doc.ID, doc.Title)
	}

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
