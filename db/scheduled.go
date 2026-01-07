package db

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

type scheduledTasks struct {
	db                     *bolt.DB
	scheduledIndex         []byte // keys are dates in "YYYY-MM-DD" format, values are encoded ScheduleTaskDb slices
	scheduledInvertedIndex []byte // keys are "docID_blockID", values are set of scheduled dates
}

var scheduledRegex = regexp.MustCompile(`/scheduled (\d{4}-\d{2}-\d{2})`)

func newScheduledTasks(db *bolt.DB) (*scheduledTasks, error) {
	scheduledIndexKey := []byte("scheduled_index")
	scheduledInvertedIndexKey := []byte("scheduled_inverted_index")

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(scheduledIndexKey)
		if err != nil {
			return err
		}

		_, err = tx.CreateBucketIfNotExists(scheduledInvertedIndexKey)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &scheduledTasks{
		db:                     db,
		scheduledIndex:         scheduledIndexKey,
		scheduledInvertedIndex: scheduledInvertedIndexKey,
	}, nil
}

func decodeScheduledDates(data []byte) (map[string]struct{}, error) {
	var dateSet map[string]struct{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&dateSet)
	if err != nil {
		return nil, err
	}
	return dateSet, nil
}

func encodeScheduledDates(dateSet map[string]struct{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(dateSet)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *scheduledTasks) setInvertedIndex(tx *bolt.Tx, doc *DocDb) error {
	bucket := tx.Bucket(s.scheduledInvertedIndex)
	if bucket == nil {
		return fmt.Errorf("scheduled inverted index bucket not found")
	}

	for _, block := range doc.Blocks {
		scheduledDates := extractScheduledDates(block.Content)
		if len(scheduledDates) == 0 {
			continue
		}

		key := fmt.Sprintf("%s_%s", doc.ID.String(), block.ID.String())
		data := bucket.Get([]byte(key))
		dateSet, err := decodeScheduledDates(data)
		if err != nil || dateSet == nil {
			dateSet = make(map[string]struct{})
		}

		for _, date := range scheduledDates {
			dateStr := date.Format("2006-01-02")
			dateSet[dateStr] = struct{}{}
		}

		// Encode and store updated date set
		encodedDateSet, err := encodeScheduledDates(dateSet)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(key), encodedDateSet)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *scheduledTasks) getInvertedIndexDates(tx *bolt.Tx, docID uuid.UUID, blockID uuid.UUID) (map[string]struct{}, error) {
	bucket := tx.Bucket(s.scheduledInvertedIndex)
	if bucket == nil {
		return nil, fmt.Errorf("scheduled inverted index bucket not found")
	}

	key := fmt.Sprintf("%s_%s", docID.String(), blockID.String())
	data := bucket.Get([]byte(key))
	if data == nil {
		return make(map[string]struct{}), nil // No scheduled dates
	}

	dateSet, err := decodeScheduledDates(data)
	if err != nil {
		return nil, err
	}

	return dateSet, nil
}

func (s *scheduledTasks) removeObsolete(tx *bolt.Tx, doc *DocDb) error {
	for _, block := range doc.Blocks {
		oldDatesSet, err := s.getInvertedIndexDates(tx, doc.ID, block.ID)
		if err != nil {
			return err
		}

		log.Printf("Removing obsolete dates %v", oldDatesSet)

		scheduledDates := extractScheduledDates(block.Content)
		log.Printf("Current scheduled dates for document ID: %s, Title: %s, Block ID: %s, Dates: %v", doc.ID, doc.Title, block.ID, scheduledDates)
		newDatesSet := make(map[string]struct{})
		for _, date := range scheduledDates {
			dateStr := date.Format("2006-01-02")
			newDatesSet[dateStr] = struct{}{}
		}

		for oldDateStr := range oldDatesSet {
			if _, exists := newDatesSet[oldDateStr]; !exists {
				oldDate, err := time.Parse("2006-01-02", oldDateStr)
				if err != nil {
					return err
				}
				log.Printf("Removing obsolete scheduled task for document ID: %s, Title: %s, Block ID: %s, Date: %s", doc.ID, doc.Title, block.ID, oldDateStr)
				err = s.removeScheduledTask(tx, oldDate, doc.ID, block.ID)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *scheduledTasks) removeScheduledTask(tx *bolt.Tx, date time.Time, docID uuid.UUID, blockID uuid.UUID) error {
	bucket := tx.Bucket(s.scheduledIndex)
	if bucket == nil {
		return fmt.Errorf("scheduled index bucket not found")
	}

	key := date.Format("2006-01-02")
	values := bucket.Get([]byte(key))
	if values == nil {
		return nil // No tasks for this date
	}

	tasks, err := decodeScheduleTasksDb(values)
	if err != nil {
		return err
	}

	var updatedTasks []ScheduleTaskDb
	for _, task := range tasks {
		if task.DocDbID == docID && task.BlockDbID == blockID {
			continue // Skip the task to be removed
		}
		updatedTasks = append(updatedTasks, task)
	}

	if len(updatedTasks) == 0 {
		return bucket.Delete([]byte(key))
	}

	encoded, err := encodeScheduleTaskDb(updatedTasks)
	if err != nil {
		return err
	}

	return bucket.Put([]byte(key), encoded)
}

func (s *scheduledTasks) save(tx *bolt.Tx, doc *DocDb) error {
	log.Printf("Updating scheduled tasks for document ID: %s, Title: %s", doc.ID, doc.Title)

	// First, remove obsolete scheduled tasks
	err := s.removeObsolete(tx, doc)
	if err != nil {
		return fmt.Errorf("error removing obsolete scheduled tasks for document ID: %s, Title: %s", doc.ID, doc.Title)
	}

	// Search in each block for ocurrences of scheduled dates in the format /scheduled YYYY-MM-DD
	for _, block := range doc.Blocks {
		scheduledDates := extractScheduledDates(block.Content)
		log.Printf("Found scheduled dates in document ID: %s, Title: %s, Block ID: %s, Dates: %v", doc.ID, doc.Title, block.ID, scheduledDates)
		for _, date := range scheduledDates {
			log.Printf("Scheduling task for document ID: %s, Title: %s, Block ID: %s, Date: %s", doc.ID, doc.Title, block.ID, date.Format("2006-01-02"))
			err := s.scheduleTask(tx, date, doc.ID, block.ID)
			if err != nil {
				return err
			}
		}
	}

	err = s.setInvertedIndex(tx, doc)
	if err != nil {
		return fmt.Errorf("failed to set inverted index: %v", err)
	}

	return nil
}

func extractScheduledDates(content string) []time.Time {
	matches := scheduledRegex.FindAllStringSubmatch(content, -1)
	var dates []time.Time
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		dateStr := match[1]
		date, err := time.Parse("2006-01-02", dateStr)
		if err == nil {
			dates = append(dates, date)
		}
	}
	return dates
}

func (s *scheduledTasks) scheduleTask(tx *bolt.Tx, date time.Time, docID uuid.UUID, blockID uuid.UUID) error {
	bucket := tx.Bucket(s.scheduledIndex)
	if bucket == nil {
		return fmt.Errorf("scheduled index bucket not found")
	}

	key := date.Format("2006-01-02")
	newTask := ScheduleTaskDb{
		ID:        uuid.New(),
		DocDbID:   docID,
		BlockDbID: blockID,
	}

	prevValues := bucket.Get([]byte(key))
	if prevValues != nil {
		// Decode existing tasks
		existingTasks, err := decodeScheduleTasksDb(prevValues)
		if err != nil {
			return err
		}

		existingTasks = append(existingTasks, newTask)

		// Re-encode and store
		encoded, err := encodeScheduleTaskDb(existingTasks)
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), encoded)
	}

	// No existing tasks, create new
	tasks := []ScheduleTaskDb{newTask}

	encoded, err := encodeScheduleTaskDb(tasks)
	if err != nil {
		return err
	}

	return bucket.Put([]byte(key), encoded)
}

func (s *scheduledTasks) getScheduledTasks(tx *bolt.Tx, date time.Time) ([]ScheduleTaskDb, error) {
	bucket := tx.Bucket(s.scheduledIndex)
	if bucket == nil {
		return nil, fmt.Errorf("scheduled index bucket not found")
	}

	key := date.Format("2006-01-02")
	values := bucket.Get([]byte(key))
	if values == nil {
		return []ScheduleTaskDb{}, nil // No tasks for this date
	}

	tasks, err := decodeScheduleTasksDb(values)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func encodeScheduleTaskDb(tasks []ScheduleTaskDb) ([]byte, error) {
	// Encode using gob
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(tasks)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeScheduleTasksDb(data []byte) ([]ScheduleTaskDb, error) {
	var tasks []ScheduleTaskDb
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
