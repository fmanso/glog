package db

import (
	"bytes"
	"encoding/gob"
	"errors"
	"glog/domain"
	"time"

	"github.com/boltdb/bolt"
	_ "github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

var ErrDocumentNotFound = errors.New("document not found")

type DocumentStore struct {
	path       string
	bolt       *bolt.DB
	bucketDocs []byte
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
		return nil
	})

	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DocumentStore{
		bolt:       db,
		path:       path,
		bucketDocs: []byte("documents"),
	}, nil
}

func (store *DocumentStore) Close() error {
	return store.bolt.Close()
}

func (store *DocumentStore) saveDoc(tx *bolt.Tx, doc *domain.Document) error {
	log.Printf("Saving document: %v\n", doc)

	// Create DocDb from domain.Document
	docDb := DocDb{
		ID:     uuid.UUID(doc.ID),
		Title:  doc.Title,
		Date:   doc.Date.UTC().Format(time.RFC3339),
		Blocks: make([]*BlockDb, len(doc.Blocks)),
	}

	for i, block := range doc.Blocks {
		docDb.Blocks[i] = &BlockDb{
			ID:      uuid.UUID(block.ID),
			Content: block.Content,
			Ident:   block.Indent,
		}
	}

	// Serialize DocDb
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(docDb)
	if err != nil {
		return err
	}

	bucket := tx.Bucket(store.bucketDocs)
	return bucket.Put([]byte(doc.ID.String()), buf.Bytes())
}

func (store *DocumentStore) Save(doc *domain.Document) error {
	return store.bolt.Update(func(tx *bolt.Tx) error {
		err := store.saveDoc(tx, doc)
		if err != nil {
			return err
		}

		return nil
	})
}

func (store *DocumentStore) LoadDocument(id domain.DocumentID) (*domain.Document, error) {
	var doc domain.Document
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketDocs)
		data := bucket.Get([]byte(id.String()))
		if data == nil {
			return ErrDocumentNotFound
		}

		// Deserialize DocDb
		var docDb DocDb
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&docDb)
		if err != nil {
			return err
		}

		// Populate domain.Document
		doc.ID = domain.DocumentID(docDb.ID)
		doc.Title = docDb.Title
		doc.Date, _ = time.Parse(time.RFC3339, docDb.Date)
		doc.Blocks = make([]*domain.Block, len(docDb.Blocks))

		for i, blockDb := range docDb.Blocks {
			doc.Blocks[i] = &domain.Block{
				ID:      domain.BlockID(blockDb.ID),
				Content: blockDb.Content,
				Indent:  blockDb.Ident,
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &doc, nil
}

type DocumentSummary struct {
	ID    domain.DocumentID
	Title string
	Date  time.Time
}

func (store *DocumentStore) ListDocuments() ([]DocumentSummary, error) {
	var summaries []DocumentSummary
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketDocs)
		return bucket.ForEach(func(k, v []byte) error {
			// Deserialize DocDb
			var docDb DocDb
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&docDb)
			if err != nil {
				return err
			}

			date, _ := time.Parse(time.RFC3339, docDb.Date)

			summaries = append(summaries, DocumentSummary{
				ID:    domain.DocumentID(docDb.ID),
				Title: docDb.Title,
				Date:  date,
			})
			return nil
		})
	})

	if err != nil {
		return nil, err
	}

	return summaries, nil
}
