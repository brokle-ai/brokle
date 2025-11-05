package billing

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	billingDomain "brokle/internal/core/domain/billing"
	"brokle/internal/workers/analytics"
	"brokle/pkg/ulid"
)

// BillingService implements billing operations for gateway usage
type BillingService struct {
	logger            *logrus.Logger
	usageRepo         billingDomain.UsageRepository
	billingRecordRepo billingDomain.BillingRecordRepository
	quotaRepo         billingDomain.QuotaRepository
	orgService        billingDomain.OrganizationService
	usageTracker      *UsageTracker
	discountCalculator *DiscountCalculator
	invoiceGenerator   *InvoiceGenerator
}

// BillingConfig holds billing service configuration
type BillingConfig struct {
	DefaultCurrency      string
	BillingPeriod       string // monthly, quarterly, annually
	PaymentGracePeriod  time.Duration
	OverageChargeRate   float64
	EnableAutoBilling   bool
	InvoiceGeneration   bool
}

// DefaultBillingConfig returns default billing configuration
func DefaultBillingConfig() *BillingConfig {
	return &BillingConfig{
		DefaultCurrency:     "USD",
		BillingPeriod:      "monthly",
		PaymentGracePeriod: 7 * 24 * time.Hour, // 7 days
		OverageChargeRate:  1.25, // 25% markup for overage
		EnableAutoBilling:  true,
		InvoiceGeneration:  true,
	}
}

// NewBillingService creates a new billing service instance
func NewBillingService(
	logger *logrus.Logger,
	config *BillingConfig,
	usageRepo billingDomain.UsageRepository,
	billingRecordRepo billingDomain.BillingRecordRepository,
	quotaRepo billingDomain.QuotaRepository,
	orgService billingDomain.OrganizationService,
) *BillingService {
	if config == nil {
		config = DefaultBillingConfig()
	}

	return &BillingService{
		logger:            logger,
		usageRepo:         usageRepo,
		billingRecordRepo: billingRecordRepo,
		quotaRepo:         quotaRepo,
		orgService:        orgService,
		usageTracker:      NewUsageTracker(logger, usageRepo, quotaRepo),
		discountCalculator: NewDiscountCalculator(logger),
		invoiceGenerator:   NewInvoiceGenerator(logger, config),
	}
}

// RecordUsage records usage for billing
func (s *BillingService) RecordUsage(ctx context.Context, usage *analytics.CostMetric) error {
	// Get organization billing tier
	billingTier, err := s.orgService.GetBillingTier(ctx, usage.OrganizationID)
	if err != nil {
		s.logger.WithError(err).WithField("org_id", usage.OrganizationID).Error("Failed to get billing tier")
		billingTier = "free" // Default fallback
	}

	// Calculate discounts
	discountRate, err := s.orgService.GetDiscountRate(ctx, usage.OrganizationID)
	if err != nil {
		s.logger.WithError(err).WithField("org_id", usage.OrganizationID).Error("Failed to get discount rate")
		discountRate = 0.0 // No discount on error
	}

	discountAmount := usage.TotalCost * discountRate
	netCost := usage.TotalCost - discountAmount

	// Create usage record
	record := &billingDomain.UsageRecord{
		ID:               ulid.New(),
		OrganizationID:   usage.OrganizationID,
		RequestID:        usage.RequestID,
		ProviderID:       usage.ProviderID,
		ModelID:          usage.ModelID,
		RequestType:      string(usage.RequestType),
		InputTokens:      usage.InputTokens,
		OutputTokens:     usage.OutputTokens,
		TotalTokens:      usage.TotalTokens,
		Cost:             usage.TotalCost,
		Currency:         usage.Currency,
		BillingTier:      billingTier,
		Discounts:        discountAmount,
		NetCost:          netCost,
		CreatedAt:        time.Now(),
	}

	// Store usage record
	if err := s.usageRepo.InsertUsageRecord(ctx, record); err != nil {
		s.logger.WithError(err).WithField("record_id", record.ID).Error("Failed to insert usage record")
		return fmt.Errorf("failed to record usage: %w", err)
	}

	// Update usage tracking
	if err := s.usageTracker.UpdateUsage(ctx, usage.OrganizationID, record); err != nil {
		s.logger.WithError(err).WithField("org_id", usage.OrganizationID).Error("Failed to update usage tracking")
		// Don't fail the entire operation for tracking errors
	}

	s.logger.WithFields(logrus.Fields{
		"org_id":     usage.OrganizationID,
		"request_id": usage.RequestID,
		"cost":       usage.TotalCost,
		"net_cost":   netCost,
	}).Debug("Recorded usage for billing")

	return nil
}

// CalculateBill generates a billing summary for an organization
func (s *BillingService) CalculateBill(ctx context.Context, orgID ulid.ULID, period string) (*analytics.BillingSummary, error) {
	// Calculate period start and end dates
	start, end := s.calculatePeriodBounds(period)

	// Get usage records for the period
	usageRecords, err := s.usageRepo.GetUsageRecords(ctx, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	if len(usageRecords) == 0 {
		return &analytics.BillingSummary{
			ID:             ulid.New(),
			OrganizationID: orgID,
			Period:         period,
			PeriodStart:    start,
			PeriodEnd:      end,
			TotalRequests:  0,
			TotalTokens:    0,
			TotalCost:      0,
			Currency:       "USD",
			NetCost:        0,
			Status:         "no_usage",
			GeneratedAt:    time.Now(),
		}, nil
	}

	// Calculate summary statistics
	summary := &analytics.BillingSummary{
		ID:                ulid.New(),
		OrganizationID:    orgID,
		Period:            period,
		PeriodStart:       start,
		PeriodEnd:         end,
		Currency:          usageRecords[0].Currency,
		ProviderBreakdown: make(map[string]float64),
		ModelBreakdown:    make(map[string]float64),
		GeneratedAt:       time.Now(),
	}

	var totalRequests int64
	var totalTokens int64
	var totalCost float64
	var totalDiscounts float64
	var totalNetCost float64

	for _, record := range usageRecords {
		totalRequests++
		totalTokens += int64(record.TotalTokens)
		totalCost += record.Cost
		totalDiscounts += record.Discounts
		totalNetCost += record.NetCost

		// Provider breakdown
		providerKey := record.ProviderID.String() // Could be enhanced with provider name
		summary.ProviderBreakdown[providerKey] += record.NetCost

		// Model breakdown
		modelKey := record.ModelID.String() // Could be enhanced with model name
		summary.ModelBreakdown[modelKey] += record.NetCost
	}

	summary.TotalRequests = totalRequests
	summary.TotalTokens = totalTokens
	summary.TotalCost = totalCost
	summary.Discounts = totalDiscounts
	summary.NetCost = totalNetCost

	// Determine billing status
	if totalNetCost > 0 {
		summary.Status = "pending"
	} else {
		summary.Status = "no_charge"
	}

	// Store the billing summary
	if err := s.billingRecordRepo.InsertBillingSummary(ctx, summary); err != nil {
		s.logger.WithError(err).WithField("org_id", orgID).Error("Failed to store billing summary")
		// Continue without failing - the summary is still valid
	}

	return summary, nil
}

// GetBillingHistory retrieves billing history for an organization
func (s *BillingService) GetBillingHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*analytics.BillingRecord, error) {
	return s.billingRecordRepo.GetBillingHistory(ctx, orgID, start, end)
}

// ProcessPayment processes a payment for a billing record
func (s *BillingService) ProcessPayment(ctx context.Context, billingRecordID ulid.ULID) error {
	// Get billing record
	record, err := s.billingRecordRepo.GetBillingRecord(ctx, billingRecordID)
	if err != nil {
		return fmt.Errorf("failed to get billing record: %w", err)
	}

	if record.Status == "paid" {
		return fmt.Errorf("billing record %s is already paid", billingRecordID)
	}

	// Get payment method
	paymentMethod, err := s.orgService.GetPaymentMethod(ctx, record.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to get payment method: %w", err)
	}

	if paymentMethod == nil {
		return fmt.Errorf("no payment method found for organization %s", record.OrganizationID)
	}

	// TODO: Integrate with payment processor (Stripe, etc.)
	// This is a placeholder for actual payment processing
	transactionID := fmt.Sprintf("txn_%s", ulid.New())
	
	// Update billing record with payment information
	now := time.Now()
	record.Status = "paid"
	record.TransactionID = &transactionID
	record.PaymentMethod = &paymentMethod.Type
	record.ProcessedAt = &now

	if err := s.billingRecordRepo.UpdateBillingRecord(ctx, billingRecordID, record); err != nil {
		return fmt.Errorf("failed to update billing record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"billing_record_id": billingRecordID,
		"organization_id":   record.OrganizationID,
		"amount":           record.Amount,
		"transaction_id":   transactionID,
	}).Info("Payment processed successfully")

	return nil
}

// CheckUsageQuotas checks if organization is within usage quotas
func (s *BillingService) CheckUsageQuotas(ctx context.Context, orgID ulid.ULID) (*billingDomain.QuotaStatus, error) {
	quota, err := s.quotaRepo.GetUsageQuota(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage quota: %w", err)
	}

	if quota == nil {
		// No quota set, assume unlimited for now
		return &billingDomain.QuotaStatus{
			OrganizationID: orgID,
			RequestsOK:     true,
			TokensOK:       true,
			CostOK:         true,
			Status:         "unlimited",
		}, nil
	}

	status := &billingDomain.QuotaStatus{
		OrganizationID: orgID,
		RequestsOK:     quota.MonthlyRequestLimit == 0 || quota.CurrentRequests < quota.MonthlyRequestLimit,
		TokensOK:       quota.MonthlyTokenLimit == 0 || quota.CurrentTokens < quota.MonthlyTokenLimit,
		CostOK:         quota.MonthlyCostLimit == 0 || quota.CurrentCost < quota.MonthlyCostLimit,
	}

	if status.RequestsOK && status.TokensOK && status.CostOK {
		status.Status = "within_limits"
	} else if quota.CurrentRequests >= quota.MonthlyRequestLimit {
		status.Status = "requests_exceeded"
	} else if quota.CurrentTokens >= quota.MonthlyTokenLimit {
		status.Status = "tokens_exceeded"
	} else if quota.CurrentCost >= quota.MonthlyCostLimit {
		status.Status = "cost_exceeded"
	}

	// Calculate usage percentages
	if quota.MonthlyRequestLimit > 0 {
		status.RequestsUsagePercent = float64(quota.CurrentRequests) / float64(quota.MonthlyRequestLimit) * 100
	}
	if quota.MonthlyTokenLimit > 0 {
		status.TokensUsagePercent = float64(quota.CurrentTokens) / float64(quota.MonthlyTokenLimit) * 100
	}
	if quota.MonthlyCostLimit > 0 {
		status.CostUsagePercent = quota.CurrentCost / quota.MonthlyCostLimit * 100
	}

	return status, nil
}

// QuotaStatus represents the current quota status for an organization

// CreateBillingRecord creates a new billing record for an organization
func (s *BillingService) CreateBillingRecord(ctx context.Context, summary *analytics.BillingSummary) (*analytics.BillingRecord, error) {
	if summary.NetCost <= 0 {
		return nil, fmt.Errorf("no charges to bill for organization %s", summary.OrganizationID)
	}

	record := &analytics.BillingRecord{
		ID:               ulid.New(),
		OrganizationID:   summary.OrganizationID,
		Period:           summary.Period,
		Amount:           summary.NetCost,
		Currency:         summary.Currency,
		Status:           "pending",
		CreatedAt:        time.Now(),
	}

	if err := s.billingRecordRepo.InsertBillingRecord(ctx, record); err != nil {
		return nil, fmt.Errorf("failed to create billing record: %w", err)
	}

	s.logger.WithFields(logrus.Fields{
		"billing_record_id": record.ID,
		"organization_id":   record.OrganizationID,
		"amount":           record.Amount,
		"period":           record.Period,
	}).Info("Created billing record")

	return record, nil
}

// Helper methods

func (s *BillingService) calculatePeriodBounds(period string) (start, end time.Time) {
	now := time.Now()
	
	switch period {
	case "daily":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		end = start.Add(24 * time.Hour)
	case "weekly":
		// Start of week (Sunday)
		weekday := int(now.Weekday())
		start = now.AddDate(0, 0, -weekday)
		start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
		end = start.Add(7 * 24 * time.Hour)
	case "monthly":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	case "yearly":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(1, 0, 0)
	default:
		// Default to current month
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		end = start.AddDate(0, 1, 0)
	}
	
	return start, end
}

// Health check
func (s *BillingService) GetHealth() map[string]interface{} {
	return map[string]interface{}{
		"service":         "billing",
		"status":          "healthy",
		"usage_tracker":   s.usageTracker.GetHealth(),
		"invoice_generator": s.invoiceGenerator.GetHealth(),
	}
}