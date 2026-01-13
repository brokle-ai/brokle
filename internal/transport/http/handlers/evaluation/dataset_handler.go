package evaluation

import (
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/response"
	"brokle/pkg/ulid"
)

type DatasetHandler struct {
	logger      *slog.Logger
	service     evaluationDomain.DatasetService
	itemService evaluationDomain.DatasetItemService
}

func NewDatasetHandler(
	logger *slog.Logger,
	service evaluationDomain.DatasetService,
	itemService evaluationDomain.DatasetItemService,
) *DatasetHandler {
	return &DatasetHandler{
		logger:      logger,
		service:     service,
		itemService: itemService,
	}
}

// @Summary Create dataset
// @Description Creates a new dataset for the project. Works for both SDK and Dashboard routes.
// @Tags Datasets, SDK - Datasets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param request body evaluation.CreateDatasetRequest true "Dataset request"
// @Success 201 {object} evaluation.DatasetResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Router /api/v1/projects/{projectId}/datasets [post]
// @Router /v1/datasets [post]
func (h *DatasetHandler) Create(c *gin.Context) {
	projectID, err := extractProjectID(c)
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	var req evaluationDomain.CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	dataset, err := h.service.Create(c.Request.Context(), projectID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.logger.Info("dataset created",
		"dataset_id", dataset.ID,
		"project_id", projectID,
		"name", dataset.Name,
	)

	response.Created(c, dataset.ToResponse())
}

// @Summary List datasets
// @Description Returns all datasets for the project.
// @Tags Datasets
// @Produce json
// @Param projectId path string true "Project ID"
// @Success 200 {array} evaluation.DatasetResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets [get]
func (h *DatasetHandler) List(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasets, err := h.service.List(c.Request.Context(), projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]*evaluationDomain.DatasetResponse, len(datasets))
	for i, dataset := range datasets {
		responses[i] = dataset.ToResponse()
	}

	response.Success(c, responses)
}

// @Summary Get dataset
// @Description Returns the dataset for a specific ID. Works for both SDK and Dashboard routes.
// @Tags Datasets, SDK - Datasets
// @Produce json
// @Security ApiKeyAuth
// @Param projectId path string false "Project ID (Dashboard routes)"
// @Param datasetId path string true "Dataset ID"
// @Success 200 {object} evaluation.DatasetResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId} [get]
// @Router /v1/datasets/{datasetId} [get]
func (h *DatasetHandler) Get(c *gin.Context) {
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

	dataset, err := h.service.GetByID(c.Request.Context(), datasetID, projectID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, dataset.ToResponse())
}

// @Summary Update dataset
// @Description Updates an existing dataset by ID.
// @Tags Datasets
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param datasetId path string true "Dataset ID"
// @Param request body evaluation.UpdateDatasetRequest true "Update request"
// @Success 200 {object} evaluation.DatasetResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 409 {object} response.ErrorResponse "Name already exists"
// @Router /api/v1/projects/{projectId}/datasets/{datasetId} [put]
func (h *DatasetHandler) Update(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	var req evaluationDomain.UpdateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	dataset, err := h.service.Update(c.Request.Context(), datasetID, projectID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, dataset.ToResponse())
}

// @Summary Delete dataset
// @Description Removes a dataset by its ID. Also deletes all items in the dataset.
// @Tags Datasets
// @Produce json
// @Param projectId path string true "Project ID"
// @Param datasetId path string true "Dataset ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId} [delete]
func (h *DatasetHandler) Delete(c *gin.Context) {
	projectID, err := ulid.Parse(c.Param("projectId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("projectId", "must be a valid ULID"))
		return
	}

	datasetID, err := ulid.Parse(c.Param("datasetId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("datasetId", "must be a valid ULID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), datasetID, projectID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}

// @Summary Batch create dataset items via SDK
// @Description Creates multiple items in a dataset using API key authentication.
// @Tags SDK - Datasets
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param datasetId path string true "Dataset ID"
// @Param request body evaluation.CreateDatasetItemsBatchRequest true "Batch items request"
// @Success 201 {object} SDKBatchCreateItemsResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/datasets/{datasetId}/items [post]
func (h *DatasetHandler) CreateItems(c *gin.Context) {
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

	var req evaluationDomain.CreateDatasetItemsBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	count, err := h.itemService.CreateBatch(c.Request.Context(), datasetID, projectID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	h.logger.Info("dataset items created",
		"dataset_id", datasetID,
		"project_id", projectID,
		"count", count,
	)

	response.Created(c, &SDKBatchCreateItemsResponse{Created: count})
}

// @Summary List dataset items via SDK
// @Description Returns items for a dataset with pagination using API key authentication.
// @Tags SDK - Datasets
// @Produce json
// @Security ApiKeyAuth
// @Param datasetId path string true "Dataset ID"
// @Param limit query int false "Limit (default 50, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} DatasetItemListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /v1/datasets/{datasetId}/items [get]
func (h *DatasetHandler) ListItems(c *gin.Context) {
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

	items, total, err := h.itemService.List(c.Request.Context(), datasetID, projectID, limit, offset)
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
