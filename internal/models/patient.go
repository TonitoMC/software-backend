package models

import "time"

// Patient represents a patient in the application's domain, potentially with embedded history.
type Patient struct {
	ID           int           `json:"id"`            // Unique identifier for the patient
	Name         string        `json:"name"`          // Patient's full name
	DateOfBirth  time.Time     `json:"date_of_birth"` // Patient's date of birth
	Phone        string        `json:"phone"`         // Patient's phone number
	Sex          string        `json:"sex"`           // Patient's sex
	Antecedentes *Antecedentes `json:"antecedentes,omitempty"`
}

// Antecedentes (Medical History) for a patient.
// Defined here as it's tightly coupled with the Patient model.
type Antecedentes struct {
	Medical string `json:"medical"`
	Family  string `json:"family"`
	Ocular  string `json:"ocular"`
	Alergic string `json:"alergic"`
	Other   string `json:"other"`
}
