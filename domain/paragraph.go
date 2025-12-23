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
	ID          ParagraphID
	Content     Content
	Indentation int
}

func NewParagraph(content Content, indentation int) *Paragraph {
	return &Paragraph{
		ID:          ParagraphID(uuid.New()),
		Content:     content,
		Indentation: indentation,
	}
}
