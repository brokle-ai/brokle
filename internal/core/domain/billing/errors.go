package billing

import (
	"errors"
	"fmt"
)

// Domain errors for billing operations
var (
	// Entity not found errors
	ErrContractNotFound = errors.New("contract not found")
	ErrBillingNotFound  = errors.New("billing record not found")
	ErrPlanNotFound     = errors.New("plan not found")
	ErrBudgetNotFound   = errors.New("budget not found")
	ErrAlertNotFound    = errors.New("alert not found")
	ErrTierNotFound     = errors.New("volume tier not found")
)

// Constructor functions for contextualized errors

func NewContractNotFoundError(id string) error {
	return fmt.Errorf("%w: %s", ErrContractNotFound, id)
}

func NewBillingNotFoundError(orgID string) error {
	return fmt.Errorf("%w: organization %s", ErrBillingNotFound, orgID)
}

func NewPlanNotFoundError(id string) error {
	return fmt.Errorf("%w: %s", ErrPlanNotFound, id)
}

func NewBudgetNotFoundError(id string) error {
	return fmt.Errorf("%w: %s", ErrBudgetNotFound, id)
}

func NewAlertNotFoundError(id string) error {
	return fmt.Errorf("%w: %s", ErrAlertNotFound, id)
}

// IsNotFoundError reports whether err is one of the billing-domain
// not-found sentinels. Used by the billing services to translate any
// not-found variant to a 404 AppError without branching per sentinel.
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrContractNotFound) ||
		errors.Is(err, ErrBillingNotFound) ||
		errors.Is(err, ErrPlanNotFound) ||
		errors.Is(err, ErrBudgetNotFound) ||
		errors.Is(err, ErrAlertNotFound) ||
		errors.Is(err, ErrTierNotFound)
}
