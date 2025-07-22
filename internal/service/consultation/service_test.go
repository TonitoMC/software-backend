package consultation

import (
	"errors"
	"testing"
	"time"

	"software-backend/internal/mocks"
	"software-backend/internal/models"

	"github.com/golang/mock/gomock"
)

func TestGetByPatientID_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockConsultationRepository(ctrl)
	svc := NewConsultationService(mockRepo)

	patientID := 123
	expected := []models.Consultation{
		{Reason: "Routine", Date: someTime()},
	}

	mockRepo.EXPECT().
		GetByPatientID(patientID).
		Return(expected, nil)

	consultations, err := svc.GetByPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(consultations) != 1 || consultations[0] != expected[0] {
		t.Errorf("unexpected consultations: %+v", consultations)
	}
}

func TestGetByPatientID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockConsultationRepository(ctrl)
	svc := NewConsultationService(mockRepo)

	patientID := 999
	mockRepo.EXPECT().
		GetByPatientID(patientID).
		Return(nil, errors.New("db error"))

	_, err := svc.GetByPatientID(patientID)
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}

// Helper for a fixed time value
func someTime() (t time.Time) {
	t, _ = time.Parse("2006-01-02", "2024-07-18")
	return
}
