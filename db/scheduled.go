package db

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
	"github.com/google/uuid"
)

type scheduledTasks struct {
	db                     *bolt.DB
	scheduledIndex         []byte
	scheduledInvertedIndex []byte
}

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
