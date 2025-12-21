package evaluation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

type experimentService struct {
	repo        evaluation.ExperimentRepository
	datasetRepo evaluation.DatasetRepository
	logger      *slog.Logger
}

func NewExperimentService(
	repo evaluation.ExperimentRepository,
	datasetRepo evaluation.DatasetRepository,
	logger *slog.Logger,
) evaluation.ExperimentService {
	return &experimentService{
		repo:        repo,
		datasetRepo: datasetRepo,
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
