package main

import (
	"glog/domain"
	"time"

	"github.com/google/uuid"
)

type ParagraphDto struct {
	ID       string         `json:"id"`
	Content  string         `json:"content"`
	Children []ParagraphDto `json:"children"`
}

type DocumentDto struct {
	ID    string         `json:"id"`
	Title string         `json:"title"`
	Date  string         `json:"date"`
	Body  []ParagraphDto `json:"body"`
}

func toDomainParagraph(paraDto ParagraphDto) (*domain.Paragraph, error) {
	id, err := uuid.Parse(paraDto.ID)
	if err != nil {
		return nil, err
	}

	para := &domain.Paragraph{
		ID:       domain.ParagraphID(id),
		Content:  domain.Content(paraDto.Content),
		Children: make([]*domain.Paragraph, 0),
	}

	for _, childDto := range paraDto.Children {
		childPara, err := toDomainParagraph(childDto)
		if err != nil {
			return nil, err
		}
		para.Children = append(para.Children, childPara)
	}

	return para, nil
}

func ToDomain(docDto *DocumentDto) (*domain.Document, error) {
	// Parse RFC3339 date
	t, err := time.Parse(time.RFC3339, docDto.Date)

	if err != nil {
		return nil, err
	}

	id, err := uuid.Parse(docDto.ID)
	if err != nil {
		return nil, err
	}

	doc := domain.Document{
		ID:    domain.DocumentID(id),
		Title: docDto.Title,
		Date:  domain.ToDateTime(t),
		Body:  make([]*domain.Paragraph, 0),
	}

	for _, paraDto := range docDto.Body {
		para, err := toDomainParagraph(paraDto)
		if err != nil {
			return nil, err
		}
		doc.Body = append(doc.Body, para)
	}

	return &doc, nil
}

func fromDomainParagraph(para *domain.Paragraph) ParagraphDto {
	children := make([]ParagraphDto, 0)
	for _, child := range para.Children {
		childDto := fromDomainParagraph(child)
		children = append(children, childDto)
	}

	return ParagraphDto{
		ID:       para.ID.String(),
		Content:  string(para.Content),
		Children: children,
	}
}

func FromDomain(doc *domain.Document) (*DocumentDto, error) {
	docDto := &DocumentDto{
		ID:    doc.ID.String(),
		Title: doc.Title,
		Date:  string(doc.Date),
		Body:  make([]ParagraphDto, 0),
	}

	for _, para := range doc.Body {
		paraDto := fromDomainParagraph(para)
		docDto.Body = append(docDto.Body, paraDto)
	}

	return docDto, nil
}
