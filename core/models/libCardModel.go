package models

import (
	"github.com/google/uuid"
	"time"
)

type JSONLibCardModel struct {
	ID           uuid.UUID `json:"id"`
	ReaderID     uuid.UUID `json:"reader_id"`
	LibCardNum   string    `json:"lib_card_num"`
	Validity     int       `json:"validity"`
	IssueDate    time.Time `json:"issue_date"`
	ActionStatus bool      `json:"action_status"`
}
