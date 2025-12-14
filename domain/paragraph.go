package domain

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/google/uuid"
)

type Content string

type ParagraphID uuid.UUID

func (id ParagraphID) String() string {
	return uuid.UUID(id).String()
}

type DocumentID uuid.UUID

func (id DocumentID) String() string {
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

type DateTime string // RFC3339 format

func ToDateTime(t time.Time) DateTime {
	return DateTime(t.Format(time.RFC3339))
}

func (dt DateTime) ToTime() (time.Time, error) {
	return time.Parse(time.RFC3339, string(dt))
}

type Document struct {
	ID    DocumentID
	Title string
	Date  DateTime
	Body  []ParagraphID
}

func NewDocument(title string, date time.Time, paras []Paragraph) *Document {
	body := make([]ParagraphID, len(paras))
	for i, para := range paras {
		body[i] = para.ID
	}

	return &Document{
		ID:    DocumentID(uuid.New()),
		Title: title,
		Date:  ToDateTime(date),
		Body:  body,
	}
}

func (doc *Document) Serialize() ([]byte, error) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)

	err := enc.Encode(*doc)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (doc *Document) Deserialize(data []byte) error {
	b := bytes.NewBuffer(data)
	dec := gob.NewDecoder(b)

	err := dec.Decode(doc)
	if err != nil {
		return err
	}

	return nil
}
