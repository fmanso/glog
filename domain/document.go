package domain

import (
	"fmt"

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

func (d *Document) GetChildren(paraID ParagraphID) ([]*Paragraph, bool) {
	for _, para := range d.Body {
		if para.ID == paraID {
			return para.Children, true
		}
	}

	// If not found at the top level, search recursively in children
	for _, para := range d.Body {
		if parentPara := d.getChildren(paraID, para); parentPara != nil {
			return parentPara.Children, true
		}
	}

	return nil, false
}

func (d *Document) getChildren(paraID ParagraphID, parent *Paragraph) *Paragraph {
	for _, child := range parent.Children {
		if child.ID == paraID {
			return child
		}
		if len(child.Children) > 0 {
			if foundParent := d.getParent(paraID, child); foundParent != nil {
				return foundParent
			}
		}
	}
	return nil
}

func (d *Document) getParent(paraID ParagraphID, parent *Paragraph) *Paragraph {
	for _, child := range parent.Children {
		if child.ID == paraID {
			return parent
		}
		if len(child.Children) > 0 {
			if foundParent := d.getParent(paraID, child); foundParent != nil {
				return foundParent
			}
		}
	}
	return nil
}

func (d *Document) GetParentOf(paraID ParagraphID) *Paragraph {
	for _, para := range d.Body {
		if para.ID == paraID {
			return nil
		}
	}

	// If not found at the top level, search recursively in children
	for _, para := range d.Body {
		if parentPara := d.getParent(paraID, para); parentPara != nil {
			return parentPara
		}
	}

	return nil
}

func (d *Document) GetParagraphByID(paraID ParagraphID) (*Paragraph, bool) {
	for _, para := range d.Body {
		if para.ID == paraID {
			return para, true
		}
	}

	// If not found at the top level, search recursively in children
	for _, para := range d.Body {
		if foundPara, found := d.recursiveSearch(paraID, para.Children); found {
			return foundPara, true
		}
	}

	return nil, false
}

func (d *Document) recursiveSearch(paraID ParagraphID, paragraphs []*Paragraph) (*Paragraph, bool) {
	for _, para := range paragraphs {
		if para.ID == paraID {
			return para, true
		}
		if len(para.Children) > 0 {
			if foundPara, found := d.recursiveSearch(paraID, para.Children); found {
				return foundPara, true
			}
		}
	}

	return nil, false
}

// Indent changes the parent of the paragraph with the given ID to be its preceding sibling.
// It returns the indented paragraph and its new parent, or an error if the operation is not possible.
func (d *Document) Indent(paraID ParagraphID) (*Paragraph, *Paragraph, error) {
	if d.Body[0].ID == paraID {
		return nil, nil, fmt.Errorf("cannot change parent of top-level paragraph")
	}

	currentParent := d.GetParentOf(paraID)
	if currentParent == nil {

		currentIndex := -1
		for i, para := range d.Body {
			if para.ID == paraID {
				currentIndex = i
				break
			}
		}

		if currentIndex == -1 {
			return nil, nil, fmt.Errorf("paragraph with ID %v not found", paraID)
		}

		newParent := d.Body[currentIndex-1]

		// Remove from top-level body
		para := d.Body[currentIndex]
		d.Body = append(d.Body[:currentIndex], d.Body[currentIndex+1:]...)

		// Add to new parent's children
		newParent.AddChild(para)
		return para, newParent, nil
	}

	currentParentIndex := -1
	for i, para := range currentParent.Children {
		if para.ID == paraID {
			currentParentIndex = i
			break
		}
	}

	if currentParentIndex <= 0 {
		return nil, nil, fmt.Errorf("no preceding sibling to indent under")
	}

	newParent := currentParent.Children[currentParentIndex-1]

	// Remove from current parent's children
	para, err := currentParent.RemoveChild(paraID)
	if err != nil {
		return nil, nil, err
	}

	// Add to new parent's children
	newParent.AddChild(para)

	return para, newParent, nil
}
