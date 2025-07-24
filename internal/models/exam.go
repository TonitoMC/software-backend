package models

import (
	"database/sql"
	"time"
)

type Exam struct {
	ID         int            `json:"id" db:"id"`
	PatientID  int            `json:"patient_id" db:"paciente_id"`
	ConsultaID int            `json:"consulta_id" db:"consulta_id"`
	Type       string         `json:"type" db:"tipo"`
	Date       time.Time      `json:"date" db:"fecha"`
	S3Key      sql.NullString `json:"s3_key,omitempty" db:"s3_key"`
	FileSize   sql.NullInt64  `json:"file_size,omitempty" db:"file_size"`
	MimeType   sql.NullString `json:"mime_type,omitempty" db:"mime_type"`
	HasFile    bool           `json:"has_file"`
}

// Helper method to check if exam has a file
func (e *Exam) SetHasFile() {
	e.HasFile = e.S3Key.Valid && e.S3Key.String != ""
}
