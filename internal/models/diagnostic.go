package models

type Diagnostic struct {
	ID             int         `json:"id"`
	Name           string      `json:"name"`
	Recommendation string      `json:"recommendation"`
	ConsultationID int         `json:"consultation_id"`
	Treatments     []Treatment `json:"treatments"`
}

type Treatment struct {
	ID              int    `json:"id"`
	DiagnosticID    int    `json:"diagnostic_id"`
	ActiveComponent string `json:"active_component"`
	Presentation    string `json:"presentation"`
	Dosage          string `json:"dosage"`
	Frequency       string `json:"frequency"`
	Duration        string `json:"duration"`
}
