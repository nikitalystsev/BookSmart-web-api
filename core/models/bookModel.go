package models

import "github.com/google/uuid"

type JSONBookModel struct {
	ID             uuid.UUID `json:"id"`
	Title          string    `json:"title"`
	Author         string    `json:"author"`
	Publisher      string    `json:"publisher"`
	CopiesNumber   uint      `json:"copies_number"`
	Rarity         string    `json:"rarity"`
	Genre          string    `json:"genre"`
	PublishingYear uint      `json:"publishing_year"`
	Language       string    `json:"language"`
	AgeLimit       uint      `json:"age_limit"`
}
