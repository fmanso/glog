package main

import (
	"fmt"
	"glog/db"
	"glog/domain"
	"time"

	"github.com/google/uuid"
)

type BlockDto struct {
	Id      string `json:"id"`
	Content string `json:"content"`
	Indent  int    `json:"indent"`
}

type DocumentDto struct {
	Id        string     `json:"id"`
	Title     string     `json:"title"`
	Blocks    []BlockDto `json:"blocks"`
	Date      string     `json:"date"`       // RFC 3339 format
	IsJournal bool       `json:"is_journal"` // Indicates if this document is a journal entry
}

func ToDocumentDto(doc *domain.Document) DocumentDto {
	blocks := make([]BlockDto, len(doc.Blocks))
	for i, b := range doc.Blocks {
		blocks[i] = BlockDto{
			Id:      b.ID.String(),
			Content: b.Content,
			Indent:  b.Indent,
		}
	}

	return DocumentDto{
		Id:        doc.ID.String(),
		Title:     doc.Title,
		Date:      doc.Date.Format(time.RFC3339),
		IsJournal: doc.IsJournal,
		Blocks:    blocks,
	}
}

func (d DocumentDto) ToDomain() (*domain.Document, error) {
	docID, err := uuid.Parse(d.Id)
	if err != nil {
		return nil, fmt.Errorf("error parsing document id: %s", err)
	}

	t, err := time.Parse(time.RFC3339, d.Date)
	if err != nil {
		return nil, fmt.Errorf("error parsing date: %s", err)
	}

	doc := &domain.Document{
		ID:        domain.DocumentID(docID),
		Title:     d.Title,
		Date:      t,
		IsJournal: d.IsJournal,
		Blocks:    make([]*domain.Block, len(d.Blocks)),
	}

	for i, b := range d.Blocks {
		blockId, err := uuid.Parse(d.Blocks[i].Id)
		if err != nil {
			return nil, fmt.Errorf("error parsing block id: %s", err)
		}

		doc.Blocks[i] = &domain.Block{
			ID:      domain.BlockID(blockId),
			Content: b.Content,
			Indent:  b.Indent,
		}
	}

	return doc, nil
}

type DocumentSummaryDto struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Date  string `json:"date"` // RFC 3339 format
}

func ToDocumentSummaryDto(summary db.DocumentSummary) DocumentSummaryDto {
	return DocumentSummaryDto{
		Id:    summary.ID.String(),
		Title: summary.Title,
		Date:  summary.Date.Format(time.RFC3339),
	}
}

type ScheduledTaskDto struct {
	Id          string `json:"id"`
	Description string `json:"description"`
	DueDate     string `json:"due_date"` // RFC 3339 format
	Title       string `json:"title"`
	BlockId     string `json:"block_id"`
	DocId       string `json:"doc_id"`
}
