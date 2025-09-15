package observability

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"brokle/internal/core/domain/observability"
	"brokle/pkg/ulid"
)

// QualityScoreRepository implements the observability.QualityScoreRepository interface
type QualityScoreRepository struct {
	db *gorm.DB
}

// NewQualityScoreRepository creates a new quality score repository instance
func NewQualityScoreRepository(db *gorm.DB) *QualityScoreRepository {
	return &QualityScoreRepository{
		db: db,
	}
}

// Create creates a new quality score in the database
func (r *QualityScoreRepository) Create(ctx context.Context, qualityScore *observability.QualityScore) error {
	if qualityScore.ID.IsZero() {
		qualityScore.ID = ulid.New()
	}
	qualityScore.CreatedAt = time.Now()
	qualityScore.UpdatedAt = time.Now()

	return r.db.WithContext(ctx).Create(qualityScore).Error
}

// GetByID retrieves a quality score by its ID
func (r *QualityScoreRepository) GetByID(ctx context.Context, id ulid.ULID) (*observability.QualityScore, error) {
	var qualityScore observability.QualityScore
	err := r.db.WithContext(ctx).First(&qualityScore, "id = ?", id.String()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get quality score by id %s: %w", id.String(), observability.ErrQualityScoreNotFound)
		}
		return nil, fmt.Errorf("get quality score by id %s: %w", id.String(), err)
	}
	return &qualityScore, nil
}

// Update updates an existing quality score
func (r *QualityScoreRepository) Update(ctx context.Context, qualityScore *observability.QualityScore) error {
	qualityScore.UpdatedAt = time.Now()

	err := r.db.WithContext(ctx).Save(qualityScore).Error
	if err != nil {
		return fmt.Errorf("update quality score %s: %w", qualityScore.ID.String(), err)
	}
	return nil
}

// Delete deletes a quality score by its ID
func (r *QualityScoreRepository) Delete(ctx context.Context, id ulid.ULID) error {
	err := r.db.WithContext(ctx).Delete(&observability.QualityScore{}, "id = ?", id.String()).Error
	if err != nil {
		return fmt.Errorf("delete quality score %s: %w", id.String(), err)
	}
	return nil
}

// List retrieves quality scores with filtering and pagination
func (r *QualityScoreRepository) List(ctx context.Context, filter observability.QualityScoreFilter) ([]*observability.QualityScore, error) {
	var qualityScores []*observability.QualityScore

	query := r.db.WithContext(ctx).Model(&observability.QualityScore{})

	// Apply filters
	if filter.TraceID != nil && !filter.TraceID.IsZero() {
		query = query.Where("trace_id = ?", filter.TraceID.String())
	}

	if filter.ObservationID != nil && !filter.ObservationID.IsZero() {
		query = query.Where("observation_id = ?", filter.ObservationID.String())
	}

	if filter.ScoreName != nil && *filter.ScoreName != "" {
		query = query.Where("score_name = ?", *filter.ScoreName)
	}

	if filter.DataType != nil && *filter.DataType != "" {
		query = query.Where("data_type = ?", *filter.DataType)
	}

	if filter.Source != nil && *filter.Source != "" {
		query = query.Where("source = ?", *filter.Source)
	}

	if filter.AuthorUserID != nil && !filter.AuthorUserID.IsZero() {
		query = query.Where("author_user_id = ?", filter.AuthorUserID.String())
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	// Apply sorting
	if filter.SortBy != "" {
		order := "ASC"
		if filter.SortOrder != "" && strings.ToUpper(filter.SortOrder) == "DESC" {
			order = "DESC"
		}
		query = query.Order(fmt.Sprintf("%s %s", filter.SortBy, order))
	} else {
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	err := query.Find(&qualityScores).Error
	if err != nil {
		return nil, fmt.Errorf("list quality scores: %w", err)
	}

	return qualityScores, nil
}

// Count returns the total number of quality scores matching the filter
func (r *QualityScoreRepository) Count(ctx context.Context, filter observability.QualityScoreFilter) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&observability.QualityScore{})

	// Apply the same filters as List (without pagination)
	if filter.TraceID != nil && !filter.TraceID.IsZero() {
		query = query.Where("trace_id = ?", filter.TraceID.String())
	}

	if filter.ObservationID != nil && !filter.ObservationID.IsZero() {
		query = query.Where("observation_id = ?", filter.ObservationID.String())
	}

	if filter.ScoreName != nil && *filter.ScoreName != "" {
		query = query.Where("score_name = ?", *filter.ScoreName)
	}

	if filter.DataType != nil && *filter.DataType != "" {
		query = query.Where("data_type = ?", *filter.DataType)
	}

	if filter.Source != nil && *filter.Source != "" {
		query = query.Where("source = ?", *filter.Source)
	}

	if filter.AuthorUserID != nil && !filter.AuthorUserID.IsZero() {
		query = query.Where("author_user_id = ?", filter.AuthorUserID.String())
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	err := query.Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("count quality scores: %w", err)
	}

	return count, nil
}

// GetByTraceID retrieves all quality scores for a specific trace
func (r *QualityScoreRepository) GetByTraceID(ctx context.Context, traceID ulid.ULID) ([]*observability.QualityScore, error) {
	var qualityScores []*observability.QualityScore

	err := r.db.WithContext(ctx).
		Where("trace_id = ?", traceID.String()).
		Order("created_at DESC").
		Find(&qualityScores).Error

	if err != nil {
		return nil, fmt.Errorf("get quality scores by trace id %s: %w", traceID.String(), err)
	}

	return qualityScores, nil
}

// GetByObservationID retrieves all quality scores for a specific observation
func (r *QualityScoreRepository) GetByObservationID(ctx context.Context, observationID ulid.ULID) ([]*observability.QualityScore, error) {
	var qualityScores []*observability.QualityScore

	err := r.db.WithContext(ctx).
		Where("observation_id = ?", observationID.String()).
		Order("created_at DESC").
		Find(&qualityScores).Error

	if err != nil {
		return nil, fmt.Errorf("get quality scores by observation id %s: %w", observationID.String(), err)
	}

	return qualityScores, nil
}

// GetLatestByScoreName retrieves the latest quality score for a specific score name and trace/observation
func (r *QualityScoreRepository) GetLatestByScoreName(ctx context.Context, traceID, observationID *ulid.ULID, scoreName string) (*observability.QualityScore, error) {
	var qualityScore observability.QualityScore

	query := r.db.WithContext(ctx).Where("score_name = ?", scoreName)

	if traceID != nil && !traceID.IsZero() {
		query = query.Where("trace_id = ?", traceID.String())
	}

	if observationID != nil && !observationID.IsZero() {
		query = query.Where("observation_id = ?", observationID.String())
	}

	err := query.Order("created_at DESC").First(&qualityScore).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get latest quality score by name %s: %w", scoreName, observability.ErrQualityScoreNotFound)
		}
		return nil, fmt.Errorf("get latest quality score by name %s: %w", scoreName, err)
	}

	return &qualityScore, nil
}

// DeleteByTraceID deletes all quality scores for a specific trace
func (r *QualityScoreRepository) DeleteByTraceID(ctx context.Context, traceID ulid.ULID) error {
	err := r.db.WithContext(ctx).Delete(&observability.QualityScore{}, "trace_id = ?", traceID.String()).Error
	if err != nil {
		return fmt.Errorf("delete quality scores by trace id %s: %w", traceID.String(), err)
	}
	return nil
}

// DeleteByObservationID deletes all quality scores for a specific observation
func (r *QualityScoreRepository) DeleteByObservationID(ctx context.Context, observationID ulid.ULID) error {
	err := r.db.WithContext(ctx).Delete(&observability.QualityScore{}, "observation_id = ?", observationID.String()).Error
	if err != nil {
		return fmt.Errorf("delete quality scores by observation id %s: %w", observationID.String(), err)
	}
	return nil
}

// GetAggregatedScores retrieves aggregated quality scores for analytics
func (r *QualityScoreRepository) GetAggregatedScores(ctx context.Context, filter observability.QualityScoreFilter) (map[string]observability.QualityScoreAggregation, error) {
	var results []struct {
		ScoreName string
		DataType  string
		Count     int64
		AvgValue  sql.NullFloat64
		MinValue  sql.NullFloat64
		MaxValue  sql.NullFloat64
	}

	query := r.db.WithContext(ctx).
		Model(&observability.QualityScore{}).
		Select("score_name, data_type, COUNT(*) as count, AVG(score_value) as avg_value, MIN(score_value) as min_value, MAX(score_value) as max_value").
		Group("score_name, data_type")

	// Apply filters (reuse logic from List)
	if filter.TraceID != nil && !filter.TraceID.IsZero() {
		query = query.Where("trace_id = ?", filter.TraceID.String())
	}

	if filter.DataType != nil && *filter.DataType != "" {
		query = query.Where("data_type = ?", *filter.DataType)
	}

	if filter.Source != nil && *filter.Source != "" {
		query = query.Where("source = ?", *filter.Source)
	}

	if filter.StartTime != nil {
		query = query.Where("created_at >= ?", *filter.StartTime)
	}

	if filter.EndTime != nil {
		query = query.Where("created_at <= ?", *filter.EndTime)
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("get aggregated quality scores: %w", err)
	}

	aggregations := make(map[string]observability.QualityScoreAggregation)
	for _, result := range results {
		agg := observability.QualityScoreAggregation{
			ScoreName: result.ScoreName,
			DataType:  observability.ScoreDataType(result.DataType),
			Count:     result.Count,
		}

		if result.AvgValue.Valid {
			agg.AvgValue = &result.AvgValue.Float64
		}
		if result.MinValue.Valid {
			agg.MinValue = &result.MinValue.Float64
		}
		if result.MaxValue.Valid {
			agg.MaxValue = &result.MaxValue.Float64
		}

		aggregations[result.ScoreName] = agg
	}

	return aggregations, nil
}

// GetByScoreName retrieves quality scores by score name with pagination
func (r *QualityScoreRepository) GetByScoreName(ctx context.Context, scoreName string, limit, offset int) ([]*observability.QualityScore, error) {
	var qualityScores []*observability.QualityScore

	query := r.db.WithContext(ctx).Where("score_name = ?", scoreName).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&qualityScores).Error
	if err != nil {
		return nil, fmt.Errorf("get quality scores by name %s: %w", scoreName, err)
	}

	return qualityScores, nil
}

// GetBySource retrieves quality scores by source with pagination
func (r *QualityScoreRepository) GetBySource(ctx context.Context, source observability.ScoreSource, limit, offset int) ([]*observability.QualityScore, error) {
	var qualityScores []*observability.QualityScore

	query := r.db.WithContext(ctx).Where("source = ?", source).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&qualityScores).Error
	if err != nil {
		return nil, fmt.Errorf("get quality scores by source %s: %w", source, err)
	}

	return qualityScores, nil
}

// GetByEvaluator retrieves quality scores by evaluator name with pagination
func (r *QualityScoreRepository) GetByEvaluator(ctx context.Context, evaluatorName string, limit, offset int) ([]*observability.QualityScore, error) {
	var qualityScores []*observability.QualityScore

	query := r.db.WithContext(ctx).Where("evaluator_name = ?", evaluatorName).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&qualityScores).Error
	if err != nil {
		return nil, fmt.Errorf("get quality scores by evaluator %s: %w", evaluatorName, err)
	}

	return qualityScores, nil
}

// GetByTraceAndScoreName retrieves a quality score by trace ID and score name
func (r *QualityScoreRepository) GetByTraceAndScoreName(ctx context.Context, traceID ulid.ULID, scoreName string) (*observability.QualityScore, error) {
	var qualityScore observability.QualityScore

	err := r.db.WithContext(ctx).
		Where("trace_id = ? AND score_name = ?", traceID.String(), scoreName).
		Order("created_at DESC").
		First(&qualityScore).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get quality score by trace %s and name %s: %w", traceID.String(), scoreName, observability.ErrQualityScoreNotFound)
		}
		return nil, fmt.Errorf("get quality score by trace %s and name %s: %w", traceID.String(), scoreName, err)
	}

	return &qualityScore, nil
}

// GetByObservationAndScoreName retrieves a quality score by observation ID and score name
func (r *QualityScoreRepository) GetByObservationAndScoreName(ctx context.Context, observationID ulid.ULID, scoreName string) (*observability.QualityScore, error) {
	var qualityScore observability.QualityScore

	err := r.db.WithContext(ctx).
		Where("observation_id = ? AND score_name = ?", observationID.String(), scoreName).
		Order("created_at DESC").
		First(&qualityScore).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("get quality score by observation %s and name %s: %w", observationID.String(), scoreName, observability.ErrQualityScoreNotFound)
		}
		return nil, fmt.Errorf("get quality score by observation %s and name %s: %w", observationID.String(), scoreName, err)
	}

	return &qualityScore, nil
}

// CreateBatch creates multiple quality scores in a batch
func (r *QualityScoreRepository) CreateBatch(ctx context.Context, scores []*observability.QualityScore) error {
	if len(scores) == 0 {
		return nil
	}

	// Set IDs and timestamps for all scores
	for _, score := range scores {
		if score.ID.IsZero() {
			score.ID = ulid.New()
		}
		score.CreatedAt = time.Now()
		score.UpdatedAt = time.Now()
	}

	err := r.db.WithContext(ctx).CreateInBatches(scores, 100).Error
	if err != nil {
		return fmt.Errorf("create quality scores batch: %w", err)
	}

	return nil
}

// UpdateBatch updates multiple quality scores in a batch
func (r *QualityScoreRepository) UpdateBatch(ctx context.Context, scores []*observability.QualityScore) error {
	if len(scores) == 0 {
		return nil
	}

	// Update timestamps
	for _, score := range scores {
		score.UpdatedAt = time.Now()
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, score := range scores {
			if err := tx.Save(score).Error; err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("update quality scores batch: %w", err)
	}

	return nil
}

// DeleteBatch deletes multiple quality scores by their IDs
func (r *QualityScoreRepository) DeleteBatch(ctx context.Context, ids []ulid.ULID) error {
	if len(ids) == 0 {
		return nil
	}

	stringIDs := make([]string, len(ids))
	for i, id := range ids {
		stringIDs[i] = id.String()
	}

	err := r.db.WithContext(ctx).Delete(&observability.QualityScore{}, "id IN ?", stringIDs).Error
	if err != nil {
		return fmt.Errorf("delete quality scores batch: %w", err)
	}

	return nil
}

// GetAverageScoreByName calculates average score for a specific score name with filters
func (r *QualityScoreRepository) GetAverageScoreByName(ctx context.Context, scoreName string, filter *observability.QualityScoreFilter) (float64, error) {
	var result sql.NullFloat64

	query := r.db.WithContext(ctx).
		Model(&observability.QualityScore{}).
		Where("score_name = ? AND data_type = ?", scoreName, observability.ScoreDataTypeNumeric).
		Select("AVG(score_value)")

	// Apply filters if provided
	if filter != nil {
		if filter.TraceID != nil && !filter.TraceID.IsZero() {
			query = query.Where("trace_id = ?", filter.TraceID.String())
		}
		if filter.ObservationID != nil && !filter.ObservationID.IsZero() {
			query = query.Where("observation_id = ?", filter.ObservationID.String())
		}
		if filter.Source != nil {
			query = query.Where("source = ?", *filter.Source)
		}
		if filter.StartTime != nil {
			query = query.Where("created_at >= ?", *filter.StartTime)
		}
		if filter.EndTime != nil {
			query = query.Where("created_at <= ?", *filter.EndTime)
		}
	}

	err := query.Scan(&result).Error
	if err != nil {
		return 0, fmt.Errorf("get average score by name %s: %w", scoreName, err)
	}

	if !result.Valid {
		return 0, nil
	}

	return result.Float64, nil
}

// GetScoreDistribution gets distribution of scores for a specific score name
func (r *QualityScoreRepository) GetScoreDistribution(ctx context.Context, scoreName string, filter *observability.QualityScoreFilter) (map[string]int, error) {
	var results []struct {
		Value string
		Count int
	}

	query := r.db.WithContext(ctx).
		Model(&observability.QualityScore{}).
		Where("score_name = ?", scoreName).
		Select("CASE WHEN data_type = ? THEN string_value ELSE CAST(score_value AS TEXT) END as value, COUNT(*) as count", observability.ScoreDataTypeCategorical).
		Group("value")

	// Apply filters if provided
	if filter != nil {
		if filter.TraceID != nil && !filter.TraceID.IsZero() {
			query = query.Where("trace_id = ?", filter.TraceID.String())
		}
		if filter.ObservationID != nil && !filter.ObservationID.IsZero() {
			query = query.Where("observation_id = ?", filter.ObservationID.String())
		}
		if filter.DataType != nil {
			query = query.Where("data_type = ?", *filter.DataType)
		}
		if filter.Source != nil {
			query = query.Where("source = ?", *filter.Source)
		}
		if filter.StartTime != nil {
			query = query.Where("created_at >= ?", *filter.StartTime)
		}
		if filter.EndTime != nil {
			query = query.Where("created_at <= ?", *filter.EndTime)
		}
	}

	err := query.Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("get score distribution for %s: %w", scoreName, err)
	}

	distribution := make(map[string]int)
	for _, result := range results {
		distribution[result.Value] = result.Count
	}

	return distribution, nil
}

// GetScoresByTimeRange retrieves quality scores within a time range with filters
func (r *QualityScoreRepository) GetScoresByTimeRange(ctx context.Context, filter *observability.QualityScoreFilter, startTime, endTime time.Time) ([]*observability.QualityScore, error) {
	var qualityScores []*observability.QualityScore

	query := r.db.WithContext(ctx).
		Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Order("created_at DESC")

	// Apply additional filters if provided
	if filter != nil {
		if filter.TraceID != nil && !filter.TraceID.IsZero() {
			query = query.Where("trace_id = ?", filter.TraceID.String())
		}
		if filter.ObservationID != nil && !filter.ObservationID.IsZero() {
			query = query.Where("observation_id = ?", filter.ObservationID.String())
		}
		if filter.ScoreName != nil && *filter.ScoreName != "" {
			query = query.Where("score_name = ?", *filter.ScoreName)
		}
		if filter.DataType != nil {
			query = query.Where("data_type = ?", *filter.DataType)
		}
		if filter.Source != nil {
			query = query.Where("source = ?", *filter.Source)
		}
		if filter.AuthorUserID != nil && !filter.AuthorUserID.IsZero() {
			query = query.Where("author_user_id = ?", filter.AuthorUserID.String())
		}
		if filter.MinScore != nil {
			query = query.Where("score_value >= ?", *filter.MinScore)
		}
		if filter.MaxScore != nil {
			query = query.Where("score_value <= ?", *filter.MaxScore)
		}

		// Apply pagination
		if filter.Limit > 0 {
			query = query.Limit(filter.Limit)
		}
		if filter.Offset > 0 {
			query = query.Offset(filter.Offset)
		}
	}

	err := query.Find(&qualityScores).Error
	if err != nil {
		return nil, fmt.Errorf("get scores by time range: %w", err)
	}

	return qualityScores, nil
}