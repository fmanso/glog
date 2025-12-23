package domain

import (
	"fmt"

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
	if index < 0 || index >= len(d.Body) {
		return fmt.Errorf("index out of range")
	}

	if index == 0 {
		return fmt.Errorf("cannot indent the first paragraph further")
	}

	prev := d.Body[index-1]
	if d.Body[index].Indentation >= prev.Indentation+1 {
		return fmt.Errorf("cannot indent paragraph ID: %s beyond the indentation level of the previous paragraph ID: %s", d.Body[index].ID.String(), prev.ID.String())
	}

	originIndentation := d.Body[index].Indentation
	d.Body[index].Indentation += 1

	for i := index + 1; i < len(d.Body); i++ {
		if d.Body[i].Indentation <= originIndentation {
			break
		}
		d.Body[i].Indentation += 1
	}

	return nil
}

func (d *Document) Outdent(index int) error {
	if index < 0 || index >= len(d.Body) {
		return fmt.Errorf("index out of range")
	}

	if d.Body[index].Indentation == 0 {
		return fmt.Errorf("paragraph ID: %s is already at the base indentation level", d.Body[index].ID.String())
	}

	originIndentation := d.Body[index].Indentation
	d.Body[index].Indentation -= 1

	for i := index + 1; i < len(d.Body); i++ {
		if d.Body[i].Indentation <= originIndentation {
			break
		}

		d.Body[i].Indentation -= 1
	}

	return nil
}

func (d *Document) DeleteParagraphAt(index int) error {
	if index < 0 || index >= len(d.Body) {
		return fmt.Errorf("index out of range")
	}

	if index == 0 {
		return fmt.Errorf("cannot delete the first paragraph")
	}

	// Append current text to the previous paragraph
	prevPara := d.Body[index-1]
	currPara := d.Body[index]
	prevPara.Content += "" + currPara.Content
	d.Body = append(d.Body[:index], d.Body[index+1:]...)
	return nil
}
