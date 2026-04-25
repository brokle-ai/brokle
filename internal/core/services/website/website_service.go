// Package website implements public-facing website operations such as contact-form submissions.
package website

import (
	"context"
	"log/slog"
	"time"

	"brokle/internal/core/domain/website"
	"brokle/pkg/email"
	appErrors "brokle/pkg/errors"
)

type WebsiteService struct {
	contactRepo website.ContactSubmissionRepository
	emailSender email.EmailSender
	notifyEmail string // from WEBSITE_NOTIFICATION_EMAIL env
	logger      *slog.Logger
}

func NewWebsiteService(
	contactRepo website.ContactSubmissionRepository,
	emailSender email.EmailSender,
	notifyEmail string,
	logger *slog.Logger,
) *WebsiteService {
	return &WebsiteService{
		contactRepo: contactRepo,
		emailSender: emailSender,
		notifyEmail: notifyEmail,
		logger:      logger.With("service", "website"),
	}
}

func (s *WebsiteService) SubmitContactForm(ctx context.Context, req *website.CreateContactSubmissionRequest, ipAddress, userAgent string) error {
	submission := website.NewContactSubmission(req, ipAddress, userAgent)

	if err := s.contactRepo.Create(ctx, submission); err != nil {
		s.logger.Error("failed to store contact submission",
			"error", err,
			"email", req.Email,
		)
		return appErrors.NewInternalError("failed to store contact submission", err)
	}

	s.logger.Info("contact form submitted",
		"submission_id", submission.ID,
		"email", req.Email,
		"subject", req.Subject,
	)

	// Send notification email (best-effort, async).
	// Use detached context with timeout since request context will be canceled when handler returns.
	if s.notifyEmail != "" {
		go func() {
			emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			s.sendNotificationEmail(emailCtx, req, ipAddress)
		}()
	}

	return nil
}

func (s *WebsiteService) sendNotificationEmail(ctx context.Context, req *website.CreateContactSubmissionRequest, ipAddress string) {
	htmlContent, textContent, err := email.BuildContactNotificationEmail(email.ContactNotificationEmailParams{
		Name:        req.Name,
		Email:       req.Email,
		Company:     req.Company,
		Subject:     req.Subject,
		Message:     req.Message,
		InquiryType: req.InquiryType,
		IPAddress:   ipAddress,
	})
	if err != nil {
		s.logger.Error("failed to build contact notification email",
			"error", err,
			"email", req.Email,
		)
		return
	}

	if err := s.emailSender.Send(ctx, email.SendEmailParams{
		To:      []string{s.notifyEmail},
		Subject: "Contact Form: " + req.Subject,
		HTML:    htmlContent,
		Text:    textContent,
		ReplyTo: req.Email,
		Tags: map[string]string{
			"type":   "contact_form",
			"source": "website",
		},
	}); err != nil {
		s.logger.Error("failed to send contact notification email",
			"error", err,
			"notify_email", s.notifyEmail,
			"from_email", req.Email,
		)
	}
}
