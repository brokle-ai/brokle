package annotation

import (
	"context"

	"github.com/google/uuid"

	annotationDomain "brokle/internal/core/domain/annotation"
	appErrors "brokle/pkg/errors"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
)

type assignmentRepository struct {
	tm *db.TxManager
}

func NewAssignmentRepository(tm *db.TxManager) annotationDomain.AssignmentRepository {
	return &assignmentRepository{tm: tm}
}

func (r *assignmentRepository) Create(ctx context.Context, a *annotationDomain.QueueAssignment) error {
	if err := r.tm.Queries(ctx).CreateAnnotationQueueAssignment(ctx, gen.CreateAnnotationQueueAssignmentParams{
		ID:         a.ID,
		QueueID:    a.QueueID,
		UserID:     a.UserID,
		Role:       string(a.Role),
		AssignedBy: a.AssignedBy,
	}); err != nil {
		if appErrors.IsUniqueViolation(err) {
			return appErrors.AlreadyExists("annotation_assignment", appErrors.WithOp("repo.annotation.assignment.create"), appErrors.WithCause(err))
		}
		return appErrors.Internal("create annotation assignment", err, appErrors.WithOp("repo.annotation.assignment.create"))
	}
	return nil
}

func (r *assignmentRepository) Delete(ctx context.Context, queueID, userID uuid.UUID) error {
	n, err := r.tm.Queries(ctx).DeleteAnnotationQueueAssignment(ctx, gen.DeleteAnnotationQueueAssignmentParams{
		QueueID: queueID,
		UserID:  userID,
	})
	if err != nil {
		return appErrors.Internal("delete annotation assignment", err, appErrors.WithOp("repo.annotation.assignment.delete"))
	}
	if n == 0 {
		return appErrors.NotFound("annotation_assignment", appErrors.WithOp("repo.annotation.assignment.delete"))
	}
	return nil
}

func (r *assignmentRepository) GetByQueueAndUser(ctx context.Context, queueID, userID uuid.UUID) (*annotationDomain.QueueAssignment, error) {
	row, err := r.tm.Queries(ctx).GetAnnotationQueueAssignmentByQueueAndUser(ctx, gen.GetAnnotationQueueAssignmentByQueueAndUserParams{
		QueueID: queueID,
		UserID:  userID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("annotation_assignment", appErrors.WithOp("repo.annotation.assignment.get_by_queue_and_user"))
		}
		return nil, appErrors.Internal("get annotation assignment", err, appErrors.WithOp("repo.annotation.assignment.get_by_queue_and_user"))
	}
	return assignmentFromRow(&row), nil
}

func (r *assignmentRepository) List(ctx context.Context, queueID uuid.UUID) ([]*annotationDomain.QueueAssignment, error) {
	rows, err := r.tm.Queries(ctx).ListAnnotationQueueAssignmentsByQueue(ctx, queueID)
	if err != nil {
		return nil, err
	}
	out := make([]*annotationDomain.QueueAssignment, 0, len(rows))
	for i := range rows {
		out = append(out, assignmentFromRow(&rows[i]))
	}
	return out, nil
}

func (r *assignmentRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*annotationDomain.QueueAssignment, error) {
	rows, err := r.tm.Queries(ctx).ListAnnotationQueueAssignmentsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]*annotationDomain.QueueAssignment, 0, len(rows))
	for i := range rows {
		out = append(out, assignmentFromRow(&rows[i]))
	}
	return out, nil
}

func (r *assignmentRepository) IsAssigned(ctx context.Context, queueID, userID uuid.UUID) (bool, error) {
	return r.tm.Queries(ctx).AnnotationQueueAssignmentExists(ctx, gen.AnnotationQueueAssignmentExistsParams{
		QueueID: queueID,
		UserID:  userID,
	})
}

// HasRole checks whether the user's assigned role meets or exceeds the
// minimum. Role hierarchy: admin > reviewer > annotator.
func (r *assignmentRepository) HasRole(ctx context.Context, queueID, userID uuid.UUID, minRole annotationDomain.AssignmentRole) (bool, error) {
	a, err := r.GetByQueueAndUser(ctx, queueID, userID)
	if err != nil {
		if appErrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return roleAtLeast(a.Role, minRole), nil
}

func roleAtLeast(actual, minimum annotationDomain.AssignmentRole) bool {
	level := map[annotationDomain.AssignmentRole]int{
		annotationDomain.RoleAnnotator: 1,
		annotationDomain.RoleReviewer:  2,
		annotationDomain.RoleAdmin:     3,
	}
	a, ok := level[actual]
	if !ok {
		return false
	}
	m, ok := level[minimum]
	if !ok {
		return false
	}
	return a >= m
}

func assignmentFromRow(row *gen.AnnotationQueueAssignment) *annotationDomain.QueueAssignment {
	return &annotationDomain.QueueAssignment{
		ID:         row.ID,
		QueueID:    row.QueueID,
		UserID:     row.UserID,
		Role:       annotationDomain.AssignmentRole(row.Role),
		AssignedAt: row.AssignedAt,
		AssignedBy: row.AssignedBy,
	}
}
