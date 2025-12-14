package db

import (
	"bytes"
	"encoding/gob"
	"errors"
	"glog/domain"
	"strings"

	"github.com/boltdb/bolt"
	_ "github.com/boltdb/bolt"
	"github.com/google/uuid"
)

type DocumentStore struct {
	path             string
	bolt             *bolt.DB
	bucketDocs       []byte
	bucketParagraphs []byte
	bucketTimeIndex  []byte
	bucketTermsIndex []byte
}

func NewDocumentStore(path string) (*DocumentStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("documents"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("time_index"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("terms_index"))
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists([]byte("paragraphs"))
		return err
	})
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DocumentStore{
		bolt:             db,
		path:             path,
		bucketDocs:       []byte("documents"),
		bucketParagraphs: []byte("paragraphs"),
		bucketTimeIndex:  []byte("time_index"),
		bucketTermsIndex: []byte("terms_index"),
	}, nil
}

func (store *DocumentStore) Close() error {
	return store.bolt.Close()
}

func (store *DocumentStore) Save(doc *domain.Document, paragraphs []domain.Paragraph) error {
	paras := make(map[domain.ParagraphID]domain.Paragraph)
	for _, para := range paragraphs {
		paras[para.ID] = para
	}

	err := store.bolt.Update(func(tx *bolt.Tx) error {
		// Add the document
		bucket := tx.Bucket(store.bucketDocs)
		data, err := doc.Serialize()
		if err != nil {
			return err
		}

		key := []byte(doc.ID.String())
		err = bucket.Put(key, data)
		if err != nil {
			return err
		}

		// Add the paragraphs
		bucket = tx.Bucket(store.bucketParagraphs)
		for _, paraID := range doc.Body {
			para, ok := paras[paraID]
			if !ok {
				return errors.New("paragraph not found in provided map")
			}

			data, err := para.Serialize()
			if err != nil {
				return err
			}

			key := []byte(para.ID.String())
			err = bucket.Put(key, data)
			if err != nil {
				return err
			}
		}

		// Add time index
		err = store.updateOrCreateSliceForTime(tx, doc.Date, doc.ID)
		if err != nil {
			return err
		}

		// Add terms index
		err = store.indexTerms(tx, doc, paragraphs)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (store *DocumentStore) indexTerms(tx *bolt.Tx, doc *domain.Document, paragraphs []domain.Paragraph) error {
	bucket := tx.Bucket(store.bucketTermsIndex)
	for _, para := range paragraphs {
		words := strings.Fields(string(para.Content))
		for _, word := range words {
			term := strings.ToLower(word)
			indices := bucket.Get([]byte(term))
			if indices == nil {
				docIds := []domain.DocumentID{doc.ID}
				// Serialize the slice
				var buf bytes.Buffer
				enc := gob.NewEncoder(&buf)
				err := enc.Encode(docIds)
				if err != nil {
					return err
				}
				// Store the serialized slice
				err = bucket.Put([]byte(term), buf.Bytes())
				if err != nil {
					return err
				}
			} else {
				// Deserialize the existing slice
				var docIds []domain.DocumentID
				buf := bytes.NewBuffer(indices)
				dec := gob.NewDecoder(buf)
				err := dec.Decode(&docIds)
				if err != nil {
					return err
				}
				// Check if doc.ID is already in docIds
				found := false
				for _, id := range docIds {
					if id == doc.ID {
						found = true
						break
					}
				}
				if !found {
					// Append the new DocumentID
					docIds = append(docIds, doc.ID)
					// Serialize the updated slice
					var outBuf bytes.Buffer
					enc := gob.NewEncoder(&outBuf)
					err = enc.Encode(docIds)
					if err != nil {
						return err
					}
					// Store the serialized slice
					err = bucket.Put([]byte(term), outBuf.Bytes())
					if err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (store *DocumentStore) updateOrCreateSliceForTime(tx *bolt.Tx, dateTime domain.DateTime, docID domain.DocumentID) error {
	// Get or create a slice of DocumentIDs for the given time
	bucket := tx.Bucket(store.bucketTimeIndex)
	documents := bucket.Get([]byte(dateTime))
	if documents == nil {
		docIds := []domain.DocumentID{docID}
		// Serialize the slice
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		err := enc.Encode(docIds)
		if err != nil {
			return err
		}
		// Store the serialized slice
		err = bucket.Put([]byte(dateTime), buf.Bytes())
		if err != nil {
			return err
		}
	} else {
		// Deserialize the existing slice
		var docIds []domain.DocumentID
		buf := bytes.NewBuffer(documents)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&docIds)
		if err != nil {
			return err
		}
		// Append the new DocumentID
		docIds = append(docIds, docID)
		// Serialize the updated slice
		var outBuf bytes.Buffer
		enc := gob.NewEncoder(&outBuf)
		err = enc.Encode(docIds)
		if err != nil {
			return err
		}
		// Store the serialized slice
		err = bucket.Put([]byte(dateTime), outBuf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

var ErrDocumentNotFound = errors.New("document not found")
var ErrParagraphNotFound = errors.New("paragraph not found")

func (store *DocumentStore) LoadDocument(id domain.DocumentID) (*domain.Document, map[domain.ParagraphID]domain.Paragraph, error) {
	var doc domain.Document
	paras := make(map[domain.ParagraphID]domain.Paragraph)
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketDocs)
		data := bucket.Get([]byte(id.String()))
		if data == nil {
			return ErrDocumentNotFound
		}

		err := doc.Deserialize(data)
		if err != nil {
			return err
		}

		bucket = tx.Bucket(store.bucketParagraphs)
		for _, paraID := range doc.Body {
			data := bucket.Get([]byte(paraID.String()))
			if data == nil {
				return ErrParagraphNotFound
			}

			var para domain.Paragraph
			err := para.Deserialize(data)
			if err != nil {
				return err
			}
			paras[paraID] = para
		}

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return &doc, paras, nil
}

func (store *DocumentStore) Load(from, to domain.DateTime) ([]domain.DocumentID, error) {
	docs := make([]domain.DocumentID, 0)
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketTimeIndex)
		fromBytes := []byte(from)
		toBytes := []byte(to)

		c := bucket.Cursor()
		for k, v := c.Seek(fromBytes); k != nil && bytes.Compare(k, toBytes) <= 0; k, v = c.Next() {
			// Deserialize the slice of DocumentIDs
			var docIds []domain.DocumentID
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&docIds)
			if err != nil {
				return err
			}
			docs = append(docs, docIds...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (store *DocumentStore) Search(terms []string) ([]domain.DocumentID, error) {
	docs := make([]domain.DocumentID, 0)
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketTermsIndex)
		for _, term := range terms {
			term = strings.ToLower(term)
			data := bucket.Get([]byte(term))
			if data == nil {
				continue
			}

			// Deserialize the slice of DocumentIDs
			var docIds []domain.DocumentID
			buf := bytes.NewBuffer(data)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&docIds)
			if err != nil {
				return err
			}
			docs = append(docs, docIds...)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (store *DocumentStore) GetReferences(documentID uuid.UUID) ([]domain.Document, error) {
	return nil, nil
}
