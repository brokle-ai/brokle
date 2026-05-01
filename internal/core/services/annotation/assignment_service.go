// Package annotation implements human-in-the-loop annotation: queues, items, assignments, and review workflow.
package annotation

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"brokle/internal/core/domain/annotation"
	appErrors "brokle/pkg/errors"
)

type AssignmentService struct {
	queueRepo      annotation.QueueRepository
	assignmentRepo annotation.AssignmentRepository
	logger         *slog.Logger
}

// NewAssignmentService creates a new AssignmentService implementation.
func NewAssignmentService(
	queueRepo annotation.QueueRepository,
	assignmentRepo annotation.AssignmentRepository,
	logger *slog.Logger,
) *AssignmentService {
	return &AssignmentService{
		queueRepo:      queueRepo,
		assignmentRepo: assignmentRepo,
		logger:         logger,
	}
}

// Assign assigns a user to a queue with the specified role.
func (s *AssignmentService) Assign(ctx context.Context, queueID, projectID, userID uuid.UUID, role annotation.AssignmentRole, assignedBy *uuid.UUID) (*annotation.QueueAssignment, error) {
	// Verify queue exists and belongs to project
	if _, err := s.queueRepo.GetByID(ctx, queueID, projectID); err != nil {
		return nil, err
	}

	// Check if already assigned
	isAssigned, err := s.assignmentRepo.IsAssigned(ctx, queueID, userID)
	if err != nil {
		return nil, appErrors.Internal("failed to check assignment", err)
	}
	if isAssigned {
		return nil, appErrors.Conflict("", "user is already assigned to this queue")
	}

	// Validate role
	if !role.IsValid() {
		return nil, appErrors.InvalidParam("role", "invalid role, must be annotator, reviewer, or admin")
	}

	// Create assignment
	assignment := annotation.NewQueueAssignment(queueID, userID, role)
	assignment.AssignedBy = assignedBy

	if validationErrors := assignment.Validate(); len(validationErrors) > 0 {
		return nil, appErrors.InvalidParam(validationErrors[0].Field, validationErrors[0].Message)
	}

	if err := s.assignmentRepo.Create(ctx, assignment); err != nil {
		return nil, err
	}

	s.logger.Info("user assigned to annotation queue",
		"assignment_id", assignment.ID,
		"queue_id", queueID,
		"user_id", userID,
		"role", role,
	)

	return assignment, nil
}

// Unassign removes a user's assignment from a queue.
func (s *AssignmentService) Unassign(ctx context.Context, queueID, projectID, userID uuid.UUID) error {
	// Verify queue exists and belongs to project
	if _, err := s.queueRepo.GetByID(ctx, queueID, projectID); err != nil {
		return err
	}

	if err := s.assignmentRepo.Delete(ctx, queueID, userID); err != nil {
		return err
	}

	s.logger.Info("user unassigned from annotation queue",
		"queue_id", queueID,
		"user_id", userID,
	)

	return nil
}

// ListAssignments retrieves all assignments for a queue.
func (s *AssignmentService) ListAssignments(ctx context.Context, queueID, projectID uuid.UUID) ([]*annotation.QueueAssignment, error) {
	// Verify queue exists and belongs to project
	if _, err := s.queueRepo.GetByID(ctx, queueID, projectID); err != nil {
		return nil, err
	}

	assignments, err := s.assignmentRepo.List(ctx, queueID)
	if err != nil {
		return nil, appErrors.Internal("failed to list assignments", err)
	}

	return assignments, nil
}

// GetUserQueues retrieves all queues a user is assigned to.
func (s *AssignmentService) GetUserQueues(ctx context.Context, userID uuid.UUID) ([]*annotation.QueueAssignment, error) {
	assignments, err := s.assignmentRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, appErrors.Internal("failed to list user queue assignments", err)
	}
	return assignments, nil
}

// CheckAccess verifies if a user has access to a queue with the minimum required role.
func (s *AssignmentService) CheckAccess(ctx context.Context, queueID, userID uuid.UUID, minRole annotation.AssignmentRole) error {
	hasRole, err := s.assignmentRepo.HasRole(ctx, queueID, userID, minRole)
	if err != nil {
		return appErrors.Internal("failed to check access", err)
	}

	if !hasRole {
		return appErrors.PermissionDenied("", "insufficient permissions for this queue")
	}

	return nil
}
