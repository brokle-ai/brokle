package auth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	authDomain "brokle/internal/core/domain/auth"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
)

// auditLogRepository is the pgx+sqlc implementation of
// authDomain.AuditLogRepository. Dynamic-filter reads live in
// audit_log_filter.go (squirrel); aggregate stats live in
// audit_log_stats.go.
//
// The domain AuditLog mirrors the Postgres schema exactly (nullable columns
// are pointers, metadata is raw JSON), so pgx/sqlc scans directly into the
// struct with no adapter helpers.
type auditLogRepository struct {
	tm *db.TxManager
}

// NewAuditLogRepository returns the pgx-backed repository.
func NewAuditLogRepository(tm *db.TxManager) authDomain.AuditLogRepository {
	return &auditLogRepository{tm: tm}
}

func (r *auditLogRepository) Create(ctx context.Context, log *authDomain.AuditLog) error {
	if err := r.tm.Queries(ctx).CreateAuditLog(ctx, gen.CreateAuditLogParams{
		ID:             log.ID,
		UserID:         log.UserID,
		OrganizationID: log.OrganizationID,
		Action:         log.Action,
		Resource:       log.Resource,
		ResourceID:     log.ResourceID,
		Metadata:       log.Metadata,
		IpAddress:      log.IPAddress,
		UserAgent:      log.UserAgent,
		CreatedAt:      log.CreatedAt,
	}); err != nil {
		return fmt.Errorf("create audit_log: %w", err)
	}
	return nil
}

func (r *auditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*authDomain.AuditLog, error) {
	row, err := r.tm.Queries(ctx).GetAuditLogByID(ctx, id)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("audit_log", appErrors.WithOp("repo.audit_log.get_by_id"))
		}
		return nil, fmt.Errorf("get audit_log by ID %s: %w", id, err)
	}
	return auditLogFromRow(&row), nil
}

func (r *auditLogRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*authDomain.AuditLog, error) {
	rows, err := r.tm.Queries(ctx).ListAuditLogsByUser(ctx, gen.ListAuditLogsByUserParams{
		UserID: &userID,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list audit_logs for user %s: %w", userID, err)
	}
	return auditLogsFromRows(rows), nil
}

func (r *auditLogRepository) GetByOrganizationID(ctx context.Context, orgID uuid.UUID, limit, offset int) ([]*authDomain.AuditLog, error) {
	rows, err := r.tm.Queries(ctx).ListAuditLogsByOrganization(ctx, gen.ListAuditLogsByOrganizationParams{
		OrganizationID: &orgID,
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list audit_logs for org %s: %w", orgID, err)
	}
	return auditLogsFromRows(rows), nil
}

func (r *auditLogRepository) GetByResource(ctx context.Context, resource, resourceID string, limit, offset int) ([]*authDomain.AuditLog, error) {
	var rid *string
	if resourceID != "" {
		rid = &resourceID
	}
	rows, err := r.tm.Queries(ctx).ListAuditLogsByResource(ctx, gen.ListAuditLogsByResourceParams{
		Resource:   resource,
		ResourceID: rid,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("list audit_logs for resource %s/%s: %w", resource, resourceID, err)
	}
	return auditLogsFromRows(rows), nil
}

// ----- gen ↔ domain boundary -----------------------------------------

// auditLogFromRow adapts a sqlc-generated row to the domain type. Field
// types are aligned so this is a direct copy, no nil-coalescing.
func auditLogFromRow(row *gen.AuditLog) *authDomain.AuditLog {
	return &authDomain.AuditLog{
		ID:             row.ID,
		UserID:         row.UserID,
		OrganizationID: row.OrganizationID,
		Action:         row.Action,
		Resource:       row.Resource,
		ResourceID:     row.ResourceID,
		Metadata:       row.Metadata,
		IPAddress:      row.IpAddress,
		UserAgent:      row.UserAgent,
		CreatedAt:      row.CreatedAt,
	}
}

func auditLogsFromRows(rows []gen.AuditLog) []*authDomain.AuditLog {
	out := make([]*authDomain.AuditLog, 0, len(rows))
	for i := range rows {
		out = append(out, auditLogFromRow(&rows[i]))
	}
	return out
}
