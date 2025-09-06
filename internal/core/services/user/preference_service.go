package user

import (
	"context"
	"fmt"
	"time"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// preferenceService implements the user.PreferenceService interface
type preferenceService struct {
	userRepo  user.Repository
	auditRepo auth.AuditLogRepository
}

// NewPreferenceService creates a new preference service instance
func NewPreferenceService(
	userRepo user.Repository,
	auditRepo auth.AuditLogRepository,
) user.PreferenceService {
	return &preferenceService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

// GetPreferences retrieves user preferences
func (s *preferenceService) GetPreferences(ctx context.Context, userID ulid.ULID) (*user.UserPreferences, error) {
	preferences, err := s.userRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("preferences not found: %w", err)
	}

	return preferences, nil
}

// UpdatePreferences updates user preferences
func (s *preferenceService) UpdatePreferences(ctx context.Context, userID ulid.ULID, req *user.UpdatePreferencesRequest) (*user.UserPreferences, error) {
	preferences, err := s.userRepo.GetPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("preferences not found: %w", err)
	}

	// Update preference fields if provided
	if req.EmailNotifications != nil {
		preferences.EmailNotifications = *req.EmailNotifications
	}
	if req.PushNotifications != nil {
		preferences.PushNotifications = *req.PushNotifications
	}
	if req.MarketingEmails != nil {
		preferences.MarketingEmails = *req.MarketingEmails
	}
	if req.WeeklyReports != nil {
		preferences.WeeklyReports = *req.WeeklyReports
	}
	if req.MonthlyReports != nil {
		preferences.MonthlyReports = *req.MonthlyReports
	}
	if req.SecurityAlerts != nil {
		preferences.SecurityAlerts = *req.SecurityAlerts
	}
	if req.BillingAlerts != nil {
		preferences.BillingAlerts = *req.BillingAlerts
	}
	if req.UsageThresholdPercent != nil {
		preferences.UsageThresholdPercent = *req.UsageThresholdPercent
	}

	preferences.UpdatedAt = time.Now()

	err = s.userRepo.UpdatePreferences(ctx, preferences)
	if err != nil {
		return nil, fmt.Errorf("failed to update preferences: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.preferences_updated", "preferences", userID.String(), "", "", ""))

	return preferences, nil
}

// ResetPreferences resets user preferences to defaults
func (s *preferenceService) ResetPreferences(ctx context.Context, userID ulid.ULID) (*user.UserPreferences, error) {
	// Create new default preferences
	defaultPreferences := user.NewUserPreferences(userID)

	err := s.userRepo.UpdatePreferences(ctx, defaultPreferences)
	if err != nil {
		return nil, fmt.Errorf("failed to reset preferences: %w", err)
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.preferences_reset", "preferences", userID.String(), "", "", ""))

	return defaultPreferences, nil
}

// GetNotificationPreferences retrieves user notification preferences
func (s *preferenceService) GetNotificationPreferences(ctx context.Context, userID ulid.ULID) (*user.NotificationPreferences, error) {
	preferences, err := s.GetPreferences(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &user.NotificationPreferences{
		EmailNotifications:      preferences.EmailNotifications,
		PushNotifications:       preferences.PushNotifications,
		SMSNotifications:        false, // Not in current model
		MarketingEmails:         preferences.MarketingEmails,
		SecurityAlerts:          preferences.SecurityAlerts,
		ProductUpdates:          false, // Not in current model
		WeeklyDigest:            preferences.WeeklyReports,
		InvitationNotifications: true, // Default value
	}, nil
}

// UpdateNotificationPreferences updates user notification preferences
func (s *preferenceService) UpdateNotificationPreferences(ctx context.Context, userID ulid.ULID, req *user.UpdateNotificationPreferencesRequest) (*user.NotificationPreferences, error) {
	// Convert notification preferences request to general preferences request
	prefReq := &user.UpdatePreferencesRequest{
		EmailNotifications: req.EmailNotifications,
		PushNotifications:  req.PushNotifications,
		MarketingEmails:    req.MarketingEmails,
		SecurityAlerts:     req.SecurityAlerts,
		WeeklyReports:      req.WeeklyDigest,
	}

	_, err := s.UpdatePreferences(ctx, userID, prefReq)
	if err != nil {
		return nil, err
	}

	return s.GetNotificationPreferences(ctx, userID)
}

// GetThemePreferences retrieves user theme preferences
func (s *preferenceService) GetThemePreferences(ctx context.Context, userID ulid.ULID) (*user.ThemePreferences, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Get profile for theme info
	profile, _ := s.userRepo.GetProfile(ctx, userID)

	themePrefs := &user.ThemePreferences{
		Theme:          existingUser.Language, // Use language as theme for now
		PrimaryColor:   "#007bff",             // Default blue
		Language:       existingUser.Language,
		TimeFormat:     "12h",       // Default
		DateFormat:     "MM/dd/yyyy", // Default
		Timezone:       existingUser.Timezone,
		CompactMode:    false, // Default
		ShowAnimations: true,  // Default
		HighContrast:   false, // Default
	}

	// Override with profile theme if available
	if profile != nil {
		themePrefs.Theme = profile.Theme
	}

	return themePrefs, nil
}

// UpdateThemePreferences updates user theme preferences
func (s *preferenceService) UpdateThemePreferences(ctx context.Context, userID ulid.ULID, req *user.UpdateThemePreferencesRequest) (*user.ThemePreferences, error) {
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
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
		profile, err := s.userRepo.GetProfile(ctx, userID)
		if err == nil {
			profile.Theme = *req.Theme
			profile.UpdatedAt = time.Now()
			s.userRepo.UpdateProfile(ctx, profile)
		}
	}

	// Audit log
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.theme_updated", "preferences", userID.String(), "", "", ""))

	return s.GetThemePreferences(ctx, userID)
}

// GetPrivacyPreferences retrieves user privacy preferences
func (s *preferenceService) GetPrivacyPreferences(ctx context.Context, userID ulid.ULID) (*user.PrivacyPreferences, error) {
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

// UpdatePrivacyPreferences updates user privacy preferences
func (s *preferenceService) UpdatePrivacyPreferences(ctx context.Context, userID ulid.ULID, req *user.UpdatePrivacyPreferencesRequest) (*user.PrivacyPreferences, error) {
	// For now, just create audit log since privacy preferences aren't fully implemented
	s.auditRepo.Create(ctx, auth.NewAuditLog(&userID, nil, "user.privacy_updated", "preferences", userID.String(), "", "", ""))

	// Return current preferences (would be updated if fully implemented)
	return s.GetPrivacyPreferences(ctx, userID)
}