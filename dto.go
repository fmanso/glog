package main

import (
	"glog/domain"
)

type ParagraphDto struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	Indentation int    `json:"indentation,omitempty"`
}

type DocumentDto struct {
	ID    string         `json:"id"`
	Title string         `json:"title"`
	Date  string         `json:"date"`
	Body  []ParagraphDto `json:"body"`
}

func FromDomainParagraph(para *domain.Paragraph, indentation int) []ParagraphDto {
	paragraphs := make([]ParagraphDto, 0)
	paraDto := ParagraphDto{
		ID:          para.ID.String(),
		Content:     string(para.Content),
		Indentation: indentation,
	}
	paragraphs = append(paragraphs, paraDto)

	for _, child := range para.Children {
		childParagraphs := FromDomainParagraph(child, indentation+1)
		paragraphs = append(paragraphs, childParagraphs...)
	}

	return paragraphs
}

func FromDomain(doc *domain.Document) (*DocumentDto, error) {
	docDto := &DocumentDto{
		ID:    doc.ID.String(),
		Title: doc.Title,
		Date:  string(doc.Date),
		Body:  make([]ParagraphDto, len(doc.Body)),
	}

	for _, para := range doc.Body {
		paraDto := FromDomainParagraph(para, 0)
		docDto.Body = append(docDto.Body, paraDto...)
	}

	return docDto, nil
}
