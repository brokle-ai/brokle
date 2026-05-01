package organization

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
)

// organizationSettingsRepository is the pgx+sqlc implementation of
// orgDomain.OrganizationSettingsRepository. Settings are stored as JSONB
// in the database and as a JSON string in the domain struct — the
// conversion is a simple []byte ↔ string at the boundary.
type organizationSettingsRepository struct {
	tm *db.TxManager
}

// NewOrganizationSettingsRepository returns the pgx-backed repository.
func NewOrganizationSettingsRepository(tm *db.TxManager) orgDomain.OrganizationSettingsRepository {
	return &organizationSettingsRepository{tm: tm}
}

func (r *organizationSettingsRepository) Create(ctx context.Context, s *orgDomain.OrganizationSettings) error {
	if err := r.tm.Queries(ctx).CreateOrganizationSetting(ctx, gen.CreateOrganizationSettingParams{
		ID:             s.ID,
		OrganizationID: s.OrganizationID,
		Key:            s.Key,
		Value:          json.RawMessage(s.Value),
		CreatedAt:      s.CreatedAt,
		UpdatedAt:      s.UpdatedAt,
	}); err != nil {
		if appErrors.IsUniqueViolation(err) {
			return appErrors.AlreadyExists("organization_settings", appErrors.WithOp("repo.organization_settings.create"))
		}
		return appErrors.Internal("create organization_setting", err,
			appErrors.WithOp("repo.organization_settings.create"))
	}
	return nil
}

func (r *organizationSettingsRepository) GetByID(ctx context.Context, id uuid.UUID) (*orgDomain.OrganizationSettings, error) {
	row, err := r.tm.Queries(ctx).GetOrganizationSettingByID(ctx, id)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("organization_settings", appErrors.WithOp("repo.organization_settings.get_by_id"))
		}
		return nil, appErrors.Internal("get organization_setting by id", err,
			appErrors.WithOp("repo.organization_settings.get_by_id"))
	}
	return organizationSettingsFromRow(&row), nil
}

func (r *organizationSettingsRepository) GetByKey(ctx context.Context, orgID uuid.UUID, key string) (*orgDomain.OrganizationSettings, error) {
	row, err := r.tm.Queries(ctx).GetOrganizationSettingByKey(ctx, gen.GetOrganizationSettingByKeyParams{
		OrganizationID: orgID,
		Key:            key,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("organization_settings", appErrors.WithOp("repo.organization_settings.get_by_key"))
		}
		return nil, appErrors.Internal("get organization_setting by key", err,
			appErrors.WithOp("repo.organization_settings.get_by_key"))
	}
	return organizationSettingsFromRow(&row), nil
}

func (r *organizationSettingsRepository) Update(ctx context.Context, s *orgDomain.OrganizationSettings) error {
	if err := r.tm.Queries(ctx).UpdateOrganizationSetting(ctx, gen.UpdateOrganizationSettingParams{
		ID:    s.ID,
		Key:   s.Key,
		Value: json.RawMessage(s.Value),
	}); err != nil {
		return appErrors.Internal("update organization_setting", err,
			appErrors.WithOp("repo.organization_settings.update"))
	}
	return nil
}

func (r *organizationSettingsRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.tm.Queries(ctx).DeleteOrganizationSetting(ctx, id); err != nil {
		return appErrors.Internal("delete organization_setting", err,
			appErrors.WithOp("repo.organization_settings.delete"))
	}
	return nil
}

func (r *organizationSettingsRepository) GetAllByOrganizationID(ctx context.Context, orgID uuid.UUID) ([]*orgDomain.OrganizationSettings, error) {
	rows, err := r.tm.Queries(ctx).ListOrganizationSettings(ctx, orgID)
	if err != nil {
		return nil, appErrors.Internal("list organization_settings", err,
			appErrors.WithOp("repo.organization_settings.list"))
	}
	return organizationSettingsFromRows(rows), nil
}

// GetSettingsMap decodes every stored setting value as JSON and returns
// a flat key → value map. Values that fail to decode are surfaced as
// raw strings so the operator can see the bytes rather than silently
// dropping them.
func (r *organizationSettingsRepository) GetSettingsMap(ctx context.Context, orgID uuid.UUID) (map[string]any, error) {
	settings, err := r.GetAllByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(settings))
	for _, s := range settings {
		var v any
		if err := json.Unmarshal([]byte(s.Value), &v); err != nil {
			v = s.Value
		}
		out[s.Key] = v
	}
	return out, nil
}

func (r *organizationSettingsRepository) DeleteByKey(ctx context.Context, orgID uuid.UUID, key string) error {
	if _, err := r.tm.Queries(ctx).DeleteOrganizationSettingByKey(ctx, gen.DeleteOrganizationSettingByKeyParams{
		OrganizationID: orgID,
		Key:            key,
	}); err != nil {
		return appErrors.Internal("delete organization_setting by key", err,
			appErrors.WithOp("repo.organization_settings.delete_by_key"))
	}
	return nil
}

// UpsertSetting creates-or-updates a single setting atomically. One
// round trip via ON CONFLICT replaces the previous get-then-create-or-
// update dance.
func (r *organizationSettingsRepository) UpsertSetting(ctx context.Context, orgID uuid.UUID, key string, value any) (*orgDomain.OrganizationSettings, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, appErrors.Internal("marshal setting value", err,
			appErrors.WithOp("repo.organization_settings.upsert"))
	}
	row, err := r.tm.Queries(ctx).UpsertOrganizationSetting(ctx, gen.UpsertOrganizationSettingParams{
		ID:             mustNewSettingID(),
		OrganizationID: orgID,
		Key:            key,
		Value:          raw,
	})
	if err != nil {
		return nil, appErrors.Internal("upsert organization_setting", err,
			appErrors.WithOp("repo.organization_settings.upsert"))
	}
	return organizationSettingsFromRow(&row), nil
}

// CreateMultiple creates several settings atomically. If any insert
// fails the entire batch rolls back via TxManager.WithinTransaction.
func (r *organizationSettingsRepository) CreateMultiple(ctx context.Context, settings []*orgDomain.OrganizationSettings) error {
	if len(settings) == 0 {
		return nil
	}
	return r.tm.WithinTransaction(ctx, func(ctx context.Context) error {
		q := r.tm.Queries(ctx)
		for _, s := range settings {
			if err := q.CreateOrganizationSetting(ctx, gen.CreateOrganizationSettingParams{
				ID:             s.ID,
				OrganizationID: s.OrganizationID,
				Key:            s.Key,
				Value:          json.RawMessage(s.Value),
				CreatedAt:      s.CreatedAt,
				UpdatedAt:      s.UpdatedAt,
			}); err != nil {
				if appErrors.IsUniqueViolation(err) {
					return appErrors.AlreadyExists("organization_settings",
						appErrors.WithOp("repo.organization_settings.create_multiple"))
				}
				return appErrors.Internal("create organization_setting", err,
					appErrors.WithOp("repo.organization_settings.create_multiple"))
			}
		}
		return nil
	})
}

func (r *organizationSettingsRepository) GetByKeys(ctx context.Context, orgID uuid.UUID, keys []string) ([]*orgDomain.OrganizationSettings, error) {
	rows, err := r.tm.Queries(ctx).ListOrganizationSettingsByKeys(ctx, gen.ListOrganizationSettingsByKeysParams{
		OrganizationID: orgID,
		Column2:        keys,
	})
	if err != nil {
		return nil, appErrors.Internal("list organization_settings by keys", err,
			appErrors.WithOp("repo.organization_settings.list_by_keys"))
	}
	return organizationSettingsFromRows(rows), nil
}

func (r *organizationSettingsRepository) DeleteMultiple(ctx context.Context, orgID uuid.UUID, keys []string) error {
	if len(keys) == 0 {
		return nil
	}
	if _, err := r.tm.Queries(ctx).DeleteOrganizationSettingsByKeys(ctx, gen.DeleteOrganizationSettingsByKeysParams{
		OrganizationID: orgID,
		Column2:        keys,
	}); err != nil {
		return appErrors.Internal("delete organization_settings by keys", err,
			appErrors.WithOp("repo.organization_settings.delete_by_keys"))
	}
	return nil
}

// organizationSettingsFromRow adapts a sqlc row into the domain type.
// The domain stores JSON as a string, not []byte, to match the legacy
// interface.
func organizationSettingsFromRow(row *gen.OrganizationSetting) *orgDomain.OrganizationSettings {
	return &orgDomain.OrganizationSettings{
		ID:             row.ID,
		OrganizationID: row.OrganizationID,
		Key:            row.Key,
		Value:          string(row.Value),
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func organizationSettingsFromRows(rows []gen.OrganizationSetting) []*orgDomain.OrganizationSettings {
	out := make([]*orgDomain.OrganizationSettings, 0, len(rows))
	for i := range rows {
		out = append(out, organizationSettingsFromRow(&rows[i]))
	}
	return out
}

// mustNewSettingID generates the ID used by UpsertOrganizationSetting
// on the INSERT path. The ON CONFLICT clause ignores it when updating,
// so the value is only observed on first-write.
func mustNewSettingID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}

