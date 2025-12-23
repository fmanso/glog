package main

import (
	"glog/domain"
	"log"
)

type ParagraphDto struct {
	ID          string `json:"id"`
	Content     string `json:"content"`
	Indentation int    `json:"indentation"`
}

type DocumentDto struct {
	ID    string         `json:"id"`
	Title string         `json:"title"`
	Date  string         `json:"date"`
	Body  []ParagraphDto `json:"body"`
}

func FromDomainParagraph(para *domain.Paragraph) ParagraphDto {
	log.Println("Converting paragraph ID:", para.ID.String())
	paraDto := ParagraphDto{
		ID:          para.ID.String(),
		Content:     string(para.Content),
		Indentation: para.Indentation,
	}
	return paraDto
}

func FromDomain(doc *domain.Document) (*DocumentDto, error) {
	docDto := &DocumentDto{
		ID:    doc.ID.String(),
		Title: doc.Title,
		Date:  string(doc.Date),
		Body:  make([]ParagraphDto, len(doc.Body)),
	}

	for i, para := range doc.Body {
		paraDto := FromDomainParagraph(para)
		docDto.Body[i] = paraDto
	}

	return docDto, nil
}
