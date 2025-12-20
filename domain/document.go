package domain

import (
	"time"

	"github.com/google/uuid"
)

type DocumentID uuid.UUID

func (id DocumentID) String() string {
	return uuid.UUID(id).String()
}

type Document struct {
	ID    DocumentID
	Title string
	Date  DateTime
	Body  []*Paragraph
}

func NewDocument(title string, date time.Time) *Document {
	return &Document{
		ID:    DocumentID(uuid.New()),
		Title: title,
		Date:  ToDateTime(date),
		Body:  make([]*Paragraph, 0),
	}
}

func (doc *Document) SetBody(paragraphs []*Paragraph) {
	doc.Body = paragraphs
}
