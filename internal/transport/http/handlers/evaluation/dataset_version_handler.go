package evaluation

import (
	"errors"
	"io"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

type DatasetVersionHandler struct {
	logger  *slog.Logger
	service evaluationDomain.DatasetVersionService
}

func NewDatasetVersionHandler(
	logger *slog.Logger,
	service evaluationDomain.DatasetVersionService,
) *DatasetVersionHandler {
	return &DatasetVersionHandler{
		logger:  logger,
		service: service,
	}
}

// @Summary Create dataset version
// @Description Creates a new version snapshot of the current dataset items. Works for both SDK and Dashboard routes.
// @Tags Dataset Versions, SDK - Datasets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param datasetId path string true "Dataset ID"
// @Param request body evaluation.CreateDatasetVersionRequest false "Version request"
// @Success 201 {object} evaluation.DatasetVersionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/versions [post]
// @Router /v1/datasets/{datasetId}/versions [post]
func (h *DatasetVersionHandler) CreateVersion(c *gin.Context) {
	projectID, err := extractProjectID(c)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	var req evaluationDomain.CreateDatasetVersionRequest
	// Allow empty body (EOF) for creating a version without description/metadata,
	// but reject malformed JSON
	if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		response.ValidationError(c, "invalid request body", err.Error())
		return
	}

	version, err := h.service.CreateVersion(c.Request.Context(), datasetID, projectID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.logger.Info("dataset version created",
		"version_id", version.ID,
		"dataset_id", datasetID,
		"project_id", projectID,
		"version", version.Version,
	)

	response.Created(c, version.ToResponse())
}

// @Summary List dataset versions
// @Description Returns all versions for a dataset. Works for both SDK and Dashboard routes.
// @Tags Dataset Versions, SDK - Datasets
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param datasetId path string true "Dataset ID"
// @Success 200 {array} evaluation.DatasetVersionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/versions [get]
// @Router /v1/datasets/{datasetId}/versions [get]
func (h *DatasetVersionHandler) ListVersions(c *gin.Context) {
	projectID, err := extractProjectID(c)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	versions, err := h.service.ListVersions(c.Request.Context(), datasetID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]*evaluationDomain.DatasetVersionResponse, len(versions))
	for i, version := range versions {
		responses[i] = version.ToResponse()
	}

	response.Success(c, responses)
}

// @Summary Get dataset version
// @Description Returns a specific version by ID. Works for both SDK and Dashboard routes.
// @Tags Dataset Versions, SDK - Datasets
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param datasetId path string true "Dataset ID"
// @Param versionId path string true "Version ID"
// @Success 200 {object} evaluation.DatasetVersionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/versions/{versionId} [get]
// @Router /v1/datasets/{datasetId}/versions/{versionId} [get]
func (h *DatasetVersionHandler) GetVersion(c *gin.Context) {
	projectID, err := extractProjectID(c)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	versionID, err := ulid.Parse(c.Param("versionId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("versionId", "must be a valid ULID"))
		return
	}

	version, err := h.service.GetVersion(c.Request.Context(), versionID, datasetID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, version.ToResponse())
}

// @Summary Get version items
// @Description Returns items for a specific version with pagination. Works for both SDK and Dashboard routes.
// @Tags Dataset Versions, SDK - Datasets
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param datasetId path string true "Dataset ID"
// @Param versionId path string true "Version ID"
// @Param limit query int false "Limit (default 50, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} DatasetItemListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/versions/{versionId}/items [get]
// @Router /v1/datasets/{datasetId}/versions/{versionId}/items [get]
func (h *DatasetVersionHandler) GetVersionItems(c *gin.Context) {
	projectID, err := extractProjectID(c)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	versionID, err := ulid.Parse(c.Param("versionId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("versionId", "must be a valid ULID"))
		return
	}

	limit := 50
	offset := 0

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	items, total, err := h.service.GetVersionItems(c.Request.Context(), versionID, datasetID, projectID, limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]*DatasetItemResponse, len(items))
	for i, item := range items {
		domainResp := item.ToResponse()
		responses[i] = &DatasetItemResponse{
			ID:            domainResp.ID,
			DatasetID:     domainResp.DatasetID,
			Input:         domainResp.Input,
			Expected:      domainResp.Expected,
			Metadata:      domainResp.Metadata,
			Source:        string(domainResp.Source),
			SourceTraceID: domainResp.SourceTraceID,
			SourceSpanID:  domainResp.SourceSpanID,
			CreatedAt:     domainResp.CreatedAt,
		}
	}

	response.Success(c, &DatasetItemListResponse{
		Items: responses,
		Total: total,
	})
}

// @Summary Pin dataset to version
// @Description Pins a dataset to a specific version. Pass null version_id to unpin. Works for both SDK and Dashboard routes.
// @Tags Dataset Versions, SDK - Datasets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param datasetId path string true "Dataset ID"
// @Param request body evaluation.PinDatasetVersionRequest true "Pin request"
// @Success 200 {object} evaluation.DatasetResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/pin [post]
// @Router /v1/datasets/{datasetId}/pin [post]
func (h *DatasetVersionHandler) PinVersion(c *gin.Context) {
	projectID, err := extractProjectID(c)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	var req evaluationDomain.PinDatasetVersionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	var versionID *ulid.ULID
	if req.VersionID != nil && *req.VersionID != "" {
		parsed, err := ulid.Parse(*req.VersionID)
		if err != nil {
			response.Error(c, appErrors.NewValidationError("version_id", "must be a valid ULID"))
			return
		}
		versionID = &parsed
	}

	dataset, err := h.service.PinVersion(c.Request.Context(), datasetID, projectID, versionID)
	if err != nil {
		response.Error(c, err)
		return
	}

	action := "unpinned"
	if versionID != nil {
		action = "pinned to version " + versionID.String()
	}
	h.logger.Info("dataset "+action,
		"dataset_id", datasetID,
		"project_id", projectID,
	)

	response.Success(c, dataset.ToResponse())
}

// @Summary Get dataset with version info
// @Description Returns a dataset with its version information (current, latest). Works for both SDK and Dashboard routes.
// @Tags Dataset Versions, SDK - Datasets
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param datasetId path string true "Dataset ID"
// @Success 200 {object} evaluation.DatasetWithVersionResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/info [get]
// @Router /v1/datasets/{datasetId}/info [get]
func (h *DatasetVersionHandler) GetDatasetWithVersionInfo(c *gin.Context) {
	projectID, err := extractProjectID(c)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	datasetWithInfo, err := h.service.GetDatasetWithVersionInfo(c.Request.Context(), datasetID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, datasetWithInfo)
}
