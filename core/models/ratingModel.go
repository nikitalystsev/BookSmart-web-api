package models

import "github.com/google/uuid"

type RatingModel struct {
	ID       uuid.UUID `json:"id"`
	ReaderID uuid.UUID `json:"reader_id"`
	BookID   uuid.UUID `json:"book_id"`
	Review   string    `json:"review"`
	Rating   int       `json:"rating"`
}
