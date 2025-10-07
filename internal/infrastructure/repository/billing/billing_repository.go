package billing

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"brokle/internal/services/billing"
	"brokle/internal/workers/analytics"
	"brokle/pkg/ulid"
)

// Repository implements the billing repository using PostgreSQL
type Repository struct {
	db     *sqlx.DB
	logger *logrus.Logger
}

// NewRepository creates a new billing repository instance
func NewRepository(db *sqlx.DB, logger *logrus.Logger) *Repository {
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
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13, $14,
			$15, $16
		)`

	_, err := r.db.ExecContext(ctx, query,
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
	)

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
		WHERE organization_id = $1 
			AND created_at >= $2 
			AND created_at < $3
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage records: %w", err)
	}
	defer rows.Close()

	var records []*billing.UsageRecord
	for rows.Next() {
		record := &billing.UsageRecord{}
		err := rows.Scan(
			&record.ID,
			&record.OrganizationID,
			&record.RequestID,
			&record.ProviderID,
			&record.ModelID,
			&record.RequestType,
			&record.InputTokens,
			&record.OutputTokens,
			&record.TotalTokens,
			&record.Cost,
			&record.Currency,
			&record.BillingTier,
			&record.Discounts,
			&record.NetCost,
			&record.CreatedAt,
			&record.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan usage record: %w", err)
		}
		records = append(records, record)
	}

	return records, rows.Err()
}

func (r *Repository) UpdateUsageRecord(ctx context.Context, recordID ulid.ULID, record *billing.UsageRecord) error {
	query := `
		UPDATE usage_records 
		SET 
			request_type = $2,
			input_tokens = $3,
			output_tokens = $4,
			total_tokens = $5,
			cost = $6,
			currency = $7,
			billing_tier = $8,
			discounts = $9,
			net_cost = $10,
			processed_at = $11
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		recordID,
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
	)

	if err != nil {
		return fmt.Errorf("failed to update usage record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
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
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10
		)`

	_, err := r.db.ExecContext(ctx, query,
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
	)

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
			period = $2,
			amount = $3,
			currency = $4,
			status = $5,
			transaction_id = $6,
			payment_method = $7,
			processed_at = $8
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		recordID,
		record.Period,
		record.Amount,
		record.Currency,
		record.Status,
		record.TransactionID,
		record.PaymentMethod,
		record.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update billing record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
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
		WHERE id = $1`

	record := &analytics.BillingRecord{}
	err := r.db.QueryRowContext(ctx, query, recordID).Scan(
		&record.ID,
		&record.OrganizationID,
		&record.Period,
		&record.Amount,
		&record.Currency,
		&record.Status,
		&record.TransactionID,
		&record.PaymentMethod,
		&record.CreatedAt,
		&record.ProcessedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
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
		WHERE organization_id = $1 
			AND created_at >= $2 
			AND created_at < $3
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing history: %w", err)
	}
	defer rows.Close()

	var records []*analytics.BillingRecord
	for rows.Next() {
		record := &analytics.BillingRecord{}
		err := rows.Scan(
			&record.ID,
			&record.OrganizationID,
			&record.Period,
			&record.Amount,
			&record.Currency,
			&record.Status,
			&record.TransactionID,
			&record.PaymentMethod,
			&record.CreatedAt,
			&record.ProcessedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan billing record: %w", err)
		}
		records = append(records, record)
	}

	return records, rows.Err()
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
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12, $13,
			$14, $15
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

	_, err = r.db.ExecContext(ctx, query,
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
	)

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
		WHERE organization_id = $1 AND period = $2
		ORDER BY period_start DESC
		LIMIT 1`

	var providerBreakdownJSON, modelBreakdownJSON []byte
	summary := &analytics.BillingSummary{}

	err := r.db.QueryRowContext(ctx, query, orgID, period).Scan(
		&summary.ID,
		&summary.OrganizationID,
		&summary.Period,
		&summary.PeriodStart,
		&summary.PeriodEnd,
		&summary.TotalRequests,
		&summary.TotalTokens,
		&summary.TotalCost,
		&summary.Currency,
		&providerBreakdownJSON,
		&modelBreakdownJSON,
		&summary.Discounts,
		&summary.NetCost,
		&summary.Status,
		&summary.GeneratedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("billing summary not found for organization %s and period %s", orgID, period)
		}
		return nil, fmt.Errorf("failed to get billing summary: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(providerBreakdownJSON, &summary.ProviderBreakdown); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider breakdown: %w", err)
	}

	if err := json.Unmarshal(modelBreakdownJSON, &summary.ModelBreakdown); err != nil {
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
		WHERE organization_id = $1 
			AND period_start >= $2 
			AND period_start < $3
		ORDER BY period_start DESC`

	rows, err := r.db.QueryContext(ctx, query, orgID, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to get billing summary history: %w", err)
	}
	defer rows.Close()

	var summaries []*analytics.BillingSummary
	for rows.Next() {
		var providerBreakdownJSON, modelBreakdownJSON []byte
		summary := &analytics.BillingSummary{}

		err := rows.Scan(
			&summary.ID,
			&summary.OrganizationID,
			&summary.Period,
			&summary.PeriodStart,
			&summary.PeriodEnd,
			&summary.TotalRequests,
			&summary.TotalTokens,
			&summary.TotalCost,
			&summary.Currency,
			&providerBreakdownJSON,
			&modelBreakdownJSON,
			&summary.Discounts,
			&summary.NetCost,
			&summary.Status,
			&summary.GeneratedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan billing summary: %w", err)
		}

		// Unmarshal JSON fields
		if err := json.Unmarshal(providerBreakdownJSON, &summary.ProviderBreakdown); err != nil {
			return nil, fmt.Errorf("failed to unmarshal provider breakdown: %w", err)
		}

		if err := json.Unmarshal(modelBreakdownJSON, &summary.ModelBreakdown); err != nil {
			return nil, fmt.Errorf("failed to unmarshal model breakdown: %w", err)
		}

		summaries = append(summaries, summary)
	}

	return summaries, rows.Err()
}

// Usage quota operations

func (r *Repository) GetUsageQuota(ctx context.Context, orgID ulid.ULID) (*billing.UsageQuota, error) {
	query := `
		SELECT 
			organization_id, billing_tier, monthly_request_limit, monthly_token_limit,
			monthly_cost_limit, current_requests, current_tokens, current_cost,
			currency, reset_date, last_updated
		FROM usage_quotas 
		WHERE organization_id = $1`

	quota := &billing.UsageQuota{}
	err := r.db.QueryRowContext(ctx, query, orgID).Scan(
		&quota.OrganizationID,
		&quota.BillingTier,
		&quota.MonthlyRequestLimit,
		&quota.MonthlyTokenLimit,
		&quota.MonthlyCostLimit,
		&quota.CurrentRequests,
		&quota.CurrentTokens,
		&quota.CurrentCost,
		&quota.Currency,
		&quota.ResetDate,
		&quota.LastUpdated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No quota found, return nil without error
		}
		return nil, fmt.Errorf("failed to get usage quota: %w", err)
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
			$1, $2, $3, $4,
			$5, $6, $7, $8,
			$9, $10, $11
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

	_, err := r.db.ExecContext(ctx, query,
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
	)

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
	query := "SELECT 1"
	var result int
	
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	err := r.db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return fmt.Errorf("billing repository health check failed: %w", err)
	}
	
	return nil
}