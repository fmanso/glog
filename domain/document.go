package domain

import (
	"time"

	"github.com/google/uuid"
)

type DocumentID uuid.UUID

func (d DocumentID) String() string {
	return uuid.UUID(d).String()
}

type BlockID uuid.UUID

func (b BlockID) String() string {
	return uuid.UUID(b).String()
}

type Document struct {
	ID        DocumentID
	Title     string
	Date      time.Time // RFC 3339 format
	IsJournal bool
	Blocks    []*Block
}

type Block struct {
	ID      BlockID
	Content string
	Indent  int
}

type ScheduleTask struct {
	ID      uuid.UUID
	DocID   DocumentID
	BlockID BlockID
	Time    time.Time
}
