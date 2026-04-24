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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	annotationDomain "brokle/internal/core/domain/annotation"
	"brokle/internal/transport/http/httpctx"
	"brokle/pkg/request"
	"brokle/pkg/response"
)

type handler struct {
	queueSvc      annotationDomain.QueueService
	itemSvc       annotationDomain.ItemService
	assignmentSvc annotationDomain.AssignmentService
	logger        *slog.Logger
}

// RegisterRoutes mounts the dashboard-plane annotation routes on r.
// Expected mount context: the authed dashboard chi group.
func RegisterRoutes(
	r chi.Router,
	queueSvc annotationDomain.QueueService,
	itemSvc annotationDomain.ItemService,
	assignmentSvc annotationDomain.AssignmentService,
	logger *slog.Logger,
) {
	h := &handler{queueSvc: queueSvc, itemSvc: itemSvc, assignmentSvc: assignmentSvc, logger: logger}

	r.Route("/api/v1/projects/{projectId}/annotation-queues", func(r chi.Router) {
		r.Post("/", h.createQueue)
		r.Get("/", h.listQueues)
		r.Route("/{queueId}", func(r chi.Router) {
			r.Get("/", h.getQueue)
			r.Get("/stats", h.getQueueWithStats)
			r.Put("/", h.updateQueue)
			r.Delete("/", h.deleteQueue)
			r.Post("/items", h.addItems)
			r.Get("/items", h.listItems)
			r.Post("/items/claim", h.claimNext)
			r.Post("/items/{itemId}/complete", h.completeItem)
			r.Post("/items/{itemId}/skip", h.skipItem)
			r.Post("/items/{itemId}/release", h.releaseLock)
			r.Delete("/items/{itemId}", h.deleteItem)
			r.Post("/assignments", h.assignUser)
			r.Get("/assignments", h.listAssignments)
			r.Delete("/assignments/{userId}", h.unassignUser)
		})
	})

	r.Get("/api/v1/annotation-queues/my-assignments", h.myAssignments)
}

// RegisterSDKRoutes mounts the SDK-plane annotation routes on r
// (project derived from the API key).
func RegisterSDKRoutes(
	r chi.Router,
	itemSvc annotationDomain.ItemService,
	logger *slog.Logger,
) {
	h := &handler{itemSvc: itemSvc, logger: logger}

	r.Route("/v1/annotation-queues/{queueId}/items", func(r chi.Router) {
		r.Post("/", h.sdkAddItems)
		r.Get("/", h.sdkListItems)
	})
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

func (h *handler) createQueue(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

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

func (h *handler) listQueues(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}

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
	response.Success(w, listQueuesResponse{Data: out, Total: total, Page: page, Limit: limit})
}

// ---- queue: get / get-with-stats ------------------------------------

func (h *handler) getQueue(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) getQueueWithStats(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) updateQueue(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) deleteQueue(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) addItems(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) doAddItems(w http.ResponseWriter, r *http.Request, projectID, queueID uuid.UUID, body AddItemsBatchRequest) {
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

func (h *handler) listItems(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	queueID, err := request.URLParamUUID(r, "queueId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
	h.doListItems(w, r, projectID, queueID)
}

func (h *handler) doListItems(w http.ResponseWriter, r *http.Request, projectID, queueID uuid.UUID) {
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
	response.Success(w, listItemsResponse{Data: out, Total: total, Page: page, Limit: limit})
}

// ---- items: claim-next ---------------------------------------------

func (h *handler) claimNext(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) completeItem(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) skipItem(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) releaseLock(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) deleteItem(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) assignUser(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) listAssignments(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) unassignUser(w http.ResponseWriter, r *http.Request) {
	projectID, err := request.URLParamUUID(r, "projectId")
	if err != nil {
		response.WriteError(w, err)
		return
	}
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

func (h *handler) myAssignments(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) sdkAddItems(w http.ResponseWriter, r *http.Request) {
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

func (h *handler) sdkListItems(w http.ResponseWriter, r *http.Request) {
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
