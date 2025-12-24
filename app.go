package main

import (
	"context"
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

		log.Printf("Created a new journal entry: %v\n", doc)
		for _, para := range doc.Body {
			log.Printf("Paragraph ID: %s, Content: %s\n", para.ID, para.Content)
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

	log.Printf("Journal loaded: %v\n", docDto)
	for _, para := range docDto.Body {
		log.Printf("Paragraph ID: %s, Content: %s, Indentation: %d\n", para.ID, para.Content, para.Indentation)
	}
	return docDto, nil
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

func (a *App) AddNewParagraph(docID string, content string, index int) (*DocumentDto, error) {
	log.Printf("Adding new paragraph at index %d in document ID: %s\n", index, docID)
	uuidDocID, err := uuid.Parse(docID)
	if err != nil {
		return nil, err
	}
	domainDocID := domain.DocumentID(uuidDocID)

	doc, err := a.docService.InsertNewParagraphAt(domainDocID, content, index)
	if err != nil {
		return nil, err
	}
	docDto, err := FromDomain(doc)
	if err != nil {
		return nil, err
	}

	return docDto, nil
}

func (a *App) Indent(docID string, index int) (*DocumentDto, error) {
	log.Printf("Indenting paragraph at index %d in document ID: %s\n", index, docID)
	uuidDocID, err := uuid.Parse(docID)
	if err != nil {
		return nil, err
	}
	domainDocID := domain.DocumentID(uuidDocID)

	domainDoc, err := a.docService.Indent(domainDocID, index)
	if err != nil {
		return nil, err
	}
	docDto, err := FromDomain(domainDoc)
	if err != nil {
		return nil, err
	}

	return docDto, nil
}

func (a *App) Outdent(docID string, index int) (*DocumentDto, error) {
	log.Printf("Unindenting paragraph at index %d in document ID: %s\n", index, docID)
	uuidDocID, err := uuid.Parse(docID)
	if err != nil {
		return nil, err
	}
	domainDocID := domain.DocumentID(uuidDocID)

	domainDoc, err := a.docService.Outdent(domainDocID, index)
	if err != nil {
		return nil, err
	}

	docDto, err := FromDomain(domainDoc)
	if err != nil {
		return nil, err
	}

	return docDto, nil
}

func (a *App) DeleteParagraphAt(docID string, index int) (*DocumentDto, error) {
	log.Printf("Deleting paragraph at index %d in document ID: %s\n", index, docID)
	uuidDocID, err := uuid.Parse(docID)
	if err != nil {
		return nil, err
	}
	domainDocID := domain.DocumentID(uuidDocID)

	domainDoc, err := a.docService.DeleteParagraph(domainDocID, index)
	if err != nil {
		return nil, err
	}

	docDto, err := FromDomain(domainDoc)
	if err != nil {
		return nil, err
	}

	return docDto, nil
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

func (a *App) NewDocument(title string) (*DocumentDto, error) {
	log.Printf("Creating new document with title: %s\n", title)
	doc, err := a.docService.NewDocument(title)
	if err != nil {
		return nil, err
	}

	docDto, err := FromDomain(doc)
	if err != nil {
		return nil, err
	}

	return docDto, nil
}

func (a *App) GetReferences(docID string) ([]*DocumentDto, error) {
	log.Printf("Getting references for document ID: %s\n", docID)
	uuidDocID, err := uuid.Parse(docID)
	if err != nil {
		return nil, err
	}
	domainDocID := domain.DocumentID(uuidDocID)

	docs, err := a.docService.GetDocumentsReferencing(domainDocID)
	if err != nil {
		return nil, err
	}

	var docDtos []*DocumentDto
	for _, doc := range docs {
		docDto, err := FromDomain(doc)
		if err != nil {
			return nil, err
		}
		docDtos = append(docDtos, docDto)
	}

	return docDtos, nil
}
