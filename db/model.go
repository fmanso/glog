package db

import "github.com/google/uuid"

type DocDb struct {
	ID     uuid.UUID
	Title  string
	Date   string
	Blocks []*BlockDb
}

type BlockDb struct {
	ID      uuid.UUID
	Content string
	Ident   int
}

type ScheduleTaskDb struct {
	ID        uuid.UUID
	DocDbID   uuid.UUID
	BlockDbID uuid.UUID
}
