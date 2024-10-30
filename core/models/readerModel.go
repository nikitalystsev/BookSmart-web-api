package models

import "github.com/google/uuid"

type JSONReaderModel struct {
	ID          uuid.UUID `json:"id"`
	Fio         string    `json:"fio"`
	PhoneNumber string    `json:"phone_number"`
	Age         uint      `json:"age"`
	Password    string    `json:"password"`
	Role        string    `json:"role"`
}
