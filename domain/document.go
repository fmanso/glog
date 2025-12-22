package domain

import (
	"fmt"
	"github.com/google/uuid"
	"log"
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

func (d *Document) AddParagraph(content string) *Paragraph {
	para := NewParagraph(Content(content))
	d.Body = append(d.Body, para)
	d.paragraphs++
	return para
}

func (d *Document) InsertParagraphAfter(paraID ParagraphID, content string) *Paragraph {
	log.Printf("Inserting paragraph after ID %v", paraID)

	parent := d.GetParentOf(paraID)
	if parent == nil {
		// Get index in top-level body
		index := -1
		for i, para := range d.Body {
			if para.ID == paraID {
				index = i
				break
			}
		}

		para := NewParagraph(Content(content))
		// Insert at index position + 1
		d.Body = append(d.Body[:index+1], append([]*Paragraph{para}, d.Body[index+1:]...)...)
		d.paragraphs++
		return para
	}

	// Get index in parent's children
	index := -1
	for i, para := range parent.Children {
		if para.ID == paraID {
			index = i
			break
		}
	}

	para := NewParagraph(Content(content))
	// Insert at index position + 1
	parent.Children = append(parent.Children[:index+1], append([]*Paragraph{para}, parent.Children[index+1:]...)...)
	d.paragraphs++
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

// UnIndent changes the parent of the paragraph with the given ID to be its current parent's parent.
// It returns the unindented paragraph and its new parent (which may be nil if it becomes a top-level paragraph), or an error if the operation is not possible.
func (d *Document) UnIndent(paraID ParagraphID) (paragraph *Paragraph, newParent *Paragraph, err error) {
	for _, para := range d.Body {
		if para.ID == paraID {
			return nil, nil, fmt.Errorf("cannot unindent top-level paragraph")
		}
	}

	currentParent := d.GetParentOf(paraID)
	if currentParent == nil {
		return nil, nil, fmt.Errorf("paragraph with ID %v not found", paraID)
	}

	grandParent := d.GetParentOf(currentParent.ID)
	var siblingsParents []*Paragraph
	if grandParent == nil {
		siblingsParents = d.Body
	} else {
		siblingsParents = grandParent.Children
	}

	currentParentIndex := -1
	for i, para := range siblingsParents {
		if para.ID == currentParent.ID {
			currentParentIndex = i
			break
		}
	}

	if currentParentIndex == -1 {
		return nil, nil, fmt.Errorf("current parent with ID %v not found among its siblings", currentParent.ID)
	}

	paragraph, err = currentParent.RemoveChild(paraID)
	if err != nil {
		return nil, nil, err
	}

	if grandParent == nil {
		d.Body = append(d.Body[:currentParentIndex+1], append([]*Paragraph{paragraph}, d.Body[currentParentIndex+1:]...)...)
		return paragraph, nil, nil
	}

	grandParent.Children = append(grandParent.Children[:currentParentIndex+1], append([]*Paragraph{paragraph}, grandParent.Children[currentParentIndex+1:]...)...)
	return paragraph, grandParent, nil
}
