package models

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
	Type      string `json:"type"` // 'entero', 'float', 'bool', 'texto'
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
