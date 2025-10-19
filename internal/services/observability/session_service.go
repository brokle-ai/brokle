package observability

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"brokle/internal/core/domain/observability"
	appErrors "brokle/pkg/errors"
	"brokle/pkg/ulid"
)

// SessionService implements business logic for session management
type SessionService struct {
	sessionRepo observability.SessionRepository
	traceRepo   observability.TraceRepository
	scoreRepo   observability.ScoreRepository
}

// NewSessionService creates a new session service instance
func NewSessionService(
	sessionRepo observability.SessionRepository,
	traceRepo observability.TraceRepository,
	scoreRepo observability.ScoreRepository,
) *SessionService {
	return &SessionService{
		sessionRepo: sessionRepo,
		traceRepo:   traceRepo,
		scoreRepo:   scoreRepo,
	}
}

// CreateSession creates a new session with validation
func (s *SessionService) CreateSession(ctx context.Context, session *observability.Session) error {
	// Validate required fields
	if session.ProjectID.IsZero() {
		return appErrors.NewValidationError("project_id is required", "session must have a valid project_id")
	}

	// Generate new ID if not provided
	if session.ID.IsZero() {
		session.ID = ulid.New()
	}

	// Create session
	if err := s.sessionRepo.Create(ctx, session); err != nil {
		return appErrors.NewInternalError("failed to create session", err)
	}

	return nil
}

// UpdateSession updates an existing session with partial update support
func (s *SessionService) UpdateSession(ctx context.Context, sessionID ulid.ULID, updateReq *observability.UpdateSessionRequest) error {
	// Validate session exists
	existing, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("session %s", sessionID.String()))
		}
		return appErrors.NewInternalError("failed to get session", err)
	}

	// Merge non-nil fields from update request into existing
	mergeSessionUpdate(existing, updateReq)

	// Preserve version for increment in repository layer
	existing.Version = existing.Version

	// Update session
	if err := s.sessionRepo.Update(ctx, existing); err != nil {
		return appErrors.NewInternalError("failed to update session", err)
	}

	return nil
}

// mergeSessionUpdate merges non-nil fields from update request into existing session
// This prevents zero-value corruption from partial JSON updates
func mergeSessionUpdate(dst *observability.Session, src *observability.UpdateSessionRequest) {
	// Immutable fields (never update):
	// - ID (primary key)
	// - ProjectID (security boundary)
	// - CreatedAt (timestamp of creation)
	// - Version (managed by repository)
	// - EventTs (managed by repository)
	// - IsDeleted (managed by Delete method)

	// Update optional fields only if non-nil (field was sent in JSON)
	if src.UserID != nil && !src.UserID.IsZero() {
		dst.UserID = src.UserID
	}
	// Allow clearing metadata by sending empty map {}
	// nil = not sent (preserve), {} = clear, {...} = update
	if src.Metadata != nil {
		dst.Metadata = src.Metadata
	}
	// Boolean fields only updated if explicitly sent (non-nil pointer)
	// nil = field not sent, preserve existing value
	// &true or &false = field explicitly sent, update value
	if src.Bookmarked != nil {
		dst.Bookmarked = *src.Bookmarked
	}
	if src.Public != nil {
		dst.Public = *src.Public
	}
}

// DeleteSession soft deletes a session
func (s *SessionService) DeleteSession(ctx context.Context, id ulid.ULID) error {
	// Validate session exists
	_, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("session %s", id.String()))
		}
		return appErrors.NewInternalError("failed to get session", err)
	}

	// Delete session
	if err := s.sessionRepo.Delete(ctx, id); err != nil {
		return appErrors.NewInternalError("failed to delete session", err)
	}

	return nil
}

// GetSessionByID retrieves a session by ID
func (s *SessionService) GetSessionByID(ctx context.Context, id ulid.ULID) (*observability.Session, error) {
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("session %s", id.String()))
		}
		return nil, appErrors.NewInternalError("failed to get session", err)
	}

	return session, nil
}

// GetSessionWithTraces retrieves a session with all its traces
func (s *SessionService) GetSessionWithTraces(ctx context.Context, id ulid.ULID) (*observability.Session, error) {
	// Get session
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("session %s", id.String()))
		}
		return nil, appErrors.NewInternalError("failed to get session", err)
	}

	// Get traces for session
	traces, err := s.traceRepo.GetBySessionID(ctx, id)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get session traces", err)
	}

	session.Traces = traces

	return session, nil
}

// GetSessionWithScores retrieves a session with all its quality scores
func (s *SessionService) GetSessionWithScores(ctx context.Context, id ulid.ULID) (*observability.Session, error) {
	// Get session
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, appErrors.NewNotFoundError(fmt.Sprintf("session %s", id.String()))
		}
		return nil, appErrors.NewInternalError("failed to get session", err)
	}

	// Get scores for session
	scores, err := s.scoreRepo.GetBySessionID(ctx, id)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get session scores", err)
	}

	session.Scores = scores

	return session, nil
}

// GetSessionsByProjectID retrieves sessions for a project with optional filtering
func (s *SessionService) GetSessionsByProjectID(ctx context.Context, projectID ulid.ULID, filter *observability.SessionFilter) ([]*observability.Session, error) {
	if projectID.IsZero() {
		return nil, appErrors.NewValidationError("project_id is required", "sessions query requires a valid project_id")
	}

	sessions, err := s.sessionRepo.GetByProjectID(ctx, projectID, filter)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get sessions", err)
	}

	return sessions, nil
}

// GetSessionsByUserID retrieves sessions for a user
func (s *SessionService) GetSessionsByUserID(ctx context.Context, userID ulid.ULID) ([]*observability.Session, error) {
	if userID.IsZero() {
		return nil, appErrors.NewValidationError("user_id is required", "sessions query requires a valid user_id")
	}

	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, appErrors.NewInternalError("failed to get sessions", err)
	}

	return sessions, nil
}

// ToggleBookmark toggles the bookmarked flag for a session
func (s *SessionService) ToggleBookmark(ctx context.Context, id ulid.ULID) error {
	// Get session
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("session %s", id.String()))
		}
		return appErrors.NewInternalError("failed to get session", err)
	}

	// Toggle bookmarked
	session.Bookmarked = !session.Bookmarked

	// Update session
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return appErrors.NewInternalError("failed to toggle bookmark", err)
	}

	return nil
}

// TogglePublic toggles the public flag for a session
func (s *SessionService) TogglePublic(ctx context.Context, id ulid.ULID) error {
	// Get session
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("session %s", id.String()))
		}
		return appErrors.NewInternalError("failed to get session", err)
	}

	// Toggle public
	session.Public = !session.Public

	// Update session
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return appErrors.NewInternalError("failed to toggle public", err)
	}

	return nil
}

// UpdateSessionMetadata updates session metadata
func (s *SessionService) UpdateSessionMetadata(ctx context.Context, id ulid.ULID, metadata map[string]string) error {
	// Get session
	session, err := s.sessionRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return appErrors.NewNotFoundError(fmt.Sprintf("session %s", id.String()))
		}
		return appErrors.NewInternalError("failed to get session", err)
	}

	// Update metadata
	if session.Metadata == nil {
		session.Metadata = make(map[string]string)
	}
	for key, value := range metadata {
		session.Metadata[key] = value
	}

	// Update session
	if err := s.sessionRepo.Update(ctx, session); err != nil {
		return appErrors.NewInternalError("failed to update session metadata", err)
	}

	return nil
}

// CountSessions returns the count of sessions matching the filter
func (s *SessionService) CountSessions(ctx context.Context, filter *observability.SessionFilter) (int64, error) {
	count, err := s.sessionRepo.Count(ctx, filter)
	if err != nil {
		return 0, appErrors.NewInternalError("failed to count sessions", err)
	}

	return count, nil
}
