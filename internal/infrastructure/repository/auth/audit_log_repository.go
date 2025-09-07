package auth

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// auditLogRepository implements auth.AuditLogRepository using GORM
type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository instance
func NewAuditLogRepository(db *gorm.DB) auth.AuditLogRepository {
	return &auditLogRepository{
		db: db,
	}
}

// Create creates a new audit log entry
func (r *auditLogRepository) Create(ctx context.Context, auditLog *auth.AuditLog) error {
	return r.db.WithContext(ctx).Create(auditLog).Error
}

// GetByID retrieves an audit log by ID
func (r *auditLogRepository) GetByID(ctx context.Context, id ulid.ULID) (*auth.AuditLog, error) {
	var auditLog auth.AuditLog
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&auditLog).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("audit log not found")
		}
		return nil, err
	}
	return &auditLog, nil
}

// GetByUserID retrieves audit logs for a user
func (r *auditLogRepository) GetByUserID(ctx context.Context, userID ulid.ULID, limit, offset int) ([]*auth.AuditLog, error) {
	var auditLogs []*auth.AuditLog
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&auditLogs).Error
	return auditLogs, err
}

// GetByOrganizationID retrieves audit logs for an organization
func (r *auditLogRepository) GetByOrganizationID(ctx context.Context, orgID ulid.ULID, limit, offset int) ([]*auth.AuditLog, error) {
	var auditLogs []*auth.AuditLog
	err := r.db.WithContext(ctx).
		Where("organization_id = ?", orgID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&auditLogs).Error
	return auditLogs, err
}

// GetByResource retrieves audit logs for a resource
func (r *auditLogRepository) GetByResource(ctx context.Context, resource, resourceID string, limit, offset int) ([]*auth.AuditLog, error) {
	var auditLogs []*auth.AuditLog
	err := r.db.WithContext(ctx).
		Where("resource = ? AND resource_id = ?", resource, resourceID).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&auditLogs).Error
	return auditLogs, err
}

// GetByFilters retrieves audit logs based on filters
func (r *auditLogRepository) GetByFilters(ctx context.Context, filters *auth.AuditLogFilters) ([]*auth.AuditLog, int, error) {
	var auditLogs []*auth.AuditLog
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&auth.AuditLog{})

	// Apply filters
	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.OrganizationID != nil {
		query = query.Where("organization_id = ?", *filters.OrganizationID)
	}
	if filters.Action != nil && *filters.Action != "" {
		query = query.Where("action = ?", *filters.Action)
	}
	if filters.Resource != nil && *filters.Resource != "" {
		query = query.Where("resource = ?", *filters.Resource)
	}
	if filters.ResourceID != nil && *filters.ResourceID != "" {
		query = query.Where("resource_id = ?", *filters.ResourceID)
	}
	if filters.IPAddress != nil && *filters.IPAddress != "" {
		query = query.Where("ip_address = ?", *filters.IPAddress)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	// Get total count
	err := query.Count(&totalCount).Error
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	switch filters.SortBy {
	case "created_at":
		if filters.SortOrder == "asc" {
			query = query.Order("created_at ASC")
		} else {
			query = query.Order("created_at DESC")
		}
	case "action":
		if filters.SortOrder == "desc" {
			query = query.Order("action DESC")
		} else {
			query = query.Order("action ASC")
		}
	default:
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err = query.Find(&auditLogs).Error
	return auditLogs, int(totalCount), err
}

// GetByAction retrieves audit logs by action
func (r *auditLogRepository) GetByAction(ctx context.Context, action string, limit, offset int) ([]*auth.AuditLog, error) {
	var auditLogs []*auth.AuditLog
	err := r.db.WithContext(ctx).
		Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auditLogs).Error
	return auditLogs, err
}

// GetByDateRange retrieves audit logs within a date range
func (r *auditLogRepository) GetByDateRange(ctx context.Context, startDate, endDate time.Time, limit, offset int) ([]*auth.AuditLog, error) {
	var auditLogs []*auth.AuditLog
	err := r.db.WithContext(ctx).
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&auditLogs).Error
	return auditLogs, err
}

// Search searches audit logs based on filters
func (r *auditLogRepository) Search(ctx context.Context, filters *auth.AuditLogFilters) ([]*auth.AuditLog, int, error) {
	// This method already exists as GetByFilters, so let's alias it
	return r.GetByFilters(ctx, filters)
}

// CleanupOldLogs removes audit logs older than specified time
func (r *auditLogRepository) CleanupOldLogs(ctx context.Context, olderThan time.Time) error {
	return r.db.WithContext(ctx).
		Where("created_at < ?", olderThan).
		Delete(&auth.AuditLog{}).Error
}

// GetAuditLogStats returns audit log statistics
func (r *auditLogRepository) GetAuditLogStats(ctx context.Context) (*auth.AuditLogStats, error) {
	stats := &auth.AuditLogStats{
		LogsByAction:   make(map[string]int64),
		LogsByResource: make(map[string]int64),
	}
	
	// Get total logs count
	err := r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Count(&stats.TotalLogs).Error
	if err != nil {
		return nil, err
	}
	
	// Get logs by action
	type actionCount struct {
		Action string
		Count  int64
	}
	var actionCounts []actionCount
	err = r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Select("action, COUNT(*) as count").
		Group("action").
		Find(&actionCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, ac := range actionCounts {
		stats.LogsByAction[ac.Action] = ac.Count
	}
	
	// Get logs by resource
	type resourceCount struct {
		Resource string
		Count    int64
	}
	var resourceCounts []resourceCount
	err = r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("resource IS NOT NULL AND resource != ''").
		Group("resource").
		Find(&resourceCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, rc := range resourceCounts {
		stats.LogsByResource[rc.Resource] = rc.Count
	}
	
	// Get last log time
	var lastLog auth.AuditLog
	err = r.db.WithContext(ctx).
		Order("created_at DESC").
		First(&lastLog).Error
	if err == nil {
		stats.LastLogTime = &lastLog.CreatedAt
	}
	
	return stats, nil
}

// GetUserAuditLogStats returns audit log statistics for a specific user
func (r *auditLogRepository) GetUserAuditLogStats(ctx context.Context, userID ulid.ULID) (*auth.AuditLogStats, error) {
	stats := &auth.AuditLogStats{
		LogsByAction:   make(map[string]int64),
		LogsByResource: make(map[string]int64),
	}
	
	// Get total logs count for user
	err := r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Where("user_id = ?", userID).
		Count(&stats.TotalLogs).Error
	if err != nil {
		return nil, err
	}
	
	// Get logs by action for user
	type actionCount struct {
		Action string
		Count  int64
	}
	var actionCounts []actionCount
	err = r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("user_id = ?", userID).
		Group("action").
		Find(&actionCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, ac := range actionCounts {
		stats.LogsByAction[ac.Action] = ac.Count
	}
	
	// Get logs by resource for user
	type resourceCount struct {
		Resource string
		Count    int64
	}
	var resourceCounts []resourceCount
	err = r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("user_id = ? AND resource IS NOT NULL AND resource != ''", userID).
		Group("resource").
		Find(&resourceCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, rc := range resourceCounts {
		stats.LogsByResource[rc.Resource] = rc.Count
	}
	
	return stats, nil
}

// GetOrganizationAuditLogStats returns audit log statistics for a specific organization
func (r *auditLogRepository) GetOrganizationAuditLogStats(ctx context.Context, orgID ulid.ULID) (*auth.AuditLogStats, error) {
	stats := &auth.AuditLogStats{
		LogsByAction:   make(map[string]int64),
		LogsByResource: make(map[string]int64),
	}
	
	// Get total logs count for organization
	err := r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Where("organization_id = ?", orgID).
		Count(&stats.TotalLogs).Error
	if err != nil {
		return nil, err
	}
	
	// Get logs by action for organization
	type actionCount struct {
		Action string
		Count  int64
	}
	var actionCounts []actionCount
	err = r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Select("action, COUNT(*) as count").
		Where("organization_id = ?", orgID).
		Group("action").
		Find(&actionCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, ac := range actionCounts {
		stats.LogsByAction[ac.Action] = ac.Count
	}
	
	// Get logs by resource for organization
	type resourceCount struct {
		Resource string
		Count    int64
	}
	var resourceCounts []resourceCount
	err = r.db.WithContext(ctx).
		Model(&auth.AuditLog{}).
		Select("resource, COUNT(*) as count").
		Where("organization_id = ? AND resource IS NOT NULL AND resource != ''", orgID).
		Group("resource").
		Find(&resourceCounts).Error
	if err != nil {
		return nil, err
	}
	
	for _, rc := range resourceCounts {
		stats.LogsByResource[rc.Resource] = rc.Count
	}
	
	return stats, nil
}