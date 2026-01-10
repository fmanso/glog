package db

import (
	"bytes"
	"encoding/gob"

	"github.com/google/uuid"
)

func decodeUUIDSet(data []byte) map[uuid.UUID]struct{} {
	var docIDs map[uuid.UUID]struct{}
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(&docIDs)
	if err != nil {
		return make(map[uuid.UUID]struct{})
	}
	return docIDs
}

func encodeUUIDSet(docIDs map[uuid.UUID]struct{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(docIDs)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
