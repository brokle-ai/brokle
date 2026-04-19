package evaluation

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
)

// ---- dashboard dataset routes ---------------------------------------

func registerDatasetRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "create-dataset",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/datasets",
		Tags:          []string{"datasets"},
		Summary:       "Create a dataset",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.dashCreateDataset)

	huma.Register(api, huma.Operation{
		OperationID: "list-datasets",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets",
		Tags:        []string{"datasets"},
		Summary:     "List datasets with item counts",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashListDatasets)

	huma.Register(api, huma.Operation{
		OperationID: "get-dataset",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}",
		Tags:        []string{"datasets"},
		Summary:     "Get a dataset",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashGetDataset)

	huma.Register(api, huma.Operation{
		OperationID: "update-dataset",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}",
		Tags:        []string{"datasets"},
		Summary:     "Update a dataset",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashUpdateDataset)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-dataset",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/datasets/{datasetId}",
		Tags:          []string{"datasets"},
		Summary:       "Delete a dataset",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.dashDeleteDataset)

	// ---- dataset items -----------------------------------------
	huma.Register(api, huma.Operation{
		OperationID: "list-dataset-items",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/items",
		Tags:        []string{"dataset-items"},
		Summary:     "List dataset items",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashListDatasetItems)

	huma.Register(api, huma.Operation{
		OperationID:   "create-dataset-item",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/datasets/{datasetId}/items",
		Tags:          []string{"dataset-items"},
		Summary:       "Create a dataset item",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.dashCreateDatasetItem)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-dataset-item",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/datasets/{datasetId}/items/{itemId}",
		Tags:          []string{"dataset-items"},
		Summary:       "Delete a dataset item",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.dashDeleteDatasetItem)

	huma.Register(api, huma.Operation{
		OperationID: "export-dataset-items",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/items/export",
		Tags:        []string{"dataset-items"},
		Summary:     "Export all dataset items",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashExportDatasetItems)

	huma.Register(api, huma.Operation{
		OperationID: "import-dataset-items-json",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/items/import-json",
		Tags:        []string{"dataset-items"},
		Summary:     "Import dataset items from JSON",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashImportItemsJSON)

	huma.Register(api, huma.Operation{
		OperationID: "import-dataset-items-csv",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/items/import-csv",
		Tags:        []string{"dataset-items"},
		Summary:     "Import dataset items from CSV",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashImportItemsCSV)

	huma.Register(api, huma.Operation{
		OperationID: "create-dataset-items-from-traces",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/items/from-traces",
		Tags:        []string{"dataset-items"},
		Summary:     "Create dataset items from production traces",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashItemsFromTraces)

	huma.Register(api, huma.Operation{
		OperationID: "create-dataset-items-from-spans",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/items/from-spans",
		Tags:        []string{"dataset-items"},
		Summary:     "Create dataset items from production spans",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashItemsFromSpans)

	// ---- dataset versions --------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "create-dataset-version",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/datasets/{datasetId}/versions",
		Tags:          []string{"dataset-versions"},
		Summary:       "Create a dataset version snapshot",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.dashCreateDatasetVersion)

	huma.Register(api, huma.Operation{
		OperationID: "list-dataset-versions",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/versions",
		Tags:        []string{"dataset-versions"},
		Summary:     "List dataset versions",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashListDatasetVersions)

	huma.Register(api, huma.Operation{
		OperationID: "get-dataset-version",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/versions/{versionId}",
		Tags:        []string{"dataset-versions"},
		Summary:     "Get a dataset version",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashGetDatasetVersion)

	huma.Register(api, huma.Operation{
		OperationID: "get-dataset-version-items",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/versions/{versionId}/items",
		Tags:        []string{"dataset-versions"},
		Summary:     "List items for a dataset version",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashGetDatasetVersionItems)

	huma.Register(api, huma.Operation{
		OperationID: "pin-dataset-version",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/pin",
		Tags:        []string{"dataset-versions"},
		Summary:     "Pin a dataset to a version (or unpin with null)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashPinDatasetVersion)

	huma.Register(api, huma.Operation{
		OperationID: "get-dataset-with-version-info",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/datasets/{datasetId}/info",
		Tags:        []string{"dataset-versions"},
		Summary:     "Get a dataset with version info",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.dashGetDatasetWithVersionInfo)
}

// ---- SDK dataset routes ---------------------------------------------

func registerSDKDatasetRoutes(api huma.API, h *handler) {
	huma.Register(api, huma.Operation{
		OperationID:   "sdk-create-dataset",
		Method:        http.MethodPost,
		Path:          "/v1/datasets",
		Tags:          []string{"SDK - datasets"},
		Summary:       "Create a dataset (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkCreateDataset)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-list-datasets",
		Method:      http.MethodGet,
		Path:        "/v1/datasets",
		Tags:        []string{"SDK - datasets"},
		Summary:     "List datasets (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkListDatasets)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-get-dataset",
		Method:      http.MethodGet,
		Path:        "/v1/datasets/{datasetId}",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Get a dataset (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkGetDataset)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-update-dataset",
		Method:      http.MethodPatch,
		Path:        "/v1/datasets/{datasetId}",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Update a dataset (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkUpdateDataset)

	huma.Register(api, huma.Operation{
		OperationID:   "sdk-delete-dataset",
		Method:        http.MethodDelete,
		Path:          "/v1/datasets/{datasetId}",
		Tags:          []string{"SDK - datasets"},
		Summary:       "Delete a dataset (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.sdkDeleteDataset)

	huma.Register(api, huma.Operation{
		OperationID:   "sdk-batch-create-dataset-items",
		Method:        http.MethodPost,
		Path:          "/v1/datasets/{datasetId}/items",
		Tags:          []string{"SDK - datasets"},
		Summary:       "Batch-create dataset items (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkBatchCreateDatasetItems)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-list-dataset-items",
		Method:      http.MethodGet,
		Path:        "/v1/datasets/{datasetId}/items",
		Tags:        []string{"SDK - datasets"},
		Summary:     "List dataset items (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkListDatasetItems)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-export-dataset-items",
		Method:      http.MethodGet,
		Path:        "/v1/datasets/{datasetId}/items/export",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Export dataset items (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkExportDatasetItems)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-import-dataset-items-json",
		Method:      http.MethodPost,
		Path:        "/v1/datasets/{datasetId}/items/import-json",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Import dataset items from JSON (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkImportItemsJSON)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-import-dataset-items-csv",
		Method:      http.MethodPost,
		Path:        "/v1/datasets/{datasetId}/items/import-csv",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Import dataset items from CSV (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkImportItemsCSV)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-dataset-items-from-traces",
		Method:      http.MethodPost,
		Path:        "/v1/datasets/{datasetId}/items/from-traces",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Create dataset items from traces (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkItemsFromTraces)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-dataset-items-from-spans",
		Method:      http.MethodPost,
		Path:        "/v1/datasets/{datasetId}/items/from-spans",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Create dataset items from spans (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkItemsFromSpans)

	huma.Register(api, huma.Operation{
		OperationID:   "sdk-create-dataset-version",
		Method:        http.MethodPost,
		Path:          "/v1/datasets/{datasetId}/versions",
		Tags:          []string{"SDK - datasets"},
		Summary:       "Create a dataset version snapshot (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkCreateDatasetVersion)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-list-dataset-versions",
		Method:      http.MethodGet,
		Path:        "/v1/datasets/{datasetId}/versions",
		Tags:        []string{"SDK - datasets"},
		Summary:     "List dataset versions (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkListDatasetVersions)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-get-dataset-version",
		Method:      http.MethodGet,
		Path:        "/v1/datasets/{datasetId}/versions/{versionId}",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Get a dataset version (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkGetDatasetVersion)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-get-dataset-version-items",
		Method:      http.MethodGet,
		Path:        "/v1/datasets/{datasetId}/versions/{versionId}/items",
		Tags:        []string{"SDK - datasets"},
		Summary:     "List dataset version items (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkGetDatasetVersionItems)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-pin-dataset-version",
		Method:      http.MethodPost,
		Path:        "/v1/datasets/{datasetId}/pin",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Pin a dataset to a version (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkPinDatasetVersion)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-get-dataset-with-version-info",
		Method:      http.MethodGet,
		Path:        "/v1/datasets/{datasetId}/info",
		Tags:        []string{"SDK - datasets"},
		Summary:     "Get dataset with version info (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkGetDatasetWithVersionInfo)
}

// ---- dataset request DTOs -------------------------------------------

type CreateDatasetRequest struct {
	Name        string         `json:"name" minLength:"1" maxLength:"255"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type UpdateDatasetRequest struct {
	Name        *string        `json:"name,omitempty" minLength:"1" maxLength:"255"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type CreateDatasetItemRequest = evaluationDomain.CreateDatasetItemRequest

type BatchCreateDatasetItemsRequest = evaluationDomain.CreateDatasetItemsBatchRequest

type ImportFromJSONRequest struct {
	Items       []map[string]any    `json:"items" minItems:"1"`
	KeysMapping *KeysMappingRequest `json:"keys_mapping,omitempty"`
	Deduplicate bool                `json:"deduplicate"`
	Source      string              `json:"source,omitempty"`
}

type CSVColumnMappingRequest struct {
	InputColumn     string   `json:"input_column"`
	ExpectedColumn  string   `json:"expected_column,omitempty"`
	MetadataColumns []string `json:"metadata_columns,omitempty"`
}

type ImportFromCSVRequest struct {
	Content       string                  `json:"content"`
	ColumnMapping CSVColumnMappingRequest `json:"column_mapping"`
	HasHeader     bool                    `json:"has_header"`
	Deduplicate   bool                    `json:"deduplicate"`
}

type CreateFromTracesRequest struct {
	TraceIDs    []string            `json:"trace_ids" minItems:"1"`
	KeysMapping *KeysMappingRequest `json:"keys_mapping,omitempty"`
	Deduplicate bool                `json:"deduplicate"`
}

type CreateFromSpansRequest struct {
	SpanIDs     []string            `json:"span_ids" minItems:"1"`
	KeysMapping *KeysMappingRequest `json:"keys_mapping,omitempty"`
	Deduplicate bool                `json:"deduplicate"`
}

type CreateDatasetVersionRequest = evaluationDomain.CreateDatasetVersionRequest

type PinDatasetVersionRequest = evaluationDomain.PinDatasetVersionRequest

// ---- shared dataset op bodies ----------------------------------------

func (h *handler) createDataset(ctx context.Context, projectID uuid.UUID, body *CreateDatasetRequest) (*evaluationDomain.DatasetResponse, error) {
	domainReq := &evaluationDomain.CreateDatasetRequest{
		Name:        body.Name,
		Description: body.Description,
		Metadata:    body.Metadata,
	}
	ds, err := h.datasetSvc.Create(ctx, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: dataset created",
		"dataset_id", ds.ID, "project_id", projectID, "name", ds.Name)
	return ds.ToResponse(), nil
}

func (h *handler) listDatasets(
	ctx context.Context,
	projectID uuid.UUID,
	search, sortBy, sortDir string,
	page, limit int,
) ([]*evaluationDomain.DatasetWithItemCountResponse, int64, pagination.Params, error) {
	filter := &evaluationDomain.DatasetFilter{}
	if search != "" {
		filter.Search = &search
	}

	params := pagination.Params{
		Page:    page,
		Limit:   limit,
		SortBy:  sortBy,
		SortDir: sortDir,
	}
	params.SetDefaults("updated_at")
	if err := params.Validate(); err != nil {
		return nil, 0, params, appErrors.NewValidationError("Invalid pagination parameters", err.Error())
	}

	datasets, total, err := h.datasetSvc.ListWithFilters(ctx, projectID, filter, params)
	if err != nil {
		return nil, 0, params, err
	}
	out := make([]*evaluationDomain.DatasetWithItemCountResponse, len(datasets))
	for i, d := range datasets {
		out[i] = d.ToResponse()
	}
	return out, total, params, nil
}

func (h *handler) updateDataset(ctx context.Context, projectID, datasetID uuid.UUID, body *UpdateDatasetRequest) (*evaluationDomain.DatasetResponse, error) {
	domainReq := &evaluationDomain.UpdateDatasetRequest{
		Name:        body.Name,
		Description: body.Description,
		Metadata:    body.Metadata,
	}
	ds, err := h.datasetSvc.Update(ctx, datasetID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return ds.ToResponse(), nil
}

// ---- dashboard handler methods ---------------------------------------

type DashCreateDatasetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Body      CreateDatasetRequest
}
type DatasetOutput struct {
	Body *evaluationDomain.DatasetResponse
}

func (h *handler) dashCreateDataset(ctx context.Context, in *DashCreateDatasetInput) (*DatasetOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	resp, err := h.createDataset(ctx, projectID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: resp}, nil
}

type DashListDatasetsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Search    string `query:"search" required:"false"`
	SortBy    string `query:"sort_by" required:"false"`
	SortDir   string `query:"sort_dir" required:"false" enum:"asc,desc"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
}
type DatasetListOutput struct {
	Body pageList[*evaluationDomain.DatasetWithItemCountResponse]
}

func (h *handler) dashListDatasets(ctx context.Context, in *DashListDatasetsInput) (*DatasetListOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	out, total, params, err := h.listDatasets(ctx, projectID, in.Search, in.SortBy, in.SortDir, in.Page, in.Limit)
	if err != nil {
		return nil, err
	}
	return &DatasetListOutput{Body: pageList[*evaluationDomain.DatasetWithItemCountResponse]{
		Data: out, Total: total, Page: params.Page, Limit: params.Limit,
	}}, nil
}

type DashGetDatasetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
}

func (h *handler) dashGetDataset(ctx context.Context, in *DashGetDatasetInput) (*DatasetOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	ds, err := h.datasetSvc.GetByID(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: ds.ToResponse()}, nil
}

type DashUpdateDatasetInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      UpdateDatasetRequest
}

func (h *handler) dashUpdateDataset(ctx context.Context, in *DashUpdateDatasetInput) (*DatasetOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	resp, err := h.updateDataset(ctx, projectID, datasetID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: resp}, nil
}

type DashDeleteDatasetInput = DashGetDatasetInput

func (h *handler) dashDeleteDataset(ctx context.Context, in *DashDeleteDatasetInput) (*EmptyOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	if err := h.datasetSvc.Delete(ctx, datasetID, projectID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

// ---- dashboard dataset items -----------------------------------------

type DashListDatasetItemsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
}
type DatasetItemListOutput struct {
	Body pageList[*DatasetItemResponse]
}

func (h *handler) dashListDatasetItems(ctx context.Context, in *DashListDatasetItemsInput) (*DatasetItemListOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	page, limit := normalizePagination(in.Page, in.Limit)
	offset := (page - 1) * limit
	items, total, err := h.datasetItemSvc.List(ctx, datasetID, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	return &DatasetItemListOutput{Body: pageList[*DatasetItemResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

type DashCreateDatasetItemInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      CreateDatasetItemRequest
}
type DatasetItemOutput struct {
	Body *DatasetItemResponse
}

func (h *handler) dashCreateDatasetItem(ctx context.Context, in *DashCreateDatasetItemInput) (*DatasetItemOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	item, err := h.datasetItemSvc.Create(ctx, datasetID, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &DatasetItemOutput{Body: toDatasetItemResponse(item)}, nil
}

type DashDeleteDatasetItemInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	ItemID    string `path:"itemId" format:"uuid"`
}

func (h *handler) dashDeleteDatasetItem(ctx context.Context, in *DashDeleteDatasetItemInput) (*EmptyOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	itemID, err := parseItemID(in.ItemID)
	if err != nil {
		return nil, err
	}
	if err := h.datasetItemSvc.Delete(ctx, itemID, datasetID, projectID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

type DashExportDatasetItemsInput = DashGetDatasetInput
type DatasetItemsOutput struct {
	Body []*DatasetItemResponse
}

func (h *handler) dashExportDatasetItems(ctx context.Context, in *DashExportDatasetItemsInput) (*DatasetItemsOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	items, err := h.datasetItemSvc.ExportItems(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	return &DatasetItemsOutput{Body: out}, nil
}

type DashImportItemsJSONInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      ImportFromJSONRequest
}
type BulkImportOutput struct {
	Body *BulkImportResponse
}

func (h *handler) dashImportItemsJSON(ctx context.Context, in *DashImportItemsJSONInput) (*BulkImportOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.importItemsJSON(ctx, datasetID, projectID, &in.Body)
}

type DashImportItemsCSVInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      ImportFromCSVRequest
}

func (h *handler) dashImportItemsCSV(ctx context.Context, in *DashImportItemsCSVInput) (*BulkImportOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.importItemsCSV(ctx, datasetID, projectID, &in.Body)
}

type DashItemsFromTracesInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      CreateFromTracesRequest
}

func (h *handler) dashItemsFromTraces(ctx context.Context, in *DashItemsFromTracesInput) (*BulkImportOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.itemsFromTraces(ctx, datasetID, projectID, &in.Body)
}

type DashItemsFromSpansInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      CreateFromSpansRequest
}

func (h *handler) dashItemsFromSpans(ctx context.Context, in *DashItemsFromSpansInput) (*BulkImportOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.itemsFromSpans(ctx, datasetID, projectID, &in.Body)
}

// ---- shared import bodies -------------------------------------------

func (h *handler) importItemsJSON(ctx context.Context, datasetID, projectID uuid.UUID, body *ImportFromJSONRequest) (*BulkImportOutput, error) {
	if body.Source != "" {
		src := evaluationDomain.DatasetItemSource(body.Source)
		if !src.IsValid() {
			return nil, appErrors.NewValidationError("Invalid source", "source must be one of: manual, trace, span, csv, json, sdk")
		}
	}
	domainReq := &evaluationDomain.ImportDatasetItemsFromJSONRequest{
		Items:       body.Items,
		Deduplicate: body.Deduplicate,
		Source:      evaluationDomain.DatasetItemSource(body.Source),
		KeysMapping: body.KeysMapping.toDomain(),
	}
	result, err := h.datasetItemSvc.ImportFromJSON(ctx, datasetID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &BulkImportOutput{Body: toBulkImportResponse(result)}, nil
}

func (h *handler) importItemsCSV(ctx context.Context, datasetID, projectID uuid.UUID, body *ImportFromCSVRequest) (*BulkImportOutput, error) {
	domainReq := &evaluationDomain.ImportDatasetItemsFromCSVRequest{
		Content:     body.Content,
		HasHeader:   body.HasHeader,
		Deduplicate: body.Deduplicate,
		ColumnMapping: evaluationDomain.CSVColumnMapping{
			InputColumn:     body.ColumnMapping.InputColumn,
			ExpectedColumn:  body.ColumnMapping.ExpectedColumn,
			MetadataColumns: body.ColumnMapping.MetadataColumns,
		},
	}
	result, err := h.datasetItemSvc.ImportFromCSV(ctx, datasetID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &BulkImportOutput{Body: toBulkImportResponse(result)}, nil
}

func (h *handler) itemsFromTraces(ctx context.Context, datasetID, projectID uuid.UUID, body *CreateFromTracesRequest) (*BulkImportOutput, error) {
	domainReq := &evaluationDomain.CreateDatasetItemsFromTracesRequest{
		TraceIDs:    body.TraceIDs,
		Deduplicate: body.Deduplicate,
		KeysMapping: body.KeysMapping.toDomain(),
	}
	result, err := h.datasetItemSvc.CreateFromTraces(ctx, datasetID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &BulkImportOutput{Body: toBulkImportResponse(result)}, nil
}

func (h *handler) itemsFromSpans(ctx context.Context, datasetID, projectID uuid.UUID, body *CreateFromSpansRequest) (*BulkImportOutput, error) {
	domainReq := &evaluationDomain.CreateDatasetItemsFromSpansRequest{
		SpanIDs:     body.SpanIDs,
		Deduplicate: body.Deduplicate,
		KeysMapping: body.KeysMapping.toDomain(),
	}
	result, err := h.datasetItemSvc.CreateFromSpans(ctx, datasetID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &BulkImportOutput{Body: toBulkImportResponse(result)}, nil
}

// ---- dashboard dataset versions -------------------------------------

type DashCreateDatasetVersionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      CreateDatasetVersionRequest
}
type DatasetVersionOutput struct {
	Body *evaluationDomain.DatasetVersionResponse
}

func (h *handler) dashCreateDatasetVersion(ctx context.Context, in *DashCreateDatasetVersionInput) (*DatasetVersionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	v, err := h.datasetVersionSvc.CreateVersion(ctx, datasetID, projectID, &body)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: dataset version created",
		"version_id", v.ID, "dataset_id", datasetID, "project_id", projectID, "version", v.Version)
	return &DatasetVersionOutput{Body: v.ToResponse()}, nil
}

type DashListDatasetVersionsInput = DashGetDatasetInput
type DatasetVersionListOutput struct {
	Body []*evaluationDomain.DatasetVersionResponse
}

func (h *handler) dashListDatasetVersions(ctx context.Context, in *DashListDatasetVersionsInput) (*DatasetVersionListOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	versions, err := h.datasetVersionSvc.ListVersions(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]*evaluationDomain.DatasetVersionResponse, len(versions))
	for i, v := range versions {
		out[i] = v.ToResponse()
	}
	return &DatasetVersionListOutput{Body: out}, nil
}

type DashGetDatasetVersionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	VersionID string `path:"versionId" format:"uuid"`
}

func (h *handler) dashGetDatasetVersion(ctx context.Context, in *DashGetDatasetVersionInput) (*DatasetVersionOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	versionID, err := parseVersionID(in.VersionID)
	if err != nil {
		return nil, err
	}
	v, err := h.datasetVersionSvc.GetVersion(ctx, versionID, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	return &DatasetVersionOutput{Body: v.ToResponse()}, nil
}

type DashGetDatasetVersionItemsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	VersionID string `path:"versionId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
}

func (h *handler) dashGetDatasetVersionItems(ctx context.Context, in *DashGetDatasetVersionItemsInput) (*DatasetItemListOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	versionID, err := parseVersionID(in.VersionID)
	if err != nil {
		return nil, err
	}
	page, limit := normalizePagination(in.Page, in.Limit)
	offset := (page - 1) * limit
	items, total, err := h.datasetVersionSvc.GetVersionItems(ctx, versionID, datasetID, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	return &DatasetItemListOutput{Body: pageList[*DatasetItemResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

type DashPinDatasetVersionInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      PinDatasetVersionRequest
}

func (h *handler) dashPinDatasetVersion(ctx context.Context, in *DashPinDatasetVersionInput) (*DatasetOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	ds, err := h.datasetVersionSvc.PinVersion(ctx, datasetID, projectID, body.VersionID)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: ds.ToResponse()}, nil
}

type DashGetDatasetWithVersionInfoInput = DashGetDatasetInput
type DatasetWithVersionInfoOutput struct {
	Body *evaluationDomain.DatasetWithVersionResponse
}

func (h *handler) dashGetDatasetWithVersionInfo(ctx context.Context, in *DashGetDatasetWithVersionInfoInput) (*DatasetWithVersionInfoOutput, error) {
	projectID, err := parseProjectID(in.ProjectID)
	if err != nil {
		return nil, err
	}
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	resp, err := h.datasetVersionSvc.GetDatasetWithVersionInfo(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	return &DatasetWithVersionInfoOutput{Body: resp}, nil
}

// ---- SDK dataset methods ---------------------------------------------

type SDKCreateDatasetInput struct {
	Body CreateDatasetRequest
}

func (h *handler) sdkCreateDataset(ctx context.Context, in *SDKCreateDatasetInput) (*DatasetOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	resp, err := h.createDataset(ctx, projectID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: resp}, nil
}

type SDKListDatasetsInput struct {
	Search  string `query:"search" required:"false"`
	SortBy  string `query:"sort_by" required:"false"`
	SortDir string `query:"sort_dir" required:"false" enum:"asc,desc"`
	Page    int    `query:"page" required:"false" minimum:"1"`
	Limit   int    `query:"limit" required:"false"`
}

func (h *handler) sdkListDatasets(ctx context.Context, in *SDKListDatasetsInput) (*DatasetListOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	out, total, params, err := h.listDatasets(ctx, projectID, in.Search, in.SortBy, in.SortDir, in.Page, in.Limit)
	if err != nil {
		return nil, err
	}
	return &DatasetListOutput{Body: pageList[*evaluationDomain.DatasetWithItemCountResponse]{
		Data: out, Total: total, Page: params.Page, Limit: params.Limit,
	}}, nil
}

type SDKDatasetInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
}

func (h *handler) sdkGetDataset(ctx context.Context, in *SDKDatasetInput) (*DatasetOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	ds, err := h.datasetSvc.GetByID(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: ds.ToResponse()}, nil
}

type SDKUpdateDatasetInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      UpdateDatasetRequest
}

func (h *handler) sdkUpdateDataset(ctx context.Context, in *SDKUpdateDatasetInput) (*DatasetOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	resp, err := h.updateDataset(ctx, projectID, datasetID, &in.Body)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: resp}, nil
}

func (h *handler) sdkDeleteDataset(ctx context.Context, in *SDKDatasetInput) (*EmptyOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	if err := h.datasetSvc.Delete(ctx, datasetID, projectID); err != nil {
		return nil, err
	}
	return &EmptyOutput{}, nil
}

type SDKBatchCreateDatasetItemsInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      BatchCreateDatasetItemsRequest
}
type CountOutput struct {
	Body *CountResponse
}

func (h *handler) sdkBatchCreateDatasetItems(ctx context.Context, in *SDKBatchCreateDatasetItemsInput) (*CountOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	count, err := h.datasetItemSvc.CreateBatch(ctx, datasetID, projectID, &body)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "evaluation: dataset items created (SDK)",
		"dataset_id", datasetID, "project_id", projectID, "count", count)
	return &CountOutput{Body: &CountResponse{Created: count}}, nil
}

type SDKListDatasetItemsInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
}

func (h *handler) sdkListDatasetItems(ctx context.Context, in *SDKListDatasetItemsInput) (*DatasetItemListOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	page, limit := normalizePagination(in.Page, in.Limit)
	offset := (page - 1) * limit
	items, total, err := h.datasetItemSvc.List(ctx, datasetID, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	return &DatasetItemListOutput{Body: pageList[*DatasetItemResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

func (h *handler) sdkExportDatasetItems(ctx context.Context, in *SDKDatasetInput) (*DatasetItemsOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	items, err := h.datasetItemSvc.ExportItems(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	return &DatasetItemsOutput{Body: out}, nil
}

type SDKImportItemsJSONInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      ImportFromJSONRequest
}

func (h *handler) sdkImportItemsJSON(ctx context.Context, in *SDKImportItemsJSONInput) (*BulkImportOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.importItemsJSON(ctx, datasetID, projectID, &in.Body)
}

type SDKImportItemsCSVInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      ImportFromCSVRequest
}

func (h *handler) sdkImportItemsCSV(ctx context.Context, in *SDKImportItemsCSVInput) (*BulkImportOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.importItemsCSV(ctx, datasetID, projectID, &in.Body)
}

type SDKItemsFromTracesInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      CreateFromTracesRequest
}

func (h *handler) sdkItemsFromTraces(ctx context.Context, in *SDKItemsFromTracesInput) (*BulkImportOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.itemsFromTraces(ctx, datasetID, projectID, &in.Body)
}

type SDKItemsFromSpansInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      CreateFromSpansRequest
}

func (h *handler) sdkItemsFromSpans(ctx context.Context, in *SDKItemsFromSpansInput) (*BulkImportOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	return h.itemsFromSpans(ctx, datasetID, projectID, &in.Body)
}

// ---- SDK dataset versions --------------------------------------------

type SDKCreateDatasetVersionInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      CreateDatasetVersionRequest
}

func (h *handler) sdkCreateDatasetVersion(ctx context.Context, in *SDKCreateDatasetVersionInput) (*DatasetVersionOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	v, err := h.datasetVersionSvc.CreateVersion(ctx, datasetID, projectID, &body)
	if err != nil {
		return nil, err
	}
	return &DatasetVersionOutput{Body: v.ToResponse()}, nil
}

func (h *handler) sdkListDatasetVersions(ctx context.Context, in *SDKDatasetInput) (*DatasetVersionListOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	versions, err := h.datasetVersionSvc.ListVersions(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]*evaluationDomain.DatasetVersionResponse, len(versions))
	for i, v := range versions {
		out[i] = v.ToResponse()
	}
	return &DatasetVersionListOutput{Body: out}, nil
}

type SDKGetDatasetVersionInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	VersionID string `path:"versionId" format:"uuid"`
}

func (h *handler) sdkGetDatasetVersion(ctx context.Context, in *SDKGetDatasetVersionInput) (*DatasetVersionOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	versionID, err := parseVersionID(in.VersionID)
	if err != nil {
		return nil, err
	}
	v, err := h.datasetVersionSvc.GetVersion(ctx, versionID, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	return &DatasetVersionOutput{Body: v.ToResponse()}, nil
}

type SDKGetDatasetVersionItemsInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	VersionID string `path:"versionId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
}

func (h *handler) sdkGetDatasetVersionItems(ctx context.Context, in *SDKGetDatasetVersionItemsInput) (*DatasetItemListOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	versionID, err := parseVersionID(in.VersionID)
	if err != nil {
		return nil, err
	}
	page, limit := normalizePagination(in.Page, in.Limit)
	offset := (page - 1) * limit
	items, total, err := h.datasetVersionSvc.GetVersionItems(ctx, versionID, datasetID, projectID, limit, offset)
	if err != nil {
		return nil, err
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	return &DatasetItemListOutput{Body: pageList[*DatasetItemResponse]{
		Data: out, Total: total, Page: page, Limit: limit,
	}}, nil
}

type SDKPinDatasetVersionInput struct {
	DatasetID string `path:"datasetId" format:"uuid"`
	Body      PinDatasetVersionRequest
}

func (h *handler) sdkPinDatasetVersion(ctx context.Context, in *SDKPinDatasetVersionInput) (*DatasetOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	body := in.Body
	ds, err := h.datasetVersionSvc.PinVersion(ctx, datasetID, projectID, body.VersionID)
	if err != nil {
		return nil, err
	}
	return &DatasetOutput{Body: ds.ToResponse()}, nil
}

func (h *handler) sdkGetDatasetWithVersionInfo(ctx context.Context, in *SDKDatasetInput) (*DatasetWithVersionInfoOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	datasetID, err := parseDatasetID(in.DatasetID)
	if err != nil {
		return nil, err
	}
	resp, err := h.datasetVersionSvc.GetDatasetWithVersionInfo(ctx, datasetID, projectID)
	if err != nil {
		return nil, err
	}
	return &DatasetWithVersionInfoOutput{Body: resp}, nil
}
