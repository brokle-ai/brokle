package user

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// profileService implements the user.ProfileService interface
type profileService struct {
	userRepo  user.Repository
	auditRepo auth.AuditLogRepository
}

// NewProfileService creates a new profile service instance
func NewProfileService(
	userRepo user.Repository,
	auditRepo auth.AuditLogRepository,
) user.ProfileService {
	return &profileService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// GetProfile retrieves user profile
func (s *profileService) GetProfile(ctx context.Context, userID ulid.ULID) (*user.UserProfile, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	return profile, nil
}

// UpdateProfile updates user profile information
func (s *profileService) UpdateProfile(ctx context.Context, userID ulid.ULID, req *user.UpdateProfileRequest) (*user.UserProfile, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// Update profile fields if provided
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

	err = s.userRepo.UpdateProfile(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.profile_updated", "profile", userID.String(), "", "", ""))

	return profile, nil
}

// UploadAvatar uploads and sets user avatar
func (s *profileService) UploadAvatar(ctx context.Context, userID ulid.ULID, imageData []byte, contentType string) (*user.UserProfile, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// TODO: Implement actual image upload to storage service
	// For now, just simulate with a placeholder URL
	avatarURL := fmt.Sprintf("https://api.example.com/avatars/%s.jpg", userID.String())
	
	profile.AvatarURL = &avatarURL
	profile.UpdatedAt = time.Now()

	err = s.userRepo.UpdateProfile(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to update avatar: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.avatar_uploaded", "profile", userID.String(), 
		fmt.Sprintf(`{"content_type": "%s"}`, contentType), "", ""))

	return s.GetProfile(ctx, userID)
}

// RemoveAvatar removes user avatar
func (s *profileService) RemoveAvatar(ctx context.Context, userID ulid.ULID) error {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	// TODO: Delete avatar from storage service
	profile.AvatarURL = nil
	profile.UpdatedAt = time.Now()

	err = s.userRepo.UpdateProfile(ctx, profile)
	if err != nil {
		return fmt.Errorf("failed to remove avatar: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.avatar_removed", "profile", userID.String(), "", "", ""))

	return nil
}

// UpdateProfileVisibility updates profile visibility settings
func (s *profileService) UpdateProfileVisibility(ctx context.Context, userID ulid.ULID, visibility user.ProfileVisibility) error {
	// TODO: Add ProfileVisibility field to User model
	// For now, just create audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.profile_visibility_updated", "profile", userID.String(),
		fmt.Sprintf(`{"visibility": "%s"}`, string(visibility)), "", ""))

	return nil
}

// GetPublicProfile retrieves public view of user profile
func (s *profileService) GetPublicProfile(ctx context.Context, userID ulid.ULID) (*user.PublicProfile, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get profile for additional info
	profile, _ := s.userRepo.GetProfile(ctx, userID)

	publicProfile := &user.PublicProfile{
		UserID:    existingUser.ID,
		FirstName: existingUser.FirstName,
		LastName:  existingUser.LastName,
		AvatarURL: nil, // Will be set from profile if available
		Bio:       nil, // Will be set from profile if available
		Location:  nil, // Will be set from profile if available
	}

	// Set profile fields if available
	if profile != nil {
		publicProfile.AvatarURL = profile.AvatarURL
		publicProfile.Bio = profile.Bio
		publicProfile.Location = profile.Location
	}

	return publicProfile, nil
}

// GetProfileCompleteness calculates profile completion status
func (s *profileService) GetProfileCompleteness(ctx context.Context, userID ulid.ULID) (*user.ProfileCompleteness, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	completeness := &user.ProfileCompleteness{
		CompletedFields: []string{},
		MissingFields:   []string{},
		Recommendations: []string{},
		Sections:        make(map[string]int),
	}

	// Get profile data for avatar check
	profile, _ := s.userRepo.GetProfile(ctx, userID)

	// Check basic info
	basicFields := 0
	totalBasicFields := 4
	if existingUser.FirstName != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "first_name")
		basicFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "first_name")
	}
	if existingUser.LastName != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "last_name")
		basicFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "last_name")
	}
	if existingUser.Email != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "email")
		basicFields++
	}
	// Check avatar from profile
	if profile != nil && profile.AvatarURL != nil && *profile.AvatarURL != "" {
		completeness.CompletedFields = append(completeness.CompletedFields, "avatar")
		basicFields++
	} else {
		completeness.MissingFields = append(completeness.MissingFields, "avatar")
		completeness.Recommendations = append(completeness.Recommendations, "Upload a profile photo")
	}
	completeness.Sections["basic"] = (basicFields * 100) / totalBasicFields

	// Check extended info from profile (already fetched above)
	extendedFields := 0
	totalExtendedFields := 3
	
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

	// Calculate overall score
	totalFields := len(completeness.CompletedFields)
	maxFields := totalBasicFields + totalExtendedFields
	completeness.OverallScore = (totalFields * 100) / maxFields

	return completeness, nil
}

// ValidateProfile validates profile data
func (s *profileService) ValidateProfile(ctx context.Context, userID ulid.ULID) (*user.ProfileValidation, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	validation := &user.ProfileValidation{
		IsValid: true,
		Errors:  []user.ProfileValidationError{},
	}

	// Validate email format
	if existingUser.Email == "" {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, user.ProfileValidationError{
			Field:   "email",
			Message: "Email is required",
		})
	}

	// Validate name fields
	if existingUser.FirstName == "" {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, user.ProfileValidationError{
			Field:   "first_name",
			Message: "First name is required",
		})
	}

	if existingUser.LastName == "" {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, user.ProfileValidationError{
			Field:   "last_name",
			Message: "Last name is required",
		})
	}

	// Validate bio length from profile
	profile, _ := s.userRepo.GetProfile(ctx, userID)
	if profile != nil && profile.Bio != nil && len(*profile.Bio) > 500 {
		validation.IsValid = false
		validation.Errors = append(validation.Errors, user.ProfileValidationError{
			Field:   "bio",
			Message: "Bio must be less than 500 characters",
		})
	}

	return validation, nil
}

// GetNotificationPreferences retrieves user notification preferences from profile
func (s *profileService) GetNotificationPreferences(ctx context.Context, userID ulid.ULID) (*user.NotificationPreferences, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	return &user.NotificationPreferences{
		EmailNotifications:      profile.EmailNotifications,
		PushNotifications:       profile.PushNotifications,
		SMSNotifications:        false, // Not stored in profile
		MarketingEmails:         profile.MarketingEmails,
		SecurityAlerts:          profile.SecurityAlerts,
		ProductUpdates:          false, // Not stored in profile
		WeeklyDigest:            profile.WeeklyReports,
		InvitationNotifications: true, // Default value
	}, nil
}

// UpdateNotificationPreferences updates user notification preferences in profile
func (s *profileService) UpdateNotificationPreferences(ctx context.Context, userID ulid.ULID, req *user.UpdateNotificationPreferencesRequest) (*user.NotificationPreferences, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// Update notification fields if provided
	if req.EmailNotifications != nil {
		profile.EmailNotifications = *req.EmailNotifications
	}
	if req.PushNotifications != nil {
		profile.PushNotifications = *req.PushNotifications
	}
	if req.MarketingEmails != nil {
		profile.MarketingEmails = *req.MarketingEmails
	}
	if req.SecurityAlerts != nil {
		profile.SecurityAlerts = *req.SecurityAlerts
	}
	if req.WeeklyDigest != nil {
		profile.WeeklyReports = *req.WeeklyDigest
	}

	err = s.userRepo.UpdateProfile(ctx, profile)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.notification_preferences_updated", "profile", userID.String(), "", "", ""))

	return s.GetNotificationPreferences(ctx, userID)
}

// GetThemePreferences retrieves user theme preferences from profile and user
func (s *profileService) GetThemePreferences(ctx context.Context, userID ulid.ULID) (*user.ThemePreferences, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	return &user.ThemePreferences{
		Theme:          profile.Theme,
		PrimaryColor:   "#007bff",             // Default blue
		Language:       existingUser.Language,
		TimeFormat:     "12h",                 // Default
		DateFormat:     "MM/dd/yyyy",          // Default
		Timezone:       existingUser.Timezone,
		CompactMode:    false,                 // Default
		ShowAnimations: true,                  // Default
		HighContrast:   false,                 // Default
	}, nil
}

// UpdateThemePreferences updates user theme preferences in both user and profile
func (s *profileService) UpdateThemePreferences(ctx context.Context, userID ulid.ULID, req *user.UpdateThemePreferencesRequest) (*user.ThemePreferences, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	// Update user fields if provided
	updated := false
	if req.Language != nil {
		existingUser.Language = *req.Language
		updated = true
	}
	if req.Timezone != nil {
		existingUser.Timezone = *req.Timezone
		updated = true
	}

	if updated {
		existingUser.UpdatedAt = time.Now()
		err = s.userRepo.Update(ctx, existingUser)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}
	}

	// Update profile theme if provided
	if req.Theme != nil {
		profile.Theme = *req.Theme
		profile.UpdatedAt = time.Now()
		err = s.userRepo.UpdateProfile(ctx, profile)
		if err != nil {
			return nil, fmt.Errorf("failed to update profile: %w", err)
		}
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.theme_updated", "profile", userID.String(), "", "", ""))

	return s.GetThemePreferences(ctx, userID)
}

// GetPrivacyPreferences retrieves user privacy preferences (stub implementation)
func (s *profileService) GetPrivacyPreferences(ctx context.Context, userID ulid.ULID) (*user.PrivacyPreferences, error) {
	// Return default privacy preferences since they're not in the current model
	return &user.PrivacyPreferences{
		ProfileVisibility:      user.ProfileVisibilityPublic, // Default
		ShowEmail:              false,                         // Default private
		ShowLastSeen:           true,                          // Default
		AllowDirectMessages:    true,                          // Default
		DataProcessingConsent:  true,                          // Required
		AnalyticsConsent:       true,                          // Default
		ThirdPartyIntegrations: false,                         // Default private
	}, nil
}

// UpdatePrivacyPreferences updates user privacy preferences (stub implementation)
func (s *profileService) UpdatePrivacyPreferences(ctx context.Context, userID ulid.ULID, req *user.UpdatePrivacyPreferencesRequest) (*user.PrivacyPreferences, error) {
	// For now, just create audit log since privacy preferences aren't fully implemented
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.privacy_updated", "profile", userID.String(), "", "", ""))

	// Return current preferences (would be updated if fully implemented)
	return s.GetPrivacyPreferences(ctx, userID)
}