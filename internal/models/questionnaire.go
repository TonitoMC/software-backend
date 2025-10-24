package models

import "time"

// models/questionnaire.go
type Questionnaire struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
	Active  bool   `json:"active"`
}

type Question struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Bilateral bool   `json:"bilateral"`
}

type QuestionWithOrder struct {
	Question
	Order int `json:"order"`
}

type QuestionnaireWithQuestions struct {
	Questionnaire
	Questions []QuestionWithOrder `json:"questions"`
}

// Add this to your models package
type QuestionnaireUpdate struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Used for incoming frontend JSON
type ConsultationFromQuestionnaireRequest struct {
	PatientID       int `json:"patientId"`
	QuestionnaireID int `json:"questionnaireId"`

	Reason  string                         `json:"reason"` // ✅ Add this
	Date    time.Time                      `json:"date"`
	Answers map[string]QuestionnaireAnswer `json:"answers"`
}

type QuestionnaireAnswer struct {
	OD      interface{} `json:"od,omitempty"`
	OI      interface{} `json:"oi,omitempty"`
	Value   interface{} `json:"value,omitempty"`
	Comment *string     `json:"comment,omitempty"`
}

// QuestionnaireFullUpdate represents the full payload for PUT updates.
type QuestionnaireFullUpdate struct {
	Name      string              `json:"name"`
	Version   string              `json:"version"`
	Active    bool                `json:"active"`
	Questions []QuestionWithOrder `json:"questions"`
}
