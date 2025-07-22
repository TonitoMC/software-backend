package patient

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"software-backend/internal/mocks"
	"software-backend/internal/models"

	"github.com/golang/mock/gomock"
)

func TestGetPatientByID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPatientRepository(ctrl)
	svc := NewPatientService(mockRepo)

	patientID := 123
	expected := &models.Patient{
		ID:          patientID,
		Name:        "Jane Doe",
		DateOfBirth: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		Phone:       "555-1234",
		Sex:         "F",
	}

	mockRepo.EXPECT().
		GetPatientByID(patientID).
		Return(expected, nil)

	patient, err := svc.GetPatientByID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if *patient != *expected {
		t.Errorf("unexpected patient: %+v", patient)
	}
}

func TestGetPatientByID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPatientRepository(ctrl)
	svc := NewPatientService(mockRepo)

	patientID := 999
	mockRepo.EXPECT().
		GetPatientByID(patientID).
		Return(nil, errors.New("not found"))

	_, err := svc.GetPatientByID(patientID)
	if err == nil || err.Error() != fmt.Sprintf("service: failed to get patient by ID %d from repository: not found", patientID) {
		t.Errorf("expected wrapped error, got %v", err)
	}
}

func TestSearchPatients_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPatientRepository(ctrl)
	svc := NewPatientService(mockRepo)

	query := "Jane"
	limit := 5
	expected := []models.Patient{
		{ID: 1, Name: "Jane Doe"},
		{ID: 2, Name: "Janet Smith"},
	}

	mockRepo.EXPECT().
		SearchPatients(query, limit).
		Return(expected, nil)

	patients, err := svc.SearchPatients(query, limit)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(patients) != 2 || patients[0] != expected[0] || patients[1] != expected[1] {
		t.Errorf("unexpected patients: %+v", patients)
	}
}

func TestSearchPatients_QueryTooShort(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockPatientRepository(ctrl)
	svc := NewPatientService(mockRepo)

	_, err := svc.SearchPatients("A", 5)
	if err == nil || err.Error() != "query too short" {
		t.Errorf("expected query too short error, got %v", err)
	}
}
