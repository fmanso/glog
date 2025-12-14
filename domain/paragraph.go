package domain

import (
	"bytes"
	"encoding/gob"

	"github.com/google/uuid"
)

type Content string

type ParagraphID uuid.UUID

func (id ParagraphID) String() string {
	return uuid.UUID(id).String()
}

type Paragraph struct {
	ID         ParagraphID
	DocumentID DocumentID
	Parent     *ParagraphID
	Children   []*ParagraphID
	Content    Content
	References []ParagraphID
}

func (paragraph *Paragraph) Serialize() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)

	err := enc.Encode(*paragraph)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (paragraph *Paragraph) Deserialize(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)

	err := dec.Decode(paragraph)
	if err != nil {
		return err
	}

	return nil
}
