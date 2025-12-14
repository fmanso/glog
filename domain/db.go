package domain

import (
	"time"

	"github.com/google/uuid"
)

type DocumentStore struct {
}

func NewDocumentStore() *DocumentStore {
	return &DocumentStore{}
}

func (store *DocumentStore) Save(paragraphs []Paragraph) error {
	return nil
}

func (store *DocumentStore) Load(from, to time.Time) ([]Document, error) {
	return nil, nil
}

func (store *DocumentStore) Search(terms []string) ([]Document, error) {
	return nil, nil
}

func (store *DocumentStore) GetReferences(documentID uuid.UUID) ([]Document, error) {
	return nil, nil
}
