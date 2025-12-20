package domain

import (
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

func (p *Paragraph) RemoveChild(child *Paragraph) {
	for i, c := range p.Children {
		if c.ID == child.ID {
			p.Children = append(p.Children[:i], p.Children[i+1:]...)
			return
		}
	}
}
