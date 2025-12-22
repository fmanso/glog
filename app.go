package main

import (
	"context"
	"fmt"
	"glog/db"
	"glog/domain"
	"glog/services"
	"log"
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

func (a *App) LoadJournal(date time.Time) (*DocumentDto, error) {
	t := date.Truncate(24 * time.Hour)
	log.Printf("Loading journal for date: %v\n", t)

	doc, err := a.store.GetDocumentFor(t)
	if err != nil {
		// Ignore
	}

	var docDto *DocumentDto
	if doc == nil {
		log.Printf("No journal found for date: %v, creating a new one\n", t)
		doc, err := a.docService.CreateSampleDocumentForToday()
		if err != nil {
			return nil, err
		}

		docDto, err = FromDomain(doc)
		if err != nil {
			return nil, err
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

func (a *App) SetParagraphContent(paraID string, content string) (string, error) {
	log.Printf("Setting content for paragraph ID: %s\n", paraID)
	uuidParaID, err := uuid.Parse(paraID)
	if err != nil {
		return "", err
	}
	domainParaID := domain.ParagraphID(uuidParaID)

	content, err = a.docService.SetParagraphContent(domainParaID, content)
	if err != nil {
		return "", err
	}

	return content, nil
}

func (a *App) SaveDocument(d *DocumentDto) error {
	//log.Printf("Saving document ID: %s, Title: %s\n", d.ID, d.Title)
	//doc, err := ToDomain(d)
	//if err != nil {
	//	return err
	//}
	//
	//err = a.store.Save(doc)
	//if err != nil {
	//	return err
	//}

	return nil
}

func (a *App) LoadTodayDocument() (*DocumentDto, error) {
	return a.LoadJournal(time.Now().UTC())
}
