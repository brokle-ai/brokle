package shared

// MessageResponse is the flat {"message": "..."} wire shape used by
// handler operations that acknowledge an action without returning an
// entity — logout, password reset, session revocation, default-org
// set, etc. Lives in handlers/shared because auth + user + future
// domains all need the same DTO; Huma's schema registry is a flat
// namespace and a per-package `messageResponse` duplicate panics at
// boot (see lint rule in scripts/lint-conventions.sh and the
// TestServer_SchemaNamesUnique regression guard).
type MessageResponse struct {
	Message string `json:"message" example:"Operation completed successfully" doc:"Human-readable result of the action"`
}
