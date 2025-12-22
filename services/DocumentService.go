package services

import (
	"fmt"
	"glog/db"
	"glog/domain"
	"time"
)

type DocumentService struct {
	store *db.DocumentStore
}

func NewDocumentService(store *db.DocumentStore) *DocumentService {
	return &DocumentService{
		store: store,
	}
}

func (s *DocumentService) CreateSampleDocumentForToday() (*domain.Document, error) {
	t := time.Now().UTC().Truncate(24 * time.Hour)
	doc := domain.NewDocument(
		fmt.Sprintf("%s, %s", t.Format("Monday"), t.Format("02/01/2006")),
		domain.ToDateTime(time.Now().UTC().Truncate(24*time.Hour)),
	)
	doc.InsertParagraph(0, "Start your journal entry here...")

	err := s.store.Save(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) SetParagraphContent(paraID domain.ParagraphID, content string) (string, error) {
	err := s.store.SetParagraphContent(paraID, domain.Content(content))
	if err != nil {
		return "", err
	}

	return content, nil
}
