// Package annotation exposes the HITL annotation queue operations on
// both surfaces:
//   - Dashboard plane (RequireAuth) — CRUD + item / assignment
//     management under /api/v1/projects/{projectId}/annotation-queues
//   - SDK plane (RequireSDKAuth) — add-items / list-items under
//     /v1/annotation-queues (project ID derived from the API key)
//
// Claim / complete / skip / release-lock flows verify the caller has
// at least the "annotator" role on the queue via AssignmentService.
package annotation

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	annotationDomain "brokle/internal/core/domain/annotation"
	annotationService "brokle/internal/core/services/annotation"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type Handler struct {
	queueSvc      *annotationService.QueueService
	itemSvc       *annotationService.ItemService
	assignmentSvc *annotationService.AssignmentService
	logger        *slog.Logger
}

// newHandler builds the shared handler.
func New(
	queueSvc *annotationService.QueueService,
	itemSvc *annotationService.ItemService,
	assignmentSvc *annotationService.AssignmentService,
	logger *slog.Logger,
) *Handler {
	return &Handler{queueSvc: queueSvc, itemSvc: itemSvc, assignmentSvc: assignmentSvc, logger: logger}
}

// ---- shared helpers --------------------------------------------------

func userIDPtr(ctx context.Context) *uuid.UUID {
	uid, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &uid
}

// ---- queue: create ---------------------------------------------------

func (h *Handler) CreateQueue(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	var body CreateQueueRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	domainReq := &annotationDomain.CreateQueueRequest{
		Name:           body.Name,
		Description:    body.Description,
		Instructions:   body.Instructions,
		ScoreConfigIDs: body.ScoreConfigIDs,
	}
	if body.Settings != nil {
		domainReq.Settings = &annotationDomain.QueueSettings{
			LockTimeoutSeconds: body.Settings.LockTimeoutSeconds,
			AutoAssignment:     body.Settings.AutoAssignment,
		}
	}

	queue, err := h.queueSvc.Create(r.Context(), projectID, userIDPtr(r.Context()), domainReq)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "annotation: queue created",
		"queue_id", queue.ID, "project_id", projectID)
	response.Created(w, toQueueResponse(queue))
}

// ---- queue: list -----------------------------------------------------

func (h *Handler) ListQueues(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())

	page, err := request.QueryInt(r, "page", 1)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if page < 1 {
		page = 1
	}
	limit, err := request.QueryInt(r, "limit", 50)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit <= 0 {
		limit = 50
	}

	filter := &annotationDomain.QueueFilter{}
	if status := r.URL.Query().Get("status"); status != "" {
		qs := annotationDomain.QueueStatus(status)
		if qs.IsValid() {
			filter.Status = &qs
		}
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}

	queues, stats, total, err := h.queueSvc.ListWithStats(r.Context(), projectID, filter, page, limit)
	if err != nil {
		response.WriteError(w, err)
		return
	}

	out := make([]*QueueWithStatsResponse, len(queues))
	for i, q := range queues {
		out[i] = &QueueWithStatsResponse{
			Queue: toQueueResponse(q),
			Stats: toStatsResponse(stats[i]),
		}
	}
	response.Success(w, listQueuesResponse{Data: out, Pagination: response.BuildPagination(page, limit, total)})
}

// ---- queue: get / get-with-stats ------------------------------------

func (h *Handler) GetQueue(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	queue, err := h.queueSvc.GetByID(r.Context(), queueID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toQueueResponse(queue))
}

func (h *Handler) GetQueueWithStats(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	queue, stats, err := h.queueSvc.GetWithStats(r.Context(), queueID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, &QueueWithStatsResponse{
		Queue: toQueueResponse(queue),
		Stats: toStatsResponse(stats),
	})
}

// ---- queue: update --------------------------------------------------

func (h *Handler) UpdateQueue(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body UpdateQueueRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}

	domainReq := &annotationDomain.UpdateQueueRequest{
		Name:           body.Name,
		Description:    body.Description,
		Instructions:   body.Instructions,
		ScoreConfigIDs: body.ScoreConfigIDs,
	}
	if body.Status != nil {
		s := annotationDomain.QueueStatus(*body.Status)
		domainReq.Status = &s
	}
	if body.Settings != nil {
		domainReq.Settings = &annotationDomain.QueueSettings{
			LockTimeoutSeconds: body.Settings.LockTimeoutSeconds,
			AutoAssignment:     body.Settings.AutoAssignment,
		}
	}

	queue, err := h.queueSvc.Update(r.Context(), queueID, projectID, domainReq)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toQueueResponse(queue))
}

// ---- queue: delete --------------------------------------------------

func (h *Handler) DeleteQueue(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.queueSvc.Delete(r.Context(), queueID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ---- items: add / list ---------------------------------------------

func (h *Handler) AddItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

	var body AddItemsBatchRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	h.doAddItems(w, r, projectID, queueID, body)
}

func (h *Handler) doAddItems(w http.ResponseWriter, r *http.Request, projectID, queueID uuid.UUID, body AddItemsBatchRequest) {
	domainReq := &annotationDomain.AddItemsBatchRequest{
		Items: make([]annotationDomain.AddItemRequest, len(body.Items)),
	}
	for i, item := range body.Items {
		domainReq.Items[i] = annotationDomain.AddItemRequest{
			ObjectID:   item.ObjectID,
			ObjectType: annotationDomain.ObjectType(item.ObjectType),
			Priority:   item.Priority,
			Metadata:   item.Metadata,
		}
	}
	count, err := h.itemSvc.AddItems(r.Context(), queueID, projectID, domainReq)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "annotation: items added",
		"queue_id", queueID, "project_id", projectID, "count", count)
	response.Created(w, &BatchAddItemsResponse{Created: count})
}

func (h *Handler) ListItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.doListItems(w, r, projectID, queueID)
}

func (h *Handler) doListItems(w http.ResponseWriter, r *http.Request, projectID, queueID uuid.UUID) {
	page, err := request.QueryInt(r, "page", 1)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if page < 1 {
		page = 1
	}
	limit, err := request.QueryInt(r, "limit", 50)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if limit <= 0 {
		limit = 50
	}
	status := r.URL.Query().Get("status")

	filter := &annotationDomain.ItemFilter{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
	if status != "" {
		s := annotationDomain.ItemStatus(status)
		filter.Status = &s
	}
	items, total, err := h.itemSvc.ListItems(r.Context(), queueID, projectID, filter)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*ItemResponse, len(items))
	for i, item := range items {
		out[i] = toItemResponse(item)
	}
	response.Success(w, listItemsResponse{Data: out, Pagination: response.BuildPagination(page, limit, total)})
}

// ---- items: claim-next ---------------------------------------------

func (h *Handler) ClaimNext(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	// Body is optional.
	var body ClaimNextRequest
	if r.ContentLength > 0 {
		if err := request.DecodeJSON(r, &body); err != nil {
			response.WriteError(w, err)
			return
		}
	}

	if err := h.assignmentSvc.CheckAccess(r.Context(), queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		response.WriteError(w, err)
		return
	}
	item, err := h.itemSvc.ClaimNext(r.Context(), queueID, projectID, userID, body.SeenItemIDs)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	response.Success(w, toItemResponse(item))
}

// ---- items: complete / skip / release-lock -------------------------

func (h *Handler) CompleteItem(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	itemID, err := request.URLParamUUID(r, "itemId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body CompleteItemRequest
	if r.ContentLength > 0 {
		if err := request.DecodeJSON(r, &body); err != nil {
			response.WriteError(w, err)
			return
		}
	}

	if err := h.assignmentSvc.CheckAccess(r.Context(), queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		response.WriteError(w, err)
		return
	}
	domainReq := &annotationDomain.CompleteItemRequest{
		Scores: make([]annotationDomain.ScoreSubmission, len(body.Scores)),
	}
	for i, s := range body.Scores {
		domainReq.Scores[i] = annotationDomain.ScoreSubmission{
			ScoreConfigID: s.ScoreConfigID,
			Value:         s.Value,
			Comment:       s.Comment,
		}
	}
	if err := h.itemSvc.Complete(r.Context(), itemID, queueID, projectID, userID, domainReq); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) SkipItem(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	itemID, err := request.URLParamUUID(r, "itemId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())

	var body SkipItemRequest
	if r.ContentLength > 0 {
		if err := request.DecodeJSON(r, &body); err != nil {
			response.WriteError(w, err)
			return
		}
	}

	if err := h.assignmentSvc.CheckAccess(r.Context(), queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.itemSvc.Skip(r.Context(), itemID, queueID, projectID, userID,
		&annotationDomain.SkipItemRequest{Reason: body.Reason}); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) ReleaseLock(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	itemID, err := request.URLParamUUID(r, "itemId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID := httpctx.MustGetUserID(r.Context())
	if err := h.assignmentSvc.CheckAccess(r.Context(), queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.itemSvc.ReleaseLock(r.Context(), itemID, queueID, projectID, userID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	itemID, err := request.URLParamUUID(r, "itemId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.itemSvc.DeleteItem(r.Context(), itemID, queueID, projectID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

// ---- assignments ---------------------------------------------------

func (h *Handler) AssignUser(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body AssignUserRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	assignment, err := h.assignmentSvc.Assign(
		r.Context(),
		queueID,
		projectID,
		body.UserID,
		annotationDomain.AssignmentRole(body.Role),
		userIDPtr(r.Context()),
	)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.logger.InfoContext(r.Context(), "annotation: user assigned",
		"queue_id", queueID, "user_id", body.UserID, "role", body.Role)
	response.Created(w, toAssignmentResponse(assignment))
}

func (h *Handler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	assignments, err := h.assignmentSvc.ListAssignments(r.Context(), queueID, projectID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*AssignmentResponse, len(assignments))
	for i, a := range assignments {
		out[i] = toAssignmentResponse(a)
	}
	response.Success(w, out)
}

func (h *Handler) UnassignUser(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	userID, err := request.URLParamUUID(r, "userId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	if err := h.assignmentSvc.Unassign(r.Context(), queueID, projectID, userID); err != nil {
		response.WriteError(w, err)
		return
	}
	response.NoContent(w)
}

func (h *Handler) MyAssignments(w http.ResponseWriter, r *http.Request) {
	userID := httpctx.MustGetUserID(r.Context())
	assignments, err := h.assignmentSvc.GetUserQueues(r.Context(), userID)
	if err != nil {
		response.WriteError(w, err)
		return
	}
	out := make([]*AssignmentResponse, len(assignments))
	for i, a := range assignments {
		out[i] = toAssignmentResponse(a)
	}
	response.Success(w, out)
}

// ---- SDK handlers --------------------------------------------------

func (h *Handler) SdkAddItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	var body AddItemsBatchRequest
	if err := request.DecodeJSON(r, &body); err != nil {
		response.WriteError(w, err)
		return
	}
	h.doAddItems(w, r, projectID, queueID, body)
}

func (h *Handler) SdkListItems(w http.ResponseWriter, r *http.Request) {
	projectID := httpctx.MustGetProjectID(r.Context())
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.doListItems(w, r, projectID, queueID)
}

// ---- mappers -------------------------------------------------------

func toQueueResponse(q *annotationDomain.AnnotationQueue) *QueueResponse {
	return &QueueResponse{
		ID:             q.ID,
		ProjectID:      q.ProjectID,
		Name:           q.Name,
		Description:    q.Description,
		Instructions:   q.Instructions,
		ScoreConfigIDs: q.ScoreConfigIDs,
		Status:         string(q.Status),
		Settings: &QueueSettings{
			LockTimeoutSeconds: q.Settings.LockTimeoutSeconds,
			AutoAssignment:     q.Settings.AutoAssignment,
		},
		CreatedBy: q.CreatedBy,
		CreatedAt: q.CreatedAt,
		UpdatedAt: q.UpdatedAt,
	}
}

func toStatsResponse(s *annotationDomain.QueueStats) *StatsResponse {
	return &StatsResponse{
		TotalItems:      s.TotalItems,
		PendingItems:    s.PendingItems,
		InProgressItems: s.InProgressItems,
		CompletedItems:  s.CompletedItems,
		SkippedItems:    s.SkippedItems,
	}
}

func toItemResponse(i *annotationDomain.QueueItem) *ItemResponse {
	return &ItemResponse{
		ID:              i.ID,
		QueueID:         i.QueueID,
		ObjectID:        i.ObjectID,
		ObjectType:      string(i.ObjectType),
		Status:          string(i.Status),
		Priority:        i.Priority,
		LockedAt:        i.LockedAt,
		LockedByUserID:  i.LockedByUserID,
		AnnotatorUserID: i.AnnotatorUserID,
		CompletedAt:     i.CompletedAt,
		Metadata:        i.Metadata,
		CreatedAt:       i.CreatedAt,
		UpdatedAt:       i.UpdatedAt,
	}
}

func toAssignmentResponse(a *annotationDomain.QueueAssignment) *AssignmentResponse {
	return &AssignmentResponse{
		ID:         a.ID,
		QueueID:    a.QueueID,
		UserID:     a.UserID,
		Role:       string(a.Role),
		AssignedAt: a.AssignedAt,
		AssignedBy: a.AssignedBy,
	}
}
