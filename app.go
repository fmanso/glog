package main

import (
	"context"
	"glog/db"
	"glog/services"
	"time"

	"github.com/google/uuid"
)

// App struct
type App struct {
	ctx        context.Context
	store      *db.DocumentStore
	docService *services.DocumentService
}

// NewApp creates a new App application struct
func NewApp(db *db.DocumentStore, docService *services.DocumentService) *App {
	return &App{
		store:      db,
		docService: docService,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

var memory = make(map[uuid.UUID]DocumentDto)

func (a *App) SaveDocument(doc DocumentDto) error {
	// Parse UUID
	id, err := uuid.Parse(doc.Id)
	if err != nil {
		return err
	}

	memory[id] = doc
	return nil
}

func (a *App) LoadJournalToday() (DocumentDto, error) {
	// Get time of today but for 6:00 AM UTC
	t := time.Now().UTC()
	t = time.Date(t.Year(), t.Month(), t.Day(), 6, 0, 0, 0, time.UTC)

	// Create title based on date (e.g., "Monday, January 2, 2006")
	title := t.Format("Monday, January 2, 2006")

	return DocumentDto{
		Id:    uuid.NewString(),
		Title: title,
		Date:  time.Now().UTC().Format(time.RFC3339),
		Blocks: []BlockDto{
			{
				Id:      uuid.NewString(),
				Content: "",
				Indent:  0,
			},
		},
	}, nil
}

func (a *App) CreateNewDocument(title string) (DocumentDto, error) {
	doc := DocumentDto{
		Id:    uuid.NewString(),
		Title: title,
		Date:  time.Now().UTC().Format(time.RFC3339),
		Blocks: []BlockDto{
			{
				Id:      uuid.NewString(),
				Content: "",
				Indent:  0,
			},
		},
	}

	memory[uuid.MustParse(doc.Id)] = doc
	return doc, nil
}

func (a *App) GetDocumentList() ([]DocumentSummaryDto, error) {
	summaries := make([]DocumentSummaryDto, len(memory))
	for _, doc := range memory {
		summaries = append(summaries, DocumentSummaryDto{
			Id:    doc.Id,
			Title: doc.Title,
			Date:  doc.Date,
		})
	}
	return summaries, nil
}

func (a *App) OpenDocument(docId string) (DocumentDto, error) {
	id, err := uuid.Parse(docId)
	if err != nil {
		return DocumentDto{}, err
	}

	doc, exists := memory[id]
	if !exists {
		return DocumentDto{}, nil
	}

	return doc, nil
}

func (a *App) SearchDocuments(search string) ([]DocumentSummaryDto, error) {
	// For simplicity, return all documents as search results
	return a.GetDocumentList()
}
