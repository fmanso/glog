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
	err := doc.InsertParagraphAt(0, "Start your journal entry here...", 0)
	if err != nil {
		return nil, err
	}

	err = s.store.Save(doc)
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

func (s *DocumentService) InsertNewParagraphAt(docID domain.DocumentID, index int) (*domain.Document, error) {
	doc, err := s.store.LoadDocument(docID)
	if err != nil {
		return nil, err
	}

	_ = doc.InsertParagraphAt(index, "", 0)

	err = s.store.Save(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) Indent(docID domain.DocumentID, index int) (*domain.Document, error) {
	doc, err := s.store.LoadDocument(docID)
	if err != nil {
		return nil, err
	}

	err = doc.Indent(index)
	if err != nil {
		return nil, err
	}

	err = s.store.Save(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}

func (s *DocumentService) Outdent(docID domain.DocumentID, index int) (*domain.Document, error) {
	doc, err := s.store.LoadDocument(docID)
	if err != nil {
		return nil, err
	}

	err = doc.Outdent(index)
	if err != nil {
		return nil, err
	}

	err = s.store.Save(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
