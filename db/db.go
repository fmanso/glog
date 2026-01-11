package db

import (
	"bytes"
	"encoding/gob"
	"errors"
	"glog/domain"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
	bolt "go.etcd.io/bbolt"
)

var ErrDocumentNotFound = errors.New("document not found")

// failedIndexEntry tracks a document that failed to index
type failedIndexEntry struct {
	docID       string
	attempts    int
	lastAttempt time.Time
	doc         *DocDb
}

// IndexHealth represents the health status of the search index
type IndexHealth struct {
	IsHealthy          bool
	FailedDocuments    int
	LastHealthCheck    time.Time
	RequiresReindex    bool
	HealthCheckMessage string
}

// DocumentStore provides thread-safe access to document storage and search.
//
// Thread-Safety:
//
// DocumentStore is safe for concurrent use. All methods that access
// the search index are protected by an internal RWMutex:
//
//   - Save() acquires a read lock when indexing (allows concurrent saves)
//   - Search() acquires a read lock (allows concurrent searches)
//   - ReindexSearch() acquires a write lock (blocks all other operations)
//   - Close() acquires a write lock (ensures clean shutdown)
//
// This ensures that ReindexSearch cannot run concurrently with any
// other search index operations, preventing race conditions.
type DocumentStore struct {
	path               string
	bolt               *bolt.DB
	bucketDocs         []byte
	bucketTimeIndex    []byte
	bucketTitleIndex   []byte
	bucketJournalIndex []byte
	search             *bleveSearch
	searchMu           sync.RWMutex // protects search index operations
	referencesIndex    *referencesIndex
	scheduledIndex     *scheduledTasks

	// Index health tracking
	failedIndexes   map[string]*failedIndexEntry
	failedIndexesMu sync.Mutex
	indexHealth     IndexHealth
	indexHealthMu   sync.RWMutex
}

func NewDocumentStore(path string) (*DocumentStore, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	docsKey := []byte("documents")
	timeIndexKey := []byte("time_index")
	titleIndexKey := []byte("title_index")
	journalIndexKey := []byte("journal_index")

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

		_, err = tx.CreateBucketIfNotExists(journalIndexKey)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		_ = db.Close()
		return nil, err
	}

	search, err := openBleveSearch(bleveIndexPath(path))
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	referencesIndex, err := newReferencesIndex(db)
	if err != nil {
		_ = db.Close()
		_ = search.Close()
		return nil, err
	}

	scheduledIndex, err := newScheduledTasks(db)
	if err != nil {
		_ = db.Close()
		_ = search.Close()
		return nil, err
	}

	store := &DocumentStore{
		bolt:               db,
		path:               path,
		bucketDocs:         docsKey,
		bucketTimeIndex:    timeIndexKey,
		bucketTitleIndex:   titleIndexKey,
		bucketJournalIndex: journalIndexKey,
		search:             search,
		referencesIndex:    referencesIndex,
		scheduledIndex:     scheduledIndex,
		failedIndexes:      make(map[string]*failedIndexEntry),
		indexHealth: IndexHealth{
			IsHealthy:       true,
			LastHealthCheck: time.Now(),
		},
	}

	// Perform initial health check
	store.checkIndexHealth()

	return store, nil
}

func (store *DocumentStore) Close() error {
	// Acquire write lock to ensure no operations are in-flight during shutdown
	store.searchMu.Lock()
	defer store.searchMu.Unlock()

	if err := store.search.Close(); err != nil {
		boltErr := store.bolt.Close()
		return errors.Join(err, boltErr)
	}
	return store.bolt.Close()
}

// checkIndexHealth performs a health check on the search index
func (store *DocumentStore) checkIndexHealth() {
	store.indexHealthMu.Lock()
	defer store.indexHealthMu.Unlock()

	store.indexHealth.LastHealthCheck = time.Now()
	store.indexHealth.FailedDocuments = len(store.failedIndexes)

	// Check if index is accessible
	if store.search == nil || store.search.index == nil {
		store.indexHealth.IsHealthy = false
		store.indexHealth.RequiresReindex = true
		store.indexHealth.HealthCheckMessage = "Search index is not initialized"
		return
	}

	// Consider unhealthy if there are failed documents
	if store.indexHealth.FailedDocuments > 0 {
		store.indexHealth.IsHealthy = false
		store.indexHealth.HealthCheckMessage = "Some documents failed to index"
	} else {
		store.indexHealth.IsHealthy = true
		store.indexHealth.HealthCheckMessage = "Index is healthy"
	}
}

// GetIndexHealth returns the current health status of the search index
func (store *DocumentStore) GetIndexHealth() IndexHealth {
	store.indexHealthMu.RLock()
	defer store.indexHealthMu.RUnlock()
	return store.indexHealth
}

// indexDocWithRetry attempts to index a document with exponential backoff
func (store *DocumentStore) indexDocWithRetry(doc *DocDb, maxAttempts int) error {
	docID := doc.ID.String()

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err := store.search.IndexDoc(doc)
		if err == nil {
			// Success - remove from failed index tracking if it was there
			store.failedIndexesMu.Lock()
			delete(store.failedIndexes, docID)
			store.failedIndexesMu.Unlock()
			return nil
		}

		// Log the error
		log.Warnf("Bleve indexing attempt %d/%d failed for document %s: %v", attempt, maxAttempts, docID, err)

		// If this is not the last attempt, wait with exponential backoff
		if attempt < maxAttempts {
			backoff := time.Duration(attempt*attempt) * 100 * time.Millisecond
			time.Sleep(backoff)
		}
	}

	// All attempts failed - track the failure
	store.failedIndexesMu.Lock()
	entry, exists := store.failedIndexes[docID]
	if !exists {
		entry = &failedIndexEntry{
			docID: docID,
			doc:   doc,
		}
		store.failedIndexes[docID] = entry
	}
	entry.attempts++
	entry.lastAttempt = time.Now()
	store.failedIndexesMu.Unlock()

	// Update health check
	store.checkIndexHealth()

	return errors.New("failed to index document after multiple attempts")
}

// RetryFailedIndexing attempts to reindex all documents that previously failed
func (store *DocumentStore) RetryFailedIndexing() (int, error) {
	store.failedIndexesMu.Lock()
	failedDocs := make([]*DocDb, 0, len(store.failedIndexes))
	for _, entry := range store.failedIndexes {
		failedDocs = append(failedDocs, entry.doc)
	}
	store.failedIndexesMu.Unlock()

	if len(failedDocs) == 0 {
		return 0, nil
	}

	successCount := 0
	store.searchMu.RLock()
	defer store.searchMu.RUnlock()

	for _, doc := range failedDocs {
		err := store.indexDocWithRetry(doc, 3)
		if err == nil {
			successCount++
		}
	}

	store.checkIndexHealth()
	return successCount, nil
}

func (store *DocumentStore) saveDoc(tx *bolt.Tx, doc *domain.Document) (*DocDb, error) {

	// Create DocDb from domain.Document
	docDb := DocDb{
		ID:        uuid.UUID(doc.ID),
		Title:     doc.Title,
		Date:      doc.Date.UTC().Format(time.RFC3339),
		IsJournal: doc.IsJournal,
		Blocks:    make([]*BlockDb, len(doc.Blocks)),
	}

	for i, block := range doc.Blocks {
		docDb.Blocks[i] = &BlockDb{
			ID:      uuid.UUID(block.ID),
			Content: block.Content,
			Indent:  block.Indent,
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
	bucket := tx.Bucket(store.bucketTitleIndex)
	titleLower := strings.ToLower(doc.Title)
	err := bucket.Put([]byte(titleLower), []byte(doc.ID.String()))
	if err != nil {
		return err
	}

	return nil
}

func (store *DocumentStore) saveJournalIndex(tx *bolt.Tx, doc *domain.Document) error {
	bucket := tx.Bucket(store.bucketJournalIndex)
	if doc.IsJournal {
		// Index by date (normalized to start of day in UTC) -> document ID
		// This allows efficient lookup of journals by date
		dateKey := time.Date(doc.Date.Year(), doc.Date.Month(), doc.Date.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
		return bucket.Put([]byte(dateKey), []byte(doc.ID.String()))
	}
	return nil
}

func (store *DocumentStore) Save(doc *domain.Document) error {
	var savedDoc *DocDb
	if err := store.bolt.Update(func(tx *bolt.Tx) error {
		docDb, err := store.saveDoc(tx, doc)
		if err != nil {
			return err
		}
		savedDoc = docDb

		err = store.saveTimeIndex(tx, doc)
		if err != nil {
			return err
		}

		err = store.saveTitleIndex(tx, doc)
		if err != nil {
			return err
		}

		err = store.saveJournalIndex(tx, doc)
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
	}); err != nil {
		return err
	}

	// Index the document in the search index with retry logic.
	// Note: This happens outside the BoltDB transaction. If indexing fails
	// after retries, the document is still saved to the database but won't
	// be searchable until RetryFailedIndexing() or ReindexSearch() is called.
	// We use RLock here to allow concurrent Save operations while preventing
	// ReindexSearch from running concurrently.
	if savedDoc != nil {
		store.searchMu.RLock()
		err := store.indexDocWithRetry(savedDoc, 3) // Retry up to 3 times
		store.searchMu.RUnlock()
		if err != nil {
			log.Errorf("Bleve indexing failed after retries: %v", err)
			// Don't return error - document is saved, just not indexed
		}
	}

	return nil
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
	doc.IsJournal = docDb.IsJournal
	doc.Blocks = make([]*domain.Block, len(docDb.Blocks))

	for i, blockDb := range docDb.Blocks {
		doc.Blocks[i] = &domain.Block{
			ID:      domain.BlockID(blockDb.ID),
			Content: blockDb.Content,
			Indent:  blockDb.Indent,
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
		bucket := tx.Bucket(store.bucketJournalIndex)
		cursor := bucket.Cursor()

		// Normalize to start of day in UTC for consistent key lookup
		startDay := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
		endDay := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

		// Iterate through the journal index in the date range
		minKey := []byte(startDay.Format(time.RFC3339))
		maxKey := []byte(endDay.Add(24 * time.Hour).Format(time.RFC3339))

		for k, v := cursor.Seek(minKey); k != nil && bytes.Compare(k, maxKey) < 0; k, v = cursor.Next() {
			id, err := uuid.Parse(string(v))
			if err != nil {
				return err
			}

			d, err := store.loadDocument(tx, domain.DocumentID(id))
			if err != nil {
				if errors.Is(err, ErrDocumentNotFound) {
					// Ignore missing document
					continue
				} else {
					return err
				}
			}

			docs = append(docs, d)
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
	store.searchMu.RLock()
	ids, err := store.search.Search(query)
	store.searchMu.RUnlock()

	if err != nil {
		return nil, err
	}

	resultIDs := make([]domain.DocumentID, 0, len(ids))
	for _, id := range ids {
		resultIDs = append(resultIDs, domain.DocumentID(id))
	}

	return resultIDs, nil
}

// ReindexSearch rebuilds the search index from scratch by re-indexing all
// documents in the database. This operation acquires an exclusive write lock,
// blocking all concurrent Save() and Search() operations until the reindex
// is complete. Use this when the search index becomes corrupted or out of sync.
func (store *DocumentStore) ReindexSearch() error {
	// Acquire write lock to ensure no other operations access the search index
	// while we're closing, deleting, and recreating it.
	store.searchMu.Lock()
	defer store.searchMu.Unlock()

	// Create new index first before closing the old one to avoid leaving
	// store.search pointing to a closed index if recreation fails.
	oldSearch := store.search

	// Close and delete the old index directory
	if err := oldSearch.Close(); err != nil {
		return err
	}
	if err := oldSearch.DeleteIndexDir(); err != nil {
		return err
	}

	// Create new index
	newSearch, err := openBleveSearch(bleveIndexPath(store.path))
	if err != nil {
		// If we can't create a new index after successfully deleting the old one,
		// surface the error to the caller rather than attempting a second,
		// potentially empty, fallback index at the same path.
		return err
	}

	// Assign new index to store
	store.search = newSearch

	// Reindex all documents
	return store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketDocs)
		return bucket.ForEach(func(k, v []byte) error {
			var docDb DocDb
			buf := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buf)
			if err := dec.Decode(&docDb); err != nil {
				return err
			}
			return store.search.IndexDoc(&docDb)
		})
	})
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
			// Normalize to start of day in local timezone, then convert to UTC for storage lookup
			scheduledTime := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location()).Add(time.Duration(d) * 24 * time.Hour)
			dbTasks, err := store.scheduledIndex.getScheduledTasks(tx, scheduledTime.UTC())
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

// Delete removes a document and all its index entries from the store.
// This includes removing from: documents bucket, time_index, title_index,
// journal_index, references_index, scheduled_index, and Bleve search index.
func (store *DocumentStore) Delete(id uuid.UUID) error {
	var docDb *DocDb

	// First, load the document to get its metadata for index cleanup
	err := store.bolt.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(store.bucketDocs)
		data := bucket.Get([]byte(id.String()))
		if data == nil {
			return ErrDocumentNotFound
		}

		var doc DocDb
		buf := bytes.NewBuffer(data)
		dec := gob.NewDecoder(buf)
		if err := dec.Decode(&doc); err != nil {
			return err
		}
		docDb = &doc
		return nil
	})

	if err != nil {
		return err
	}

	// Delete from all BoltDB indexes in a single transaction
	if err := store.bolt.Update(func(tx *bolt.Tx) error {
		// Delete from documents bucket
		docsBucket := tx.Bucket(store.bucketDocs)
		if err := docsBucket.Delete([]byte(id.String())); err != nil {
			return err
		}

		// Delete from time_index
		timeBucket := tx.Bucket(store.bucketTimeIndex)
		if timeBucket != nil {
			date, _ := time.Parse(time.RFC3339, docDb.Date)
			timeKey := date.UTC().Format(time.RFC3339)
			_ = timeBucket.Delete([]byte(timeKey))
		}

		// Delete from title_index
		titleBucket := tx.Bucket(store.bucketTitleIndex)
		if titleBucket != nil {
			titleKey := strings.ToLower(docDb.Title)
			_ = titleBucket.Delete([]byte(titleKey))
		}

		// Delete from journal_index (if it's a journal)
		if docDb.IsJournal {
			journalBucket := tx.Bucket(store.bucketJournalIndex)
			if journalBucket != nil {
				date, _ := time.Parse(time.RFC3339, docDb.Date)
				dateKey := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
				_ = journalBucket.Delete([]byte(dateKey))
			}
		}

		// Delete from references_index
		if err := store.referencesIndex.delete(tx, docDb); err != nil {
			return err
		}

		// Delete from scheduled_index
		if err := store.scheduledIndex.delete(tx, docDb); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// Delete from Bleve search index
	store.searchMu.RLock()
	err = store.search.DeleteDoc(id.String())
	store.searchMu.RUnlock()

	if err != nil {
		log.Warnf("Failed to delete document from search index: %v", err)
		// Don't return error - document is deleted from main storage
	}

	// Remove from failed indexes tracking if present
	store.failedIndexesMu.Lock()
	delete(store.failedIndexes, id.String())
	store.failedIndexesMu.Unlock()

	return nil
}
