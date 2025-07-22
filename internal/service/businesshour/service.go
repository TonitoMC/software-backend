// service/business_hours_service.go
package businesshour

import (
	"time"

	"software-backend/internal/models"
	bh "software-backend/internal/repository/business_hour"
)

// BussinessHoursService interface defines the methods expected from the service
type BusinessHoursService interface {
	GetBusinessHoursForDate(date time.Time) ([]models.BusinessHourInterval, error)
}

// Struct to manage dependencies
type businessHoursService struct {
	repo bh.BusinessHoursRepository
}

// Constructor to pass on dependencies
func NewBusinessHoursService(repo bh.BusinessHoursRepository) BusinessHoursService {
	return &businessHoursService{repo: repo}
}

// Get business hours for a specific day, date in Go Time format
func (s *businessHoursService) GetBusinessHoursForDate(date time.Time) ([]models.BusinessHourInterval, error) {
	return s.repo.GetBusinessHoursForDate(date)
}
