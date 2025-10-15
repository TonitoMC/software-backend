package models

import (
	"time"
)

// WhatsAppConfig stores WhatsApp Business API credentials and configuration
type WhatsAppConfig struct {
	ID                   int       `json:"id" db:"id"`
	PhoneNumberID        string    `json:"phone_number_id" db:"phone_number_id"` // WhatsApp Business Phone Number ID
	AccessToken          string    `json:"-" db:"access_token"`                  // Bearer token (not exposed in JSON)
	BusinessAccountID    string    `json:"business_account_id" db:"business_account_id"`
	WebhookVerifyToken   string    `json:"-" db:"webhook_verify_token"` // Token for webhook verification
	IsActive             bool      `json:"is_active" db:"is_active"`
	ReminderEnabled      bool      `json:"reminder_enabled" db:"reminder_enabled"`
	Reminder3DaysBefore  bool      `json:"reminder_3_days_before" db:"reminder_3_days_before"`
	Reminder1DayBefore   bool      `json:"reminder_1_day_before" db:"reminder_1_day_before"`
	Reminder2HoursBefore bool      `json:"reminder_2_hours_before" db:"reminder_2_hours_before"`
	TemplateNameReminder string    `json:"template_name_reminder" db:"template_name_reminder"` // e.g., "recordatorio_cita"
	TemplateLangCode     string    `json:"template_lang_code" db:"template_lang_code"`         // e.g., "es" or "es_MX"
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// WhatsAppNotification tracks sent WhatsApp reminders
type WhatsAppNotification struct {
	ID            int        `json:"id" db:"id"`
	AppointmentID int        `json:"appointment_id" db:"appointment_id"`
	PatientID     int        `json:"patient_id" db:"patient_id"`
	PhoneNumber   string     `json:"phone_number" db:"phone_number"`
	MessageType   string     `json:"message_type" db:"message_type"` // "3_days", "1_day", "2_hours"
	Status        string     `json:"status" db:"status"`             // "pending", "sent", "delivered", "read", "failed"
	WhatsAppMsgID string     `json:"whatsapp_msg_id" db:"whatsapp_msg_id"`
	ErrorMessage  string     `json:"error_message" db:"error_message"`
	SentAt        time.Time  `json:"sent_at" db:"sent_at"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	ReadAt        *time.Time `json:"read_at,omitempty" db:"read_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// WhatsAppWebhookLog logs incoming webhooks for debugging
type WhatsAppWebhookLog struct {
	ID        int       `json:"id" db:"id"`
	EventType string    `json:"event_type" db:"event_type"` // "message", "status", etc.
	Payload   string    `json:"payload" db:"payload"`       // JSON payload
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
