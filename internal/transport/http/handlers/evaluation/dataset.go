package evaluation

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	evaluationDomain "brokle/internal/core/domain/evaluation"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/pagination"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

// ---- shared bodies --------------------------------------------------

func (h *Handler) createDatasetCore(ctx context.Context, projectID uuid.UUID, body *CreateDatasetRequest) (*evaluationDomain.DatasetResponse, error) {
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

func (h *Handler) listDatasetsCore(
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
		return nil, 0, params, appErrors.InvalidParam("pagination", err.Error())
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

func (h *Handler) updateDatasetCore(ctx context.Context, projectID, datasetID uuid.UUID, body *UpdateDatasetRequest) (*evaluationDomain.DatasetResponse, error) {
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

// ---- shared import bodies -------------------------------------------

func (h *Handler) importItemsJSONCore(ctx context.Context, datasetID, projectID uuid.UUID, body *ImportFromJSONRequest) (*BulkImportResponse, error) {
	if body.Source != "" {
		src := evaluationDomain.DatasetItemSource(body.Source)
		if !src.IsValid() {
			return nil, appErrors.InvalidParam("source", "must be one of: manual, trace, span, csv, json, sdk")
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
	return toBulkImportResponse(result), nil
}

func (h *Handler) importItemsCSVCore(ctx context.Context, datasetID, projectID uuid.UUID, body *ImportFromCSVRequest) (*BulkImportResponse, error) {
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
	return toBulkImportResponse(result), nil
}

func (h *Handler) itemsFromTracesCore(ctx context.Context, datasetID, projectID uuid.UUID, body *CreateFromTracesRequest) (*BulkImportResponse, error) {
	domainReq := &evaluationDomain.CreateDatasetItemsFromTracesRequest{
		TraceIDs:    body.TraceIDs,
		Deduplicate: body.Deduplicate,
		KeysMapping: body.KeysMapping.toDomain(),
	}
	result, err := h.datasetItemSvc.CreateFromTraces(ctx, datasetID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return toBulkImportResponse(result), nil
}

func (h *Handler) itemsFromSpansCore(ctx context.Context, datasetID, projectID uuid.UUID, body *CreateFromSpansRequest) (*BulkImportResponse, error) {
	domainReq := &evaluationDomain.CreateDatasetItemsFromSpansRequest{
		SpanIDs:     body.SpanIDs,
		Deduplicate: body.Deduplicate,
		KeysMapping: body.KeysMapping.toDomain(),
	}
	result, err := h.datasetItemSvc.CreateFromSpans(ctx, datasetID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return toBulkImportResponse(result), nil
}

// ---- dashboard handlers ---------------------------------------------

func (h *Handler) DashCreateDataset(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	var body CreateDatasetRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.createDatasetCore(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, ds)
}

func (h *Handler) DashListDatasets(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	out, total, params, err := h.listDatasetsCore(r.Context(), projectID,
		q.Get("search"), q.Get("sort_by"), q.Get("sort_dir"), page, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, pageList[*evaluationDomain.DatasetWithItemCountResponse]{
		Data: out, Pagination: response.BuildPagination(params.Page, params.Limit, total),
	})
}

func (h *Handler) DashGetDataset(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.datasetSvc.GetByID(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ds.ToResponse())
}

func (h *Handler) DashUpdateDataset(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body UpdateDatasetRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.updateDatasetCore(r.Context(), projectID, datasetID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ds)
}

func (h *Handler) DashDeleteDataset(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.datasetSvc.Delete(r.Context(), datasetID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) DashListDatasetItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	offset := (page - 1) * limit
	items, total, err := h.datasetItemSvc.List(r.Context(), datasetID, projectID, limit, offset)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	response.Success(w, pageList[*DatasetItemResponse]{
		Data: out, Pagination: response.BuildPagination(page, limit, total),
	})
}

func (h *Handler) DashCreateDatasetItem(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body CreateDatasetItemRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	item, err := h.datasetItemSvc.Create(r.Context(), datasetID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, toDatasetItemResponse(item))
}

func (h *Handler) DashDeleteDatasetItem(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	itemID, err := request.URLParamUUID(r, "itemId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.datasetItemSvc.Delete(r.Context(), itemID, datasetID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) DashExportDatasetItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	items, err := h.datasetItemSvc.ExportItems(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	response.Success(w, out)
}

// ---- dashboard import flows ----------------------------------------

func (h *Handler) DashImportItemsJSON(w http.ResponseWriter, r *http.Request) {
	h.runDashImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body ImportFromJSONRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.importItemsJSONCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) DashImportItemsCSV(w http.ResponseWriter, r *http.Request) {
	h.runDashImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body ImportFromCSVRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.importItemsCSVCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) DashItemsFromTraces(w http.ResponseWriter, r *http.Request) {
	h.runDashImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body CreateFromTracesRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.itemsFromTracesCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) DashItemsFromSpans(w http.ResponseWriter, r *http.Request) {
	h.runDashImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body CreateFromSpansRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.itemsFromSpansCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) runDashImport(
	w http.ResponseWriter,
	r *http.Request,
	do func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error),
) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := do(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

// ---- dashboard dataset versions ------------------------------------

func (h *Handler) DashCreateDatasetVersion(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body CreateDatasetVersionRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	v, err := h.datasetVersionSvc.CreateVersion(r.Context(), datasetID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "evaluation: dataset version created",
		"version_id", v.ID, "dataset_id", datasetID, "project_id", projectID, "version", v.Version)
	response.Created(w, v.ToResponse())
}

func (h *Handler) DashListDatasetVersions(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versions, err := h.datasetVersionSvc.ListVersions(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*evaluationDomain.DatasetVersionResponse, len(versions))
	for i, v := range versions {
		out[i] = v.ToResponse()
	}
	response.Success(w, out)
}

func (h *Handler) DashGetDatasetVersion(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versionID, err := request.URLParamUUID(r, "versionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	v, err := h.datasetVersionSvc.GetVersion(r.Context(), versionID, datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, v.ToResponse())
}

func (h *Handler) DashGetDatasetVersionItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versionID, err := request.URLParamUUID(r, "versionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	offset := (page - 1) * limit
	items, total, err := h.datasetVersionSvc.GetVersionItems(r.Context(), versionID, datasetID, projectID, limit, offset)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	response.Success(w, pageList[*DatasetItemResponse]{
		Data: out, Pagination: response.BuildPagination(page, limit, total),
	})
}

func (h *Handler) DashPinDatasetVersion(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body PinDatasetVersionRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.datasetVersionSvc.PinVersion(r.Context(), datasetID, projectID, body.VersionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ds.ToResponse())
}

func (h *Handler) DashGetDatasetWithVersionInfo(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	resp, err := h.datasetVersionSvc.GetDatasetWithVersionInfo(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}

// ---- SDK dataset handlers -------------------------------------------

func (h *Handler) SdkCreateDataset(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	var body CreateDatasetRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.createDatasetCore(r.Context(), projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, ds)
}

func (h *Handler) SdkListDatasets(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	q := r.URL.Query()
	out, total, params, err := h.listDatasetsCore(r.Context(), projectID,
		q.Get("search"), q.Get("sort_by"), q.Get("sort_dir"), page, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, pageList[*evaluationDomain.DatasetWithItemCountResponse]{
		Data: out, Pagination: response.BuildPagination(params.Page, params.Limit, total),
	})
}

func (h *Handler) SdkGetDataset(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.datasetSvc.GetByID(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ds.ToResponse())
}

func (h *Handler) SdkUpdateDataset(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body UpdateDatasetRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.updateDatasetCore(r.Context(), projectID, datasetID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ds)
}

func (h *Handler) SdkDeleteDataset(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.datasetSvc.Delete(r.Context(), datasetID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) SdkBatchCreateDatasetItems(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body BatchCreateDatasetItemsRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	count, err := h.datasetItemSvc.CreateBatch(r.Context(), datasetID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "evaluation: dataset items created (SDK)",
		"dataset_id", datasetID, "project_id", projectID, "count", count)
	response.Created(w, &CountResponse{Created: count})
}

func (h *Handler) SdkListDatasetItems(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	offset := (page - 1) * limit
	items, total, err := h.datasetItemSvc.List(r.Context(), datasetID, projectID, limit, offset)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	response.Success(w, pageList[*DatasetItemResponse]{
		Data: out, Pagination: response.BuildPagination(page, limit, total),
	})
}

func (h *Handler) SdkExportDatasetItems(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	items, err := h.datasetItemSvc.ExportItems(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	response.Success(w, out)
}

func (h *Handler) SdkImportItemsJSON(w http.ResponseWriter, r *http.Request) {
	h.runSDKImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body ImportFromJSONRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.importItemsJSONCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) SdkImportItemsCSV(w http.ResponseWriter, r *http.Request) {
	h.runSDKImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body ImportFromCSVRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.importItemsCSVCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) SdkItemsFromTraces(w http.ResponseWriter, r *http.Request) {
	h.runSDKImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body CreateFromTracesRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.itemsFromTracesCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) SdkItemsFromSpans(w http.ResponseWriter, r *http.Request) {
	h.runSDKImport(w, r, func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error) {
		var body CreateFromSpansRequest
		if err := request.DecodeJSON(r, &body); err != nil {
			return nil, err
		}
		return h.itemsFromSpansCore(ctx, datasetID, projectID, &body)
	})
}

func (h *Handler) runSDKImport(
	w http.ResponseWriter,
	r *http.Request,
	do func(ctx context.Context, datasetID, projectID uuid.UUID) (*BulkImportResponse, error),
) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	res, err := do(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, res)
}

// ---- SDK dataset versions ------------------------------------------

func (h *Handler) SdkCreateDatasetVersion(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body CreateDatasetVersionRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	v, err := h.datasetVersionSvc.CreateVersion(r.Context(), datasetID, projectID, &body)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Created(w, v.ToResponse())
}

func (h *Handler) SdkListDatasetVersions(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versions, err := h.datasetVersionSvc.ListVersions(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*evaluationDomain.DatasetVersionResponse, len(versions))
	for i, v := range versions {
		out[i] = v.ToResponse()
	}
	response.Success(w, out)
}

func (h *Handler) SdkGetDatasetVersion(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versionID, err := request.URLParamUUID(r, "versionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	v, err := h.datasetVersionSvc.GetVersion(r.Context(), versionID, datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, v.ToResponse())
}

func (h *Handler) SdkGetDatasetVersionItems(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	versionID, err := request.URLParamUUID(r, "versionId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	page, limit, err := readPagination(r)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	offset := (page - 1) * limit
	items, total, err := h.datasetVersionSvc.GetVersionItems(r.Context(), versionID, datasetID, projectID, limit, offset)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*DatasetItemResponse, len(items))
	for i, it := range items {
		out[i] = toDatasetItemResponse(it)
	}
	response.Success(w, pageList[*DatasetItemResponse]{
		Data: out, Pagination: response.BuildPagination(page, limit, total),
	})
}

func (h *Handler) SdkPinDatasetVersion(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body PinDatasetVersionRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	ds, err := h.datasetVersionSvc.PinVersion(r.Context(), datasetID, projectID, body.VersionID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, ds.ToResponse())
}

func (h *Handler) SdkGetDatasetWithVersionInfo(w http.ResponseWriter, r *http.Request) {
	projectID := projectIDForSDK(r)
	datasetID, err := request.URLParamUUID(r, "datasetId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	resp, err := h.datasetVersionSvc.GetDatasetWithVersionInfo(r.Context(), datasetID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, resp)
}
