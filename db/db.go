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

type DocDb struct {
	ID    uuid.UUID
	Title string
	Date  string
	Body  []uuid.UUID
}

type ParagraphDb struct {
	ID          uuid.UUID
	DocumentID  uuid.UUID
	Content     string
	Indentation int
}

func (store *DocumentStore) saveDoc(tx *bolt.Tx, doc *domain.Document) error {
	log.Printf("Saving document: %v\n", doc)

	// Create DocDb from domain.Document
	docDb := DocDb{
		ID:    uuid.UUID(doc.ID),
		Title: doc.Title,
		Date:  string(doc.Date),
		Body:  make([]uuid.UUID, len(doc.Body)),
	}

	for i, para := range doc.Body {
		docDb.Body[i] = uuid.UUID(para.ID)
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

func (store *DocumentStore) SetParagraphContent(id domain.ParagraphID, content domain.Content) error {
	bucket := store.bucketParagraphs
	return store.bolt.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucket)
		data := b.Get([]byte(id.String()))
		if data == nil {
			return ErrParagraphNotFound
		}

		var paraDb ParagraphDb
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&paraDb)
		if err != nil {
			return err
		}

		paraDb.Content = string(content)

		var outBuf bytes.Buffer
		enc := gob.NewEncoder(&outBuf)
		err = enc.Encode(paraDb)
		if err != nil {
			return err
		}

		return b.Put([]byte(paraDb.ID.String()), outBuf.Bytes())
	})
}

func (store *DocumentStore) saveParagraphs(tx *bolt.Tx, doc *domain.Document) error {
	bucket := tx.Bucket(store.bucketParagraphs)

	for _, para := range doc.Body {
		paraDb := ParagraphDb{
			ID:          uuid.UUID(para.ID),
			DocumentID:  uuid.UUID(doc.ID),
			Content:     string(para.Content),
			Indentation: para.Indentation,
		}

		// Serialize ParagraphDb
		var paraBuf bytes.Buffer
		paraEnc := gob.NewEncoder(&paraBuf)
		err := paraEnc.Encode(paraDb)
		if err != nil {
			return err
		}

		// Save to BoltDB
		err = bucket.Put([]byte(paraDb.ID.String()), paraBuf.Bytes())
		if err != nil {
			return err
		}
	}

	return nil
}

func (store *DocumentStore) Save(doc *domain.Document) error {
	return store.bolt.Update(func(tx *bolt.Tx) error {
		err := store.saveDoc(tx, doc)
		if err != nil {
			return err
		}

		err = store.saveParagraphs(tx, doc)
		if err != nil {
			return err
		}

		err = store.updateOrCreateSliceForTime(tx, doc.Date, doc.ID)
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
		doc.Date = domain.DateTime(docDb.Date)
		doc.Body = make([]*domain.Paragraph, 0)

		bucket = tx.Bucket(store.bucketParagraphs)
		for _, paraID := range docDb.Body {
			data := bucket.Get([]byte(paraID.String()))
			if data == nil {
				return ErrParagraphNotFound
			}

			var paraDb ParagraphDb
			buf := bytes.NewBuffer(data)
			dec := gob.NewDecoder(buf)
			err := dec.Decode(&paraDb)
			if err != nil {
				return err
			}

			para := &domain.Paragraph{
				ID:          domain.ParagraphID(paraDb.ID),
				Content:     domain.Content(paraDb.Content),
				Indentation: paraDb.Indentation,
			}
			doc.Body = append(doc.Body, para)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &doc, nil
}

func (store *DocumentStore) GetDocumentFor(time time.Time) (*domain.Document, error) {
	date := domain.ToDateTime(time)
	var docIds map[domain.DocumentID]struct{}
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketTimeIndex)
		data := bucket.Get([]byte(date))
		if data == nil {
			return nil
		}

		// Deserialize the slice of DocumentIDs
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&docIds)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for docID := range docIds {
		doc, err := store.LoadDocument(docID)
		if err != nil {
			return nil, err
		}
		return doc, nil
	}

	return nil, ErrDocumentNotFound
}

//func (store *DocumentStore) indexTerms(tx *bolt.Tx, doc *domain.Document, paragraphs []domain.Paragraph) error {
//	bucket := tx.Bucket(store.bucketTermsIndex)
//	for _, para := range paragraphs {
//		words := strings.Fields(string(para.Content))
//		for _, word := range words {
//			term := strings.ToLower(word)
//			indices := bucket.Get([]byte(term))
//			if indices == nil {
//				docIds := []domain.DocumentID{doc.ID}
//				// Serialize the slice
//				var buf bytes.Buffer
//				enc := gob.NewEncoder(&buf)
//				err := enc.Encode(docIds)
//				if err != nil {
//					return err
//				}
//				// Store the serialized slice
//				err = bucket.Put([]byte(term), buf.Bytes())
//				if err != nil {
//					return err
//				}
//			} else {
//				// Deserialize the existing slice
//				var docIds []domain.DocumentID
//				buf := bytes.NewBuffer(indices)
//				dec := gob.NewDecoder(buf)
//				err := dec.Decode(&docIds)
//				if err != nil {
//					return err
//				}
//				// Check if doc.ID is already in docIds
//				found := false
//				for _, id := range docIds {
//					if id == doc.ID {
//						found = true
//						break
//					}
//				}
//				if !found {
//					// Append the new DocumentID
//					docIds = append(docIds, doc.ID)
//					// Serialize the updated slice
//					var outBuf bytes.Buffer
//					enc := gob.NewEncoder(&outBuf)
//					err = enc.Encode(docIds)
//					if err != nil {
//						return err
//					}
//					// Store the serialized slice
//					err = bucket.Put([]byte(term), outBuf.Bytes())
//					if err != nil {
//						return err
//					}
//				}
//			}
//		}
//	}
//
//	return nil
//}
//

func (store *DocumentStore) updateOrCreateSliceForTime(tx *bolt.Tx, dateTime domain.DateTime, docID domain.DocumentID) error {
	// Get or create a slice of DocumentIDs for the given time
	bucket := tx.Bucket(store.bucketTimeIndex)
	documents := bucket.Get([]byte(dateTime))
	if documents == nil {
		docIds := map[domain.DocumentID]struct{}{}
		docIds[docID] = struct{}{}
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
		var docIds map[domain.DocumentID]struct{}
		buf := bytes.NewBuffer(documents)
		dec := gob.NewDecoder(buf)
		err := dec.Decode(&docIds)
		if err != nil {
			return err
		}
		docIds[docID] = struct{}{}
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

//func (store *DocumentStore) LoadDocument(id domain.DocumentID) (*domain.Document, map[domain.ParagraphID]domain.Paragraph, error) {
//	var doc domain.Document
//	paras := make(map[domain.ParagraphID]domain.Paragraph)
//	err := store.bolt.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket(store.bucketDocs)
//		data := bucket.Get([]byte(id.String()))
//		if data == nil {
//			return ErrDocumentNotFound
//		}
//
//		err := doc.Deserialize(data)
//		if err != nil {
//			return err
//		}
//
//		bucket = tx.Bucket(store.bucketParagraphs)
//		for _, paraID := range doc.Body {
//			data := bucket.Get([]byte(paraID.String()))
//			if data == nil {
//				return ErrParagraphNotFound
//			}
//
//			var para domain.Paragraph
//			err := para.Deserialize(data)
//			if err != nil {
//				return err
//			}
//			paras[paraID] = para
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return &doc, paras, nil
//}
//
//func (store *DocumentStore) LoadDate(date domain.DateTime) (*domain.Document, map[domain.ParagraphID]domain.Paragraph, error) {
//	var docIds []domain.DocumentID
//	err := store.bolt.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket(store.bucketTimeIndex)
//		data := bucket.Get([]byte(date))
//		if data == nil {
//			return nil
//		}
//
//		// Deserialize the slice of DocumentIDs
//		buf := bytes.NewBuffer(data)
//		dec := gob.NewDecoder(buf)
//		err := dec.Decode(&docIds)
//		if err != nil {
//			return err
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		return nil, nil, err
//	}
//
//	if len(docIds) == 0 {
//		return nil, nil, ErrDocumentNotFound
//	}
//
//	doc, paras, err := store.LoadDocument(docIds[0])
//	if err != nil {
//		return nil, nil, err
//	}
//
//	return doc, paras, nil
//}
//
//func (store *DocumentStore) Load(from, to domain.DateTime) ([]domain.DocumentID, error) {
//	docs := make([]domain.DocumentID, 0)
//	err := store.bolt.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket(store.bucketTimeIndex)
//		fromBytes := []byte(from)
//		toBytes := []byte(to)
//
//		c := bucket.Cursor()
//		for k, v := c.Seek(fromBytes); k != nil && bytes.Compare(k, toBytes) <= 0; k, v = c.Next() {
//			// Deserialize the slice of DocumentIDs
//			var docIds []domain.DocumentID
//			buf := bytes.NewBuffer(v)
//			dec := gob.NewDecoder(buf)
//			err := dec.Decode(&docIds)
//			if err != nil {
//				return err
//			}
//			docs = append(docs, docIds...)
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		return nil, err
//	}
//
//	return docs, nil
//}
//
//func (store *DocumentStore) Search(terms []string) ([]domain.DocumentID, error) {
//	docs := make([]domain.DocumentID, 0)
//	err := store.bolt.View(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket(store.bucketTermsIndex)
//		for _, term := range terms {
//			term = strings.ToLower(term)
//			data := bucket.Get([]byte(term))
//			if data == nil {
//				continue
//			}
//
//			// Deserialize the slice of DocumentIDs
//			var docIds []domain.DocumentID
//			buf := bytes.NewBuffer(data)
//			dec := gob.NewDecoder(buf)
//			err := dec.Decode(&docIds)
//			if err != nil {
//				return err
//			}
//			docs = append(docs, docIds...)
//		}
//
//		return nil
//	})
//
//	if err != nil {
//		return nil, err
//	}
//
//	return docs, nil
//}
//
//func (store *DocumentStore) GetReferences(documentID uuid.UUID) ([]domain.Document, error) {
//	return nil, nil
//}
