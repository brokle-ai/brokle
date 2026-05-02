package evaluation

import (
	"context"

	"github.com/google/uuid"

	evalDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
)

type experimentConfigRepository struct {
	tm *db.TxManager
}

func NewExperimentConfigRepository(tm *db.TxManager) evalDomain.ExperimentConfigRepository {
	return &experimentConfigRepository{tm: tm}
}

func (r *experimentConfigRepository) Create(ctx context.Context, c *evalDomain.ExperimentConfig) error {
	modelCfg, err := marshalEvalJSON(c.ModelConfig)
	if err != nil {
		return err
	}
	variableMapping, err := marshalEvalJSON(c.VariableMapping)
	if err != nil {
		return err
	}
	evaluators, err := marshalEvalJSON(c.Evaluators)
	if err != nil {
		return err
	}
	return r.tm.Queries(ctx).CreateExperimentConfig(ctx, gen.CreateExperimentConfigParams{
		ID:               c.ID,
		ExperimentID:     c.ExperimentID,
		PromptID:         c.PromptID,
		PromptVersionID:  c.PromptVersionID,
		ModelConfig:      modelCfg,
		DatasetID:        c.DatasetID,
		DatasetVersionID: c.DatasetVersionID,
		VariableMapping:  variableMapping,
		Evaluators:       evaluators,
	})
}

func (r *experimentConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*evalDomain.ExperimentConfig, error) {
	row, err := r.tm.Queries(ctx).GetExperimentConfigByID(ctx, id)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("experiment_config", appErrors.WithOp("repo.experiment_config.get_by_id"))
		}
		return nil, appErrors.Internal("get experiment config", err, appErrors.WithOp("repo.experiment_config.get_by_id"))
	}
	return experimentConfigFromRow(&row)
}

func (r *experimentConfigRepository) GetByExperimentID(ctx context.Context, experimentID uuid.UUID) (*evalDomain.ExperimentConfig, error) {
	row, err := r.tm.Queries(ctx).GetExperimentConfigByExperimentID(ctx, experimentID)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("experiment_config", appErrors.WithOp("repo.experiment_config.get_by_experiment_id"))
		}
		return nil, appErrors.Internal("get experiment config", err, appErrors.WithOp("repo.experiment_config.get_by_experiment_id"))
	}
	return experimentConfigFromRow(&row)
}

func (r *experimentConfigRepository) Update(ctx context.Context, c *evalDomain.ExperimentConfig) error {
	modelCfg, err := marshalEvalJSON(c.ModelConfig)
	if err != nil {
		return err
	}
	variableMapping, err := marshalEvalJSON(c.VariableMapping)
	if err != nil {
		return err
	}
	evaluators, err := marshalEvalJSON(c.Evaluators)
	if err != nil {
		return err
	}
	n, err := r.tm.Queries(ctx).UpdateExperimentConfig(ctx, gen.UpdateExperimentConfigParams{
		ID:               c.ID,
		PromptID:         c.PromptID,
		PromptVersionID:  c.PromptVersionID,
		ModelConfig:      modelCfg,
		DatasetID:        c.DatasetID,
		DatasetVersionID: c.DatasetVersionID,
		VariableMapping:  variableMapping,
		Evaluators:       evaluators,
	})
	if err != nil {
		return appErrors.Internal("update experiment config", err, appErrors.WithOp("repo.experiment_config.update"))
	}
	if n == 0 {
		return appErrors.NotFound("experiment_config", appErrors.WithOp("repo.experiment_config.update"))
	}
	return nil
}

func (r *experimentConfigRepository) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := r.tm.Queries(ctx).DeleteExperimentConfig(ctx, id)
	if err != nil {
		return appErrors.Internal("delete experiment config", err, appErrors.WithOp("repo.experiment_config.delete"))
	}
	if n == 0 {
		return appErrors.NotFound("experiment_config", appErrors.WithOp("repo.experiment_config.delete"))
	}
	return nil
}

func experimentConfigFromRow(row *gen.ExperimentConfig) (*evalDomain.ExperimentConfig, error) {
	c := &evalDomain.ExperimentConfig{
		ID:               row.ID,
		ExperimentID:     row.ExperimentID,
		PromptID:         row.PromptID,
		PromptVersionID:  row.PromptVersionID,
		DatasetID:        row.DatasetID,
		DatasetVersionID: row.DatasetVersionID,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}
	if err := unmarshalEvalJSON(row.ModelConfig, &c.ModelConfig); err != nil {
		return nil, err
	}
	if err := unmarshalEvalJSON(row.VariableMapping, &c.VariableMapping); err != nil {
		return nil, err
	}
	if err := unmarshalEvalJSON(row.Evaluators, &c.Evaluators); err != nil {
		return nil, err
	}
	return c, nil
}
