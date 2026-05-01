package evaluation

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	sq "github.com/Masterminds/squirrel"

	evalDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/infrastructure/db"
	"brokle/internal/infrastructure/db/gen"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

type datasetRepository struct {
	tm *db.TxManager
}

func NewDatasetRepository(tm *db.TxManager) evalDomain.DatasetRepository {
	return &datasetRepository{tm: tm}
}

func (r *datasetRepository) Create(ctx context.Context, d *evalDomain.Dataset) error {
	meta, err := marshalEvalJSON(d.Metadata)
	if err != nil {
		return err
	}
	if err := r.tm.Queries(ctx).CreateDataset(ctx, gen.CreateDatasetParams{
		ID:               d.ID,
		ProjectID:        d.ProjectID,
		Name:             d.Name,
		Description:      d.Description,
		Metadata:         meta,
		CurrentVersionID: d.CurrentVersionID,
	}); err != nil {
		if appErrors.IsUniqueViolation(err) {
			return appErrors.AlreadyExists("dataset", appErrors.WithOp("repo.dataset.create"))
		}
		return appErrors.Internal("create dataset", err, appErrors.WithOp("repo.dataset.create"))
	}
	return nil
}

func (r *datasetRepository) GetByID(ctx context.Context, id, projectID uuid.UUID) (*evalDomain.Dataset, error) {
	row, err := r.tm.Queries(ctx).GetDatasetByID(ctx, gen.GetDatasetByIDParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, appErrors.NotFound("dataset", appErrors.WithOp("repo.dataset.get_by_id"))
		}
		return nil, appErrors.Internal("get dataset", err, appErrors.WithOp("repo.dataset.get_by_id"))
	}
	return datasetFromRow(&row)
}

func (r *datasetRepository) GetByName(ctx context.Context, projectID uuid.UUID, name string) (*evalDomain.Dataset, error) {
	row, err := r.tm.Queries(ctx).GetDatasetByName(ctx, gen.GetDatasetByNameParams{
		ProjectID: projectID,
		Name:      name,
	})
	if err != nil {
		if db.IsNoRows(err) {
			return nil, nil
		}
		return nil, err
	}
	return datasetFromRow(&row)
}

func (r *datasetRepository) List(ctx context.Context, projectID uuid.UUID, filter *evalDomain.DatasetFilter, offset, limit int) ([]*evalDomain.Dataset, int64, error) {
	base := sq.Select().From("datasets").Where(sq.Eq{"project_id": projectID})
	if filter != nil && filter.Search != nil && *filter.Search != "" {
		p := "%" + strings.ToLower(*filter.Search) + "%"
		base = base.Where(sq.Or{
			sq.Expr("LOWER(name) LIKE ?", p),
			sq.Expr("LOWER(description) LIKE ?", p),
		})
	}

	cntSQL, cntArgs, err := base.Columns("COUNT(*)").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.tm.DB(ctx).QueryRow(ctx, cntSQL, cntArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	selSQL, selArgs, err := base.Columns(datasetColumns...).
		OrderBy("created_at DESC").Offset(uint64(offset)).Limit(uint64(limit)).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.tm.DB(ctx).Query(ctx, selSQL, selArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*evalDomain.Dataset, 0)
	for rows.Next() {
		d, err := scanDataset(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, rows.Err()
}

func (r *datasetRepository) Update(ctx context.Context, d *evalDomain.Dataset, projectID uuid.UUID) error {
	meta, err := marshalEvalJSON(d.Metadata)
	if err != nil {
		return err
	}
	n, err := r.tm.Queries(ctx).UpdateDataset(ctx, gen.UpdateDatasetParams{
		ID:               d.ID,
		ProjectID:        projectID,
		Name:             d.Name,
		Description:      d.Description,
		Metadata:         meta,
		CurrentVersionID: d.CurrentVersionID,
	})
	if err != nil {
		if appErrors.IsUniqueViolation(err) {
			return appErrors.AlreadyExists("dataset", appErrors.WithOp("repo.dataset.update"))
		}
		return appErrors.Internal("update dataset", err, appErrors.WithOp("repo.dataset.update"))
	}
	if n == 0 {
		return appErrors.NotFound("dataset", appErrors.WithOp("repo.dataset.update"))
	}
	return nil
}

func (r *datasetRepository) Delete(ctx context.Context, id, projectID uuid.UUID) error {
	n, err := r.tm.Queries(ctx).DeleteDataset(ctx, gen.DeleteDatasetParams{
		ID:        id,
		ProjectID: projectID,
	})
	if err != nil {
		return appErrors.Internal("delete dataset", err, appErrors.WithOp("repo.dataset.delete"))
	}
	if n == 0 {
		return appErrors.NotFound("dataset", appErrors.WithOp("repo.dataset.delete"))
	}
	return nil
}

func (r *datasetRepository) ExistsByName(ctx context.Context, projectID uuid.UUID, name string) (bool, error) {
	return r.tm.Queries(ctx).DatasetExistsByName(ctx, gen.DatasetExistsByNameParams{
		ProjectID: projectID,
		Name:      name,
	})
}

// ListWithFilters returns datasets with item counts, via a LEFT JOIN
// against a count subquery so sort-by-item_count works for pagination.
func (r *datasetRepository) ListWithFilters(
	ctx context.Context,
	projectID uuid.UUID,
	filter *evalDomain.DatasetFilter,
	params pagination.Params,
) ([]*evalDomain.DatasetWithItemCount, int64, error) {
	allowed := []string{"name", "created_at", "updated_at", "item_count"}
	params.SetDefaults("updated_at")
	if _, err := pagination.ValidateSortField(params.SortBy, allowed); err != nil {
		params.SortBy = "updated_at"
	}

	base := sq.Select().From("datasets d").Where(sq.Eq{"d.project_id": projectID})
	if filter != nil && filter.Search != nil && *filter.Search != "" {
		p := "%" + strings.ToLower(*filter.Search) + "%"
		base = base.Where(sq.Expr("LOWER(d.name) LIKE ?", p))
	}

	cntSQL, cntArgs, err := base.Columns("COUNT(*)").PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}
	var total int64
	if err := r.tm.DB(ctx).QueryRow(ctx, cntSQL, cntArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*evalDomain.DatasetWithItemCount{}, 0, nil
	}

	sortDir := strings.ToUpper(params.SortDir)
	if sortDir != "ASC" && sortDir != "DESC" {
		sortDir = "DESC"
	}
	sortField := params.SortBy
	if sortField != "item_count" {
		sortField = "d." + sortField
	}

	list := base.Columns(
		"d.id", "d.project_id", "d.name", "d.description", "d.metadata",
		"d.created_at", "d.updated_at", "d.current_version_id",
		"COALESCE(item_counts.count, 0) AS item_count",
	).LeftJoin(
		"(SELECT dataset_id, COUNT(*) AS count FROM dataset_items GROUP BY dataset_id) item_counts ON item_counts.dataset_id = d.id",
	).OrderBy(fmt.Sprintf("%s %s, d.id %s", sortField, sortDir, sortDir)).
		Offset(uint64(params.GetOffset())).Limit(uint64(params.Limit))

	selSQL, selArgs, err := list.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.tm.DB(ctx).Query(ctx, selSQL, selArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := make([]*evalDomain.DatasetWithItemCount, 0)
	for rows.Next() {
		var (
			d     evalDomain.Dataset
			count int64
			meta  []byte
		)
		if err := rows.Scan(
			&d.ID, &d.ProjectID, &d.Name, &d.Description, &meta,
			&d.CreatedAt, &d.UpdatedAt, &d.CurrentVersionID, &count,
		); err != nil {
			return nil, 0, err
		}
		if err := unmarshalEvalJSON(meta, &d.Metadata); err != nil {
			return nil, 0, err
		}
		out = append(out, &evalDomain.DatasetWithItemCount{Dataset: d, ItemCount: count})
	}
	return out, total, rows.Err()
}

var datasetColumns = []string{
	"id", "project_id", "name", "description", "metadata",
	"created_at", "updated_at", "current_version_id",
}

func scanDataset(row interface {
	Scan(dest ...any) error
}) (*evalDomain.Dataset, error) {
	var (
		d    evalDomain.Dataset
		meta []byte
	)
	if err := row.Scan(
		&d.ID, &d.ProjectID, &d.Name, &d.Description, &meta,
		&d.CreatedAt, &d.UpdatedAt, &d.CurrentVersionID,
	); err != nil {
		return nil, err
	}
	if err := unmarshalEvalJSON(meta, &d.Metadata); err != nil {
		return nil, err
	}
	return &d, nil
}

func datasetFromRow(row *gen.Dataset) (*evalDomain.Dataset, error) {
	d := &evalDomain.Dataset{
		ID:               row.ID,
		ProjectID:        row.ProjectID,
		Name:             row.Name,
		Description:      row.Description,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
		CurrentVersionID: row.CurrentVersionID,
	}
	if err := unmarshalEvalJSON(row.Metadata, &d.Metadata); err != nil {
		return nil, err
	}
	return d, nil
}
