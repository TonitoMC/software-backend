package businesshour

import (
	"database/sql"
	"fmt"
	"time"

	"software-backend/internal/models"
)

// Interface defines methods to interact with repository
type BusinessHoursRepository interface {
	GetBusinessHoursForDate(date time.Time) ([]models.BusinessHourInterval, error)
}

// Struct to manage dependencies
type businessHoursRepository struct {
	db *sql.DB
}

// Constructor to pass on dependencies
func NewBusinessHoursRepository(db *sql.DB) BusinessHoursRepository {
	return &businessHoursRepository{db: db}
}

// Get business hours for a specific date
func (r *businessHoursRepository) GetBusinessHoursForDate(date time.Time) ([]models.BusinessHourInterval, error) {
	var intervals []models.BusinessHourInterval

	// Check for special hours (Holiday, etc.)
	rows, err := r.db.Query(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_especiales
        WHERE fecha = $1
    `, date.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("query special hours: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var start, end time.Time
		if err := rows.Scan(&start, &end); err != nil {
			return nil, fmt.Errorf("scan special hours: %w", err)
		}
		intervals = append(intervals, models.BusinessHourInterval{
			Start: start.Format("15:04"),
			End:   end.Format("15:04"),
		})
	}
	// Return special working hours if found
	if len(intervals) > 0 {
		return intervals, nil
	}

	// If no special work-hours go for normal schedule
	weekday := int(date.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	// Get normal hours for weekday
	rows, err = r.db.Query(`
        SELECT hora_apertura, hora_cierre
        FROM horarios_laborales
        WHERE dia_semana = $1
    `, weekday)
	if err != nil {
		return nil, fmt.Errorf("query regular hours: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var start, end time.Time
		if err := rows.Scan(&start, &end); err != nil {
			return nil, fmt.Errorf("scan regular hours: %w", err)
		}
		intervals = append(intervals, models.BusinessHourInterval{
			Start: start.Format("15:04"),
			End:   end.Format("15:04"),
		})
	}
	// Return regular hours
	return intervals, nil
}
