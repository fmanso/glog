package db

import "github.com/google/uuid"

type DocDb struct {
	ID        uuid.UUID
	Title     string
	Date      string
	IsJournal bool
	Blocks    []*BlockDb
}

type BlockDb struct {
	ID      uuid.UUID
	Content string
	Indent  int
}

type ScheduleTaskDb struct {
	ID        uuid.UUID
	DocDbID   uuid.UUID
	BlockDbID uuid.UUID
}
