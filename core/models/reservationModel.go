package models

import (
	"github.com/google/uuid"
	"time"
)

type JSONReservationModel struct {
	ID         uuid.UUID `json:"id"`
	ReaderID   uuid.UUID `json:"reader_id"`
	BookID     uuid.UUID `json:"book_id"`
	IssueDate  time.Time `json:"issue_date"`
	ReturnDate time.Time `json:"return_date"`
	State      string    `json:"state"`
}
