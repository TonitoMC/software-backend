package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type Exam struct {
	ID         int            `json:"id" db:"id"`
	PatientID  int            `json:"patient_id" db:"paciente_id"`
	ConsultaID sql.NullInt64  `json:"consulta_id" db:"consulta_id"`
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

// examJSON is a helper struct for custom JSON marshaling
type examJSON struct {
	ID         int       `json:"id"`
	PatientID  int       `json:"patient_id"`
	ConsultaID *int64    `json:"consulta_id,omitempty"` // Use pointer for null, omitempty for not present
	Type       string    `json:"type"`
	Date       time.Time `json:"date"`
	S3Key      *string   `json:"s3_key,omitempty"`    // Use pointer for null
	FileSize   *int64    `json:"file_size,omitempty"` // Use pointer for null
	MimeType   *string   `json:"mime_type,omitempty"` // Use pointer for null
	HasFile    bool      `json:"has_file"`
}

// MarshalJSON implements the json.Marshaler interface for Exam
func (e *Exam) MarshalJSON() ([]byte, error) {
	// Call SetHasFile before marshaling to ensure HasFile is correct
	e.SetHasFile()

	ej := examJSON{
		ID:        e.ID,
		PatientID: e.PatientID,
		Type:      e.Type,
		Date:      e.Date,
		HasFile:   e.HasFile,
	}

	if e.ConsultaID.Valid {
		ej.ConsultaID = &e.ConsultaID.Int64
	}
	if e.S3Key.Valid {
		ej.S3Key = &e.S3Key.String
	}
	if e.FileSize.Valid {
		ej.FileSize = &e.FileSize.Int64
	}
	if e.MimeType.Valid {
		ej.MimeType = &e.MimeType.String
	}

	return json.Marshal(ej)
}
