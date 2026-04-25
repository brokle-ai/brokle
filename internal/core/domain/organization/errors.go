package organization

import "errors"

var (
	// Organization errors
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")

	// Member errors
	ErrMemberNotFound      = errors.New("member not found")
	ErrMemberAlreadyExists = errors.New("member already exists")
	ErrInsufficientRole    = errors.New("insufficient role")

	// Project errors
	ErrProjectNotFound      = errors.New("project not found")
	ErrProjectAlreadyExists = errors.New("project already exists")

	// Invitation errors
	ErrInvitationNotFound = errors.New("invitation not found")
	// ErrInvitationResendLimit is returned when the configured maximum
	// number of resend attempts has been reached on an invitation token.
	ErrInvitationResendLimit = errors.New("maximum invitation resend attempts reached")
	// ErrInvitationResendCooldown is returned when a resend is requested
	// before the per-invitation cooldown has elapsed.
	ErrInvitationResendCooldown = errors.New("must wait before resending invitation")

	// Settings errors
	ErrSettingsNotFound      = errors.New("settings not found")
	ErrSettingsAlreadyExists = errors.New("settings already exist")
)
