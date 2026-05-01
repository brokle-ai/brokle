// Package user implements user profiles, password lifecycle, and account
// administration. The two services in this package — UserService and
// ProfileService — split responsibility along the canonical lines: User
// owns auth-adjacent identity data (password, email-verification status,
// default org, last-login), Profile owns the public-facing profile
// (bio, social links, completeness).
package user

import (
	"context"
	"time"

	"github.com/google/uuid"

	userDomain "brokle/internal/core/domain/user"
	appErrors "brokle/pkg/errors"
)

// ProfileService manages the user profile (bio, social links,
// completeness scoring). Notification, theme, and privacy preferences
// are NOT in scope — those routes do not exist in the dashboard plane
// today (CLAUDE.md scaffolded-but-unreachable rule).
type ProfileService struct {
	userRepo userDomain.Repository
}

// NewProfileService creates a new profile service instance.
func NewProfileService(userRepo userDomain.Repository) *ProfileService {
	return &ProfileService{userRepo: userRepo}
}

// GetProfile retrieves a user's editable profile record.
func (s *ProfileService) GetProfile(ctx context.Context, userID uuid.UUID) (*userDomain.UserProfile, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, appErrors.NotFound("profile")
	}
	return profile, nil
}

// UpdateProfile applies a partial update to the user's profile.
// Pointer fields on the request mean "only update if provided".
func (s *ProfileService) UpdateProfile(ctx context.Context, userID uuid.UUID, req *userDomain.UpdateUserProfileRequest) (*userDomain.UserProfile, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, appErrors.NotFound("profile")
	}

	if req.Bio != nil {
		profile.Bio = req.Bio
	}
	if req.Location != nil {
		profile.Location = req.Location
	}
	if req.Website != nil {
		profile.Website = req.Website
	}
	if req.TwitterURL != nil {
		profile.TwitterURL = req.TwitterURL
	}
	if req.LinkedInURL != nil {
		profile.LinkedInURL = req.LinkedInURL
	}
	if req.GitHubURL != nil {
		profile.GitHubURL = req.GitHubURL
	}
	if req.Timezone != nil {
		profile.Timezone = *req.Timezone
	}
	if req.Language != nil {
		profile.Language = *req.Language
	}
	if req.Theme != nil {
		profile.Theme = *req.Theme
	}
	profile.UpdatedAt = time.Now()

	if err := s.userRepo.UpdateProfile(ctx, profile); err != nil {
		return nil, appErrors.Internal("failed to update profile", err)
	}
	return profile, nil
}

// GetProfileCompleteness reports the user's profile-completion score
// and the per-section breakdown surfaced by the dashboard onboarding
// checklist. Score is a percentage of completed/total fields across
// the basic and extended sections.
func (s *ProfileService) GetProfileCompleteness(ctx context.Context, userID uuid.UUID) (*userDomain.ProfileCompleteness, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, appErrors.NotFound("user")
	}

	completeness := &userDomain.ProfileCompleteness{
		CompletedFields: []string{},
		MissingFields:   []string{},
		Recommendations: []string{},
		Sections:        make(map[string]int),
	}

	profile, _ := s.userRepo.GetProfile(ctx, userID)

	// Basic-info section.
	const totalBasicFields = 4
	basicFields := 0
	if user.FirstName != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "first_name")
		basicFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "first_name")
	}
	if user.LastName != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "last_name")
		basicFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "last_name")
	}
	if user.Email != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "email")
		basicFields++
	}
	if profile != nil && profile.AvatarURL != nil && *profile.AvatarURL != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "avatar")
		basicFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "avatar")
		completeness.Recommendations = append(completeness.Recommendations, "Upload a profile photo")
	}
	completeness.Sections["basic"] = (basicFields * 100) / totalBasicFields

	// Extended-info section.
	const totalExtendedFields = 3
	extendedFields := 0
	if profile != nil && profile.Bio != nil && *profile.Bio != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "bio")
		extendedFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "bio")
		completeness.Recommendations = append(completeness.Recommendations, "Add a bio to tell others about yourself")
	}
	if profile != nil && profile.Location != nil && *profile.Location != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "location")
		extendedFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "location")
	}
	if profile != nil && profile.Website != nil && *profile.Website != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "website")
		extendedFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "website")
	}
	completeness.Sections["extended"] = (extendedFields * 100) / totalExtendedFields

	// Overall percentage across both sections.
	totalFields := len(completeness.CompletedFields)
	maxFields := totalBasicFields + totalExtendedFields
	completeness.OverallScore = (totalFields * 100) / maxFields

	return completeness, nil
}
