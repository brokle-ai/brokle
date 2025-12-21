package evaluation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

type experimentItemService struct {
	itemRepo        evaluation.ExperimentItemRepository
	experimentRepo  evaluation.ExperimentRepository
	datasetItemRepo evaluation.DatasetItemRepository
	logger          *slog.Logger
}

func NewExperimentItemService(
	itemRepo evaluation.ExperimentItemRepository,
	experimentRepo evaluation.ExperimentRepository,
	datasetItemRepo evaluation.DatasetItemRepository,
	logger *slog.Logger,
) evaluation.ExperimentItemService {
	return &experimentItemService{
		itemRepo:        itemRepo,
		experimentRepo:  experimentRepo,
		datasetItemRepo: datasetItemRepo,
		logger:          logger,
	}
}

func (s *experimentItemService) CreateBatch(ctx context.Context, experimentID ulid.ULID, projectID ulid.ULID, req *evaluation.CreateExperimentItemsBatchRequest) (int, error) {
	experiment, err := s.experimentRepo.GetByID(ctx, experimentID, projectID)
	if err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return 0, appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", experimentID))
		}
		return 0, appErrors.NewInternalError("failed to verify experiment", err)
	}

	if len(req.Items) == 0 {
		return 0, appErrors.NewValidationError("items", "items array cannot be empty")
	}

	items := make([]*evaluation.ExperimentItem, 0, len(req.Items))
	for i, itemReq := range req.Items {
		item := evaluation.NewExperimentItem(experimentID, itemReq.Input)
		item.Output = itemReq.Output
		item.Expected = itemReq.Expected
		item.TraceID = itemReq.TraceID
		if itemReq.Metadata != nil {
			item.Metadata = itemReq.Metadata
		}
		if itemReq.TrialNumber != nil {
			item.TrialNumber = *itemReq.TrialNumber
		}

		if itemReq.DatasetItemID != nil {
			if experiment.DatasetID == nil {
				return 0, appErrors.NewValidationError(
					fmt.Sprintf("items[%d].dataset_item_id", i),
					"cannot reference dataset items when experiment has no dataset",
				)
			}

			datasetItemID, err := ulid.Parse(*itemReq.DatasetItemID)
			if err != nil {
				return 0, appErrors.NewValidationError(
					fmt.Sprintf("items[%d].dataset_item_id", i),
					"must be a valid ULID",
				)
			}

			if _, err := s.datasetItemRepo.GetByID(ctx, datasetItemID, *experiment.DatasetID); err != nil {
				if errors.Is(err, evaluation.ErrDatasetItemNotFound) {
					return 0, appErrors.NewValidationError(
						fmt.Sprintf("items[%d].dataset_item_id", i),
						fmt.Sprintf("dataset item %s not found in experiment's dataset", datasetItemID),
					)
				}
				return 0, appErrors.NewInternalError("failed to verify dataset item", err)
			}

			item.DatasetItemID = &datasetItemID
		}

		if validationErrors := item.Validate(); len(validationErrors) > 0 {
			return 0, appErrors.NewValidationError(
				fmt.Sprintf("items[%d].%s", i, validationErrors[0].Field),
				validationErrors[0].Message,
			)
		}
		items = append(items, item)
	}

	if err := s.itemRepo.CreateBatch(ctx, items); err != nil {
		return 0, appErrors.NewInternalError("failed to create experiment items", err)
	}

	s.logger.Info("experiment items batch created",
		"experiment_id", experimentID,
		"count", len(items),
	)

	return len(items), nil
}

func (s *experimentItemService) List(ctx context.Context, experimentID ulid.ULID, projectID ulid.ULID, limit, offset int) ([]*evaluation.ExperimentItem, int64, error) {
	if _, err := s.experimentRepo.GetByID(ctx, experimentID, projectID); err != nil {
		if errors.Is(err, evaluation.ErrExperimentNotFound) {
			return nil, 0, appErrors.NewNotFoundError(fmt.Sprintf("experiment %s", experimentID))
		}
		return nil, 0, appErrors.NewInternalError("failed to verify experiment", err)
	}

	items, total, err := s.itemRepo.List(ctx, experimentID, limit, offset)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list experiment items", err)
	}
	return items, total, nil
}
