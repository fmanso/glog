package db

import (
	"bytes"
	"encoding/gob"
	"glog/domain"
	"log"
	"regexp"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

type referencesDb struct {
	bolt             *bolt.DB
	bucketReferences []byte
}

func newReferencesDb(boltDb *bolt.DB) (*referencesDb, error) {
	refKey := []byte("references")
	err := boltDb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(refKey)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &referencesDb{
		bolt:             boltDb,
		bucketReferences: refKey,
	}, nil
}

func (r *referencesDb) handleReferences(tx *bolt.Tx, paragraph *ParagraphDb) error {
	// A reference is a markup in the content with this format: [[document_id:title]].
	// This function parses the paragraph content, extracts all referenced document IDs,
	// and stores them in the "references" bucket. For each referenced document ID, the
	// bucket key is the document ID (as a string) and the value is a gob-encoded
	// map[uuid.UUID]struct{} that acts as a set of paragraph IDs referencing that document.
	// Note: this function only adds/updates entries for the current paragraph; it does not
	// remove references for paragraphs that have been deleted or whose content no longer
	// contains a given reference, so stale references may remain unless cleaned up elsewhere.

	// Search using regexp
	docIds := getReferences(paragraph.Content)
	bucket := tx.Bucket(r.bucketReferences)
	for _, docId := range docIds {
		docKey := []byte(docId.String())
		data := bucket.Get(docKey)
		var docReferences map[uuid.UUID]struct{}
		if data == nil {
			docReferences = map[uuid.UUID]struct{}{}
		} else {
			buf := bytes.NewBuffer(data)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&docReferences)
			if err != nil {
				return err
			}
		}

		docReferences[paragraph.ID] = struct{}{}
		log.Println("Storing reference from paragraph", paragraph.ID.String(), "to document", docId.String())

		var outBuf bytes.Buffer
		enc := gob.NewEncoder(&outBuf)
		err := enc.Encode(docReferences)
		if err != nil {
			return err
		}
		err = bucket.Put(docKey, outBuf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil

}

func getReferences(text string) []domain.DocumentID {
	re := regexp.MustCompile(`\[\[([^:]+):([^\]]+)\]\]`)
	matches := re.FindAllStringSubmatch(text, -1)

	var references []domain.DocumentID
	for _, match := range matches {
		if len(match) >= 2 {
			docId, err := uuid.Parse(match[1])
			if err != nil {
				continue
			}
			references = append(references, domain.DocumentID(docId))
		}
	}
	return references
}

func (r *referencesDb) getParagraphIds(tx *bolt.Tx, docID uuid.UUID) ([]uuid.UUID, error) {
	bucket := tx.Bucket(r.bucketReferences)
	docKey := []byte(docID.String())
	data := bucket.Get(docKey)
	if data == nil {
		return []uuid.UUID{}, nil
	}

	var docReferences map[uuid.UUID]struct{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&docReferences)
	if err != nil {
		return nil, err
	}

	references := make([]uuid.UUID, len(docReferences))
	i := 0
	for paraID := range docReferences {
		references[i] = paraID
		i++
	}

	return references, nil
}
