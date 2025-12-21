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

type datasetItemService struct {
	itemRepo    evaluation.DatasetItemRepository
	datasetRepo evaluation.DatasetRepository
	logger      *slog.Logger
}

func NewDatasetItemService(
	itemRepo evaluation.DatasetItemRepository,
	datasetRepo evaluation.DatasetRepository,
	logger *slog.Logger,
) evaluation.DatasetItemService {
	return &datasetItemService{
		itemRepo:    itemRepo,
		datasetRepo: datasetRepo,
		logger:      logger,
	}
}

func (s *datasetItemService) Create(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *evaluation.CreateDatasetItemRequest) (*evaluation.DatasetItem, error) {
	if _, err := s.datasetRepo.GetByID(ctx, datasetID, projectID); err != nil {
		if errors.Is(err, evaluation.ErrDatasetNotFound) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("dataset %s", datasetID))
		}
		return nil, appErrors.NewInternalError("failed to verify dataset", err)
	}

	item := evaluation.NewDatasetItem(datasetID, req.Input)
	item.Expected = req.Expected
	if req.Metadata != nil {
		item.Metadata = req.Metadata
	}

	if validationErrors := item.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.NewValidationError(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, appErrors.NewInternalError("failed to create dataset item", err)
	}

	s.logger.Info("dataset item created",
		"item_id", item.ID,
		"dataset_id", datasetID,
	)

	return item, nil
}

func (s *datasetItemService) CreateBatch(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, req *evaluation.CreateDatasetItemsBatchRequest) (int, error) {
	if _, err := s.datasetRepo.GetByID(ctx, datasetID, projectID); err != nil {
		if errors.Is(err, evaluation.ErrDatasetNotFound) {
			return 0, appErrors.NewNotFoundError(fmt.Sprintf("dataset %s", datasetID))
		}
		return 0, appErrors.NewInternalError("failed to verify dataset", err)
	}

	if len(req.Items) == 0 {
		return 0, appErrors.NewValidationError("items", "items array cannot be empty")
	}

	items := make([]*evaluation.DatasetItem, 0, len(req.Items))
	for i, itemReq := range req.Items {
		item := evaluation.NewDatasetItem(datasetID, itemReq.Input)
		item.Expected = itemReq.Expected
		if itemReq.Metadata != nil {
			item.Metadata = itemReq.Metadata
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
		return 0, appErrors.NewInternalError("failed to create dataset items", err)
	}

	s.logger.Info("dataset items batch created",
		"dataset_id", datasetID,
		"count", len(items),
	)

	return len(items), nil
}

func (s *datasetItemService) List(ctx context.Context, datasetID ulid.ULID, projectID ulid.ULID, limit, offset int) ([]*evaluation.DatasetItem, int64, error) {
	if _, err := s.datasetRepo.GetByID(ctx, datasetID, projectID); err != nil {
		if errors.Is(err, evaluation.ErrDatasetNotFound) {
			return nil, 0, appErrors.NewNotFoundError(fmt.Sprintf("dataset %s", datasetID))
		}
		return nil, 0, appErrors.NewInternalError("failed to verify dataset", err)
	}

	items, total, err := s.itemRepo.List(ctx, datasetID, limit, offset)
	if err != nil {
		return nil, 0, appErrors.NewInternalError("failed to list dataset items", err)
	}
	return items, total, nil
}

func (s *datasetItemService) Delete(ctx context.Context, id ulid.ULID, datasetID ulid.ULID, projectID ulid.ULID) error {
	if _, err := s.datasetRepo.GetByID(ctx, datasetID, projectID); err != nil {
		if errors.Is(err, evaluation.ErrDatasetNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("dataset %s", datasetID))
		}
		return appErrors.NewInternalError("failed to verify dataset", err)
	}

	if err := s.itemRepo.Delete(ctx, id, datasetID); err != nil {
		if errors.Is(err, evaluation.ErrDatasetItemNotFound) {
			return appErrors.NewNotFoundError(fmt.Sprintf("dataset item %s", id))
		}
		return appErrors.NewInternalError("failed to delete dataset item", err)
	}

	s.logger.Info("dataset item deleted",
		"item_id", id,
		"dataset_id", datasetID,
	)

	return nil
}
