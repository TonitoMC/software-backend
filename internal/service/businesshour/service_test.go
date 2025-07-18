package businesshour

import (
	"errors"
	"testing"
	"time"

	"software-backend/internal/mocks"
	"software-backend/internal/models"

	"github.com/golang/mock/gomock"
)

func TestGetBusinessHoursForDate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBusinessHoursRepository(ctrl)
	svc := NewBusinessHoursService(mockRepo)

	testDate := time.Date(2024, 7, 18, 0, 0, 0, 0, time.UTC)
	expected := []models.BusinessHourInterval{
		{Start: "09:00", End: "17:00"},
	}

	mockRepo.EXPECT().
		GetBusinessHoursForDate(testDate).
		Return(expected, nil)

	intervals, err := svc.GetBusinessHoursForDate(testDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(intervals) != 1 || intervals[0] != expected[0] {
		t.Errorf("unexpected intervals: %+v", intervals)
	}
}

func TestGetBusinessHoursForDate_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockBusinessHoursRepository(ctrl)
	svc := NewBusinessHoursService(mockRepo)

	testDate := time.Date(2024, 7, 18, 0, 0, 0, 0, time.UTC)
	mockRepo.EXPECT().
		GetBusinessHoursForDate(testDate).
		Return(nil, errors.New("db error"))

	_, err := svc.GetBusinessHoursForDate(testDate)
	if err == nil || err.Error() != "db error" {
		t.Errorf("expected db error, got %v", err)
	}
}
