package models

import "time"

type Consultation struct {
	ID              int       `json:"id"`
	PatientID       int       `json:"patient_id"`
	QuestionnaireID *int      `json:"questionnaire_id,omitempty"`
	Reason          string    `json:"reason"`
	Date            time.Time `json:"date"`
}

type ConsultationQuestion struct {
	ID             int      `json:"id"`
	ConsultationID int      `json:"consultation_id"`
	QuestionID     int      `json:"question_id"`
	QuestionName   string   `json:"question_name,omitempty"`
	QuestionType   string   `json:"question_type,omitempty"`
	Bilateral      bool     `json:"bilateral,omitempty"`
	TextValues     []string `json:"text_values,omitempty"`
	IntValues      []int    `json:"int_values,omitempty"`
	BoolValues     []bool   `json:"bool_values,omitempty"`
	TextValue      *string  `json:"text_value,omitempty"`
	IntValue       *int     `json:"int_value,omitempty"`
	BoolValue      *bool    `json:"bool_value,omitempty"`
	Comment        *string  `json:"comment,omitempty"`
}

type CompleteConsultation struct {
	Consultation
	Questions []ConsultationQuestion `json:"questions,omitempty"`
	Diagnoses []Diagnostic           `json:"diagnoses,omitempty"`
}

type ConsultationWithDetails struct {
	Consultation
	Diagnoses []Diagnostic `json:"diagnoses,omitempty"`
}
