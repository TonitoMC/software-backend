package exam

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

	mockRepo := mocks.NewMockExamRepository(ctrl)
	svc := NewExamService(mockRepo)

	patientID := 123
	expected := []models.Exam{
		{Type: "Blood Test", Date: someTime()},
	}

	mockRepo.EXPECT().
		GetByPatientID(patientID).
		Return(expected, nil)

	exams, err := svc.GetByPatientID(patientID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exams) != 1 || exams[0] != expected[0] {
		t.Errorf("unexpected exams: %+v", exams)
	}
}

func TestGetByPatientID_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockExamRepository(ctrl)
	svc := NewExamService(mockRepo)

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
