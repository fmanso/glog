package main

import (
	"context"
	"errors"
	"glog/db"
	"glog/domain"
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
	// Get time of today but for 6:00 AM UTC
	t := time.Now().UTC()
	t = time.Date(t.Year(), t.Month(), t.Day(), 6, 0, 0, 0, time.UTC)

	doc, err := a.db.LoadDocumentByTime(t)
	if err == nil {
		return ToDocumentDto(doc), nil
	}

	if errors.Is(err, db.ErrDocumentNotFound) {
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

	return DocumentDto{}, err
}

func (a *App) createNewDocument(title string) (DocumentDto, error) {
	doc := domain.Document{
		ID:    domain.DocumentID(uuid.New()),
		Title: title,
		Date:  time.Now().UTC(),
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
