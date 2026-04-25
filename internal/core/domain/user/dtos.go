package user

// User-domain DTOs surfaced to handlers + analytics views.
//
// Profile-related types live here because they cross the
// service/handler boundary (GetProfileCompleteness response,
// GetUserStats response). Anything declared but unused has been
// removed during the 2026-04-25 standardization sweep.

// ProfileCompleteness is the per-section breakdown surfaced by the
// dashboard onboarding checklist. Sections are scored 0–100 each;
// OverallScore is the aggregate completion percentage.
type ProfileCompleteness struct {
	Sections        map[string]int `json:"sections"`
	CompletedFields []string       `json:"completed_fields"`
	MissingFields   []string       `json:"missing_fields"`
	Recommendations []string       `json:"recommendations"`
	OverallScore    int            `json:"overall_score"`
}

// UserStats is the aggregate user-counter snapshot the dashboard's
// platform overview consumes. Computed on read by the user repository
// against the users table; numbers are point-in-time and not cached.
type UserStats struct {
	TotalUsers        int64 `json:"total_users"`
	ActiveUsers       int64 `json:"active_users"`
	VerifiedUsers     int64 `json:"verified_users"`
	NewUsersToday     int64 `json:"new_users_today"`
	NewUsersThisWeek  int64 `json:"new_users_this_week"`
	NewUsersThisMonth int64 `json:"new_users_this_month"`
}
