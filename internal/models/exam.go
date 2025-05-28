package models

import "time"

// Consultas / Medical checkups for a patient
type Exam struct {
	Type string    `json:"type"`
	Date time.Time `json:"date"`
}
