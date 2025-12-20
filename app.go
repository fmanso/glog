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
	ctx   context.Context
	store *db.DocumentStore
}

// NewApp creates a new App application struct
func NewApp(db *db.DocumentStore) *App {
	return &App{
		store: db,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

var docDto *DocumentDto

func (a *App) LoadTodayDocument() (*DocumentDto, error) {
	doc, err := a.store.GetDocumentForToday(time.Now().Truncate(24 * time.Hour))
	if err != nil {
		if !errors.Is(err, db.ErrDocumentNotFound) {
			return nil, err
		}
	}

	if doc == nil {
		docDto = &DocumentDto{
			ID:    uuid.NewString(),
			Title: "Sample Document",
			Date:  string(domain.ToDateTime(time.Now().Truncate(24 * time.Hour))),
			Body: []ParagraphDto{
				{
					ID:      uuid.NewString(),
					Content: "This is the first paragraph.",
					Children: []ParagraphDto{
						{
							ID:      uuid.NewString(),
							Content: "This is a child paragraph.",
						},
					},
				},
				{
					ID:      uuid.NewString(),
					Content: "This is the second paragraph.",
				},
			},
		}
	} else {
		docDto, err = FromDomain(doc)
		if err != nil {
			return nil, err
		}
	}

	return docDto, nil
}

func (a *App) SaveDocument(d *DocumentDto) error {
	doc, err := ToDomain(d)
	if err != nil {
		return err
	}

	err = a.store.Save(doc)
	if err != nil {
		return err
	}

	return nil
}
