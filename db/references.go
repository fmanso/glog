package db

import (
	"bytes"
	"encoding/gob"
	"glog/domain"
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

func (r *referencesDb) handleReferences(paragraph *domain.Paragraph) error {
	// A reference is a markup in the content with this format: [[document_id:title]]
	// This function will parse the content and extract the references
	// and store them in the references bucket

	contentStr := string(paragraph.Content)
	// Search using regexp
	docIds := getReferences(contentStr)
	err := r.bolt.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(r.bucketReferences)
		for _, docId := range docIds {
			docKey := []byte(docId.String())
			data := bucket.Get(docKey)
			var docReferences map[domain.ParagraphID]struct{}
			if data == nil {
				docReferences = map[domain.ParagraphID]struct{}{}
			} else {
				buf := bytes.NewBuffer(data)
				dec := gob.NewDecoder(buf)
				err := dec.Decode(&docReferences)
				if err != nil {
					return err
				}
			}

			docReferences[paragraph.ID] = struct{}{}
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
	})

	return err
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
