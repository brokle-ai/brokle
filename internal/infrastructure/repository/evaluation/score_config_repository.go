package evaluation

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	evalDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
)

// float64PtrToDecimal / decimalPtrToFloat64 bridge between the domain's
// *float64 score bounds and the generated *decimal.Decimal. Lossless
// for the typical 0-1 or 0-100 range of score configs.
func float64PtrToDecimal(p *float64) *decimal.Decimal {
	if p == nil {
		return nil
	}
	d := decimal.NewFromFloat(*p)
	return &d
}

func decimalPtrToFloat64(p *decimal.Decimal) *float64 {
	if p == nil {
		return nil
	}
	f := p.InexactFloat64()
	return &f
}

type scoreConfigRepository struct {
	tm *db.TxManager
}

func NewScoreConfigRepository(tm *db.TxManager) evalDomain.ScoreConfigRepository {
	return &scoreConfigRepository{tm: tm}
}

func (r *scoreConfigRepository) Create(ctx context.Context, c *evalDomain.ScoreConfig) error {
	cats, err := marshalEvalJSON(c.Categories)
	if err != nil {
		return err
	}
	meta, err := marshalEvalJSON(c.Metadata)
	if err != nil {
		return err
	}
	if err := r.tm.Queries(ctx).CreateScoreConfig(ctx, gen.CreateScoreConfigParams{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		Name:        c.Name,
		Description: c.Description,
		Type:        string(c.Type),
		MinValue:    float64PtrToDecimal(c.MinValue),
		MaxValue:    float64PtrToDecimal(c.MaxValue),
		Categories:  cats,
		Metadata:    meta,
	}); err != nil {
		if appErrors.IsUniqueViolation(err) {
			return appErrors.AlreadyExists("score_config", appErrors.WithOp("repo.score_config.create"))
		}
		return appErrors.Internal("create score config", err, appErrors.WithOp("repo.score_config.create"))
	}
	return nil
}

func (r *scoreConfigRepository) GetByID(ctx context.Context, id, projectID uuid.UUID) (*evalDomain.ScoreConfig, error) {
	row, err := r.tm.Queries(ctx).GetScoreConfigByID(ctx, gen.GetScoreConfigByIDParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("score_config", appErrors.WithOp("repo.score_config.get_by_id"))
		}
		return nil, appErrors.Internal("get score config", err, appErrors.WithOp("repo.score_config.get_by_id"))
	}
	return scoreConfigFromRow(&row)
}

func (r *scoreConfigRepository) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*evalDomain.ScoreConfig, error) {
	row, err := r.tm.Queries(ctx).GetScoreConfigByName(ctx, gen.GetScoreConfigByNameParams{
		ProjectID: projectID,
		Name:      name,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return scoreConfigFromRow(&row)
}

func (r *scoreConfigRepository) List(ctx context.Context, projectID uuid.UUID, offset, limit int) ([]*evalDomain.ScoreConfig, int64, error) {
	total, err := r.tm.Queries(ctx).CountScoreConfigsByProject(ctx, projectID)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.tm.Queries(ctx).ListScoreConfigsByProject(ctx, gen.ListScoreConfigsByProjectParams{
		ProjectID: projectID,
		Offset:    int32(offset),
		Limit:     int32(limit),
	})
	if err != nil {
		return nil, 0, err
	}
	out := make([]*evalDomain.ScoreConfig, 0, len(rows))
	for i := range rows {
		c, err := scoreConfigFromRow(&rows[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, c)
	}
	return out, total, nil
}

func (r *scoreConfigRepository) Update(ctx context.Context, c *evalDomain.ScoreConfig, projectID uuid.UUID) error {
	cats, err := marshalEvalJSON(c.Categories)
	if err != nil {
		return err
	}
	meta, err := marshalEvalJSON(c.Metadata)
	if err != nil {
		return err
	}
	n, err := r.tm.Queries(ctx).UpdateScoreConfig(ctx, gen.UpdateScoreConfigParams{
		ID:          c.ID,
		ProjectID:   projectID,
		Name:        c.Name,
		Description: c.Description,
		Type:        string(c.Type),
		MinValue:    float64PtrToDecimal(c.MinValue),
		MaxValue:    float64PtrToDecimal(c.MaxValue),
		Categories:  cats,
		Metadata:    meta,
	})
	if err != nil {
		if appErrors.IsUniqueViolation(err) {
			return appErrors.AlreadyExists("score_config", appErrors.WithOp("repo.score_config.update"))
		}
		return appErrors.Internal("update score config", err, appErrors.WithOp("repo.score_config.update"))
	}
	if n == 0 {
		return appErrors.NotFound("score_config", appErrors.WithOp("repo.score_config.update"))
	}
	return nil
}

func (r *scoreConfigRepository) Delete(ctx context.Context, id, projectID uuid.UUID) error {
	n, err := r.tm.Queries(ctx).DeleteScoreConfig(ctx, gen.DeleteScoreConfigParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		return appErrors.Internal("delete score config", err, appErrors.WithOp("repo.score_config.delete"))
	}
	if n == 0 {
		return appErrors.NotFound("score_config", appErrors.WithOp("repo.score_config.delete"))
	}
	return nil
}

func (r *scoreConfigRepository) ExistsByName(ctx context.Context, projectID uuid.UUID, name string) (bool, error) {
	return r.tm.Queries(ctx).ScoreConfigExistsByName(ctx, gen.ScoreConfigExistsByNameParams{
		ProjectID: projectID,
		Name:      name,
	})
}

func scoreConfigFromRow(row *gen.ScoreConfig) (*evalDomain.ScoreConfig, error) {
	c := &evalDomain.ScoreConfig{
		ID:          row.ID,
		ProjectID:   row.ProjectID,
		Name:        row.Name,
		Description: row.Description,
		Type:        evalDomain.ScoreType(row.Type),
		MinValue:    decimalPtrToFloat64(row.MinValue),
		MaxValue:    decimalPtrToFloat64(row.MaxValue),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
	if err := unmarshalEvalJSON(row.Categories, &c.Categories); err != nil {
		return nil, fmt.Errorf("unmarshal categories: %w", err)
	}
	if err := unmarshalEvalJSON(row.Metadata, &c.Metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metadata: %w", err)
	}
	return c, nil
}

// marshalEvalJSON handles the domain-side JSONB serialization. Returns
// nil for empty/nil values so the column stores NULL.
func marshalEvalJSON(v any) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

func unmarshalEvalJSON(raw json.RawMessage, dst any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, dst)
}
