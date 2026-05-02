// Package website wires the marketing-site contact form endpoint
// onto the dashboard plane. Public route; no auth required. IP-based
// rate limiting is applied at the route group level in routes.go.
package website

import (
	"log/slog"
	"net/http"

	"brokle/internal/core/domain/website"
	websiteService "brokle/internal/core/services/website"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	svc    *websiteService.WebsiteService
	logger *slog.Logger
}

// New constructs a Handler with all required services.
func New(svc *websiteService.WebsiteService, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// submitContact accepts a marketing contact submission, logs
// rejections for the abuse/audit trail, and returns 201 with a
// human-readable confirmation the browser can render verbatim.
//
// IP + User-Agent are read from the request context populated by the
// RequestMetadata middleware in the global chain; when that
// middleware did not run (direct handler tests without the full
// chain) both return "" and the submission records empty values
// rather than crashing.
func (h *Handler) SubmitContact(w http.ResponseWriter, r *http.Request) {
	var body submitContactBody
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	req := &website.CreateContactSubmissionRequest{
		Name:        body.Name,
		Email:       body.Email,
		Company:     body.Company,
		Subject:     body.Subject,
		Message:     body.Message,
		InquiryType: body.InquiryType,
	}

	if err := h.svc.SubmitContactForm(r.Context(), req,
		httpctx.ClientIP(r.Context()), httpctx.UserAgent(r.Context())); err != nil {
		h.logger.WarnContext(r.Context(), "contact-form submission rejected",
			"error", err,
			"email", req.Email,
			"subject", req.Subject,
		)
		response.WriteError(w, err)
		return
	}

	response.Created(w, submitContactResponse{
		Message: "Thank you for your message. We'll get back to you soon.",
	})
}
