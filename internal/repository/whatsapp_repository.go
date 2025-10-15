package repository

import (
	"context"
	"database/sql"
	"fmt"

	"software-backend/internal/models"
)

type WhatsAppRepository interface {
	GetConfig(ctx context.Context) (*models.WhatsAppConfig, error)
	UpdateConfig(ctx context.Context, config *models.WhatsAppConfig) error
	CreateNotification(ctx context.Context, notification *models.WhatsAppNotification) error
	UpdateNotificationStatus(ctx context.Context, msgID string, status string, deliveredAt, readAt *sql.NullTime) error
	GetPendingNotifications(ctx context.Context, limit int) ([]*models.WhatsAppNotification, error)
	LogWebhook(ctx context.Context, eventType string, payload string) error
	GetNotificationsByAppointment(ctx context.Context, appointmentID int) ([]*models.WhatsAppNotification, error)
}

type whatsAppRepository struct {
	db *sql.DB
}

func NewWhatsAppRepository(db *sql.DB) WhatsAppRepository {
	return &whatsAppRepository{db: db}
}

func (r *whatsAppRepository) GetConfig(ctx context.Context) (*models.WhatsAppConfig, error) {
	query := `
		SELECT id, phone_number_id, access_token, business_account_id, 
		       webhook_verify_token, is_active, reminder_enabled,
		       reminder_3_days_before, reminder_1_day_before, reminder_2_hours_before,
		       template_name_reminder, template_lang_code, created_at, updated_at
		FROM whatsapp_config
		ORDER BY id DESC
		LIMIT 1
	`

	var config models.WhatsAppConfig
	err := r.db.QueryRowContext(ctx, query).Scan(
		&config.ID,
		&config.PhoneNumberID,
		&config.AccessToken,
		&config.BusinessAccountID,
		&config.WebhookVerifyToken,
		&config.IsActive,
		&config.ReminderEnabled,
		&config.Reminder3DaysBefore,
		&config.Reminder1DayBefore,
		&config.Reminder2HoursBefore,
		&config.TemplateNameReminder,
		&config.TemplateLangCode,
		&config.CreatedAt,
		&config.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no WhatsApp configuration found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get WhatsApp config: %w", err)
	}

	return &config, nil
}

func (r *whatsAppRepository) UpdateConfig(ctx context.Context, config *models.WhatsAppConfig) error {
	query := `
		UPDATE whatsapp_config
		SET phone_number_id = $1,
		    access_token = $2,
		    business_account_id = $3,
		    webhook_verify_token = $4,
		    is_active = $5,
		    reminder_enabled = $6,
		    reminder_3_days_before = $7,
		    reminder_1_day_before = $8,
		    reminder_2_hours_before = $9,
		    template_name_reminder = $10,
		    template_lang_code = $11,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $12
	`

	_, err := r.db.ExecContext(ctx, query,
		config.PhoneNumberID,
		config.AccessToken,
		config.BusinessAccountID,
		config.WebhookVerifyToken,
		config.IsActive,
		config.ReminderEnabled,
		config.Reminder3DaysBefore,
		config.Reminder1DayBefore,
		config.Reminder2HoursBefore,
		config.TemplateNameReminder,
		config.TemplateLangCode,
		config.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update WhatsApp config: %w", err)
	}

	return nil
}

func (r *whatsAppRepository) CreateNotification(ctx context.Context, notification *models.WhatsAppNotification) error {
	query := `
		INSERT INTO whatsapp_notifications (
			appointment_id, patient_id, phone_number, message_type, 
			status, whatsapp_msg_id, error_message, sent_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (appointment_id, message_type) DO UPDATE
		SET status = EXCLUDED.status,
		    whatsapp_msg_id = EXCLUDED.whatsapp_msg_id,
		    error_message = EXCLUDED.error_message,
		    sent_at = EXCLUDED.sent_at,
		    updated_at = CURRENT_TIMESTAMP
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRowContext(ctx, query,
		notification.AppointmentID,
		notification.PatientID,
		notification.PhoneNumber,
		notification.MessageType,
		notification.Status,
		notification.WhatsAppMsgID,
		notification.ErrorMessage,
		notification.SentAt,
	).Scan(&notification.ID, &notification.CreatedAt, &notification.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create notification: %w", err)
	}

	return nil
}

func (r *whatsAppRepository) UpdateNotificationStatus(ctx context.Context, msgID string, status string, deliveredAt, readAt *sql.NullTime) error {
	query := `
		UPDATE whatsapp_notifications
		SET status = $1,
		    delivered_at = COALESCE($2, delivered_at),
		    read_at = COALESCE($3, read_at),
		    updated_at = CURRENT_TIMESTAMP
		WHERE whatsapp_msg_id = $4
	`

	_, err := r.db.ExecContext(ctx, query, status, deliveredAt, readAt, msgID)
	if err != nil {
		return fmt.Errorf("failed to update notification status: %w", err)
	}

	return nil
}

func (r *whatsAppRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*models.WhatsAppNotification, error) {
	query := `
		SELECT id, appointment_id, patient_id, phone_number, message_type, 
		       status, whatsapp_msg_id, error_message, sent_at, 
		       delivered_at, read_at, created_at, updated_at
		FROM whatsapp_notifications
		WHERE status = 'pending'
		ORDER BY created_at ASC
		LIMIT $1
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.WhatsAppNotification
	for rows.Next() {
		var n models.WhatsAppNotification
		err := rows.Scan(
			&n.ID,
			&n.AppointmentID,
			&n.PatientID,
			&n.PhoneNumber,
			&n.MessageType,
			&n.Status,
			&n.WhatsAppMsgID,
			&n.ErrorMessage,
			&n.SentAt,
			&n.DeliveredAt,
			&n.ReadAt,
			&n.CreatedAt,
			&n.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, &n)
	}

	return notifications, nil
}

func (r *whatsAppRepository) LogWebhook(ctx context.Context, eventType string, payload string) error {
	query := `
		INSERT INTO whatsapp_webhook_logs (event_type, payload)
		VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(ctx, query, eventType, payload)
	if err != nil {
		return fmt.Errorf("failed to log webhook: %w", err)
	}

	return nil
}

func (r *whatsAppRepository) GetNotificationsByAppointment(ctx context.Context, appointmentID int) ([]*models.WhatsAppNotification, error) {
	query := `
		SELECT id, appointment_id, patient_id, phone_number, message_type, 
		       status, whatsapp_msg_id, error_message, sent_at, 
		       delivered_at, read_at, created_at, updated_at
		FROM whatsapp_notifications
		WHERE appointment_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.WhatsAppNotification
	for rows.Next() {
		var n models.WhatsAppNotification
		err := rows.Scan(
			&n.ID,
			&n.AppointmentID,
			&n.PatientID,
			&n.PhoneNumber,
			&n.MessageType,
			&n.Status,
			&n.WhatsAppMsgID,
			&n.ErrorMessage,
			&n.SentAt,
			&n.DeliveredAt,
			&n.ReadAt,
			&n.CreatedAt,
			&n.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}
		notifications = append(notifications, &n)
	}

	return notifications, nil
}
