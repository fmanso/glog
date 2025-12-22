package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type Content string

type ParagraphID uuid.UUID

func (id ParagraphID) String() string {
	return uuid.UUID(id).String()
}

type Paragraph struct {
	ID       ParagraphID
	Children []*Paragraph
	Content  Content
}

func NewParagraph(content Content) *Paragraph {
	return &Paragraph{
		ID:       ParagraphID(uuid.New()),
		Content:  content,
		Children: []*Paragraph{},
	}
}

func (p *Paragraph) AddChild(child *Paragraph) {
	p.Children = append(p.Children, child)
}

func (p *Paragraph) RemoveChild(paraID ParagraphID) (*Paragraph, error) {
	for i, c := range p.Children {
		if c.ID == paraID {
			p.Children = append(p.Children[:i], p.Children[i+1:]...)
			return c, nil
		}
	}

	return nil, fmt.Errorf("paragraph does not contain paragraph with ID %v", paraID)
}
