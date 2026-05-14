package comment

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	commentDomain "brokle/internal/core/domain/comment"
	appErrors "brokle/pkg/errors"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
)

type commentRepository struct {
	tm *db.TxManager
}

func NewCommentRepository(tm *db.TxManager) commentDomain.Repository {
	return &commentRepository{tm: tm}
}

func (r *commentRepository) Create(ctx context.Context, c *commentDomain.Comment) error {
	if err := r.tm.Queries(ctx).CreateComment(ctx, gen.CreateCommentParams{
		ID:         c.ID,
		EntityType: gen.CommentEntityType(c.EntityType),
		EntityID:   c.EntityID,
		ProjectID:  c.ProjectID,
		Content:    c.Content,
		ParentID:   c.ParentID,
		CreatedBy:  c.CreatedBy,
		UpdatedBy:  c.UpdatedBy,
	}); err != nil {
		return appErrors.Internal("create comment", err, appErrors.WithOp("repo.comment.create"))
	}
	return nil
}

func (r *commentRepository) GetByID(ctx context.Context, id uuid.UUID) (*commentDomain.Comment, error) {
	row, err := r.tm.Queries(ctx).GetCommentByID(ctx, id)
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("comment", appErrors.WithOp("repo.comment.get_by_id"))
		}
		return nil, appErrors.Internal(fmt.Sprintf("get comment %s", id), err, appErrors.WithOp("repo.comment.get_by_id"))
	}
	return commentFromRow(&row), nil
}

func (r *commentRepository) GetByIDWithUser(ctx context.Context, id uuid.UUID) (*commentDomain.CommentWithUser, error) {
	c, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	ids := collectUserIDs([]*commentDomain.Comment{c})
	users, err := r.loadUsers(ctx, ids)
	if err != nil {
		return nil, err
	}
	return withUser(c, users), nil
}

func (r *commentRepository) Update(ctx context.Context, c *commentDomain.Comment) error {
	n, err := r.tm.Queries(ctx).UpdateComment(ctx, gen.UpdateCommentParams{
		ID:        c.ID,
		Content:   c.Content,
		UpdatedBy: c.UpdatedBy,
	})
	if err != nil {
		return appErrors.Internal(fmt.Sprintf("update comment %s", c.ID), err, appErrors.WithOp("repo.comment.update"))
	}
	if n == 0 {
		return appErrors.NotFound("comment", appErrors.WithOp("repo.comment.update"))
	}
	return nil
}

func (r *commentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	n, err := r.tm.Queries(ctx).SoftDeleteComment(ctx, id)
	if err != nil {
		return appErrors.Internal(fmt.Sprintf("delete comment %s", id), err, appErrors.WithOp("repo.comment.delete"))
	}
	if n == 0 {
		return appErrors.NotFound("comment", appErrors.WithOp("repo.comment.delete"))
	}
	return nil
}

func (r *commentRepository) ListByEntity(ctx context.Context, entityType commentDomain.EntityType, entityID string, projectID uuid.UUID) ([]*commentDomain.CommentWithUser, error) {
	rows, err := r.tm.Queries(ctx).ListCommentsByEntity(ctx, gen.ListCommentsByEntityParams{
		EntityType: gen.CommentEntityType(entityType),
		EntityID:   entityID,
		ProjectID:  projectID,
	})
	if err != nil {
		return nil, err
	}
	comments := make([]*commentDomain.Comment, 0, len(rows))
	for i := range rows {
		comments = append(comments, commentFromRow(&rows[i]))
	}
	users, err := r.loadUsers(ctx, collectUserIDs(comments))
	if err != nil {
		return nil, err
	}
	out := make([]*commentDomain.CommentWithUser, 0, len(comments))
	for _, c := range comments {
		out = append(out, withUser(c, users))
	}
	return out, nil
}

func (r *commentRepository) CountByEntity(ctx context.Context, entityType commentDomain.EntityType, entityID string, projectID uuid.UUID) (int64, error) {
	return r.tm.Queries(ctx).CountCommentsByEntity(ctx, gen.CountCommentsByEntityParams{
		EntityType: gen.CommentEntityType(entityType),
		EntityID:   entityID,
		ProjectID:  projectID,
	})
}

func (r *commentRepository) ListReplies(ctx context.Context, parentIDs []uuid.UUID) (map[string][]*commentDomain.CommentWithUser, error) {
	out := make(map[string][]*commentDomain.CommentWithUser, len(parentIDs))
	for _, id := range parentIDs {
		out[id.String()] = []*commentDomain.CommentWithUser{}
	}
	if len(parentIDs) == 0 {
		return out, nil
	}
	rows, err := r.tm.Queries(ctx).ListRepliesByParents(ctx, parentIDs)
	if err != nil {
		return nil, err
	}
	replies := make([]*commentDomain.Comment, 0, len(rows))
	for i := range rows {
		replies = append(replies, commentFromRow(&rows[i]))
	}
	users, err := r.loadUsers(ctx, collectUserIDs(replies))
	if err != nil {
		return nil, err
	}
	for _, c := range replies {
		key := c.ParentID.String()
		out[key] = append(out[key], withUser(c, users))
	}
	return out, nil
}

func (r *commentRepository) CountReplies(ctx context.Context, parentIDs []uuid.UUID) (map[string]int, error) {
	out := make(map[string]int, len(parentIDs))
	for _, id := range parentIDs {
		out[id.String()] = 0
	}
	if len(parentIDs) == 0 {
		return out, nil
	}
	rows, err := r.tm.Queries(ctx).CountRepliesByParents(ctx, parentIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		if row.ParentID != nil {
			out[row.ParentID.String()] = int(row.Count)
		}
	}
	return out, nil
}

// ----- helpers --------------------------------------------------------

func collectUserIDs(comments []*commentDomain.Comment) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	for _, c := range comments {
		if c.CreatedBy != nil {
			seen[*c.CreatedBy] = struct{}{}
		}
		if c.UpdatedBy != nil {
			seen[*c.UpdatedBy] = struct{}{}
		}
	}
	out := make([]uuid.UUID, 0, len(seen))
	for id := range seen {
		out = append(out, id)
	}
	return out
}

func (r *commentRepository) loadUsers(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*commentDomain.CommentUser, error) {
	out := make(map[uuid.UUID]*commentDomain.CommentUser)
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := r.tm.Queries(ctx).ListUsersForCommentEnrichment(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("load comment users: %w", err)
	}
	for _, row := range rows {
		out[row.ID] = &commentDomain.CommentUser{
			ID:        row.ID,
			Name:      row.FirstName + " " + row.LastName,
			Email:     row.Email,
			AvatarURL: row.AvatarUrl,
		}
	}
	return out, nil
}

func withUser(c *commentDomain.Comment, users map[uuid.UUID]*commentDomain.CommentUser) *commentDomain.CommentWithUser {
	cwu := &commentDomain.CommentWithUser{Comment: *c}
	if c.CreatedBy != nil {
		cwu.Author = users[*c.CreatedBy]
	}
	if c.UpdatedBy != nil {
		cwu.Editor = users[*c.UpdatedBy]
	}
	return cwu
}

// ----- gen ↔ domain boundary -----------------------------------------

func commentFromRow(row *gen.TraceComment) *commentDomain.Comment {
	return &commentDomain.Comment{
		ID:         row.ID,
		EntityType: commentDomain.EntityType(row.EntityType),
		EntityID:   row.EntityID,
		ProjectID:  row.ProjectID,
		ParentID:   row.ParentID,
		Content:    row.Content,
		CreatedBy:  row.CreatedBy,
		UpdatedBy:  row.UpdatedBy,
		CreatedAt:  row.CreatedAt,
		UpdatedAt:  row.UpdatedAt,
		DeletedAt:  row.DeletedAt,
	}
}
