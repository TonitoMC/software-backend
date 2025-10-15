package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"software-backend/internal/models"
)

const (
	WhatsAppAPIBaseURL = "https://graph.facebook.com/v20.0"
	RequestTimeout     = 30 * time.Second
)

// Client handles WhatsApp Business Cloud API requests
type Client struct {
	config     *models.WhatsAppConfig
	httpClient *http.Client
}

// NewClient creates a new WhatsApp API client
func NewClient(config *models.WhatsAppConfig) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: RequestTimeout,
		},
	}
}

// TemplateMessage represents a WhatsApp template message payload
type TemplateMessage struct {
	MessagingProduct string   `json:"messaging_product"`
	RecipientType    string   `json:"recipient_type"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Template         Template `json:"template"`
}

type Template struct {
	Name       string                   `json:"name"`
	Language   TemplateLanguage         `json:"language"`
	Components []map[string]interface{} `json:"components"`
}

type TemplateLanguage struct {
	Code string `json:"code"`
}

// SendMessageResponse represents the API response
type SendMessageResponse struct {
	MessagingProduct string `json:"messaging_product"`
	Contacts         []struct {
		Input string `json:"input"`
		WaID  string `json:"wa_id"`
	} `json:"contacts"`
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

// ErrorResponse represents an error from WhatsApp API
type ErrorResponse struct {
	Error struct {
		Message      string `json:"message"`
		Type         string `json:"type"`
		Code         int    `json:"code"`
		ErrorSubcode int    `json:"error_subcode"`
		FbtraceID    string `json:"fbtrace_id"`
	} `json:"error"`
}

// SendTemplateMessage sends a template message to a WhatsApp number
func (c *Client) SendTemplateMessage(ctx context.Context, phoneNumber string, patientName string, appointmentDate string, appointmentTime string) (*SendMessageResponse, error) {
	if !c.config.IsActive {
		return nil, fmt.Errorf("WhatsApp integration is not active")
	}

	// Build the message payload
	msg := TemplateMessage{
		MessagingProduct: "whatsapp",
		RecipientType:    "individual",
		To:               phoneNumber,
		Type:             "template",
		Template: Template{
			Name: c.config.TemplateNameReminder,
			Language: TemplateLanguage{
				Code: c.config.TemplateLangCode,
			},
			Components: []map[string]interface{}{
				{
					"type": "body",
					"parameters": []map[string]interface{}{
						{
							"type": "text",
							"text": patientName,
						},
						{
							"type": "text",
							"text": appointmentDate,
						},
						{
							"type": "text",
							"text": appointmentTime,
						},
					},
				},
			},
		},
	}

	// Marshal to JSON
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %w", err)
	}

	// Build request
	url := fmt.Sprintf("%s/%s/messages", WhatsAppAPIBaseURL, c.config.PhoneNumberID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}
		return nil, fmt.Errorf("WhatsApp API error (code %d): %s", errResp.Error.Code, errResp.Error.Message)
	}

	// Parse success response
	var result SendMessageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// VerifyWebhook verifies incoming webhook requests from Meta
func VerifyWebhook(verifyToken, hubMode, hubVerifyToken, hubChallenge string) (string, error) {
	if hubMode != "subscribe" {
		return "", fmt.Errorf("invalid hub.mode: %s", hubMode)
	}

	if hubVerifyToken != verifyToken {
		return "", fmt.Errorf("invalid verify token")
	}

	return hubChallenge, nil
}

// WebhookPayload represents incoming webhook data
type WebhookPayload struct {
	Object string `json:"object"`
	Entry  []struct {
		ID      string `json:"id"`
		Changes []struct {
			Value struct {
				MessagingProduct string `json:"messaging_product"`
				Metadata         struct {
					DisplayPhoneNumber string `json:"display_phone_number"`
					PhoneNumberID      string `json:"phone_number_id"`
				} `json:"metadata"`
				Contacts []struct {
					Profile struct {
						Name string `json:"name"`
					} `json:"profile"`
					WaID string `json:"wa_id"`
				} `json:"contacts"`
				Messages []struct {
					From      string `json:"from"`
					ID        string `json:"id"`
					Timestamp string `json:"timestamp"`
					Type      string `json:"type"`
					Text      struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
				Statuses []struct {
					ID           string `json:"id"`
					Status       string `json:"status"` // "sent", "delivered", "read"
					Timestamp    string `json:"timestamp"`
					RecipientID  string `json:"recipient_id"`
					Conversation struct {
						ID     string `json:"id"`
						Origin struct {
							Type string `json:"type"`
						} `json:"origin"`
					} `json:"conversation"`
				} `json:"statuses"`
			} `json:"value"`
			Field string `json:"field"`
		} `json:"changes"`
	} `json:"entry"`
}
