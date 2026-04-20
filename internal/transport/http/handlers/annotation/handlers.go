// Package annotation exposes the HITL annotation queue operations
// on both surfaces:
//   - Dashboard plane (apiAdmin, RequireAuth) — CRUD + item / assignment
//     management under /api/v1/projects/{projectId}/annotation-queues
//   - SDK plane (apiPublic, RequireSDKAuth) — add-items / list-items
//     under /v1/annotation-queues (project ID derived from the API key)
//
// Claim / complete / skip / release-lock flows verify the caller has
// at least the "annotator" role on the queue via AssignmentService.
package annotation

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"

	annotationDomain "brokle/internal/core/domain/annotation"
	"brokle/internal/transport/http/httpctx"
	appErrors "brokle/pkg/errors"
)

type handler struct {
	queueSvc      annotationDomain.QueueService
	itemSvc       annotationDomain.ItemService
	assignmentSvc annotationDomain.AssignmentService
	logger        *slog.Logger
}

// RegisterRoutes wires the dashboard-plane annotation operations.
func RegisterRoutes(
	api huma.API,
	queueSvc annotationDomain.QueueService,
	itemSvc annotationDomain.ItemService,
	assignmentSvc annotationDomain.AssignmentService,
	logger *slog.Logger,
) {
	h := &handler{queueSvc: queueSvc, itemSvc: itemSvc, assignmentSvc: assignmentSvc, logger: logger}

	// ---- queues --------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "create-annotation-queue",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/annotation-queues",
		Tags:          []string{"annotation-queues"},
		Summary:       "Create an annotation queue",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.createQueue)

	huma.Register(api, huma.Operation{
		OperationID: "list-annotation-queues",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/annotation-queues",
		Tags:        []string{"annotation-queues"},
		Summary:     "List annotation queues with statistics",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listQueues)

	huma.Register(api, huma.Operation{
		OperationID: "get-annotation-queue",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/annotation-queues/{queueId}",
		Tags:        []string{"annotation-queues"},
		Summary:     "Get an annotation queue",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getQueue)

	huma.Register(api, huma.Operation{
		OperationID: "get-annotation-queue-with-stats",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/annotation-queues/{queueId}/stats",
		Tags:        []string{"annotation-queues"},
		Summary:     "Get an annotation queue with statistics",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.getQueueWithStats)

	huma.Register(api, huma.Operation{
		OperationID: "update-annotation-queue",
		Method:      http.MethodPut,
		Path:        "/api/v1/projects/{projectId}/annotation-queues/{queueId}",
		Tags:        []string{"annotation-queues"},
		Summary:     "Update an annotation queue",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.updateQueue)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-annotation-queue",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}",
		Tags:          []string{"annotation-queues"},
		Summary:       "Delete an annotation queue (cascades items + assignments)",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteQueue)

	// ---- items ---------------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "add-annotation-items",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}/items",
		Tags:          []string{"annotation-items"},
		Summary:       "Add items (traces or spans) to an annotation queue",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.addItems)

	huma.Register(api, huma.Operation{
		OperationID: "list-annotation-items",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/annotation-queues/{queueId}/items",
		Tags:        []string{"annotation-items"},
		Summary:     "List items in an annotation queue",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listItems)

	huma.Register(api, huma.Operation{
		OperationID: "claim-next-annotation-item",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{projectId}/annotation-queues/{queueId}/items/claim",
		Tags:        []string{"annotation-items"},
		Summary:     "Claim the next pending item (5-minute lock)",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.claimNext)

	huma.Register(api, huma.Operation{
		OperationID:   "complete-annotation-item",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}/items/{itemId}/complete",
		Tags:          []string{"annotation-items"},
		Summary:       "Mark an item complete and optionally submit scores",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.completeItem)

	huma.Register(api, huma.Operation{
		OperationID:   "skip-annotation-item",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}/items/{itemId}/skip",
		Tags:          []string{"annotation-items"},
		Summary:       "Skip an item",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.skipItem)

	huma.Register(api, huma.Operation{
		OperationID:   "release-annotation-item-lock",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}/items/{itemId}/release",
		Tags:          []string{"annotation-items"},
		Summary:       "Release the lock on an item, returning it to the pending pool",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.releaseLock)

	huma.Register(api, huma.Operation{
		OperationID:   "delete-annotation-item",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}/items/{itemId}",
		Tags:          []string{"annotation-items"},
		Summary:       "Delete an item",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.deleteItem)

	// ---- assignments --------------------------------------------
	huma.Register(api, huma.Operation{
		OperationID:   "assign-annotation-queue-user",
		Method:        http.MethodPost,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}/assignments",
		Tags:          []string{"annotation-assignments"},
		Summary:       "Assign a user to an annotation queue",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.assignUser)

	huma.Register(api, huma.Operation{
		OperationID: "list-annotation-queue-assignments",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{projectId}/annotation-queues/{queueId}/assignments",
		Tags:        []string{"annotation-assignments"},
		Summary:     "List users assigned to an annotation queue",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.listAssignments)

	huma.Register(api, huma.Operation{
		OperationID:   "unassign-annotation-queue-user",
		Method:        http.MethodDelete,
		Path:          "/api/v1/projects/{projectId}/annotation-queues/{queueId}/assignments/{userId}",
		Tags:          []string{"annotation-assignments"},
		Summary:       "Remove a user's annotation queue assignment",
		Security:      []map[string][]string{{"bearerAuth": {}}},
		DefaultStatus: http.StatusNoContent,
	}, h.unassignUser)

	huma.Register(api, huma.Operation{
		OperationID: "list-my-annotation-assignments",
		Method:      http.MethodGet,
		Path:        "/api/v1/annotation-queues/my-assignments",
		Tags:        []string{"annotation-assignments"},
		Summary:     "List my annotation queue assignments",
		Security:    []map[string][]string{{"bearerAuth": {}}},
	}, h.myAssignments)
}

// RegisterSDKRoutes wires the SDK-plane annotation operations
// (project derived from the API key).
func RegisterSDKRoutes(
	api huma.API,
	itemSvc annotationDomain.ItemService,
	logger *slog.Logger,
) {
	h := &handler{itemSvc: itemSvc, logger: logger}

	huma.Register(api, huma.Operation{
		OperationID:   "sdk-add-annotation-items",
		Method:        http.MethodPost,
		Path:          "/v1/annotation-queues/{queueId}/items",
		Tags:          []string{"SDK - annotation-items"},
		Summary:       "Add items to an annotation queue (SDK)",
		Security:      []map[string][]string{{"apiKey": {}}},
		DefaultStatus: http.StatusCreated,
	}, h.sdkAddItems)

	huma.Register(api, huma.Operation{
		OperationID: "sdk-list-annotation-items",
		Method:      http.MethodGet,
		Path:        "/v1/annotation-queues/{queueId}/items",
		Tags:        []string{"SDK - annotation-items"},
		Summary:     "List items in an annotation queue (SDK)",
		Security:    []map[string][]string{{"apiKey": {}}},
	}, h.sdkListItems)
}

// ---- shared parsers -------------------------------------------------

func parseProject(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid project ID", "projectId must be a valid UUID")
	}
	return id, nil
}

func parseQueue(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid queue ID", "queueId must be a valid UUID")
	}
	return id, nil
}

func parseItem(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid item ID", "itemId must be a valid UUID")
	}
	return id, nil
}

func parseUser(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, appErrors.NewValidationError("Invalid user ID", "userId must be a valid UUID")
	}
	return id, nil
}

func userIDPtr(ctx context.Context) *uuid.UUID {
	uid, ok := httpctx.UserID(ctx)
	if !ok {
		return nil
	}
	return &uid
}

// ---- queue: create ---------------------------------------------------

type CreateQueueInput struct {
	ProjectID string             `path:"projectId" format:"uuid"`
	Body      CreateQueueRequest
}

type CreateQueueOutput struct {
	Body *QueueResponse
}

func (h *handler) createQueue(ctx context.Context, in *CreateQueueInput) (*CreateQueueOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}

	domainReq := &annotationDomain.CreateQueueRequest{
		Name:           in.Body.Name,
		Description:    in.Body.Description,
		Instructions:   in.Body.Instructions,
		ScoreConfigIDs: in.Body.ScoreConfigIDs,
	}
	if in.Body.Settings != nil {
		domainReq.Settings = &annotationDomain.QueueSettings{
			LockTimeoutSeconds: in.Body.Settings.LockTimeoutSeconds,
			AutoAssignment:     in.Body.Settings.AutoAssignment,
		}
	}

	queue, err := h.queueSvc.Create(ctx, projectID, userIDPtr(ctx), domainReq)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "annotation: queue created", "queue_id", queue.ID, "project_id", projectID)
	return &CreateQueueOutput{Body: toQueueResponse(queue)}, nil
}

// ---- queue: list -----------------------------------------------------

type ListQueuesInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
	Status    string `query:"status" required:"false" enum:"active,paused,archived"`
	Search    string `query:"search" required:"false"`
}

type ListQueuesOutput struct {
	Body listQueuesResponse
}

type listQueuesResponse struct {
	Data  []*QueueWithStatsResponse `json:"data"`
	Total int64                     `json:"total"`
	Page  int                       `json:"page"`
	Limit int                       `json:"limit"`
}

func (h *handler) listQueues(ctx context.Context, in *ListQueuesInput) (*ListQueuesOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}

	page := in.Page
	if page < 1 {
		page = 1
	}
	limit := in.Limit
	if limit <= 0 {
		limit = 50
	}

	filter := &annotationDomain.QueueFilter{}
	if in.Status != "" {
		qs := annotationDomain.QueueStatus(in.Status)
		if qs.IsValid() {
			filter.Status = &qs
		}
	}
	if in.Search != "" {
		s := in.Search
		filter.Search = &s
	}

	queues, stats, total, err := h.queueSvc.ListWithStats(ctx, projectID, filter, page, limit)
	if err != nil {
		return nil, err
	}

	out := make([]*QueueWithStatsResponse, len(queues))
	for i, q := range queues {
		out[i] = &QueueWithStatsResponse{
			Queue: toQueueResponse(q),
			Stats: toStatsResponse(stats[i]),
		}
	}
	return &ListQueuesOutput{Body: listQueuesResponse{Data: out, Total: total, Page: page, Limit: limit}}, nil
}

// ---- queue: get / get-with-stats ------------------------------------

type GetQueueInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	QueueID   string `path:"queueId" format:"uuid"`
}

type GetQueueOutput struct {
	Body *QueueResponse
}

func (h *handler) getQueue(ctx context.Context, in *GetQueueInput) (*GetQueueOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	queue, err := h.queueSvc.GetByID(ctx, queueID, projectID)
	if err != nil {
		return nil, err
	}
	return &GetQueueOutput{Body: toQueueResponse(queue)}, nil
}

type GetQueueWithStatsOutput struct {
	Body *QueueWithStatsResponse
}

func (h *handler) getQueueWithStats(ctx context.Context, in *GetQueueInput) (*GetQueueWithStatsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	queue, stats, err := h.queueSvc.GetWithStats(ctx, queueID, projectID)
	if err != nil {
		return nil, err
	}
	return &GetQueueWithStatsOutput{Body: &QueueWithStatsResponse{
		Queue: toQueueResponse(queue),
		Stats: toStatsResponse(stats),
	}}, nil
}

// ---- queue: update --------------------------------------------------

type UpdateQueueInput struct {
	ProjectID string             `path:"projectId" format:"uuid"`
	QueueID   string             `path:"queueId" format:"uuid"`
	Body      UpdateQueueRequest
}

type UpdateQueueOutput struct {
	Body *QueueResponse
}

func (h *handler) updateQueue(ctx context.Context, in *UpdateQueueInput) (*UpdateQueueOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}

	domainReq := &annotationDomain.UpdateQueueRequest{
		Name:           in.Body.Name,
		Description:    in.Body.Description,
		Instructions:   in.Body.Instructions,
		ScoreConfigIDs: in.Body.ScoreConfigIDs,
	}
	if in.Body.Status != nil {
		s := annotationDomain.QueueStatus(*in.Body.Status)
		domainReq.Status = &s
	}
	if in.Body.Settings != nil {
		domainReq.Settings = &annotationDomain.QueueSettings{
			LockTimeoutSeconds: in.Body.Settings.LockTimeoutSeconds,
			AutoAssignment:     in.Body.Settings.AutoAssignment,
		}
	}

	queue, err := h.queueSvc.Update(ctx, queueID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	return &UpdateQueueOutput{Body: toQueueResponse(queue)}, nil
}

// ---- queue: delete --------------------------------------------------

type DeleteQueueInput = GetQueueInput
type DeleteQueueOutput struct{}

func (h *handler) deleteQueue(ctx context.Context, in *DeleteQueueInput) (*DeleteQueueOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	if err := h.queueSvc.Delete(ctx, queueID, projectID); err != nil {
		return nil, err
	}
	return &DeleteQueueOutput{}, nil
}

// ---- items: add / list ---------------------------------------------

type AddItemsInput struct {
	ProjectID string               `path:"projectId" format:"uuid"`
	QueueID   string               `path:"queueId" format:"uuid"`
	Body      AddItemsBatchRequest
}

type AddItemsOutput struct {
	Body *BatchAddItemsResponse
}

func (h *handler) addItems(ctx context.Context, in *AddItemsInput) (*AddItemsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	return h.doAddItems(ctx, projectID, queueID, in.Body)
}

func (h *handler) doAddItems(ctx context.Context, projectID, queueID uuid.UUID, body AddItemsBatchRequest) (*AddItemsOutput, error) {
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
	count, err := h.itemSvc.AddItems(ctx, queueID, projectID, domainReq)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "annotation: items added", "queue_id", queueID, "project_id", projectID, "count", count)
	return &AddItemsOutput{Body: &BatchAddItemsResponse{Created: count}}, nil
}

type ListItemsInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	QueueID   string `path:"queueId" format:"uuid"`
	Page      int    `query:"page" required:"false" minimum:"1"`
	Limit     int    `query:"limit" required:"false"`
	Status    string `query:"status" required:"false" enum:"pending,completed,skipped"`
}

type ListItemsOutput struct {
	Body listItemsResponse
}

type listItemsResponse struct {
	Data  []*ItemResponse `json:"data"`
	Total int64           `json:"total"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
}

func (h *handler) listItems(ctx context.Context, in *ListItemsInput) (*ListItemsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	return h.doListItems(ctx, projectID, queueID, in.Page, in.Limit, in.Status)
}

func (h *handler) doListItems(ctx context.Context, projectID, queueID uuid.UUID, page, limit int, status string) (*ListItemsOutput, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	filter := &annotationDomain.ItemFilter{
		Limit:  limit,
		Offset: (page - 1) * limit,
	}
	if status != "" {
		s := annotationDomain.ItemStatus(status)
		filter.Status = &s
	}
	items, total, err := h.itemSvc.ListItems(ctx, queueID, projectID, filter)
	if err != nil {
		return nil, err
	}
	out := make([]*ItemResponse, len(items))
	for i, item := range items {
		out[i] = toItemResponse(item)
	}
	return &ListItemsOutput{Body: listItemsResponse{Data: out, Total: total, Page: page, Limit: limit}}, nil
}

// ---- items: claim-next ---------------------------------------------

type ClaimNextInput struct {
	ProjectID string           `path:"projectId" format:"uuid"`
	QueueID   string           `path:"queueId" format:"uuid"`
	Body      ClaimNextRequest `required:"false"`
}

type ClaimNextOutput struct {
	Body *ItemResponse
}

func (h *handler) claimNext(ctx context.Context, in *ClaimNextInput) (*ClaimNextOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)

	if err := h.assignmentSvc.CheckAccess(ctx, queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		return nil, err
	}
	item, err := h.itemSvc.ClaimNext(ctx, queueID, projectID, userID, in.Body.SeenItemIDs)
	if err != nil {
		return nil, err
	}
	return &ClaimNextOutput{Body: toItemResponse(item)}, nil
}

// ---- items: complete / skip / release-lock -------------------------

type CompleteItemInput struct {
	ProjectID string              `path:"projectId" format:"uuid"`
	QueueID   string              `path:"queueId" format:"uuid"`
	ItemID    string              `path:"itemId" format:"uuid"`
	Body      CompleteItemRequest `required:"false"`
}

type CompleteItemOutput struct{}

func (h *handler) completeItem(ctx context.Context, in *CompleteItemInput) (*CompleteItemOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	itemID, err := parseItem(in.ItemID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.assignmentSvc.CheckAccess(ctx, queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		return nil, err
	}
	domainReq := &annotationDomain.CompleteItemRequest{
		Scores: make([]annotationDomain.ScoreSubmission, len(in.Body.Scores)),
	}
	for i, s := range in.Body.Scores {
		domainReq.Scores[i] = annotationDomain.ScoreSubmission{
			ScoreConfigID: s.ScoreConfigID,
			Value:         s.Value,
			Comment:       s.Comment,
		}
	}
	if err := h.itemSvc.Complete(ctx, itemID, queueID, projectID, userID, domainReq); err != nil {
		return nil, err
	}
	return &CompleteItemOutput{}, nil
}

type SkipItemInput struct {
	ProjectID string          `path:"projectId" format:"uuid"`
	QueueID   string          `path:"queueId" format:"uuid"`
	ItemID    string          `path:"itemId" format:"uuid"`
	Body      SkipItemRequest `required:"false"`
}

type SkipItemOutput struct{}

func (h *handler) skipItem(ctx context.Context, in *SkipItemInput) (*SkipItemOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	itemID, err := parseItem(in.ItemID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.assignmentSvc.CheckAccess(ctx, queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		return nil, err
	}
	if err := h.itemSvc.Skip(ctx, itemID, queueID, projectID, userID, &annotationDomain.SkipItemRequest{Reason: in.Body.Reason}); err != nil {
		return nil, err
	}
	return &SkipItemOutput{}, nil
}

type ReleaseLockInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	QueueID   string `path:"queueId" format:"uuid"`
	ItemID    string `path:"itemId" format:"uuid"`
}

type ReleaseLockOutput struct{}

func (h *handler) releaseLock(ctx context.Context, in *ReleaseLockInput) (*ReleaseLockOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	itemID, err := parseItem(in.ItemID)
	if err != nil {
		return nil, err
	}
	userID := httpctx.MustGetUserID(ctx)
	if err := h.assignmentSvc.CheckAccess(ctx, queueID, userID, annotationDomain.RoleAnnotator); err != nil {
		return nil, err
	}
	if err := h.itemSvc.ReleaseLock(ctx, itemID, queueID, projectID, userID); err != nil {
		return nil, err
	}
	return &ReleaseLockOutput{}, nil
}

type DeleteItemInput = ReleaseLockInput
type DeleteItemOutput struct{}

func (h *handler) deleteItem(ctx context.Context, in *DeleteItemInput) (*DeleteItemOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	itemID, err := parseItem(in.ItemID)
	if err != nil {
		return nil, err
	}
	if err := h.itemSvc.DeleteItem(ctx, itemID, queueID, projectID); err != nil {
		return nil, err
	}
	return &DeleteItemOutput{}, nil
}

// ---- assignments ---------------------------------------------------

type AssignQueueUserInput struct {
	ProjectID string            `path:"projectId" format:"uuid"`
	QueueID   string            `path:"queueId" format:"uuid"`
	Body      AssignUserRequest
}

type AssignQueueUserOutput struct {
	Body *AssignmentResponse
}

func (h *handler) assignUser(ctx context.Context, in *AssignQueueUserInput) (*AssignQueueUserOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	assignment, err := h.assignmentSvc.Assign(
		ctx,
		queueID,
		projectID,
		in.Body.UserID,
		annotationDomain.AssignmentRole(in.Body.Role),
		userIDPtr(ctx),
	)
	if err != nil {
		return nil, err
	}
	h.logger.InfoContext(ctx, "annotation: user assigned", "queue_id", queueID, "user_id", in.Body.UserID, "role", in.Body.Role)
	return &AssignQueueUserOutput{Body: toAssignmentResponse(assignment)}, nil
}

type ListAssignmentsInput = GetQueueInput
type ListAssignmentsOutput struct {
	Body []*AssignmentResponse
}

func (h *handler) listAssignments(ctx context.Context, in *ListAssignmentsInput) (*ListAssignmentsOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	assignments, err := h.assignmentSvc.ListAssignments(ctx, queueID, projectID)
	if err != nil {
		return nil, err
	}
	out := make([]*AssignmentResponse, len(assignments))
	for i, a := range assignments {
		out[i] = toAssignmentResponse(a)
	}
	return &ListAssignmentsOutput{Body: out}, nil
}

type UnassignQueueUserInput struct {
	ProjectID string `path:"projectId" format:"uuid"`
	QueueID   string `path:"queueId" format:"uuid"`
	UserID    string `path:"userId" format:"uuid"`
}

type UnassignQueueUserOutput struct{}

func (h *handler) unassignUser(ctx context.Context, in *UnassignQueueUserInput) (*UnassignQueueUserOutput, error) {
	projectID, err := parseProject(in.ProjectID)
	if err != nil {
		return nil, err
	}
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	userID, err := parseUser(in.UserID)
	if err != nil {
		return nil, err
	}
	if err := h.assignmentSvc.Unassign(ctx, queueID, projectID, userID); err != nil {
		return nil, err
	}
	return &UnassignQueueUserOutput{}, nil
}

type MyAssignmentsOutput struct {
	Body []*AssignmentResponse
}

func (h *handler) myAssignments(ctx context.Context, _ *struct{}) (*MyAssignmentsOutput, error) {
	userID := httpctx.MustGetUserID(ctx)
	assignments, err := h.assignmentSvc.GetUserQueues(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]*AssignmentResponse, len(assignments))
	for i, a := range assignments {
		out[i] = toAssignmentResponse(a)
	}
	return &MyAssignmentsOutput{Body: out}, nil
}

// ---- SDK handlers --------------------------------------------------

type SDKAddItemsInput struct {
	QueueID string               `path:"queueId" format:"uuid"`
	Body    AddItemsBatchRequest
}

func (h *handler) sdkAddItems(ctx context.Context, in *SDKAddItemsInput) (*AddItemsOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	return h.doAddItems(ctx, projectID, queueID, in.Body)
}

type SDKListItemsInput struct {
	QueueID string `path:"queueId" format:"uuid"`
	Page    int    `query:"page" required:"false" minimum:"1"`
	Limit   int    `query:"limit" required:"false"`
	Status  string `query:"status" required:"false" enum:"pending,completed,skipped"`
}

func (h *handler) sdkListItems(ctx context.Context, in *SDKListItemsInput) (*ListItemsOutput, error) {
	projectID := httpctx.MustGetProjectID(ctx)
	queueID, err := parseQueue(in.QueueID)
	if err != nil {
		return nil, err
	}
	return h.doListItems(ctx, projectID, queueID, in.Page, in.Limit, in.Status)
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
