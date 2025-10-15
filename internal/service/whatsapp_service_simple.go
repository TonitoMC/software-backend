package service

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"software-backend/internal/models"
	"software-backend/internal/repository"
	appointment_repo "software-backend/internal/repository/appointment"
	patient_repo "software-backend/internal/repository/patient"
	"software-backend/internal/whatsapp"
)

// WhatsAppService simplified interface
type WhatsAppServiceSimple interface {
	GetConfig(ctx context.Context) (*models.WhatsAppConfig, error)
	UpdateConfig(ctx context.Context, config *models.WhatsAppConfig) error
	CheckAndScheduleReminders(ctx context.Context) error
	ProcessWebhookStatus(ctx context.Context, payload *whatsapp.WebhookPayload) error
}

type whatsAppServiceSimple struct {
	whatsAppRepo    repository.WhatsAppRepository
	appointmentRepo appointment_repo.AppointmentRepository
	patientRepo     patient_repo.PatientRepository
}

func NewWhatsAppServiceSimple(
	whatsAppRepo repository.WhatsAppRepository,
	appointmentRepo appointment_repo.AppointmentRepository,
	patientRepo patient_repo.PatientRepository,
) WhatsAppServiceSimple {
	return &whatsAppServiceSimple{
		whatsAppRepo:    whatsAppRepo,
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
	}
}

func (s *whatsAppServiceSimple) GetConfig(ctx context.Context) (*models.WhatsAppConfig, error) {
	return s.whatsAppRepo.GetConfig(ctx)
}

func (s *whatsAppServiceSimple) UpdateConfig(ctx context.Context, config *models.WhatsAppConfig) error {
	return s.whatsAppRepo.UpdateConfig(ctx, config)
}

func (s *whatsAppServiceSimple) CheckAndScheduleReminders(ctx context.Context) error {
	config, err := s.whatsAppRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if !config.IsActive || !config.ReminderEnabled {
		return nil // Silently skip if disabled
	}

	now := time.Now()

	// Define time windows
	type TimeWindow struct {
		Start       time.Time
		End         time.Time
		MessageType string
		Enabled     bool
	}

	var windows []TimeWindow

	if config.Reminder3DaysBefore {
		windows = append(windows, TimeWindow{
			Start:       now.Add(72 * time.Hour),
			End:         now.Add(73 * time.Hour),
			MessageType: "3_days",
			Enabled:     true,
		})
	}

	if config.Reminder1DayBefore {
		windows = append(windows, TimeWindow{
			Start:       now.Add(24 * time.Hour),
			End:         now.Add(25 * time.Hour),
			MessageType: "1_day",
			Enabled:     true,
		})
	}

	if config.Reminder2HoursBefore {
		windows = append(windows, TimeWindow{
			Start:       now.Add(2 * time.Hour),
			End:         now.Add(3 * time.Hour),
			MessageType: "2_hours",
			Enabled:     true,
		})
	}

	// Check each window
	for _, window := range windows {
		appointments, err := s.appointmentRepo.ListAppointmentsInDateRange(window.Start, window.End)
		if err != nil {
			log.Printf("Error fetching appointments for %s: %v", window.MessageType, err)
			continue
		}

		for _, appointment := range appointments {
			// Check if already sent
			existing, _ := s.whatsAppRepo.GetNotificationsByAppointment(ctx, appointment.ID)
			alreadySent := false
			for _, n := range existing {
				if n.MessageType == window.MessageType && (n.Status == "sent" || n.Status == "delivered" || n.Status == "read") {
					alreadySent = true
					break
				}
			}

			if alreadySent {
				continue
			}

			// Get patient
			patient, err := s.patientRepo.GetPatientByID(appointment.PatientID)
			if err != nil {
				log.Printf("Error fetching patient %d: %v", appointment.PatientID, err)
				continue
			}

			// Send reminder
			if err := s.sendReminder(ctx, config, &appointment, patient, window.MessageType); err != nil {
				log.Printf("Error sending %s reminder for appointment %d: %v", window.MessageType, appointment.ID, err)
			} else {
				log.Printf("✅ Sent %s reminder for appointment %d to %s", window.MessageType, appointment.ID, patient.Name)
			}
		}
	}

	return nil
}

func (s *whatsAppServiceSimple) sendReminder(ctx context.Context, config *models.WhatsAppConfig, appointment *models.Appointment, patient *models.Patient, messageType string) error {
	// Validate phone
	if patient.Phone == "" {
		return fmt.Errorf("patient has no phone number")
	}

	phoneNumber := patient.Phone
	if phoneNumber[0] != '+' {
		phoneNumber = "+" + phoneNumber
	}

	// Format date and time
	appointmentDate := appointment.Start.Format("02/01/2006")
	appointmentTime := appointment.Start.Format("15:04")

	// Create client and send
	client := whatsapp.NewClient(config)
	response, err := client.SendTemplateMessage(ctx, phoneNumber, patient.Name, appointmentDate, appointmentTime)

	// Create notification record
	notification := &models.WhatsAppNotification{
		AppointmentID: appointment.ID,
		PatientID:     patient.ID,
		PhoneNumber:   phoneNumber,
		MessageType:   messageType,
		SentAt:        time.Now(),
	}

	if err != nil {
		notification.Status = "failed"
		notification.ErrorMessage = err.Error()
		_ = s.whatsAppRepo.CreateNotification(ctx, notification)
		return err
	}

	notification.Status = "sent"
	if len(response.Messages) > 0 {
		notification.WhatsAppMsgID = response.Messages[0].ID
	}

	_ = s.whatsAppRepo.CreateNotification(ctx, notification)
	return nil
}

func (s *whatsAppServiceSimple) ProcessWebhookStatus(ctx context.Context, payload *whatsapp.WebhookPayload) error {
	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, status := range change.Value.Statuses {
				var deliveredAt, readAt *sql.NullTime

				switch status.Status {
				case "delivered":
					t := time.Now()
					deliveredAt = &sql.NullTime{Time: t, Valid: true}
				case "read":
					t := time.Now()
					readAt = &sql.NullTime{Time: t, Valid: true}
				}

				err := s.whatsAppRepo.UpdateNotificationStatus(ctx, status.ID, status.Status, deliveredAt, readAt)
				if err != nil {
					log.Printf("Error updating status for %s: %v", status.ID, err)
				}
			}
		}
	}
	return nil
}
