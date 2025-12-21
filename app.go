package main

import (
	"context"
	"errors"
	"fmt"
	"glog/db"
	"glog/domain"
	"log"
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

func (a *App) LoadJournal(date time.Time) (*DocumentDto, error) {
	t := date.Truncate(24 * time.Hour)
	log.Printf("Loading journal for date: %v\n", t)

	doc, err := a.store.GetDocumentForToday(t)
	if err != nil {
		if !errors.Is(err, db.ErrDocumentNotFound) {
			return nil, err
		}
		log.Printf("Error retrieving document: %v\n", err)
		return nil, err
	}

	var docDto *DocumentDto
	if doc == nil {
		log.Printf("No journal found for date: %v, creating a new one\n", t)
		docDto = &DocumentDto{
			ID:    uuid.NewString(),
			Title: fmt.Sprintf("%s, %s", t.Format("Monday"), t.Format("02/01/2006")),
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
		log.Printf("Journal found for date: %v, loading existing journal\n", t)
		docDto, err = FromDomain(doc)
		if err != nil {
			return nil, err
		}
	}

	return docDto, nil
}

func (a *App) LoadJournalsFromTo(from time.Time, to time.Time) ([]*DocumentDto, error) {
	fmt.Printf("Loading journals from %v to %v\n", from, to)
	fromTruncate := from.Truncate(24 * time.Hour)
	toTruncate := to.Truncate(24 * time.Hour)
	docs := make([]*DocumentDto, 0)
	current := fromTruncate
	for !current.Before(toTruncate) {
		doc, err := a.LoadJournal(current)
		if err != nil {
			return nil, err
		}
		current = current.Add(-24 * time.Hour)
		docs = append(docs, doc)
	}
	return docs, nil
}

func (a *App) SaveDocument(d *DocumentDto) error {
	log.Printf("Saving document ID: %s, Title: %s\n", d.ID, d.Title)
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

func (a *App) LoadTodayDocument() (*DocumentDto, error) {
	return a.LoadJournal(time.Now())
}
