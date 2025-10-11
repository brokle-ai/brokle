package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"brokle/internal/services/billing"
	"brokle/internal/workers/analytics"
	"brokle/pkg/ulid"
)

// Repository implements the billing repository using PostgreSQL
type Repository struct {
	db     *gorm.DB
	logger *logrus.Logger
}

// NewRepository creates a new billing repository instance
func NewRepository(db *gorm.DB, logger *logrus.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

// Usage record operations

func (r *Repository) InsertUsageRecord(ctx context.Context, record *billing.UsageRecord) error {
	query := `
		INSERT INTO usage_records (
			id, organization_id, request_id, provider_id, model_id,
			request_type, input_tokens, output_tokens, total_tokens,
			cost, currency, billing_tier, discounts, net_cost,
			created_at, processed_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?, ?,
			?, ?
		)`

	err := r.db.WithContext(ctx).Exec(query,
		record.ID,
		record.OrganizationID,
		record.RequestID,
		record.ProviderID,
		record.ModelID,
		record.RequestType,
		record.InputTokens,
		record.OutputTokens,
		record.TotalTokens,
		record.Cost,
		record.Currency,
		record.BillingTier,
		record.Discounts,
		record.NetCost,
		record.CreatedAt,
		record.ProcessedAt,
	).Error

	if err != nil {
		r.logger.WithError(err).WithField("record_id", record.ID).Error("Failed to insert usage record")
		return fmt.Errorf("failed to insert usage record: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"record_id":       record.ID,
		"organization_id": record.OrganizationID,
		"net_cost":        record.NetCost,
	}).Debug("Inserted usage record")

	return nil
}

func (r *Repository) GetUsageRecords(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*billing.UsageRecord, error) {
	query := `
		SELECT
			id, organization_id, request_id, provider_id, model_id,
			request_type, input_tokens, output_tokens, total_tokens,
			cost, currency, billing_tier, discounts, net_cost,
			created_at, processed_at
		FROM usage_records
		WHERE organization_id = ?
			AND created_at >= ?
			AND created_at < ?
		ORDER BY created_at DESC`

	var records []*billing.UsageRecord
	err := r.db.WithContext(ctx).Raw(query, orgID, start, end).Scan(&records).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}

	return records, nil
}

func (r *Repository) UpdateUsageRecord(ctx context.Context, recordID ulid.ULID, record *billing.UsageRecord) error {
	query := `
		UPDATE usage_records
		SET
			request_type = ?,
			input_tokens = ?,
			output_tokens = ?,
			total_tokens = ?,
			cost = ?,
			currency = ?,
			billing_tier = ?,
			discounts = ?,
			net_cost = ?,
			processed_at = ?
		WHERE id = ?`

	result := r.db.WithContext(ctx).Exec(query,
		record.RequestType,
		record.InputTokens,
		record.OutputTokens,
		record.TotalTokens,
		record.Cost,
		record.Currency,
		record.BillingTier,
		record.Discounts,
		record.NetCost,
		record.ProcessedAt,
		recordID,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update usage record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("usage record not found: %s", recordID)
	}

	return nil
}

// Billing record operations

func (r *Repository) InsertBillingRecord(ctx context.Context, record *analytics.BillingRecord) error {
	query := `
		INSERT INTO billing_records (
			id, organization_id, period, amount, currency,
			status, transaction_id, payment_method, created_at, processed_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?, ?
		)`

	err := r.db.WithContext(ctx).Exec(query,
		record.ID,
		record.OrganizationID,
		record.Period,
		record.Amount,
		record.Currency,
		record.Status,
		record.TransactionID,
		record.PaymentMethod,
		record.CreatedAt,
		record.ProcessedAt,
	).Error

	if err != nil {
		r.logger.WithError(err).WithField("record_id", record.ID).Error("Failed to insert billing record")
		return fmt.Errorf("failed to insert billing record: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"record_id":       record.ID,
		"organization_id": record.OrganizationID,
		"amount":          record.Amount,
		"period":          record.Period,
	}).Debug("Inserted billing record")

	return nil
}

func (r *Repository) UpdateBillingRecord(ctx context.Context, recordID ulid.ULID, record *analytics.BillingRecord) error {
	query := `
		UPDATE billing_records
		SET
			period = ?,
			amount = ?,
			currency = ?,
			status = ?,
			transaction_id = ?,
			payment_method = ?,
			processed_at = ?
		WHERE id = ?`

	result := r.db.WithContext(ctx).Exec(query,
		record.Period,
		record.Amount,
		record.Currency,
		record.Status,
		record.TransactionID,
		record.PaymentMethod,
		record.ProcessedAt,
		recordID,
	)

	if result.Error != nil {
		return fmt.Errorf("failed to update billing record: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("billing record not found: %s", recordID)
	}

	return nil
}

func (r *Repository) GetBillingRecord(ctx context.Context, recordID ulid.ULID) (*analytics.BillingRecord, error) {
	query := `
		SELECT
			id, organization_id, period, amount, currency,
			status, transaction_id, payment_method, created_at, processed_at
		FROM billing_records
		WHERE id = ?`

	record := &analytics.BillingRecord{}
	err := r.db.WithContext(ctx).Raw(query, recordID).Scan(record).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("billing record not found: %s", recordID)
		}
		return nil, fmt.Errorf("failed to get billing record: %w", err)
	}

	return record, nil
}

func (r *Repository) GetBillingHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*analytics.BillingRecord, error) {
	query := `
		SELECT
			id, organization_id, period, amount, currency,
			status, transaction_id, payment_method, created_at, processed_at
		FROM billing_records
		WHERE organization_id = ?
			AND created_at >= ?
			AND created_at < ?
		ORDER BY created_at DESC`

	var records []*analytics.BillingRecord
	err := r.db.WithContext(ctx).Raw(query, orgID, start, end).Scan(&records).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get billing history: %w", err)
	}

	return records, nil
}

// Billing summary operations

func (r *Repository) InsertBillingSummary(ctx context.Context, summary *analytics.BillingSummary) error {
	providerBreakdownJSON, err := json.Marshal(summary.ProviderBreakdown)
	if err != nil {
		return fmt.Errorf("failed to marshal provider breakdown: %w", err)
	}

	modelBreakdownJSON, err := json.Marshal(summary.ModelBreakdown)
	if err != nil {
		return fmt.Errorf("failed to marshal model breakdown: %w", err)
	}

	query := `
		INSERT INTO billing_summaries (
			id, organization_id, period, period_start, period_end,
			total_requests, total_tokens, total_cost, currency,
			provider_breakdown, model_breakdown, discounts, net_cost,
			status, generated_at
		) VALUES (
			?, ?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?
		)
		ON CONFLICT (organization_id, period, period_start)
		DO UPDATE SET
			total_requests = EXCLUDED.total_requests,
			total_tokens = EXCLUDED.total_tokens,
			total_cost = EXCLUDED.total_cost,
			provider_breakdown = EXCLUDED.provider_breakdown,
			model_breakdown = EXCLUDED.model_breakdown,
			discounts = EXCLUDED.discounts,
			net_cost = EXCLUDED.net_cost,
			status = EXCLUDED.status,
			generated_at = EXCLUDED.generated_at`

	// Generate ID if not provided
	if summary.ID.IsZero() {
		summary.ID = ulid.New()
	}

	err = r.db.WithContext(ctx).Exec(query,
		summary.ID,
		summary.OrganizationID,
		summary.Period,
		summary.PeriodStart,
		summary.PeriodEnd,
		summary.TotalRequests,
		summary.TotalTokens,
		summary.TotalCost,
		summary.Currency,
		providerBreakdownJSON,
		modelBreakdownJSON,
		summary.Discounts,
		summary.NetCost,
		summary.Status,
		summary.GeneratedAt,
	).Error

	if err != nil {
		r.logger.WithError(err).WithField("summary_id", summary.ID).Error("Failed to insert billing summary")
		return fmt.Errorf("failed to insert billing summary: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"summary_id":      summary.ID,
		"organization_id": summary.OrganizationID,
		"period":          summary.Period,
		"net_cost":        summary.NetCost,
	}).Debug("Inserted billing summary")

	return nil
}

func (r *Repository) GetBillingSummary(ctx context.Context, orgID ulid.ULID, period string) (*analytics.BillingSummary, error) {
	query := `
		SELECT
			id, organization_id, period, period_start, period_end,
			total_requests, total_tokens, total_cost, currency,
			provider_breakdown, model_breakdown, discounts, net_cost,
			status, generated_at
		FROM billing_summaries
		WHERE organization_id = ? AND period = ?
		ORDER BY period_start DESC
		LIMIT 1`

	type BillingSummaryRow struct {
		ID                   ulid.ULID
		OrganizationID       ulid.ULID
		Period               string
		PeriodStart          time.Time
		PeriodEnd            time.Time
		TotalRequests        int64
		TotalTokens          int64
		TotalCost            float64
		Currency             string
		ProviderBreakdownRaw []byte `gorm:"column:provider_breakdown"`
		ModelBreakdownRaw    []byte `gorm:"column:model_breakdown"`
		Discounts            float64
		NetCost              float64
		Status               string
		GeneratedAt          time.Time
	}

	var row BillingSummaryRow
	err := r.db.WithContext(ctx).Raw(query, orgID, period).Scan(&row).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound || err == sql.ErrNoRows {
			return nil, fmt.Errorf("billing summary not found for organization %s and period %s", orgID, period)
		}
		return nil, fmt.Errorf("failed to get billing summary: %w", err)
	}

	// Check if we got empty result
	if row.ID.IsZero() {
		return nil, fmt.Errorf("billing summary not found for organization %s and period %s", orgID, period)
	}

	summary := &analytics.BillingSummary{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		Period:         row.Period,
		PeriodStart:    row.PeriodStart,
		PeriodEnd:      row.PeriodEnd,
		TotalRequests:  row.TotalRequests,
		TotalTokens:    row.TotalTokens,
		TotalCost:      row.TotalCost,
		Currency:       row.Currency,
		Discounts:      row.Discounts,
		NetCost:        row.NetCost,
		Status:         row.Status,
		GeneratedAt:    row.GeneratedAt,
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(row.ProviderBreakdownRaw, &summary.ProviderBreakdown); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider breakdown: %w", err)
	}

	if err := json.Unmarshal(row.ModelBreakdownRaw, &summary.ModelBreakdown); err != nil {
		return nil, fmt.Errorf("failed to unmarshal model breakdown: %w", err)
	}

	return summary, nil
}

func (r *Repository) GetBillingSummaryHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*analytics.BillingSummary, error) {
	query := `
		SELECT
			id, organization_id, period, period_start, period_end,
			total_requests, total_tokens, total_cost, currency,
			provider_breakdown, model_breakdown, discounts, net_cost,
			status, generated_at
		FROM billing_summaries
		WHERE organization_id = ?
			AND period_start >= ?
			AND period_start < ?
		ORDER BY period_start DESC`

	type BillingSummaryRow struct {
		ID                   ulid.ULID
		OrganizationID       ulid.ULID
		Period               string
		PeriodStart          time.Time
		PeriodEnd            time.Time
		TotalRequests        int64
		TotalTokens          int64
		TotalCost            float64
		Currency             string
		ProviderBreakdownRaw []byte `gorm:"column:provider_breakdown"`
		ModelBreakdownRaw    []byte `gorm:"column:model_breakdown"`
		Discounts            float64
		NetCost              float64
		Status               string
		GeneratedAt          time.Time
	}

	var rows []BillingSummaryRow
	err := r.db.WithContext(ctx).Raw(query, orgID, start, end).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get billing summary history: %w", err)
	}

	var summaries []*analytics.BillingSummary
	for _, row := range rows {
		summary := &analytics.BillingSummary{
			ID:             row.ID,
			OrganizationID: row.OrganizationID,
			Period:         row.Period,
			PeriodStart:    row.PeriodStart,
			PeriodEnd:      row.PeriodEnd,
			TotalRequests:  row.TotalRequests,
			TotalTokens:    row.TotalTokens,
			TotalCost:      row.TotalCost,
			Currency:       row.Currency,
			Discounts:      row.Discounts,
			NetCost:        row.NetCost,
			Status:         row.Status,
			GeneratedAt:    row.GeneratedAt,
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(row.ProviderBreakdownRaw, &summary.ProviderBreakdown); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider breakdown: %w", err)
		}

		if err := json.Unmarshal(row.ModelBreakdownRaw, &summary.ModelBreakdown); err != nil {
			return nil, fmt.Errorf("failed to unmarshal model breakdown: %w", err)
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Usage quota operations

func (r *Repository) GetUsageQuota(ctx context.Context, orgID ulid.ULID) (*billing.UsageQuota, error) {
	query := `
		SELECT
			organization_id, billing_tier, monthly_request_limit, monthly_token_limit,
			monthly_cost_limit, current_requests, current_tokens, current_cost,
			currency, reset_date, last_updated
		FROM usage_quotas
		WHERE organization_id = ?`

	quota := &billing.UsageQuota{}
	err := r.db.WithContext(ctx).Raw(query, orgID).Scan(quota).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound || err == sql.ErrNoRows {
			return nil, nil // No quota found, return nil without error
		}
		return nil, fmt.Errorf("failed to get usage quota: %w", err)
	}

	// Check if we got empty result
	if quota.OrganizationID.IsZero() {
		return nil, nil
	}

	return quota, nil
}

func (r *Repository) UpdateUsageQuota(ctx context.Context, orgID ulid.ULID, quota *billing.UsageQuota) error {
	query := `
		INSERT INTO usage_quotas (
			organization_id, billing_tier, monthly_request_limit, monthly_token_limit,
			monthly_cost_limit, current_requests, current_tokens, current_cost,
			currency, reset_date, last_updated
		) VALUES (
			?, ?, ?, ?,
			?, ?, ?, ?,
			?, ?, ?
		)
		ON CONFLICT (organization_id)
		DO UPDATE SET
			billing_tier = EXCLUDED.billing_tier,
			monthly_request_limit = EXCLUDED.monthly_request_limit,
			monthly_token_limit = EXCLUDED.monthly_token_limit,
			monthly_cost_limit = EXCLUDED.monthly_cost_limit,
			current_requests = EXCLUDED.current_requests,
			current_tokens = EXCLUDED.current_tokens,
			current_cost = EXCLUDED.current_cost,
			currency = EXCLUDED.currency,
			reset_date = EXCLUDED.reset_date,
			last_updated = EXCLUDED.last_updated`

	err := r.db.WithContext(ctx).Exec(query,
		quota.OrganizationID,
		quota.BillingTier,
		quota.MonthlyRequestLimit,
		quota.MonthlyTokenLimit,
		quota.MonthlyCostLimit,
		quota.CurrentRequests,
		quota.CurrentTokens,
		quota.CurrentCost,
		quota.Currency,
		quota.ResetDate,
		quota.LastUpdated,
	).Error

	if err != nil {
		r.logger.WithError(err).WithField("org_id", orgID).Error("Failed to update usage quota")
		return fmt.Errorf("failed to update usage quota: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"org_id":        orgID,
		"billing_tier":  quota.BillingTier,
		"request_limit": quota.MonthlyRequestLimit,
		"cost_limit":    quota.MonthlyCostLimit,
	}).Debug("Updated usage quota")

	return nil
}

// Health check
func (r *Repository) GetHealth(ctx context.Context) error {
	var result int

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := r.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error
	if err != nil {
		return fmt.Errorf("billing repository health check failed: %w", err)
	}

	return nil
}
