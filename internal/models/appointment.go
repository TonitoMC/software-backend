package models

import "time"

// Represents an appointment
type Appointment struct {
	ID        int           `json:"id"`
	PatientID int           `json:"patient_id"`
	Name      string        `json:"name"`
	Start     time.Time     `json:"start"`
	Duration  time.Duration `json:"duration"`
}
