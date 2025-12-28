package main

import (
	"context"
	"glog/db"
	"glog/services"

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

func (a *App) SaveDocument(doc DocumentDto) error {
	return nil
}

func (a *App) LoadJournalToday() (DocumentDto, error) {
	return DocumentDto{
		Id:    uuid.NewString(),
		Title: "Today's Journal",
		Blocks: []BlockDto{
			{
				Id:      uuid.NewString(),
				Content: "This is a sample journal entry.",
				Indent:  0,
			},
		},
	}, nil
}
