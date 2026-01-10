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

	// Conflict errors
	ErrContractAlreadyActive = errors.New("organization already has an active contract")
	ErrBillingAlreadyExists  = errors.New("billing record already exists")

	// Validation errors
	ErrInvalidContractDates = errors.New("invalid contract dates")
	ErrInvalidTierConfig    = errors.New("invalid volume tier configuration")
	ErrInvalidBudgetConfig  = errors.New("invalid budget configuration")
)

// Error codes for structured API responses
const (
	ErrCodeContractNotFound    = "CONTRACT_NOT_FOUND"
	ErrCodeBillingNotFound     = "BILLING_NOT_FOUND"
	ErrCodePlanNotFound        = "PLAN_NOT_FOUND"
	ErrCodeBudgetNotFound      = "BUDGET_NOT_FOUND"
	ErrCodeContractConflict    = "CONTRACT_CONFLICT"
	ErrCodeBillingExists       = "BILLING_EXISTS"
	ErrCodeInvalidContractDate = "INVALID_CONTRACT_DATE"
	ErrCodeInvalidTierConfig   = "INVALID_TIER_CONFIG"
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

// Classification helpers

// IsNotFoundError returns true if the error is a billing not-found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrContractNotFound) ||
		errors.Is(err, ErrBillingNotFound) ||
		errors.Is(err, ErrPlanNotFound) ||
		errors.Is(err, ErrBudgetNotFound) ||
		errors.Is(err, ErrAlertNotFound) ||
		errors.Is(err, ErrTierNotFound)
}

// IsConflictError returns true if the error is a billing conflict error
func IsConflictError(err error) bool {
	return errors.Is(err, ErrContractAlreadyActive) ||
		errors.Is(err, ErrBillingAlreadyExists)
}

// IsValidationError returns true if the error is a billing validation error
func IsValidationError(err error) bool {
	return errors.Is(err, ErrInvalidContractDates) ||
		errors.Is(err, ErrInvalidTierConfig) ||
		errors.Is(err, ErrInvalidBudgetConfig)
}
