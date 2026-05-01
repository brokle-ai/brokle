package annotation

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"brokle/internal/core/domain/annotation"
	appErrors "brokle/pkg/errors"
)

type QueueService struct {
	queueRepo      annotation.QueueRepository
	itemRepo       annotation.ItemRepository
	assignmentRepo annotation.AssignmentRepository
	logger         *slog.Logger
}

// NewQueueService creates a new QueueService implementation.
func NewQueueService(
	queueRepo annotation.QueueRepository,
	itemRepo annotation.ItemRepository,
	assignmentRepo annotation.AssignmentRepository,
	logger *slog.Logger,
) *QueueService {
	return &QueueService{
		queueRepo:      queueRepo,
		itemRepo:       itemRepo,
		assignmentRepo: assignmentRepo,
		logger:         logger,
	}
}

// Create creates a new annotation queue.
func (s *QueueService) Create(ctx context.Context, projectID uuid.UUID, userID *uuid.UUID, req *annotation.CreateQueueRequest) (*annotation.AnnotationQueue, error) {
	queue := annotation.NewAnnotationQueue(projectID, req.Name)
	queue.Description = req.Description
	queue.Instructions = req.Instructions
	queue.CreatedBy = userID

	if req.ScoreConfigIDs != nil {
		queue.ScoreConfigIDs = req.ScoreConfigIDs
	}

	if req.Settings != nil {
		queue.Settings = *req.Settings
	}

	if validationErrors := queue.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.InvalidParam(validationErrors[0].Field, validationErrors[0].Message)
	}

	exists, err := s.queueRepo.ExistsByName(ctx, projectID, req.Name)
	if err != nil {
		return nil, appErrors.Internal("failed to check name uniqueness", err)
	}
	if exists {
		return nil, appErrors.Conflict("", fmt.Sprintf("annotation queue '%s' already exists in this project", req.Name))
	}

	if err := s.queueRepo.Create(ctx, queue); err != nil {
		return nil, err
	}

	s.logger.Info("annotation queue created",
		"queue_id", queue.ID,
		"project_id", projectID,
		"name", queue.Name,
	)

	return queue, nil
}

// GetByID retrieves an annotation queue by its ID.
func (s *QueueService) GetByID(ctx context.Context, id, projectID uuid.UUID) (*annotation.AnnotationQueue, error) {
	return s.queueRepo.GetByID(ctx, id, projectID)
}

// List retrieves all annotation queues for a project with optional filtering and pagination.
func (s *QueueService) List(ctx context.Context, projectID uuid.UUID, filter *annotation.QueueFilter, page, limit int) ([]*annotation.AnnotationQueue, int64, error) {
	offset := (page - 1) * limit
	queues, total, err := s.queueRepo.List(ctx, projectID, filter, offset, limit)
	if err != nil {
		return nil, 0, appErrors.Internal("failed to list annotation queues", err)
	}
	return queues, total, nil
}

// ListWithStats retrieves all annotation queues for a project with their statistics and pagination.
func (s *QueueService) ListWithStats(ctx context.Context, projectID uuid.UUID, filter *annotation.QueueFilter, page, limit int) ([]*annotation.AnnotationQueue, []*annotation.QueueStats, int64, error) {
	offset := (page - 1) * limit
	queues, total, err := s.queueRepo.List(ctx, projectID, filter, offset, limit)
	if err != nil {
		return nil, nil, 0, appErrors.Internal("failed to list annotation queues", err)
	}

	stats := make([]*annotation.QueueStats, len(queues))
	for i, queue := range queues {
		queueStats, err := s.itemRepo.GetStats(ctx, queue.ID, queue.Settings.LockTimeoutSeconds)
		if err != nil {
			// Log error but continue with empty stats
			s.logger.Warn("failed to get stats for queue",
				"queue_id", queue.ID,
				"error", err,
			)
			queueStats = &annotation.QueueStats{}
		}
		stats[i] = queueStats
	}

	return queues, stats, total, nil
}

// Update updates an existing annotation queue.
func (s *QueueService) Update(ctx context.Context, id, projectID uuid.UUID, req *annotation.UpdateQueueRequest) (*annotation.AnnotationQueue, error) {
	queue, err := s.queueRepo.GetByID(ctx, id, projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil && *req.Name != queue.Name {
		exists, err := s.queueRepo.ExistsByName(ctx, projectID, *req.Name)
		if err != nil {
			return nil, appErrors.Internal("failed to check name uniqueness", err)
		}
		if exists {
			return nil, appErrors.Conflict("", fmt.Sprintf("annotation queue '%s' already exists in this project", *req.Name))
		}
		queue.Name = *req.Name
	}

	if req.Description != nil {
		queue.Description = req.Description
	}
	if req.Instructions != nil {
		queue.Instructions = req.Instructions
	}
	if req.ScoreConfigIDs != nil {
		queue.ScoreConfigIDs = *req.ScoreConfigIDs // Dereference: empty slice clears, populated slice sets values
	}
	if req.Status != nil {
		queue.Status = *req.Status
	}
	if req.Settings != nil {
		queue.Settings = *req.Settings
	}

	queue.UpdatedAt = time.Now()

	if validationErrors := queue.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.InvalidParam(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.queueRepo.Update(ctx, queue); err != nil {
		return nil, err
	}

	s.logger.Info("annotation queue updated",
		"queue_id", id,
		"project_id", projectID,
	)

	return queue, nil
}

// Delete removes an annotation queue by ID.
// Also deletes all items and assignments in the queue (via CASCADE).
func (s *QueueService) Delete(ctx context.Context, id, projectID uuid.UUID) error {
	queue, err := s.queueRepo.GetByID(ctx, id, projectID)
	if err != nil {
		return err
	}

	if err := s.queueRepo.Delete(ctx, id, projectID); err != nil {
		return err
	}

	s.logger.Info("annotation queue deleted",
		"queue_id", id,
		"project_id", projectID,
		"name", queue.Name,
	)

	return nil
}

// GetWithStats retrieves a queue with its statistics.
func (s *QueueService) GetWithStats(ctx context.Context, id, projectID uuid.UUID) (*annotation.AnnotationQueue, *annotation.QueueStats, error) {
	queue, err := s.queueRepo.GetByID(ctx, id, projectID)
	if err != nil {
		return nil, nil, err
	}

	stats, err := s.itemRepo.GetStats(ctx, id, queue.Settings.LockTimeoutSeconds)
	if err != nil {
		return nil, nil, appErrors.Internal("failed to get queue stats", err)
	}

	return queue, stats, nil
}
