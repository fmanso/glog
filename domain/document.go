package domain

import (
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type DocumentID uuid.UUID

func (id DocumentID) String() string {
	return uuid.UUID(id).String()
}

type Document struct {
	ID         DocumentID
	Title      string
	Date       DateTime
	Body       []*Paragraph
	paragraphs int
}

func NewDocument(title string, date DateTime) *Document {
	return &Document{
		ID:    DocumentID(uuid.New()),
		Title: title,
		Date:  date,
		Body:  []*Paragraph{},
	}
}

func (d *Document) InsertParagraph(index int, content string) *Paragraph {
	log.Debugf("Inserting paragraph at index %d with content: %s", index, content)
	log.Debugf("Current paragraphs %v", d.Body)
	para := NewParagraph(Content(content))

	// Insert at the specified index
	if index < 0 || index > len(d.Body) {
		d.Body = append(d.Body, para)
	} else {
		d.Body = append(d.Body[:index], append([]*Paragraph{para}, d.Body[index:]...)...)
	}

	d.paragraphs++
	log.Debugf("Paragraphs %v", d.paragraphs)
	return para
}
