package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/preview"
	"brokle/pkg/ulid"
)

// ObservationService implements business logic for OTEL observation (span) management
type ObservationService struct {
	observationRepo    observability.ObservationRepository
	traceRepo          observability.TraceRepository
	scoreRepo          observability.ScoreRepository
	blobStorageService *BlobStorageService
	logger             *logrus.Logger
}

// NewObservationService creates a new observation service instance
func NewObservationService(
	observationRepo observability.ObservationRepository,
	traceRepo observability.TraceRepository,
	scoreRepo observability.ScoreRepository,
	blobStorageService *BlobStorageService,
	logger *logrus.Logger,
) *ObservationService {
	return &ObservationService{
		observationRepo:    observationRepo,
		traceRepo:          traceRepo,
		scoreRepo:          scoreRepo,
		blobStorageService: blobStorageService,
		logger:             logger,
	}
}

// CreateObservation creates a new OTEL observation (span) with validation
func (s *ObservationService) CreateObservation(ctx context.Context, obs *observability.Observation) error {
	// Validate required fields
	if obs.TraceID == "" {
		return appErrors.NewValidationError("trace_id is required", "observation must be linked to a trace")
	}
	if obs.ProjectID == "" {
		return appErrors.NewValidationError("project_id is required", "observation must have a valid project_id")
	}
	if obs.Name == "" {
		return appErrors.NewValidationError("name is required", "observation name cannot be empty")
	}
	if obs.ID == "" {
		return appErrors.NewValidationError("id is required", "observation must have OTEL span_id")
	}

	// Validate OTEL span_id format (16 hex chars)
	if len(obs.ID) != 16 {
		return appErrors.NewValidationError("invalid span_id", "OTEL span_id must be 16 hex characters")
	}

	// Validate trace exists
	_, err := s.traceRepo.GetByID(ctx, obs.TraceID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("trace %s", obs.TraceID))
	}

	// Note: We do NOT validate parent observation existence here
	// Async processing means parent may arrive after children - eventual consistency model
	// Database foreign key relationship will be preserved

	// Set defaults
	if obs.StatusCode == "" {
		obs.StatusCode = observability.StatusCodeUnset
	}
	if obs.SpanKind == "" {
		obs.SpanKind = string(observability.SpanKindInternal)
	}
	if obs.Type == "" {
		obs.Type = observability.ObservationTypeSpan
	}
	if obs.Level == "" {
		obs.Level = observability.ObservationLevelDefault
	}
	if obs.Attributes == "" {
		obs.Attributes = "{}"
	}
	if obs.Provider == "" {
		obs.Provider = ""
	}
	if obs.CreatedAt.IsZero() {
		obs.CreatedAt = time.Now()
	}

	// Initialize maps if nil
	if obs.Metadata == nil {
		obs.Metadata = make(map[string]string)
	}
	if obs.ProvidedUsageDetails == nil {
		obs.ProvidedUsageDetails = make(map[string]uint64)
	}
	if obs.UsageDetails == nil {
		obs.UsageDetails = make(map[string]uint64)
	}
	if obs.ProvidedCostDetails == nil {
		obs.ProvidedCostDetails = make(map[string]float64)
	}
	if obs.CostDetails == nil {
		obs.CostDetails = make(map[string]float64)
	}

	// Calculate duration if not set
	obs.CalculateDuration()

	// Auto-offload large payloads to S3 with type-aware preview generation
	// CRITICAL: Preview is ALWAYS generated and stored, regardless of offloading
	if s.blobStorageService != nil {
		// Handle input: Always generate preview, conditionally offload
		if obs.Input != nil && *obs.Input != "" {
			// Generate type-aware preview (always)
			inputPreview := preview.GeneratePreview(*obs.Input)
			obs.InputPreview = &inputPreview

			// Check if payload is large enough to offload
			if s.blobStorageService.ShouldOffload(*obs.Input) {
				// Large payload - upload to S3
				blob, _, err := s.blobStorageService.UploadToS3WithPreview(
					ctx,
					*obs.Input,
					obs.ProjectID,
					"observation",
					obs.ID,
					ulid.New().String(),
				)
				if err != nil {
					s.logger.WithError(err).WithField("observation_id", obs.ID).Warn("Failed to upload input to S3, storing inline")
				} else {
					s.logger.WithFields(logrus.Fields{
						"observation_id": obs.ID,
						"blob_id":        blob.ID,
						"original_size":  len(*obs.Input),
						"preview_size":   len(inputPreview),
					}).Info("Offloaded input to S3 with preview")
					obs.InputBlobStorageID = &blob.ID
					obs.Input = nil // NULL in ClickHouse
					// InputPreview already set above
				}
			}
			// else: small payload stays inline, preview still populated
		}

		// Handle output: Always generate preview, conditionally offload
		if obs.Output != nil && *obs.Output != "" {
			// Generate type-aware preview (always)
			outputPreview := preview.GeneratePreview(*obs.Output)
			obs.OutputPreview = &outputPreview

			// Check if payload is large enough to offload
			if s.blobStorageService.ShouldOffload(*obs.Output) {
				// Large payload - upload to S3
				blob, _, err := s.blobStorageService.UploadToS3WithPreview(
					ctx,
					*obs.Output,
					obs.ProjectID,
					"observation",
					obs.ID,
					ulid.New().String(),
				)
				if err != nil {
					s.logger.WithError(err).WithField("observation_id", obs.ID).Warn("Failed to upload output to S3, storing inline")
				} else {
					s.logger.WithFields(logrus.Fields{
						"observation_id": obs.ID,
						"blob_id":        blob.ID,
						"original_size":  len(*obs.Output),
						"preview_size":   len(outputPreview),
					}).Info("Offloaded output to S3 with preview")
					obs.OutputBlobStorageID = &blob.ID
					obs.Output = nil // NULL in ClickHouse
					// OutputPreview already set above
				}
			}
			// else: small payload stays inline, preview still populated
		}
	}

	// Create observation (with blob references if offloaded)
	if err := s.observationRepo.Create(ctx, obs); err != nil {
		return appErrors.NewInternalError("failed to create observation", err)
	}

	return nil
}

// UpdateObservation updates an existing observation
func (s *ObservationService) UpdateObservation(ctx context.Context, obs *observability.Observation) error {
	// Validate observation exists
	existing, err := s.observationRepo.GetByID(ctx, obs.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", obs.ID))
		}
		return appErrors.NewInternalError("failed to get observation", err)
	}

	// Merge non-zero fields from incoming observation into existing
	mergeObservationFields(existing, obs)

	// Preserve version for increment in repository layer
	existing.Version = existing.Version

	// Calculate duration if end time updated
	existing.CalculateDuration()

	// Auto-offload large payloads on update with preview generation
	if s.blobStorageService != nil {
		// Handle output if newly added or updated
		if existing.Output != nil && *existing.Output != "" && existing.OutputBlobStorageID == nil {
			// Generate type-aware preview (always)
			outputPreview := preview.GeneratePreview(*existing.Output)
			existing.OutputPreview = &outputPreview

			// Check if payload is large enough to offload
			if s.blobStorageService.ShouldOffload(*existing.Output) {
				blob, _, err := s.blobStorageService.UploadToS3WithPreview(
					ctx,
					*existing.Output,
					existing.ProjectID,
					"observation",
					existing.ID,
					ulid.New().String(),
				)
				if err != nil {
					s.logger.WithError(err).WithField("observation_id", existing.ID).Warn("Failed to upload output to S3 on update, storing inline")
				} else {
					s.logger.WithFields(logrus.Fields{
						"observation_id": existing.ID,
						"blob_id":        blob.ID,
						"original_size":  len(*existing.Output),
						"preview_size":   len(outputPreview),
					}).Info("Offloaded output to S3 on update with preview")
					existing.OutputBlobStorageID = &blob.ID
					existing.Output = nil
					// OutputPreview already set above
				}
			}
			// else: small payload stays inline, preview still populated
		}
	}

	// Update observation
	if err := s.observationRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update observation", err)
	}

	return nil
}

// SetObservationCost sets cost details for an observation
func (s *ObservationService) SetObservationCost(ctx context.Context, observationID string, inputCost, outputCost float64) error {
	obs, err := s.observationRepo.GetByID(ctx, observationID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", observationID))
	}

	// Set cost details (updates both Maps and Brokle extension fields)
	obs.SetCostDetails(inputCost, outputCost)

	// Update observation
	if err := s.observationRepo.Update(ctx, obs); err != nil {
		return appErrors.NewInternalError("failed to update observation cost", err)
	}

	return nil
}

// SetObservationUsage sets usage details for an observation
func (s *ObservationService) SetObservationUsage(ctx context.Context, observationID string, promptTokens, completionTokens uint32) error {
	obs, err := s.observationRepo.GetByID(ctx, observationID)
	if err != nil {
		return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", observationID))
	}

	// Set usage details (populates usage_details Map)
	obs.SetUsageDetails(uint64(promptTokens), uint64(completionTokens))

	// Update observation
	if err := s.observationRepo.Update(ctx, obs); err != nil {
		return appErrors.NewInternalError("failed to update observation usage", err)
	}

	return nil
}

// mergeObservationFields merges non-zero fields from src into dst
func mergeObservationFields(dst *observability.Observation, src *observability.Observation) {
	// Update optional fields only if non-zero
	if src.Name != "" {
		dst.Name = src.Name
	}
	if src.SpanKind != "" {
		dst.SpanKind = src.SpanKind
	}
	if src.Type != "" {
		dst.Type = src.Type
	}
	if !src.StartTime.IsZero() {
		dst.StartTime = src.StartTime
	}
	if src.EndTime != nil {
		dst.EndTime = src.EndTime
	}
	if src.StatusCode != "" {
		dst.StatusCode = src.StatusCode
	}
	if src.StatusMessage != nil {
		dst.StatusMessage = src.StatusMessage
	}
	if src.Attributes != "" {
		dst.Attributes = src.Attributes
	}
	if src.Input != nil {
		dst.Input = src.Input
	}
	if src.Output != nil {
		dst.Output = src.Output
	}
	if src.Metadata != nil {
		dst.Metadata = src.Metadata
	}
	if src.Level != "" {
		dst.Level = src.Level
	}

	// Model fields
	if src.ModelName != nil {
		dst.ModelName = src.ModelName
	}
	if src.Provider != "" {
		dst.Provider = src.Provider
	}
	if src.InternalModelID != nil {
		dst.InternalModelID = src.InternalModelID
	}
	if src.ModelParameters != nil {
		dst.ModelParameters = src.ModelParameters
	}

	// Usage & Cost Maps
	if src.ProvidedUsageDetails != nil {
		dst.ProvidedUsageDetails = src.ProvidedUsageDetails
	}
	if src.UsageDetails != nil {
		dst.UsageDetails = src.UsageDetails
	}
	if src.ProvidedCostDetails != nil {
		dst.ProvidedCostDetails = src.ProvidedCostDetails
	}
	if src.CostDetails != nil {
		dst.CostDetails = src.CostDetails
	}
	if src.TotalCost != nil {
		dst.TotalCost = src.TotalCost
	}

	// Prompt management
	if src.PromptID != nil {
		dst.PromptID = src.PromptID
	}
	if src.PromptName != nil {
		dst.PromptName = src.PromptName
	}
	if src.PromptVersion != nil {
		dst.PromptVersion = src.PromptVersion
	}

	// Blob storage references
	if src.InputBlobStorageID != nil {
		dst.InputBlobStorageID = src.InputBlobStorageID
	}
	if src.OutputBlobStorageID != nil {
		dst.OutputBlobStorageID = src.OutputBlobStorageID
	}
}

// DeleteObservation soft deletes an observation
func (s *ObservationService) DeleteObservation(ctx context.Context, id string) error {
	// Validate observation exists
	_, err := s.observationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("observation %s", id))
		}
		return appErrors.NewInternalError("failed to get observation", err)
	}

	// Delete observation
	if err := s.observationRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete observation", err)
	}

	return nil
}

// GetObservationByID retrieves an observation by its OTEL span_id
func (s *ObservationService) GetObservationByID(ctx context.Context, id string) (*observability.Observation, error) {
	obs, err := s.observationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("observation %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get observation", err)
	}

	return obs, nil
}

// GetObservationsByTraceID retrieves all observations for a trace
func (s *ObservationService) GetObservationsByTraceID(ctx context.Context, traceID string) ([]*observability.Observation, error) {
	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observations", err)
	}

	return observations, nil
}

// GetRootSpan retrieves the root span for a trace (parent_observation_id IS NULL)
func (s *ObservationService) GetRootSpan(ctx context.Context, traceID string) (*observability.Observation, error) {
	rootSpan, err := s.observationRepo.GetRootSpan(ctx, traceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("root span for trace %s", traceID))
		}
		return nil, appErrors.NewInternalError("failed to get root span", err)
	}

	return rootSpan, nil
}

// GetObservationTreeByTraceID retrieves all observations in a tree structure
func (s *ObservationService) GetObservationTreeByTraceID(ctx context.Context, traceID string) ([]*observability.Observation, error) {
	observations, err := s.observationRepo.GetTreeByTraceID(ctx, traceID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observation tree", err)
	}

	return observations, nil
}

// GetChildObservations retrieves child observations of a parent
func (s *ObservationService) GetChildObservations(ctx context.Context, parentObservationID string) ([]*observability.Observation, error) {
	observations, err := s.observationRepo.GetChildren(ctx, parentObservationID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get child observations", err)
	}

	return observations, nil
}

// GetObservationsByFilter retrieves observations by filter criteria
func (s *ObservationService) GetObservationsByFilter(ctx context.Context, filter *observability.ObservationFilter) ([]*observability.Observation, error) {
	observations, err := s.observationRepo.GetByFilter(ctx, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get observations", err)
	}

	return observations, nil
}

// CreateObservationBatch creates multiple observations in a batch
func (s *ObservationService) CreateObservationBatch(ctx context.Context, observations []*observability.Observation) error {
	if len(observations) == 0 {
		return nil
	}

	// Validate all observations
	for i, obs := range observations {
		if obs.TraceID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("observation[%d].trace_id", i), "trace_id is required")
		}
		if obs.ProjectID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("observation[%d].project_id", i), "project_id is required")
		}
		if obs.Name == "" {
			return appErrors.NewValidationError(fmt.Sprintf("observation[%d].name", i), "name is required")
		}
		if obs.ID == "" {
			return appErrors.NewValidationError(fmt.Sprintf("observation[%d].id", i), "OTEL span_id is required")
		}

		// Set defaults
		if obs.StatusCode == "" {
			obs.StatusCode = observability.StatusCodeUnset
		}
		if obs.SpanKind == "" {
			obs.SpanKind = string(observability.SpanKindInternal)
		}
		if obs.Type == "" {
			obs.Type = observability.ObservationTypeSpan
		}
		if obs.Level == "" {
			obs.Level = observability.ObservationLevelDefault
		}
		if obs.Attributes == "" {
			obs.Attributes = "{}"
		}
		if obs.CreatedAt.IsZero() {
			obs.CreatedAt = time.Now()
		}

		// Initialize maps if nil
		if obs.Metadata == nil {
			obs.Metadata = make(map[string]string)
		}
		if obs.ProvidedUsageDetails == nil {
			obs.ProvidedUsageDetails = make(map[string]uint64)
		}
		if obs.UsageDetails == nil {
			obs.UsageDetails = make(map[string]uint64)
		}
		if obs.ProvidedCostDetails == nil {
			obs.ProvidedCostDetails = make(map[string]float64)
		}
		if obs.CostDetails == nil {
			obs.CostDetails = make(map[string]float64)
		}

		// Calculate duration
		obs.CalculateDuration()
	}

	// Create batch
	if err := s.observationRepo.CreateBatch(ctx, observations); err != nil {
		return appErrors.NewInternalError("failed to create observation batch", err)
	}

	return nil
}

// CountObservations returns the count of observations matching the filter
func (s *ObservationService) CountObservations(ctx context.Context, filter *observability.ObservationFilter) (int64, error) {
	count, err := s.observationRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count observations", err)
	}

	return count, nil
}

// CalculateTraceCost calculates total cost for all observations in a trace
func (s *ObservationService) CalculateTraceCost(ctx context.Context, traceID string) (float64, error) {
	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to get observations", err)
	}

	var totalCost float64
	for _, obs := range observations {
		totalCost += obs.GetTotalCost()
	}

	return totalCost, nil
}

// CalculateTraceTokens calculates total tokens for all observations in a trace
func (s *ObservationService) CalculateTraceTokens(ctx context.Context, traceID string) (uint32, error) {
	observations, err := s.observationRepo.GetByTraceID(ctx, traceID)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to get observations", err)
	}

	var totalTokens uint64
	for _, obs := range observations {
		totalTokens += obs.GetTotalTokens()
	}

	return uint32(totalTokens), nil
}

// GetObservationWithFullContent fetches observation and loads full content from S3 if needed
// This method should be used for detail views where full content is required
// For list views, use GetObservationByID which returns previews only
func (s *ObservationService) GetObservationWithFullContent(ctx context.Context, id string) (*observability.Observation, error) {
	// 1. Fetch observation from ClickHouse (includes preview fields)
	obs, err := s.observationRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("observation %s", id))
		}
		return nil, appErrors.NewInternalError("failed to get observation", err)
	}

	// 2. Load input from S3 if offloaded
	if obs.InputBlobStorageID != nil && obs.Input == nil {
		if s.blobStorageService != nil {
			content, err := s.blobStorageService.DownloadFromS3(ctx, *obs.InputBlobStorageID)
			if err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"observation_id": obs.ID,
					"blob_id":        *obs.InputBlobStorageID,
				}).Warn("Failed to fetch input from S3, using preview")
				// Graceful fallback to preview
				obs.Input = obs.InputPreview
			} else {
				obs.Input = &content
			}
		} else {
			// S3 client not initialized, fallback to preview
			s.logger.WithField("observation_id", obs.ID).Warn("S3 client not initialized, using input preview")
			obs.Input = obs.InputPreview
		}
	}

	// 3. Load output from S3 if offloaded
	if obs.OutputBlobStorageID != nil && obs.Output == nil {
		if s.blobStorageService != nil {
			content, err := s.blobStorageService.DownloadFromS3(ctx, *obs.OutputBlobStorageID)
			if err != nil {
				s.logger.WithError(err).WithFields(logrus.Fields{
					"observation_id": obs.ID,
					"blob_id":        *obs.OutputBlobStorageID,
				}).Warn("Failed to fetch output from S3, using preview")
				// Graceful fallback to preview
				obs.Output = obs.OutputPreview
			} else {
				obs.Output = &content
			}
		} else {
			// S3 client not initialized, fallback to preview
			s.logger.WithField("observation_id", obs.ID).Warn("S3 client not initialized, using output preview")
			obs.Output = obs.OutputPreview
		}
	}

	return obs, nil
}
