package organization

import "errors"

// Repository-level errors for invitation operations
var (
	// ErrResendLimitReached is returned when maximum resend attempts have been reached
	ErrResendLimitReached = errors.New("maximum resend attempts reached")

	// ErrResendCooldown is returned when trying to resend before cooldown period has elapsed
	ErrResendCooldown = errors.New("must wait before resending")
)
