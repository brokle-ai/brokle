package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SendGridConfig contains configuration for the SendGrid client
type SendGridConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
	Timeout   time.Duration
}

// SendGridClient implements EmailSender using SendGrid API v3
type SendGridClient struct {
	apiKey     string
	fromEmail  string
	fromName   string
	httpClient *http.Client
	baseURL    string
}

// NewSendGridClient creates a new SendGrid API client
func NewSendGridClient(cfg SendGridConfig) *SendGridClient {
	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return &SendGridClient{
		apiKey:    cfg.APIKey,
		fromEmail: cfg.FromEmail,
		fromName:  cfg.FromName,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL: "https://api.sendgrid.com/v3",
	}
}

// sendGridRequest represents the SendGrid v3 Mail Send API request
type sendGridRequest struct {
	Personalizations []sendGridPersonalization `json:"personalizations"`
	From             sendGridEmail             `json:"from"`
	Subject          string                    `json:"subject"`
	Content          []sendGridContent         `json:"content"`
	ReplyTo          *sendGridEmail            `json:"reply_to,omitempty"`
}

type sendGridPersonalization struct {
	To []sendGridEmail `json:"to"`
}

type sendGridEmail struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type sendGridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// sendGridError represents SendGrid API error response
type sendGridError struct {
	Errors []struct {
		Message string `json:"message"`
		Field   string `json:"field"`
		Help    string `json:"help"`
	} `json:"errors"`
}

// Send sends an email using the SendGrid API
func (c *SendGridClient) Send(ctx context.Context, params SendEmailParams) error {
	if len(params.To) == 0 {
		return fmt.Errorf("email: recipient list is empty")
	}
	if params.Subject == "" {
		return fmt.Errorf("email: subject is required")
	}
	if params.HTML == "" && params.Text == "" {
		return fmt.Errorf("email: either HTML or Text content is required")
	}

	// Build recipients
	recipients := make([]sendGridEmail, 0, len(params.To))
	for _, email := range params.To {
		recipients = append(recipients, sendGridEmail{Email: email})
	}

	// Build content (SendGrid requires at least one content type)
	content := make([]sendGridContent, 0, 2)
	if params.Text != "" {
		content = append(content, sendGridContent{
			Type:  "text/plain",
			Value: params.Text,
		})
	}
	if params.HTML != "" {
		content = append(content, sendGridContent{
			Type:  "text/html",
			Value: params.HTML,
		})
	}

	// Build request
	req := sendGridRequest{
		Personalizations: []sendGridPersonalization{
			{To: recipients},
		},
		From: sendGridEmail{
			Email: c.fromEmail,
			Name:  c.fromName,
		},
		Subject: params.Subject,
		Content: content,
	}

	// Add reply-to if provided
	if params.ReplyTo != "" {
		req.ReplyTo = &sendGridEmail{Email: params.ReplyTo}
	}

	// Marshal request body
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("email: failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/mail/send", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("email: failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("email: request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("email: failed to read response: %w", err)
	}

	// Check for errors (202 Accepted is success for SendGrid)
	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		var errResp sendGridError
		if err := json.Unmarshal(respBody, &errResp); err == nil && len(errResp.Errors) > 0 {
			return fmt.Errorf("email: SendGrid API error (%d): %s", resp.StatusCode, errResp.Errors[0].Message)
		}
		return fmt.Errorf("email: SendGrid API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// IsConfigured returns true if the client has required configuration
func (c *SendGridClient) IsConfigured() bool {
	return c.apiKey != "" && c.fromEmail != ""
}
