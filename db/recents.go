package db

import (
	"bytes"
	"encoding/gob"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
)

type recentsDocs struct {
	db            *bolt.DB
	recentsBucket []byte
}

func newRecentsDocs(db *bolt.DB) (*recentsDocs, error) {
	recentsBucket := []byte("recents_index")

	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(recentsBucket)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &recentsDocs{
		db:            db,
		recentsBucket: recentsBucket,
	}, nil
}

func deserializeRecents(data []byte) ([]uuid.UUID, error) {
	var dateSet []uuid.UUID
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&dateSet)
	if err != nil {
		return nil, err
	}
	return dateSet, nil
}

func serializeRecents(recents []uuid.UUID) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(recents)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (r *recentsDocs) Get() ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(r.recentsBucket)
		data := bucket.Get([]byte("recents_list"))
		if data != nil {
			r, err := deserializeRecents(data)
			if err != nil {
				return err
			}
			ids = r
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *recentsDocs) delete(tx *bolt.Tx, id uuid.UUID) error {
	bucket := tx.Bucket(r.recentsBucket)
	data := bucket.Get([]byte("recents_list"))
	if data == nil {
		return nil
	}

	recentsList, err := deserializeRecents(data)
	if err != nil {
		return err
	}

	// Remove id from ids
	for i, savedID := range recentsList {
		if savedID == id {
			recentsList = append(recentsList[:i], recentsList[i+1:]...)
			updatedRecentsData, err := serializeRecents(recentsList)
			if err != nil {
				return err
			}

			err = bucket.Put([]byte("recents_list"), updatedRecentsData)
			if err != nil {
				return err
			}

			break
		}
	}

	return nil
}

func (r *recentsDocs) update(tx *bolt.Tx, id uuid.UUID) error {
	bucket := tx.Bucket(r.recentsBucket)

	// Deserialize existing recents list
	recentsData := bucket.Get([]byte("recents_list"))
	var recentsList []uuid.UUID
	if recentsData != nil {
		r, err := deserializeRecents(recentsData)
		if err != nil {
			return err
		}

		recentsList = r
	}

	// Remove doc ID if it already exists
	for i, savedID := range recentsList {
		if savedID == id {
			recentsList = append(recentsList[:i], recentsList[i+1:]...)
			break
		}
	}

	// Prepend the new doc ID
	recentsList = append([]uuid.UUID{id}, recentsList...)

	// Limit the size of the recents list (e.g., to 100 items)
	if len(recentsList) > 100 {
		recentsList = recentsList[:100]
	}

	// Serialize and store the updated recents list
	updatedRecentsData, err := serializeRecents(recentsList)
	if err != nil {
		return err
	}

	return bucket.Put([]byte("recents_list"), updatedRecentsData)
}
