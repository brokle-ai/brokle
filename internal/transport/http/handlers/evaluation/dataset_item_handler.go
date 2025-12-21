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

type DatasetItemHandler struct {
	logger  *slog.Logger
	service evaluationDomain.DatasetItemService
}

func NewDatasetItemHandler(
	logger *slog.Logger,
	service evaluationDomain.DatasetItemService,
) *DatasetItemHandler {
	return &DatasetItemHandler{
		logger:  logger,
		service: service,
	}
}

// @Summary List dataset items
// @Description Returns items for a dataset with pagination.
// @Tags Dataset Items
// @Produce json
// @Param projectId path string true "Project ID"
// @Param datasetId path string true "Dataset ID"
// @Param limit query int false "Limit (default 50, max 100)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} DatasetItemListResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/items [get]
func (h *DatasetItemHandler) List(c *gin.Context) {
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

	items, total, err := h.service.List(c.Request.Context(), datasetID, projectID, limit, offset)
	if err != nil {
		response.Error(c, err)
		return
	}

	responses := make([]*DatasetItemResponse, len(items))
	for i, item := range items {
		domainResp := item.ToResponse()
		responses[i] = &DatasetItemResponse{
			ID:        domainResp.ID,
			DatasetID: domainResp.DatasetID,
			Input:     domainResp.Input,
			Expected:  domainResp.Expected,
			Metadata:  domainResp.Metadata,
			CreatedAt: domainResp.CreatedAt,
		}
	}

	response.Success(c, &DatasetItemListResponse{
		Items: responses,
		Total: total,
	})
}

// @Summary Create dataset item
// @Description Creates a new item in the dataset.
// @Tags Dataset Items
// @Accept json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param datasetId path string true "Dataset ID"
// @Param request body evaluation.CreateDatasetItemRequest true "Item request"
// @Success 201 {object} DatasetItemResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/items [post]
func (h *DatasetItemHandler) Create(c *gin.Context) {
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

	var req evaluationDomain.CreateDatasetItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, "Invalid request body", err.Error())
		return
	}

	item, err := h.service.Create(c.Request.Context(), datasetID, projectID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	domainResp := item.ToResponse()
	response.Created(c, &DatasetItemResponse{
		ID:        domainResp.ID,
		DatasetID: domainResp.DatasetID,
		Input:     domainResp.Input,
		Expected:  domainResp.Expected,
		Metadata:  domainResp.Metadata,
		CreatedAt: domainResp.CreatedAt,
	})
}

// @Summary Delete dataset item
// @Description Removes an item from the dataset.
// @Tags Dataset Items
// @Produce json
// @Param projectId path string true "Project ID"
// @Param datasetId path string true "Dataset ID"
// @Param itemId path string true "Item ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Router /api/v1/projects/{projectId}/datasets/{datasetId}/items/{itemId} [delete]
func (h *DatasetItemHandler) Delete(c *gin.Context) {
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

	itemID, err := ulid.Parse(c.Param("itemId"))
	if err != nil {
		response.Error(c, appErrors.NewValidationError("itemId", "must be a valid ULID"))
		return
	}

	if err := h.service.Delete(c.Request.Context(), itemID, datasetID, projectID); err != nil {
		response.Error(c, err)
		return
	}

	response.NoContent(c)
}
