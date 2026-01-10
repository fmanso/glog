package main

import (
	"context"
	"errors"
	"glog/db"
	"glog/domain"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// App struct
type App struct {
	ctx context.Context
	db  *db.DocumentStore
}

// NewApp creates a new App application struct
func NewApp(db *db.DocumentStore) *App {
	return &App{
		db: db,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SaveDocument(doc DocumentDto) error {
	domainDoc, err := doc.ToDomain()
	if err != nil {
		return err
	}

	err = a.db.Save(domainDoc)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) LoadJournalToday() (DocumentDto, error) {
	// Get current local time and normalize to start of day
	now := time.Now()
	t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	from := t
	to := t.Add(24 * time.Hour)
	docs, err := a.db.LoadJournals(from, to)
	if err == nil && docs != nil && len(docs) > 0 {
		return ToDocumentDto(docs[0]), nil
	}

	// Create title based on date (e.g., "Monday, January 2, 2006")
	title := t.Format("Monday, January 2, 2006")
	return DocumentDto{
		Id:    uuid.NewString(),
		Title: title,
		Date:  t.Format(time.RFC3339),
		Blocks: []BlockDto{
			{
				Id:      uuid.NewString(),
				Content: "",
				Indent:  0,
			},
		},
	}, nil
}

func (a *App) LoadJournals(from string, to string) ([]DocumentDto, error) {
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return nil, err
	}

	toTime, err := time.Parse(time.RFC3339, to)
	if err != nil {
		return nil, err
	}

	docs, err := a.db.LoadJournals(fromTime, toTime)
	if err != nil {
		return nil, err
	}

	docDtos := make([]DocumentDto, len(docs))
	for i, doc := range docs {
		docDtos[i] = ToDocumentDto(doc)
	}

	return docDtos, nil
}

func (a *App) createNewDocument(title string) (DocumentDto, error) {
	doc := domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: title,
		Date:  time.Now(), // Use local time
		Blocks: []*domain.Block{
			{
				ID:      domain.BlockID(uuid.New()),
				Content: "",
				Indent:  0,
			},
		},
	}

	err := a.db.Save(&doc)
	if err != nil {
		return DocumentDto{}, err
	}

	return ToDocumentDto(&doc), nil
}

func (a *App) GetDocumentList() ([]DocumentSummaryDto, error) {
	docs, err := a.db.ListDocuments()
	if err != nil {
		return nil, err
	}

	summaries := make([]DocumentSummaryDto, len(docs))
	for i, doc := range docs {
		summaries[i] = DocumentSummaryDto{
			Id:    doc.ID.String(),
			Title: doc.Title,
			Date:  doc.Date.Format(time.RFC3339),
		}
	}

	return summaries, nil
}

func (a *App) OpenDocument(docId string) (DocumentDto, error) {
	id, err := uuid.Parse(docId)
	if err != nil {
		return DocumentDto{}, err
	}

	domainDoc, err := a.db.LoadDocument(domain.DocumentID(id))
	if err != nil {
		return DocumentDto{}, err
	}

	doc := ToDocumentDto(domainDoc)
	return doc, nil
}

func (a *App) OpenDocumentByTitle(title string) (DocumentDto, error) {
	domainDoc, err := a.db.LoadDocumentByTitle(title)
	if err != nil {
		if errors.Is(err, db.ErrDocumentNotFound) {
			return a.createNewDocument(title)
		}
		return DocumentDto{}, err
	}

	doc := ToDocumentDto(domainDoc)
	return doc, nil
}

func (a *App) SearchDocuments(search string) ([]DocumentSummaryDto, error) {
	docIDs, err := a.db.Search(search)
	if err != nil {
		return nil, err
	}

	summaries := make([]DocumentSummaryDto, 0, len(docIDs))
	for _, id := range docIDs {
		domainDoc, err := a.db.LoadDocument(domain.DocumentID(id))
		if err != nil {
			return nil, err
		}

		summaries = append(summaries, DocumentSummaryDto{
			Id:    domainDoc.ID.String(),
			Title: domainDoc.Title,
			Date:  domainDoc.Date.Format(time.RFC3339),
		})
	}

	return summaries, nil
}

type DocumentReferenceDto struct {
	Id     string
	Title  string
	Blocks []BlockReferenceDto
}

type BlockReferenceDto struct {
	Id      string
	Content string
	Indent  int
}

func (a *App) GetReferences(title string) ([]DocumentReferenceDto, error) {
	titleLower := strings.ToLower(title)
	docIDs, err := a.db.GetReferences(title)
	if err != nil {
		return nil, err
	}

	log.Println("Found references for title", title, ":", docIDs)

	var references []DocumentReferenceDto
	for _, id := range docIDs {
		domainDoc, err := a.db.LoadDocument(id)
		if err != nil {
			return nil, err
		}

		var blocks []BlockReferenceDto
		for i := 0; i < len(domainDoc.Blocks); i++ {
			block := domainDoc.Blocks[i]
			if strings.Contains(strings.ToLower(block.Content), titleLower) {
				blocks = append(blocks, BlockReferenceDto{
					Id:      block.ID.String(),
					Content: block.Content,
					Indent:  block.Indent,
				})
			}
		}

		references = append(references, DocumentReferenceDto{
			Id:     domainDoc.ID.String(),
			Title:  domainDoc.Title,
			Blocks: blocks,
		})
	}

	log.Println("Found references for title", title, ":", references)
	return references, nil
}

func (a *App) GetScheduledTasks() ([]ScheduledTaskDto, error) {
	scheduleTasks, err := a.db.GetScheduledTasks(time.Now(), 5)
	if err != nil {
		return nil, err
	}

	var scheduledTaskDtos []ScheduledTaskDto
	for _, task := range scheduleTasks {
		doc, err := a.db.LoadDocument(task.DocID)
		if err != nil {
			return nil, err
		}

		// Get the block content
		var blockContent string
		for _, block := range doc.Blocks {
			if block.ID == task.BlockID {
				blockContent = block.Content
				break
			}
		}

		scheduledTaskDtos = append(scheduledTaskDtos, ScheduledTaskDto{
			Id:          task.ID.String(),
			Title:       doc.Title,
			DocId:       task.DocID.String(),
			BlockId:     task.BlockID.String(),
			Description: blockContent,
			DueDate:     task.Time.Format(time.RFC3339),
		})
	}

	return scheduledTaskDtos, nil
}

// IndexHealthDto represents the health status of the search index for the UI
type IndexHealthDto struct {
	IsHealthy          bool   `json:"isHealthy"`
	FailedDocuments    int    `json:"failedDocuments"`
	LastHealthCheck    string `json:"lastHealthCheck"`
	RequiresReindex    bool   `json:"requiresReindex"`
	HealthCheckMessage string `json:"healthCheckMessage"`
}

// GetIndexHealth returns the current health status of the search index
func (a *App) GetIndexHealth() IndexHealthDto {
	health := a.db.GetIndexHealth()
	return IndexHealthDto{
		IsHealthy:          health.IsHealthy,
		FailedDocuments:    health.FailedDocuments,
		LastHealthCheck:    health.LastHealthCheck.Format(time.RFC3339),
		RequiresReindex:    health.RequiresReindex,
		HealthCheckMessage: health.HealthCheckMessage,
	}
}

// ReindexSearch rebuilds the entire search index from scratch
func (a *App) ReindexSearch() error {
	return a.db.ReindexSearch()
}

// RetryFailedIndexing attempts to reindex documents that previously failed
func (a *App) RetryFailedIndexing() (int, error) {
	return a.db.RetryFailedIndexing()
}
