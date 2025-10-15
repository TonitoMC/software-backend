package api

import (
	"encoding/json"
	"net/http"

	"software-backend/internal/middleware"
	"software-backend/internal/service"
	"software-backend/internal/whatsapp"

	"github.com/labstack/echo/v4"
)

type WhatsAppHandler struct {
	service service.WhatsAppService
}

func NewWhatsAppHandler(service service.WhatsAppService) *WhatsAppHandler {
	return &WhatsAppHandler{service: service}
}

func (h *WhatsAppHandler) RegisterRoutes(e *echo.Group) {
	// Admin routes - protected
	adminGroup := e.Group("/whatsapp")
	adminGroup.Use(middleware.JWTAuth())

	adminGroup.GET("/config", h.GetConfig)
	adminGroup.PUT("/config", h.UpdateConfig)

	// Webhook - unprotected (verified by Meta)
	e.GET("/whatsapp/webhook", h.VerifyWebhook)
	e.POST("/whatsapp/webhook", h.HandleWebhook)
}

// GetConfig returns the current WhatsApp configuration
func (h *WhatsAppHandler) GetConfig(c echo.Context) error {
	config, err := h.service.GetConfig(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get configuration",
		})
	}

	// Don't expose sensitive tokens in response
	config.AccessToken = ""
	config.WebhookVerifyToken = ""

	return c.JSON(http.StatusOK, config)
}

// UpdateConfig updates WhatsApp configuration
func (h *WhatsAppHandler) UpdateConfig(c echo.Context) error {
	var config map[string]interface{}
	if err := c.Bind(&config); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Get existing config
	existingConfig, err := h.service.GetConfig(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get existing configuration",
		})
	}

	// Update fields if provided
	if v, ok := config["phone_number_id"].(string); ok {
		existingConfig.PhoneNumberID = v
	}
	if v, ok := config["access_token"].(string); ok && v != "" {
		existingConfig.AccessToken = v
	}
	if v, ok := config["business_account_id"].(string); ok {
		existingConfig.BusinessAccountID = v
	}
	if v, ok := config["webhook_verify_token"].(string); ok && v != "" {
		existingConfig.WebhookVerifyToken = v
	}
	if v, ok := config["is_active"].(bool); ok {
		existingConfig.IsActive = v
	}
	if v, ok := config["reminder_enabled"].(bool); ok {
		existingConfig.ReminderEnabled = v
	}
	if v, ok := config["reminder_3_days_before"].(bool); ok {
		existingConfig.Reminder3DaysBefore = v
	}
	if v, ok := config["reminder_1_day_before"].(bool); ok {
		existingConfig.Reminder1DayBefore = v
	}
	if v, ok := config["reminder_2_hours_before"].(bool); ok {
		existingConfig.Reminder2HoursBefore = v
	}
	if v, ok := config["template_name_reminder"].(string); ok {
		existingConfig.TemplateNameReminder = v
	}
	if v, ok := config["template_lang_code"].(string); ok {
		existingConfig.TemplateLangCode = v
	}

	if err := h.service.UpdateConfig(c.Request().Context(), existingConfig); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update configuration",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Configuration updated successfully",
	})
}

// VerifyWebhook handles webhook verification from Meta
func (h *WhatsAppHandler) VerifyWebhook(c echo.Context) error {
	mode := c.QueryParam("hub.mode")
	token := c.QueryParam("hub.verify_token")
	challenge := c.QueryParam("hub.challenge")

	config, err := h.service.GetConfig(c.Request().Context())
	if err != nil {
		return c.String(http.StatusInternalServerError, "Configuration error")
	}

	result, err := whatsapp.VerifyWebhook(config.WebhookVerifyToken, mode, token, challenge)
	if err != nil {
		return c.String(http.StatusForbidden, "Verification failed")
	}

	return c.String(http.StatusOK, result)
}

// HandleWebhook processes incoming webhooks from WhatsApp
func (h *WhatsAppHandler) HandleWebhook(c echo.Context) error {
	var payload whatsapp.WebhookPayload
	if err := c.Bind(&payload); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid webhook payload",
		})
	}

	// Log the webhook (async)
	go func() {
		payloadBytes, _ := json.Marshal(payload)
		// We would log this to database here
		_ = payloadBytes
	}()

	// Process status updates
	if err := h.service.ProcessWebhookStatus(c.Request().Context(), &payload); err != nil {
		// Log error but return 200 to Meta to acknowledge receipt
		c.Logger().Errorf("Failed to process webhook: %v", err)
	}

	// Always return 200 to Meta
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
