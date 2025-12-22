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

func (s *DocumentService) AddNewParagraph(docID domain.DocumentID, index int) (*domain.Paragraph, error) {
	doc, err := s.store.LoadDocument(docID)
	if err != nil {
		return nil, err
	}

	para := doc.InsertParagraph(index, "")

	err = s.store.Save(doc)
	if err != nil {
		return nil, err
	}

	return para, nil
}

func (s *DocumentService) Indent(docId domain.DocumentID, paraID domain.ParagraphID) (*domain.Paragraph, error) {
	doc, err := s.store.LoadDocument(docId)
	if err != nil {
		return nil, err
	}

	para, parent, err := doc.Indent(paraID)
	if err != nil {
		return nil, err
	}

	if parent == nil {
		return nil, fmt.Errorf("cannot indent paragraph with ID %v", paraID)
	}

	err = s.store.ChangeParentID(paraID, &parent.ID)
	if err != nil {
		return nil, err
	}

	return para, nil
}
