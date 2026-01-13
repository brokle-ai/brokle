package evaluation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"brokle/internal/core/domain/evaluation"
	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

type experimentService struct {
	repo        evaluation.ExperimentRepository
	datasetRepo evaluation.DatasetRepository
	scoreRepo   observability.ScoreRepository
	logger      *slog.Logger
}

func NewExperimentService(
	repo evaluation.ExperimentRepository,
	datasetRepo evaluation.DatasetRepository,
	scoreRepo observability.ScoreRepository,
	logger *slog.Logger,
) evaluation.ExperimentService {
	return &experimentService{
		repo:        repo,
		datasetRepo: datasetRepo,
		scoreRepo:   scoreRepo,
		logger:      logger,
	}
}

func (s *experimentService) Create(ctx context.Context, projectID ulid.ULID, req *evaluation.CreateExperimentRequest) (*evaluation.Experiment, error) {
	experiment := evaluation.NewExperiment(projectID, req.Name)
	experiment.Description = req.Description
	if req.Metadata != nil {
		experiment.Metadata = req.Metadata
	}

	if req.DatasetID != nil {
		datasetID, err := ulid.Parse(*req.DatasetID)
		if err != nil {
			return nil, appErrors.NewValidationError("dataset_id", "must be a valid ULID")
		}
		if _, err := s.datasetRepo.GetByID(ctx, datasetID, projectID); err != nil {
			if errors.Is(err, evaluation.ErrDatasetNotFound) {
				return nil, appErrors.NewNotFoundError(fmt.Sprintf("dataset %s", *req.DatasetID))
			}
			return nil, appErrors.NewInternalError("failed to verify dataset", err)
		}
		experiment.DatasetID = &datasetID
	}

	if validationErrors := experiment.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.repo.Create(ctx, experiment); err != nil {
		return nil, appErrors.NewInternalError("failed to create experiment", err)
	}

	s.logger.Info("experiment created",
		"experiment_id", experiment.ID,
		"project_id", projectID,
		"name", experiment.Name,
	)

	return experiment, nil
}

func (s *experimentService) Update(ctx context.Context, id ulid.ULID, projectID ulid.ULID, req *evaluation.UpdateExperimentRequest) (*evaluation.Experiment, error) {
	experiment, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get experiment", err)
	}

	if req.Name != nil {
		experiment.Name = *req.Name
	}
	if req.Description != nil {
		experiment.Description = req.Description
	}
	if req.Metadata != nil {
		experiment.Metadata = req.Metadata
	}
	if req.Status != nil {
		oldStatus := experiment.Status
		experiment.Status = *req.Status

		now := time.Now()
		if oldStatus == evaluation.ExperimentStatusPending && *req.Status == evaluation.ExperimentStatusRunning {
			experiment.StartedAt = &now
		}
		if (*req.Status == evaluation.ExperimentStatusCompleted || *req.Status == evaluation.ExperimentStatusFailed) &&
			experiment.CompletedAt == nil {
			experiment.CompletedAt = &now
		}
	}

	experiment.UpdatedAt = time.Now()

	if validationErrors := experiment.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.repo.Update(ctx, experiment, projectID); err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", id))
		}
		return nil, appErrors.NewInternalError("failed to update experiment", err)
	}

	s.logger.Info("experiment updated",
		"experiment_id", id,
		"project_id", projectID,
		"status", experiment.Status,
	)

	return experiment, nil
}

func (s *experimentService) Delete(ctx context.Context, id ulid.ULID, projectID ulid.ULID) error {
	experiment, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", id))
		}
		return appErrors.NewInternalError("failed to get experiment", err)
	}

	if err := s.repo.Delete(ctx, id, projectID); err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", id))
		}
		return appErrors.NewInternalError("failed to delete experiment", err)
	}

	s.logger.Info("experiment deleted",
		"experiment_id", id,
		"project_id", projectID,
		"name", experiment.Name,
	)

	return nil
}

func (s *experimentService) GetByID(ctx context.Context, id ulid.ULID, projectID ulid.ULID) (*evaluation.Experiment, error) {
	experiment, err := s.repo.GetByID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get experiment", err)
	}
	return experiment, nil
}

func (s *experimentService) List(ctx context.Context, projectID ulid.ULID, filter *evaluation.ExperimentFilter) ([]*evaluation.Experiment, error) {
	experiments, err := s.repo.List(ctx, projectID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to list experiments", err)
	}
	return experiments, nil
}

// Rerun creates a new experiment based on an existing one, using the same dataset.
func (s *experimentService) Rerun(ctx context.Context, sourceID ulid.ULID, projectID ulid.ULID, req *evaluation.RerunExperimentRequest) (*evaluation.Experiment, error) {
	// Get the source experiment
	sourceExp, err := s.repo.GetByID(ctx, sourceID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", sourceID))
		}
		return nil, appErrors.NewInternalError("failed to get source experiment", err)
	}

	// Generate default name if not provided: "Original Name (Re-run)"
	name := fmt.Sprintf("%s (Re-run)", sourceExp.Name)
	if req.Name != nil && *req.Name != "" {
		name = *req.Name
	}

	// Create the new experiment
	newExp := evaluation.NewExperiment(projectID, name)
	newExp.DatasetID = sourceExp.DatasetID
	newExp.Description = req.Description
	if newExp.Description == nil {
		newExp.Description = sourceExp.Description
	}
	if req.Metadata != nil {
		newExp.Metadata = req.Metadata
	} else if sourceExp.Metadata != nil {
		// Copy source metadata and add rerun reference
		newExp.Metadata = make(map[string]interface{})
		for k, v := range sourceExp.Metadata {
			newExp.Metadata[k] = v
		}
	}
	// Add reference to source experiment (ensure map is initialized)
	if newExp.Metadata == nil {
		newExp.Metadata = make(map[string]interface{})
	}
	newExp.Metadata["source_experiment_id"] = sourceID.String()

	if validationErrors := newExp.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.repo.Create(ctx, newExp); err != nil {
		return nil, appErrors.NewInternalError("failed to create experiment", err)
	}

	s.logger.Info("experiment rerun created",
		"experiment_id", newExp.ID,
		"source_experiment_id", sourceID,
		"project_id", projectID,
		"name", newExp.Name,
	)

	return newExp, nil
}

// CompareExperiments compares score metrics across multiple experiments
func (s *experimentService) CompareExperiments(
	ctx context.Context,
	projectID ulid.ULID,
	experimentIDs []ulid.ULID,
	baselineID *ulid.ULID,
) (*evaluation.CompareExperimentsResponse, error) {
	if len(experimentIDs) < 2 {
		return nil, appErrors.NewValidationError("experiment_ids", "at least 2 experiments required for comparison")
	}

	// 1. Validate all experiments exist and belong to the project
	experimentSummaries := make(map[string]*evaluation.ExperimentSummary)
	experimentIDStrings := make([]string, len(experimentIDs))

	for i, expID := range experimentIDs {
		exp, err := s.repo.GetByID(ctx, expID, projectID)
		if err != nil {
			if errors.Is(err, evaluation.ErrExperimentNotFound) {
				return nil, appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", expID))
			}
			return nil, appErrors.NewInternalError("failed to get experiment", err)
		}

		experimentSummaries[expID.String()] = &evaluation.ExperimentSummary{
			Name:   exp.Name,
			Status: string(exp.Status),
		}
		experimentIDStrings[i] = expID.String()
	}

	// 2. Validate baseline is in the list (if provided)
	if baselineID != nil {
		found := false
		for _, expID := range experimentIDs {
			if expID == *baselineID {
				found = true
				break
			}
		}
		if !found {
			return nil, appErrors.NewValidationError("baseline_id", "baseline must be one of the compared experiments")
		}
	}

	// 3. Get score aggregations from ClickHouse
	scoreAggregations, err := s.scoreRepo.GetAggregationsByExperiments(ctx, projectID.String(), experimentIDStrings)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get score aggregations", err)
	}

	// 4. Convert observability.ScoreAggregation to evaluation.ScoreAggregation
	scores := make(map[string]map[string]*evaluation.ScoreAggregation)
	for scoreName, expScores := range scoreAggregations {
		scores[scoreName] = make(map[string]*evaluation.ScoreAggregation)
		for expID, agg := range expScores {
			scores[scoreName][expID] = &evaluation.ScoreAggregation{
				Mean:   agg.Mean,
				StdDev: agg.StdDev,
				Min:    agg.Min,
				Max:    agg.Max,
				Count:  agg.Count,
			}
		}
	}

	// 5. Calculate diffs if baseline is provided
	var diffs map[string]map[string]*evaluation.ScoreDiff
	if baselineID != nil {
		diffs = make(map[string]map[string]*evaluation.ScoreDiff)
		baselineIDStr := baselineID.String()

		for scoreName, expScores := range scores {
			baselineAgg := expScores[baselineIDStr]
			if baselineAgg == nil {
				continue
			}

			diffs[scoreName] = make(map[string]*evaluation.ScoreDiff)
			for expID, agg := range expScores {
				if expID == baselineIDStr {
					continue // Don't diff baseline against itself
				}
				diffs[scoreName][expID] = evaluation.CalculateDiff(baselineAgg, agg)
			}
		}
	}

	s.logger.Info("experiments compared",
		"project_id", projectID,
		"experiment_count", len(experimentIDs),
		"score_names", len(scores),
	)

	return &evaluation.CompareExperimentsResponse{
		Experiments: experimentSummaries,
		Scores:      scores,
		Diffs:       diffs,
	}, nil
}
