package service

import (
	"fmt"

	"software-backend/internal/models"
	"software-backend/internal/repository"
)

// TODO custom errors for service

// Interface PatientService defines methods expected from the service
type PatientService interface {
	GetPatientByID(patientID int) (*models.Patient, error)
	SearchPatients(query string, limit int) ([]models.Patient, error)
}

// Struct to manage dependencies
type patientService struct {
	patientRepo repository.PatientRepository
}

// Constructor to pass on dependencies
func NewPatientService(patientRepo repository.PatientRepository) PatientService {
	return &patientService{
		patientRepo: patientRepo,
	}
}

// Get a patient by their PatientID
func (s *patientService) GetPatientByID(patientID int) (*models.Patient, error) {
	// Call repository to get the patient
	patient, err := s.patientRepo.GetPatientByID(patientID)
	if err != nil {
		// Handle repository errors, custom errors still pending
		return nil, fmt.Errorf("service: failed to get patient by ID %d from repository: %w", patientID, err)
	}

	return patient, nil
}

// Fuzzy-search patient names & return patients that match
func (s *patientService) SearchPatients(query string, limit int) ([]models.Patient, error) {
	if len(query) < 2 {
		return nil, fmt.Errorf("query too short")
	}
	return s.patientRepo.SearchPatients(query, limit)
}
