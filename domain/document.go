package domain

import (
	"fmt"
	"log"

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

func NewDocument(title string, date DateTime) *Document {
	return &Document{
		ID:    DocumentID(uuid.New()),
		Title: title,
		Date:  date,
		Body:  []*Paragraph{},
	}
}

func (d *Document) InsertParagraphAt(index int, content string, indentation int) error {
	para := NewParagraph(Content(content), indentation)
	if index < 0 || index > len(d.Body) {
		d.Body = append(d.Body, para)
	} else {
		d.Body = append(d.Body[:index], append([]*Paragraph{para}, d.Body[index:]...)...)
	}

	return nil
}

func (d *Document) Indent(index int) error {
	if index <= 0 || index >= len(d.Body) {
		return fmt.Errorf("index out of range")
	}

	para := d.Body[index]
	para.Indentation += 1

	return nil
}

func (d *Document) Outdent(index int) error {
	if index < 0 || index >= len(d.Body) {
		return fmt.Errorf("index out of range")
	}

	para := d.Body[index]
	if para.Indentation > 0 {
		para.Indentation -= 1
	} else {
		log.Printf("Paragraph ID: %s is already at indentation level 0", para.ID.String())
	}

	return nil
}
