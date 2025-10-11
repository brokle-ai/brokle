package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/pkg/ulid"
)

// DiscountCalculator handles discount calculations for billing
type DiscountCalculator struct {
	logger *logrus.Logger
}

// DiscountRule represents a discount rule
type DiscountRule struct {
	ID                ulid.ULID          `json:"id"`
	OrganizationID    *ulid.ULID         `json:"organization_id,omitempty"` // nil for global rules
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	Type              DiscountType       `json:"type"`
	Value             float64            `json:"value"` // percentage (0.1 = 10%) or fixed amount
	MinimumAmount     float64            `json:"minimum_amount"`
	MaximumDiscount   float64            `json:"maximum_discount"`
	Conditions        *DiscountCondition `json:"conditions,omitempty"`
	ValidFrom         time.Time          `json:"valid_from"`
	ValidUntil        *time.Time         `json:"valid_until,omitempty"`
	UsageLimit        *int               `json:"usage_limit,omitempty"`
	UsageCount        int                `json:"usage_count"`
	IsActive          bool               `json:"is_active"`
	Priority          int                `json:"priority"` // Higher priority rules are applied first
	CreatedAt         time.Time          `json:"created_at"`
	UpdatedAt         time.Time          `json:"updated_at"`
}

// DiscountType represents the type of discount
type DiscountType string

const (
	DiscountTypePercentage DiscountType = "percentage"
	DiscountTypeFixed      DiscountType = "fixed"
	DiscountTypeTiered     DiscountType = "tiered"
)

// DiscountCondition represents conditions for applying discounts
type DiscountCondition struct {
	BillingTiers     []string           `json:"billing_tiers,omitempty"`     // Apply only to specific tiers
	MinUsage         *UsageThreshold    `json:"min_usage,omitempty"`         // Minimum usage requirements
	RequestTypes     []string           `json:"request_types,omitempty"`     // Specific request types
	Providers        []ulid.ULID        `json:"providers,omitempty"`         // Specific providers
	Models           []ulid.ULID        `json:"models,omitempty"`            // Specific models
	TimeOfDay        *TimeRange         `json:"time_of_day,omitempty"`       // Time-based discounts
	DaysOfWeek       []time.Weekday     `json:"days_of_week,omitempty"`      // Day-based discounts
	FirstTimeCustomer bool              `json:"first_time_customer"`         // First-time customer discount
	VolumeThreshold  *VolumeDiscount    `json:"volume_threshold,omitempty"`  // Volume-based discounts
}

// UsageThreshold represents minimum usage requirements
type UsageThreshold struct {
	Requests int64   `json:"requests"`
	Tokens   int64   `json:"tokens"`
	Cost     float64 `json:"cost"`
}

// TimeRange represents a time range for discounts
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// VolumeDiscount represents volume-based discount tiers
type VolumeDiscount struct {
	Tiers []VolumeTier `json:"tiers"`
}

// VolumeTier represents a single volume discount tier
type VolumeTier struct {
	MinAmount float64 `json:"min_amount"`
	Discount  float64 `json:"discount"` // percentage or fixed amount
}

// DiscountCalculation represents the result of discount calculation
type DiscountCalculation struct {
	OriginalAmount   float64          `json:"original_amount"`
	TotalDiscount    float64          `json:"total_discount"`
	NetAmount        float64          `json:"net_amount"`
	AppliedDiscounts []AppliedDiscount `json:"applied_discounts"`
	Currency         string           `json:"currency"`
}

// AppliedDiscount represents a discount that was applied
type AppliedDiscount struct {
	RuleID      ulid.ULID    `json:"rule_id"`
	RuleName    string       `json:"rule_name"`
	Type        DiscountType `json:"type"`
	Value       float64      `json:"value"`
	Amount      float64      `json:"amount"`
	Description string       `json:"description"`
}

// DiscountContext provides context for discount calculation
type DiscountContext struct {
	OrganizationID  ulid.ULID  `json:"organization_id"`
	BillingTier     string     `json:"billing_tier"`
	UsageSummary    *UsageData `json:"usage_summary"`
	RequestType     *string    `json:"request_type,omitempty"`
	ProviderID      *ulid.ULID `json:"provider_id,omitempty"`
	ModelID         *ulid.ULID `json:"model_id,omitempty"`
	IsFirstCustomer bool       `json:"is_first_customer"`
	Timestamp       time.Time  `json:"timestamp"`
}

// UsageData represents usage data for discount calculation
type UsageData struct {
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	Currency      string  `json:"currency"`
}

// NewDiscountCalculator creates a new discount calculator
func NewDiscountCalculator(logger *logrus.Logger) *DiscountCalculator {
	return &DiscountCalculator{
		logger: logger,
	}
}

// CalculateDiscounts calculates applicable discounts for a billing amount
func (c *DiscountCalculator) CalculateDiscounts(
	ctx context.Context,
	amount float64,
	currency string,
	discountContext *DiscountContext,
	rules []*DiscountRule,
) (*DiscountCalculation, error) {
	
	calculation := &DiscountCalculation{
		OriginalAmount:   amount,
		TotalDiscount:    0,
		NetAmount:        amount,
		AppliedDiscounts: []AppliedDiscount{},
		Currency:         currency,
	}

	// Filter and sort applicable rules
	applicableRules := c.filterApplicableRules(rules, discountContext)
	if len(applicableRules) == 0 {
		return calculation, nil
	}

	c.logger.WithFields(logrus.Fields{
		"org_id":           discountContext.OrganizationID,
		"original_amount":  amount,
		"applicable_rules": len(applicableRules),
	}).Debug("Calculating discounts")

	// Apply discounts in priority order
	currentAmount := amount
	
	for _, rule := range applicableRules {
		discount, err := c.calculateSingleDiscount(rule, currentAmount, discountContext)
		if err != nil {
			c.logger.WithError(err).WithField("rule_id", rule.ID).Error("Failed to calculate discount")
			continue
		}

		if discount.Amount > 0 {
			// Apply maximum discount limit if set
			if rule.MaximumDiscount > 0 && discount.Amount > rule.MaximumDiscount {
				discount.Amount = rule.MaximumDiscount
			}

			calculation.AppliedDiscounts = append(calculation.AppliedDiscounts, *discount)
			calculation.TotalDiscount += discount.Amount
			currentAmount -= discount.Amount

			// Ensure we don't go negative
			if currentAmount < 0 {
				calculation.TotalDiscount += currentAmount // Reduce total discount by the negative amount
				currentAmount = 0
				break
			}
		}
	}

	calculation.NetAmount = currentAmount

	c.logger.WithFields(logrus.Fields{
		"org_id":          discountContext.OrganizationID,
		"original_amount": amount,
		"total_discount":  calculation.TotalDiscount,
		"net_amount":      calculation.NetAmount,
		"discounts_count": len(calculation.AppliedDiscounts),
	}).Debug("Discount calculation completed")

	return calculation, nil
}

// GetOrganizationDiscountRate gets the default discount rate for an organization
func (c *DiscountCalculator) GetOrganizationDiscountRate(
	ctx context.Context,
	orgID ulid.ULID,
	billingTier string,
) (float64, error) {
	// Default discount rates by billing tier
	discountRates := map[string]float64{
		"free":       0.0,  // No discount for free tier
		"pro":        0.05, // 5% discount for pro tier
		"business":   0.10, // 10% discount for business tier
		"enterprise": 0.15, // 15% discount for enterprise tier
	}

	if rate, exists := discountRates[billingTier]; exists {
		return rate, nil
	}

	return 0.0, nil // Default to no discount
}

// CreateVolumeDiscountRule creates a volume-based discount rule
func (c *DiscountCalculator) CreateVolumeDiscountRule(
	orgID *ulid.ULID,
	name string,
	description string,
	tiers []VolumeTier,
	validFrom time.Time,
	validUntil *time.Time,
) *DiscountRule {
	return &DiscountRule{
		ID:              ulid.New(),
		OrganizationID:  orgID,
		Name:            name,
		Description:     description,
		Type:            DiscountTypeTiered,
		Value:           0, // Value is determined by tiers
		MinimumAmount:   0,
		MaximumDiscount: 0, // No maximum for volume discounts
		Conditions: &DiscountCondition{
			VolumeThreshold: &VolumeDiscount{
				Tiers: tiers,
			},
		},
		ValidFrom:  validFrom,
		ValidUntil: validUntil,
		IsActive:   true,
		Priority:   100, // High priority for volume discounts
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// CreateFirstTimeCustomerDiscount creates a first-time customer discount rule
func (c *DiscountCalculator) CreateFirstTimeCustomerDiscount(
	percentage float64,
	maxDiscount float64,
	validFor time.Duration,
) *DiscountRule {
	validUntil := time.Now().Add(validFor)
	
	return &DiscountRule{
		ID:              ulid.New(),
		OrganizationID:  nil, // Global rule
		Name:            "First Time Customer Discount",
		Description:     fmt.Sprintf("%.0f%% discount for first-time customers", percentage*100),
		Type:            DiscountTypePercentage,
		Value:           percentage,
		MinimumAmount:   0,
		MaximumDiscount: maxDiscount,
		Conditions: &DiscountCondition{
			FirstTimeCustomer: true,
		},
		ValidFrom:  time.Now(),
		ValidUntil: &validUntil,
		IsActive:   true,
		Priority:   200, // High priority for first-time customers
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Internal methods

func (c *DiscountCalculator) filterApplicableRules(
	rules []*DiscountRule,
	context *DiscountContext,
) []*DiscountRule {
	var applicable []*DiscountRule
	now := context.Timestamp

	for _, rule := range rules {
		if !c.isRuleApplicable(rule, context, now) {
			continue
		}
		applicable = append(applicable, rule)
	}

	// Sort by priority (highest first)
	for i := 0; i < len(applicable)-1; i++ {
		for j := i + 1; j < len(applicable); j++ {
			if applicable[i].Priority < applicable[j].Priority {
				applicable[i], applicable[j] = applicable[j], applicable[i]
			}
		}
	}

	return applicable
}

func (c *DiscountCalculator) isRuleApplicable(
	rule *DiscountRule,
	context *DiscountContext,
	now time.Time,
) bool {
	// Check if rule is active
	if !rule.IsActive {
		return false
	}

	// Check validity period
	if now.Before(rule.ValidFrom) {
		return false
	}
	if rule.ValidUntil != nil && now.After(*rule.ValidUntil) {
		return false
	}

	// Check usage limit
	if rule.UsageLimit != nil && rule.UsageCount >= *rule.UsageLimit {
		return false
	}

	// Check organization-specific rule
	if rule.OrganizationID != nil && *rule.OrganizationID != context.OrganizationID {
		return false
	}

	// Check conditions if they exist
	if rule.Conditions != nil {
		return c.checkDiscountConditions(rule.Conditions, context, now)
	}

	return true
}

func (c *DiscountCalculator) checkDiscountConditions(
	conditions *DiscountCondition,
	context *DiscountContext,
	now time.Time,
) bool {
	// Check billing tier
	if len(conditions.BillingTiers) > 0 {
		found := false
		for _, tier := range conditions.BillingTiers {
			if tier == context.BillingTier {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check minimum usage
	if conditions.MinUsage != nil && context.UsageSummary != nil {
		usage := context.UsageSummary
		if usage.TotalRequests < conditions.MinUsage.Requests ||
			usage.TotalTokens < conditions.MinUsage.Tokens ||
			usage.TotalCost < conditions.MinUsage.Cost {
			return false
		}
	}

	// Check request types
	if len(conditions.RequestTypes) > 0 && context.RequestType != nil {
		found := false
		for _, reqType := range conditions.RequestTypes {
			if reqType == *context.RequestType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check providers
	if len(conditions.Providers) > 0 && context.ProviderID != nil {
		found := false
		for _, providerID := range conditions.Providers {
			if providerID == *context.ProviderID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check models
	if len(conditions.Models) > 0 && context.ModelID != nil {
		found := false
		for _, modelID := range conditions.Models {
			if modelID == *context.ModelID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check first-time customer
	if conditions.FirstTimeCustomer && !context.IsFirstCustomer {
		return false
	}

	// Check time of day
	if conditions.TimeOfDay != nil {
		// Simple time range check (ignoring timezone complexities for now)
		currentTime := now.Format("15:04")
		startTime := conditions.TimeOfDay.Start.Format("15:04")
		endTime := conditions.TimeOfDay.End.Format("15:04")
		
		if currentTime < startTime || currentTime > endTime {
			return false
		}
	}

	// Check days of week
	if len(conditions.DaysOfWeek) > 0 {
		currentDay := now.Weekday()
		found := false
		for _, day := range conditions.DaysOfWeek {
			if day == currentDay {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

func (c *DiscountCalculator) calculateSingleDiscount(
	rule *DiscountRule,
	amount float64,
	context *DiscountContext,
) (*AppliedDiscount, error) {
	// Check minimum amount requirement
	if amount < rule.MinimumAmount {
		return &AppliedDiscount{Amount: 0}, nil
	}

	var discountAmount float64
	var description string

	switch rule.Type {
	case DiscountTypePercentage:
		discountAmount = amount * rule.Value
		description = fmt.Sprintf("%.1f%% discount", rule.Value*100)

	case DiscountTypeFixed:
		discountAmount = rule.Value
		description = fmt.Sprintf("$%.2f fixed discount", rule.Value)

	case DiscountTypeTiered:
		if rule.Conditions != nil && rule.Conditions.VolumeThreshold != nil {
			discountAmount = c.calculateVolumeDiscount(amount, rule.Conditions.VolumeThreshold)
			description = "Volume-based discount"
		}

	default:
		return nil, fmt.Errorf("unsupported discount type: %s", rule.Type)
	}

	// Ensure discount doesn't exceed the original amount
	if discountAmount > amount {
		discountAmount = amount
	}

	return &AppliedDiscount{
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Type:        rule.Type,
		Value:       rule.Value,
		Amount:      discountAmount,
		Description: description,
	}, nil
}

func (c *DiscountCalculator) calculateVolumeDiscount(amount float64, volumeDiscount *VolumeDiscount) float64 {
	var totalDiscount float64
	remainingAmount := amount

	// Sort tiers by minimum amount (ascending)
	tiers := make([]VolumeTier, len(volumeDiscount.Tiers))
	copy(tiers, volumeDiscount.Tiers)
	
	for i := 0; i < len(tiers)-1; i++ {
		for j := i + 1; j < len(tiers); j++ {
			if tiers[i].MinAmount > tiers[j].MinAmount {
				tiers[i], tiers[j] = tiers[j], tiers[i]
			}
		}
	}

	// Apply tiered discounts
	for i, tier := range tiers {
		if remainingAmount <= 0 {
			break
		}

		tierAmount := remainingAmount
		if i < len(tiers)-1 {
			// Not the last tier, calculate amount in this tier
			nextTierMin := tiers[i+1].MinAmount
			if amount > nextTierMin {
				tierAmount = nextTierMin - tier.MinAmount
				if tierAmount > remainingAmount {
					tierAmount = remainingAmount
				}
			}
		}

		if tierAmount > 0 {
			totalDiscount += tierAmount * tier.Discount
			remainingAmount -= tierAmount
		}
	}

	return totalDiscount
}

// Health check
func (c *DiscountCalculator) GetHealth() map[string]interface{} {
	return map[string]interface{}{
		"service": "discount_calculator",
		"status":  "healthy",
	}
}