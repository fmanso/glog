package db

import (
	"bytes"
	"encoding/gob"
	"errors"
	"glog/domain"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	_ "github.com/boltdb/bolt"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

var ErrDocumentNotFound = errors.New("document not found")

type DocumentStore struct {
	path                     string
	bolt                     *bolt.DB
	bucketDocs               []byte
	bucketTimeIndex          []byte
	bucketTitleIndex         []byte
	bucketTitleInvertedIndex []byte
	bucketWordIndex          []byte
	wordBlockIndex           *wordIndex
	wordTitleIndex           *wordIndex
	referencesIndex          *referencesIndex
	scheduledIndex           *scheduledTasks
}

func NewDocumentStore(path string) (*DocumentStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	docsKey := []byte("documents")
	timeIndexKey := []byte("time_index")
	titleIndexKey := []byte("title_index")
	titleInvertedIndexKey := []byte("title_inverted_index")
	wordIndexKey := []byte("word_index")

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(docsKey)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(timeIndexKey)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(titleIndexKey)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(titleInvertedIndexKey)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(wordIndexKey)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		_ = db.Close()
		return nil, err
	}

	wordBlockIndex, err := newWordIndex(db, "word_block_index", "doc_word_block_index", getWordsFromBlocks)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	wordTitleIndex, err := newWordIndex(db, "word_title_index", "doc_word_title_index", getWordsFromTitle)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	referencesIndex, err := newReferencesIndex(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	scheduledIndex, err := newScheduledTasks(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return &DocumentStore{
		bolt:             db,
		path:             path,
		bucketDocs:       docsKey,
		bucketTimeIndex:  timeIndexKey,
		bucketTitleIndex: titleIndexKey,
		wordBlockIndex:   wordBlockIndex,
		wordTitleIndex:   wordTitleIndex,
		referencesIndex:  referencesIndex,
		scheduledIndex:   scheduledIndex,
	}, nil
}

func (store *DocumentStore) Close() error {
	return store.bolt.Close()
}

func (store *DocumentStore) saveDoc(tx *bolt.Tx, doc *domain.Document) (*DocDb, error) {
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
		return nil, err
	}

	bucket := tx.Bucket(store.bucketDocs)
	return &docDb, bucket.Put([]byte(doc.ID.String()), buf.Bytes())
}

func (store *DocumentStore) saveTimeIndex(tx *bolt.Tx, doc *domain.Document) error {
	log.Printf("Saving time index: %v\n", doc)
	bucket := tx.Bucket(store.bucketTimeIndex)
	// The chance of accidentally overwriting an existing key is negligible
	// since we use RFC3339 formatted timestamps with nanosecond precision.
	err := bucket.Put([]byte(doc.Date.UTC().Format(time.RFC3339)), []byte(doc.ID.String()))
	if err != nil {
		return err
	}

	return nil
}

func (store *DocumentStore) saveTitleIndex(tx *bolt.Tx, doc *domain.Document) error {
	log.Printf("Saving title index: %v\n", doc)

	bucket := tx.Bucket(store.bucketTitleIndex)
	titleLower := strings.ToLower(doc.Title)
	err := bucket.Put([]byte(titleLower), []byte(doc.ID.String()))
	if err != nil {
		return err
	}

	return nil
}

func (store *DocumentStore) Save(doc *domain.Document) error {
	return store.bolt.Update(func(tx *bolt.Tx) error {
		docDb, err := store.saveDoc(tx, doc)
		if err != nil {
			return err
		}
		err = store.saveTimeIndex(tx, doc)
		if err != nil {
			return err
		}

		err = store.saveTitleIndex(tx, doc)
		if err != nil {
			return err
		}

		err = store.wordBlockIndex.save(tx, docDb)
		if err != nil {
			return err
		}

		err = store.wordTitleIndex.save(tx, docDb)
		if err != nil {
			return err
		}

		err = store.referencesIndex.save(tx, docDb)
		if err != nil {
			return err
		}

		err = store.scheduledIndex.save(tx, docDb)
		if err != nil {
			return err
		}

		return nil
	})
}

func (store *DocumentStore) loadDocument(tx *bolt.Tx, id domain.DocumentID) (*domain.Document, error) {
	var doc domain.Document
	bucket := tx.Bucket(store.bucketDocs)
	data := bucket.Get([]byte(id.String()))
	if data == nil {
		return nil, ErrDocumentNotFound
	}

	// Deserialize DocDb
	var docDb DocDb
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&docDb)
	if err != nil {
		return nil, err
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

	return &doc, nil
}

func (store *DocumentStore) LoadDocument(id domain.DocumentID) (*domain.Document, error) {
	var doc *domain.Document
	err := store.bolt.View(func(tx *bolt.Tx) error {
		d, err := store.loadDocument(tx, id)
		if err != nil {
			return err
		}
		doc = d
		return nil
	})

	if err != nil {
		return nil, err
	}

	return doc, nil
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

func (store *DocumentStore) LoadJournals(from time.Time, to time.Time) ([]*domain.Document, error) {
	var docs []*domain.Document
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketTimeIndex)
		// Start from 'from' setting hours, minutes, seconds, nanoseconds to zero
		current := time.Date(from.Year(), from.Month(), from.Day(), 6, 0, 0, 0, time.UTC)
		end := time.Date(to.Year(), to.Month(), to.Day(), 6, 0, 0, 0, time.UTC)

		for !current.After(end) {
			data := bucket.Get([]byte(current.Format(time.RFC3339)))
			if data != nil {
				id, err := uuid.Parse(string(data))
				if err != nil {
					return err
				}

				d, err := store.loadDocument(tx, domain.DocumentID(id))
				if err != nil {
					if errors.Is(err, ErrDocumentNotFound) {
						// Ignore missing document
					} else {
						return err
					}
				} else {
					docs = append(docs, d)
				}
			}

			current = current.Add(24 * time.Hour)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return docs, nil
}

func (store *DocumentStore) LoadDocumentByTitle(title string) (*domain.Document, error) {
	var docId domain.DocumentID
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketTitleIndex)
		data := bucket.Get([]byte(strings.ToLower(title)))
		if data == nil {
			return ErrDocumentNotFound
		}

		id, err := uuid.Parse(string(data))
		if err != nil {
			return err
		}

		docId = domain.DocumentID(id)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return store.LoadDocument(docId)
}

func (store *DocumentStore) Search(query string) ([]domain.DocumentID, error) {
	var resultIDs []domain.DocumentID
	err := store.bolt.View(func(tx *bolt.Tx) error {
		ids, err := store.wordBlockIndex.Search(tx, query)
		if err != nil {
			return err
		}

		for _, id := range ids {
			resultIDs = append(resultIDs, domain.DocumentID(id))
		}

		ids, err = store.wordTitleIndex.Search(tx, query)
		if err != nil {
			return err
		}

		for _, id := range ids {
			resultIDs = append(resultIDs, domain.DocumentID(id))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return resultIDs, nil
}

func (store *DocumentStore) GetReferences(title string) ([]domain.DocumentID, error) {
	var resultIDs []domain.DocumentID
	ids, err := store.referencesIndex.getReferences(title)
	if err != nil {
		return nil, err
	}

	for _, id := range ids {
		resultIDs = append(resultIDs, domain.DocumentID(id))
	}

	return resultIDs, nil
}

func (store *DocumentStore) ScheduleTask(date time.Time, docID domain.DocumentID, blockID domain.BlockID) error {
	// Create new time to set hours, minutes, seconds, nanoseconds to zero
	scheduledTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	err := store.bolt.Update(func(tx *bolt.Tx) error {
		return store.scheduledIndex.scheduleTask(tx, scheduledTime, uuid.UUID(docID), uuid.UUID(blockID))
	})
	return err
}

// GetScheduledTasks retrieves scheduled tasks for a specific date and the next 'days' days.
func (store *DocumentStore) GetScheduledTasks(date time.Time, days int) ([]domain.ScheduleTask, error) {
	var tasks []domain.ScheduleTask
	err := store.bolt.View(func(tx *bolt.Tx) error {
		for d := 0; d < days; d++ {
			scheduledTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Add(time.Duration(d) * 24 * time.Hour)
			dbTasks, err := store.scheduledIndex.getScheduledTasks(tx, scheduledTime)
			if err != nil {
				return err
			}

			for _, dbTask := range dbTasks {
				tasks = append(tasks, domain.ScheduleTask{
					ID:      dbTask.ID,
					DocID:   domain.DocumentID(dbTask.DocDbID),
					BlockID: domain.BlockID(dbTask.BlockDbID),
					Time:    scheduledTime,
				})
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}
