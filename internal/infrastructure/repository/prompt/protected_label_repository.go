package prompt

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	promptDomain "brokle/internal/core/domain/prompt"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
)

type protectedLabelRepository struct {
	tm *db.TxManager
}

func NewProtectedLabelRepository(tm *db.TxManager) promptDomain.ProtectedLabelRepository {
	return &protectedLabelRepository{tm: tm}
}

func (r *protectedLabelRepository) Create(ctx context.Context, l *promptDomain.ProtectedLabel) error {
	if err := r.tm.Queries(ctx).CreateProtectedPromptLabel(ctx, gen.CreateProtectedPromptLabelParams{
		ID:        l.ID,
		ProjectID: l.ProjectID,
		LabelName: l.LabelName,
		CreatedBy: l.CreatedBy,
	}); err != nil {
		if appErrors.IsUniqueViolation(err) {
			return appErrors.AlreadyExists("prompt_label", appErrors.WithMessage("protected label already exists"), appErrors.WithOp("repo.protected_prompt_label.create"))
		}
		return fmt.Errorf("create protected label %s: %w", l.LabelName, err)
	}
	return nil
}

func (r *protectedLabelRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.tm.Queries(ctx).DeleteProtectedPromptLabel(ctx, id)
}

func (r *protectedLabelRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]*promptDomain.ProtectedLabel, error) {
	rows, err := r.tm.Queries(ctx).ListProtectedPromptLabelsByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]*promptDomain.ProtectedLabel, 0, len(rows))
	for i := range rows {
		out = append(out, protectedLabelFromRow(&rows[i]))
	}
	return out, nil
}

func (r *protectedLabelRepository) IsProtected(ctx context.Context, projectID uuid.UUID, labelName string) (bool, error) {
	return r.tm.Queries(ctx).ProtectedPromptLabelExists(ctx, gen.ProtectedPromptLabelExistsParams{
		ProjectID: projectID,
		LabelName: labelName,
	})
}

// SetProtectedLabels atomically replaces the full protected-label set
// for a project: delete-then-insert inside a single transaction.
func (r *protectedLabelRepository) SetProtectedLabels(ctx context.Context, projectID uuid.UUID, labels []string, createdBy *uuid.UUID) error {
	return r.tm.WithinTransaction(ctx, func(ctx context.Context) error {
		q := r.tm.Queries(ctx)
		if err := q.DeleteProtectedPromptLabelsByProject(ctx, projectID); err != nil {
			return err
		}
		for _, name := range labels {
			l := promptDomain.NewProtectedLabel(projectID, name, createdBy)
			if err := q.CreateProtectedPromptLabel(ctx, gen.CreateProtectedPromptLabelParams{
				ID:        l.ID,
				ProjectID: l.ProjectID,
				LabelName: l.LabelName,
				CreatedBy: l.CreatedBy,
			}); err != nil {
				return err
			}
		}
		return nil
	})
}

func protectedLabelFromRow(row *gen.PromptProtectedLabel) *promptDomain.ProtectedLabel {
	return &promptDomain.ProtectedLabel{
		ID:        row.ID,
		ProjectID: row.ProjectID,
		LabelName: row.LabelName,
		CreatedBy: row.CreatedBy,
		CreatedAt: row.CreatedAt,
	}
}
