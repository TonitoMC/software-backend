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

type WhatsAppService interface {
	GetConfig(ctx context.Context) (*models.WhatsAppConfig, error)
	UpdateConfig(ctx context.Context, config *models.WhatsAppConfig) error
	SendReminder(ctx context.Context, appointment *models.Appointment, patient *models.Patient, messageType string) error
	ProcessPendingReminders(ctx context.Context) error
	CheckAndScheduleReminders(ctx context.Context) error
	ProcessWebhookStatus(ctx context.Context, payload *whatsapp.WebhookPayload) error
}

type whatsAppService struct {
	whatsAppRepo    repository.WhatsAppRepository
	appointmentRepo appointment_repo.AppointmentRepository
	patientRepo     patient_repo.PatientRepository
}

func NewWhatsAppService(
	whatsAppRepo repository.WhatsAppRepository,
	appointmentRepo appointment_repo.AppointmentRepository,
	patientRepo patient_repo.PatientRepository,
) WhatsAppService {
	return &whatsAppService{
		whatsAppRepo:    whatsAppRepo,
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
	}
}

func (s *whatsAppService) GetConfig(ctx context.Context) (*models.WhatsAppConfig, error) {
	return s.whatsAppRepo.GetConfig(ctx)
}

func (s *whatsAppService) UpdateConfig(ctx context.Context, config *models.WhatsAppConfig) error {
	return s.whatsAppRepo.UpdateConfig(ctx, config)
}

func (s *whatsAppService) SendReminder(ctx context.Context, appointment *models.Appointment, patient *models.Patient, messageType string) error {
	config, err := s.whatsAppRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get WhatsApp config: %w", err)
	}

	if !config.IsActive || !config.ReminderEnabled {
		return fmt.Errorf("WhatsApp reminders are not enabled")
	}

	// Validate phone number
	if patient.Phone == "" {
		return fmt.Errorf("patient has no phone number")
	}

	// Format phone number (ensure it starts with +)
	phoneNumber := patient.Phone
	if phoneNumber[0] != '+' {
		phoneNumber = "+" + phoneNumber
	}

	// Format appointment date and time
	appointmentDate := appointment.Date.Format("02/01/2006")
	appointmentTime := appointment.Date.Format("15:04")

	// Create WhatsApp client
	client := whatsapp.NewClient(config)

	// Send the message
	response, err := client.SendTemplateMessage(ctx, phoneNumber, patient.Name, appointmentDate, appointmentTime)
	
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
		return fmt.Errorf("failed to send WhatsApp message: %w", err)
	}

	// Save notification record
	notification.Status = "sent"
	if len(response.Messages) > 0 {
		notification.WhatsAppMsgID = response.Messages[0].ID
	}

	if err := s.whatsAppRepo.CreateNotification(ctx, notification); err != nil {
		log.Printf("Warning: Failed to save notification record: %v", err)
	}

	return nil
}

func (s *whatsAppService) CheckAndScheduleReminders(ctx context.Context) error {
	config, err := s.whatsAppRepo.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	if !config.IsActive || !config.ReminderEnabled {
		return nil // Silently skip if disabled
	}

	now := time.Now()
	
	// Define time windows for checking upcoming appointments
	var timeWindows []struct {
		start       time.Time
		end         time.Time
		messageType string
		enabled     bool
	}

	if config.Reminder3DaysBefore {
		// 3 days before (between 72-73 hours from now)
		start3d := now.Add(72 * time.Hour)
		end3d := now.Add(73 * time.Hour)
		timeWindows = append(timeWindows, struct {
			start       time.Time
			end         time.Time
			messageType string
			enabled     bool
		}{start3d, end3d, "3_days", true})
	}

	if config.Reminder1DayBefore {
		// 1 day before (between 24-25 hours from now)
		start1d := now.Add(24 * time.Hour)
		end1d := now.Add(25 * time.Hour)
		timeWindows = append(timeWindows, struct {
			start       time.Time
			end         time.Time
			messageType string
			enabled     bool
		}{start1d, end1d, "1_day", true})
	}

	if config.Reminder2HoursBefore {
		// 2 hours before (between 2-3 hours from now)
		start2h := now.Add(2 * time.Hour)
		end2h := now.Add(3 * time.Hour)
		timeWindows = append(timeWindows, struct {
			start       time.Time
			end         time.Time
			messageType string
			enabled     bool
		}{start2h, end2h, "2_hours", true})
	}

	// For each time window, find appointments and schedule reminders
	for _, window := range timeWindows {
		appointments, err := s.appointmentRepo.GetAppointmentsByDateRange(ctx, window.start, window.end)
		if err != nil {
			log.Printf("Error fetching appointments for %s reminders: %v", window.messageType, err)
			continue
		}

		for _, appointment := range appointments {
			// Check if notification already exists for this appointment and type
			existing, _ := s.whatsAppRepo.GetNotificationsByAppointment(ctx, appointment.ID)
			alreadySent := false
			for _, n := range existing {
				if n.MessageType == window.messageType && (n.Status == "sent" || n.Status == "delivered" || n.Status == "read") {
					alreadySent = true
					break
				}
			}

			if alreadySent {
				continue // Skip if already sent
			}

			// Get patient details
			patient, err := s.patientRepo.GetPatientByID(ctx, appointment.PatientID)
			if err != nil {
				log.Printf("Error fetching patient %d: %v", appointment.PatientID, err)
				continue
			}

			// Send reminder
			if err := s.SendReminder(ctx, appointment, patient, window.messageType); err != nil {
				log.Printf("Error sending %s reminder for appointment %d: %v", window.messageType, appointment.ID, err)
			} else {
				log.Printf("Successfully sent %s reminder for appointment %d to patient %s", window.messageType, appointment.ID, patient.Name)
			}
		}
	}

	return nil
}

func (s *whatsAppService) ProcessPendingReminders(ctx context.Context) error {
	// Get pending notifications that failed to send
	notifications, err := s.whatsAppRepo.GetPendingNotifications(ctx, 50)
	if err != nil {
		return fmt.Errorf("failed to get pending notifications: %w", err)
	}

	for _, notification := range notifications {
		// Get appointment and patient
		appointment, err := s.appointmentRepo.GetAppointmentByID(ctx, notification.AppointmentID)
		if err != nil {
			log.Printf("Error fetching appointment %d: %v", notification.AppointmentID, err)
			continue
		}

		patient, err := s.patientRepo.GetPatientByID(ctx, appointment.PatientID)
		if err != nil {
			log.Printf("Error fetching patient %d: %v", appointment.PatientID, err)
			continue
		}

		// Retry sending
		if err := s.SendReminder(ctx, appointment, patient, notification.MessageType); err != nil {
			log.Printf("Failed to retry notification %d: %v", notification.ID, err)
		}
	}

	return nil
}

func (s *whatsAppService) ProcessWebhookStatus(ctx context.Context, payload *whatsapp.WebhookPayload) error {
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
					log.Printf("Error updating notification status for message %s: %v", status.ID, err)
				}
			}
		}
	}

	return nil
}
