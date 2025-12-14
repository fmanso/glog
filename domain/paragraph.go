package domain

import (
	"time"

	"github.com/google/uuid"
)

type Content string

type Paragraph struct {
	ID         uuid.UUID
	Parent     *uuid.UUID
	Children   []*Content
	Content    Content
	References []uuid.UUID
}

type Document struct {
	ID   uuid.UUID
	Date time.Time
	Body []Paragraph
}
