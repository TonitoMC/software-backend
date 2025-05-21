// service/business_hours_service.go
package service

import (
	"time"

	"software-backend/internal/models"
	"software-backend/internal/repository"
)

// BussinessHoursService interface defines the methods expected from the service
type BusinessHoursService interface {
	GetBusinessHoursForDate(date time.Time) ([]models.BusinessHourInterval, error)
}

// Struct to manage dependencies
type businessHoursService struct {
	repo repository.BusinessHoursRepository
}

// Constructor to pass on dependencies
func NewBusinessHoursService(repo repository.BusinessHoursRepository) BusinessHoursService {
	return &businessHoursService{repo: repo}
}

// Get business hours for a specific day, date in Go Time format
func (s *businessHoursService) GetBusinessHoursForDate(date time.Time) ([]models.BusinessHourInterval, error) {
	return s.repo.GetBusinessHoursForDate(date)
}
