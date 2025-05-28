package models

import "time"

// Consultas / Medical checkups for a patient
type Consultation struct {
	Date   time.Time `json:"date"`
	Reason string    `json:"motive"`
}
