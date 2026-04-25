package shared

// MessageResponse is the flat {"message": "..."} wire shape used by
// handler operations that acknowledge an action without returning an
// entity — logout, password reset, session revocation, default-org
// set, etc. Lives in handlers/shared because auth, user, and future
// domains all need the same DTO and a single canonical type avoids
// per-package duplication.
type MessageResponse struct {
	Message string `json:"message"`
}
