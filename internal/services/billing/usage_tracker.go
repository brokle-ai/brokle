package billing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/pkg/ulid"
)

// UsageTracker manages organization usage tracking and quotas
type UsageTracker struct {
	logger     *logrus.Logger
	repository BillingRepository
	
	// In-memory cache for frequently accessed quotas
	quotaCache map[ulid.ULID]*UsageQuota
	cacheMutex sync.RWMutex
	
	// Configuration
	cacheExpiry time.Duration
	syncInterval time.Duration
	
	// Background sync
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// UsageUpdate represents an update to organization usage
type UsageUpdate struct {
	OrganizationID ulid.ULID
	Requests       int64
	Tokens         int64
	Cost           float64
	Currency       string
	Timestamp      time.Time
}

// NewUsageTracker creates a new usage tracker instance
func NewUsageTracker(logger *logrus.Logger, repository BillingRepository) *UsageTracker {
	tracker := &UsageTracker{
		logger:       logger,
		repository:   repository,
		quotaCache:   make(map[ulid.ULID]*UsageQuota),
		cacheExpiry:  5 * time.Minute,
		syncInterval: 1 * time.Minute,
		stopCh:       make(chan struct{}),
	}
	
	// Start background sync
	tracker.wg.Add(1)
	go tracker.backgroundSync()
	
	return tracker
}

// UpdateUsage updates usage tracking for an organization
func (t *UsageTracker) UpdateUsage(ctx context.Context, orgID ulid.ULID, record *UsageRecord) error {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()
	
	// Get or load quota
	quota, err := t.getQuotaLocked(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}
	
	if quota == nil {
		// No quota exists, create default one
		quota = &UsageQuota{
			OrganizationID:      orgID,
			BillingTier:         record.BillingTier,
			MonthlyRequestLimit: 0, // Unlimited by default
			MonthlyTokenLimit:   0, // Unlimited by default
			MonthlyCostLimit:    0, // Unlimited by default
			Currency:            record.Currency,
			ResetDate:           t.getNextResetDate(),
			LastUpdated:         time.Now(),
		}
	}
	
	// Update current usage
	quota.CurrentRequests++
	quota.CurrentTokens += int64(record.TotalTokens)
	quota.CurrentCost += record.NetCost
	quota.LastUpdated = time.Now()
	
	// Check if we need to reset monthly counters
	if time.Now().After(quota.ResetDate) {
		quota.CurrentRequests = 1 // This request
		quota.CurrentTokens = int64(record.TotalTokens)
		quota.CurrentCost = record.NetCost
		quota.ResetDate = t.getNextResetDate()
	}
	
	// Update cache
	t.quotaCache[orgID] = quota
	
	// Persist to database (async to avoid blocking)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		if err := t.repository.UpdateUsageQuota(ctx, orgID, quota); err != nil {
			t.logger.WithError(err).WithField("org_id", orgID).Error("Failed to persist usage quota")
		}
	}()
	
	return nil
}

// GetUsageQuota retrieves current usage quota for an organization
func (t *UsageTracker) GetUsageQuota(ctx context.Context, orgID ulid.ULID) (*UsageQuota, error) {
	t.cacheMutex.RLock()
	defer t.cacheMutex.RUnlock()
	
	return t.getQuotaLocked(ctx, orgID)
}

// SetUsageQuota sets usage quota limits for an organization
func (t *UsageTracker) SetUsageQuota(ctx context.Context, orgID ulid.ULID, quota *UsageQuota) error {
	quota.LastUpdated = time.Now()
	
	// Update database
	if err := t.repository.UpdateUsageQuota(ctx, orgID, quota); err != nil {
		return fmt.Errorf("failed to update usage quota: %w", err)
	}
	
	// Update cache
	t.cacheMutex.Lock()
	t.quotaCache[orgID] = quota
	t.cacheMutex.Unlock()
	
	t.logger.WithFields(logrus.Fields{
		"org_id":             orgID,
		"request_limit":      quota.MonthlyRequestLimit,
		"token_limit":        quota.MonthlyTokenLimit,
		"cost_limit":         quota.MonthlyCostLimit,
	}).Info("Updated usage quota")
	
	return nil
}

// CheckQuotaExceeded checks if organization has exceeded any quotas
func (t *UsageTracker) CheckQuotaExceeded(ctx context.Context, orgID ulid.ULID) (*QuotaStatus, error) {
	quota, err := t.GetUsageQuota(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage quota: %w", err)
	}
	
	if quota == nil {
		// No quota set, allow unlimited usage
		return &QuotaStatus{
			OrganizationID: orgID,
			RequestsOK:     true,
			TokensOK:       true,
			CostOK:         true,
			Status:         "unlimited",
		}, nil
	}
	
	status := &QuotaStatus{
		OrganizationID: orgID,
	}
	
	// Check request limits
	if quota.MonthlyRequestLimit > 0 {
		status.RequestsOK = quota.CurrentRequests < quota.MonthlyRequestLimit
		status.RequestsUsagePercent = float64(quota.CurrentRequests) / float64(quota.MonthlyRequestLimit) * 100
	} else {
		status.RequestsOK = true
	}
	
	// Check token limits
	if quota.MonthlyTokenLimit > 0 {
		status.TokensOK = quota.CurrentTokens < quota.MonthlyTokenLimit
		status.TokensUsagePercent = float64(quota.CurrentTokens) / float64(quota.MonthlyTokenLimit) * 100
	} else {
		status.TokensOK = true
	}
	
	// Check cost limits
	if quota.MonthlyCostLimit > 0 {
		status.CostOK = quota.CurrentCost < quota.MonthlyCostLimit
		status.CostUsagePercent = quota.CurrentCost / quota.MonthlyCostLimit * 100
	} else {
		status.CostOK = true
	}
	
	// Determine overall status
	if !status.RequestsOK {
		status.Status = "requests_exceeded"
	} else if !status.TokensOK {
		status.Status = "tokens_exceeded"
	} else if !status.CostOK {
		status.Status = "cost_exceeded"
	} else {
		status.Status = "within_limits"
	}
	
	return status, nil
}

// GetUsageHistory retrieves usage history for an organization
func (t *UsageTracker) GetUsageHistory(ctx context.Context, orgID ulid.ULID, start, end time.Time) ([]*UsageRecord, error) {
	return t.repository.GetUsageRecords(ctx, orgID, start, end)
}

// ResetMonthlyUsage resets monthly usage counters for an organization
func (t *UsageTracker) ResetMonthlyUsage(ctx context.Context, orgID ulid.ULID) error {
	quota, err := t.GetUsageQuota(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get usage quota: %w", err)
	}
	
	if quota == nil {
		return fmt.Errorf("no usage quota found for organization %s", orgID)
	}
	
	// Reset counters
	quota.CurrentRequests = 0
	quota.CurrentTokens = 0
	quota.CurrentCost = 0
	quota.ResetDate = t.getNextResetDate()
	quota.LastUpdated = time.Now()
	
	// Update database and cache
	if err := t.SetUsageQuota(ctx, orgID, quota); err != nil {
		return fmt.Errorf("failed to reset usage quota: %w", err)
	}
	
	t.logger.WithField("org_id", orgID).Info("Reset monthly usage counters")
	return nil
}

// Stop stops the usage tracker background processes
func (t *UsageTracker) Stop() {
	close(t.stopCh)
	t.wg.Wait()
}

// Health check
func (t *UsageTracker) GetHealth() map[string]interface{} {
	t.cacheMutex.RLock()
	cacheSize := len(t.quotaCache)
	t.cacheMutex.RUnlock()
	
	return map[string]interface{}{
		"service":              "usage_tracker",
		"status":               "healthy",
		"cached_quotas":        cacheSize,
		"cache_expiry_minutes": t.cacheExpiry.Minutes(),
		"sync_interval_seconds": t.syncInterval.Seconds(),
	}
}

// Internal methods

func (t *UsageTracker) getQuotaLocked(ctx context.Context, orgID ulid.ULID) (*UsageQuota, error) {
	// Check cache first
	if quota, exists := t.quotaCache[orgID]; exists {
		// Check if cache entry is still valid
		if time.Since(quota.LastUpdated) < t.cacheExpiry {
			return quota, nil
		}
		// Cache expired, remove it
		delete(t.quotaCache, orgID)
	}
	
	// Load from database
	quota, err := t.repository.GetUsageQuota(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to load usage quota from database: %w", err)
	}
	
	// Update cache if quota exists
	if quota != nil {
		t.quotaCache[orgID] = quota
	}
	
	return quota, nil
}

func (t *UsageTracker) getNextResetDate() time.Time {
	now := time.Now()
	// Reset on the first of next month
	return time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
}

func (t *UsageTracker) backgroundSync() {
	defer t.wg.Done()
	
	ticker := time.NewTicker(t.syncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			t.syncQuotas()
		}
	}
}

func (t *UsageTracker) syncQuotas() {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()
	
	// Check for expired quotas and quotas that need monthly reset
	now := time.Now()
	var expiredOrgs []ulid.ULID
	
	for orgID, quota := range t.quotaCache {
		// Check if cache entry expired
		if now.Sub(quota.LastUpdated) > t.cacheExpiry {
			expiredOrgs = append(expiredOrgs, orgID)
			continue
		}
		
		// Check if monthly reset is needed
		if now.After(quota.ResetDate) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			
			// Reset monthly counters
			quota.CurrentRequests = 0
			quota.CurrentTokens = 0
			quota.CurrentCost = 0
			quota.ResetDate = t.getNextResetDate()
			quota.LastUpdated = now
			
			// Persist reset
			if err := t.repository.UpdateUsageQuota(ctx, orgID, quota); err != nil {
				t.logger.WithError(err).WithField("org_id", orgID).Error("Failed to sync quota reset")
			} else {
				t.logger.WithField("org_id", orgID).Info("Monthly usage quota reset")
			}
			
			cancel()
		}
	}
	
	// Remove expired cache entries
	for _, orgID := range expiredOrgs {
		delete(t.quotaCache, orgID)
	}
	
	if len(expiredOrgs) > 0 {
		t.logger.WithField("expired_count", len(expiredOrgs)).Debug("Cleared expired quota cache entries")
	}
}

// GetUsageMetrics returns usage metrics for monitoring
func (t *UsageTracker) GetUsageMetrics(ctx context.Context, orgID ulid.ULID) (map[string]interface{}, error) {
	quota, err := t.GetUsageQuota(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage quota: %w", err)
	}
	
	if quota == nil {
		return map[string]interface{}{
			"organization_id": orgID,
			"has_quota":       false,
			"unlimited":       true,
		}, nil
	}
	
	return map[string]interface{}{
		"organization_id":         orgID,
		"has_quota":              true,
		"billing_tier":           quota.BillingTier,
		"current_requests":       quota.CurrentRequests,
		"current_tokens":         quota.CurrentTokens,
		"current_cost":           quota.CurrentCost,
		"monthly_request_limit":  quota.MonthlyRequestLimit,
		"monthly_token_limit":    quota.MonthlyTokenLimit,
		"monthly_cost_limit":     quota.MonthlyCostLimit,
		"currency":               quota.Currency,
		"reset_date":             quota.ResetDate,
		"last_updated":           quota.LastUpdated,
		"requests_usage_percent": func() float64 {
			if quota.MonthlyRequestLimit > 0 {
				return float64(quota.CurrentRequests) / float64(quota.MonthlyRequestLimit) * 100
			}
			return 0
		}(),
		"tokens_usage_percent": func() float64 {
			if quota.MonthlyTokenLimit > 0 {
				return float64(quota.CurrentTokens) / float64(quota.MonthlyTokenLimit) * 100
			}
			return 0
		}(),
		"cost_usage_percent": func() float64 {
			if quota.MonthlyCostLimit > 0 {
				return quota.CurrentCost / quota.MonthlyCostLimit * 100
			}
			return 0
		}(),
	}, nil
}