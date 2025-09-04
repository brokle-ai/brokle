package services

import (
	"context"
	"fmt"
	"time"

	"brokle/pkg/ulid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"brokle/internal/config"
	"brokle/internal/core/domain/user"
	"brokle/internal/infrastructure/repository/redis"
)

// UserService implements user.Service interface
type UserService struct {
	config     *config.Config
	logger     *logrus.Logger
	repository user.Repository
	cache      *redis.CacheRepository
}

// NewUserService creates a new user service
func NewUserService(
	config *config.Config,
	logger *logrus.Logger,
	repository user.Repository,
	cache *redis.CacheRepository,
) user.Service {
	return &UserService{
		config:     config,
		logger:     logger,
		repository: repository,
		cache:      cache,
	}
}

// Create creates a new user
func (s *UserService) Create(ctx context.Context, u *user.User) error {
	s.logger.WithFields(logrus.Fields{
		"email":      u.Email,
		"first_name": u.FirstName,
		"last_name":  u.LastName,
	}).Info("Creating new user")

	// Validate user data
	if err := s.validateUser(u); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.repository.GetByEmail(ctx, u.Email)
	if err == nil && existingUser != nil {
		return user.ErrUserAlreadyExists
	}

	// Set timestamps
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now

	// Set user as active and verified by default
	u.IsActive = true
	u.IsEmailVerified = false // Will be verified via email

	// Create user in repository
	if err := s.repository.Create(ctx, u); err != nil {
		s.logger.WithError(err).Error("Failed to create user in repository")
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Cache user data
	if err := s.cacheUser(ctx, u); err != nil {
		s.logger.WithError(err).Warn("Failed to cache user data")
		// Don't fail the operation if caching fails
	}

	s.logger.WithField("user_id", u.ID).Info("User created successfully")
	return nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id string) (*user.User, error) {
	// Try cache first
	var cachedUser user.User
	if err := s.cache.Get(ctx, s.userCacheKey(id), &cachedUser); err == nil {
		s.logger.WithField("user_id", id).Debug("User found in cache")
		return &cachedUser, nil
	}

	// Convert string to ULID
	userID, err := ulid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	// Get from repository
	u, err := s.repository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache for next time
	if err := s.cacheUser(ctx, u); err != nil {
		s.logger.WithError(err).Warn("Failed to cache user data")
	}

	return u, nil
}

// GetByEmail retrieves a user by email
func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.repository.GetByEmail(ctx, email)
}

// Update updates a user
func (s *UserService) Update(ctx context.Context, u *user.User) error {
	s.logger.WithField("user_id", u.ID).Info("Updating user")

	// Validate user data
	if err := s.validateUser(u); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	// Set updated timestamp
	u.UpdatedAt = time.Now()

	// Update in repository
	if err := s.repository.Update(ctx, u); err != nil {
		s.logger.WithError(err).Error("Failed to update user in repository")
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Update cache
	if err := s.cacheUser(ctx, u); err != nil {
		s.logger.WithError(err).Warn("Failed to update user cache")
	}

	s.logger.WithField("user_id", u.ID).Info("User updated successfully")
	return nil
}

// Delete deletes a user by ID
func (s *UserService) Delete(ctx context.Context, id string) error {
	s.logger.WithField("user_id", id).Info("Deleting user")

	// Convert string to ULID
	userID, err := ulid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid user ID format: %w", err)
	}

	// Delete from repository
	if err := s.repository.Delete(ctx, userID); err != nil {
		s.logger.WithError(err).Error("Failed to delete user from repository")
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Remove from cache
	if err := s.cache.Delete(ctx, s.userCacheKey(id)); err != nil {
		s.logger.WithError(err).Warn("Failed to remove user from cache")
	}

	s.logger.WithField("user_id", id).Info("User deleted successfully")
	return nil
}

// ListUsers retrieves users with pagination and filtering
func (s *UserService) ListUsers(ctx context.Context, filters *user.ListFilters) ([]*user.User, int, error) {
	// Get users from repository
	users, totalCount, err := s.repository.List(ctx, filters)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	return users, totalCount, nil
}

// Authenticate authenticates a user with email and password
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*user.User, error) {
	s.logger.WithField("email", email).Info("Authenticating user")

	// Get user by email
	u, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		if err == user.ErrUserNotFound {
			// Don't reveal that user doesn't exist
			return nil, user.ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Check if user is active
	if !u.IsActive {
		s.logger.WithField("user_id", u.ID).Warn("Attempted login for inactive user")
		return nil, user.ErrUserInactive
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		s.logger.WithField("user_id", u.ID).Warn("Invalid password attempt")
		return nil, user.ErrInvalidCredentials
	}

	// Update last login time
	u.LastLoginAt = &time.Time{}
	*u.LastLoginAt = time.Now()
	if err := s.repository.Update(ctx, u); err != nil {
		s.logger.WithError(err).Warn("Failed to update last login time")
		// Don't fail authentication if this fails
	}

	s.logger.WithField("user_id", u.ID).Info("User authenticated successfully")
	return u, nil
}



// ChangePassword changes user password
func (s *UserService) ChangePassword(ctx context.Context, userID ulid.ULID, currentPassword, newPassword string) error {
	s.logger.WithField("user_id", userID).Info("Changing user password")

	// Get user
	u, err := s.repository.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(currentPassword)); err != nil {
		return user.ErrInvalidCredentials
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	u.Password = string(hashedPassword)
	u.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, u); err != nil {
		s.logger.WithError(err).Error("Failed to update user password")
		return fmt.Errorf("failed to update password: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("User password changed successfully")
	return nil
}

// ResetPassword resets user password (admin function)
func (s *UserService) ResetPassword(ctx context.Context, token, newPassword string) error {
	s.logger.WithField("token", token).Info("Resetting user password")

	// TODO: Implement proper token validation logic
	// For now, assume token is a user ID string (should be replaced with proper token validation)
	userID, err := ulid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid token format: %w", err)
	}

	// Validate new password
	if err := s.validatePassword(newPassword); err != nil {
		return err
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Get user and update password
	u, err := s.repository.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	u.Password = string(hashedPassword)
	u.UpdatedAt = time.Now()

	if err := s.repository.Update(ctx, u); err != nil {
		s.logger.WithError(err).Error("Failed to reset user password")
		return fmt.Errorf("failed to reset password: %w", err)
	}

	s.logger.WithField("user_id", userID).Info("User password reset successfully")
	return nil
}

// Validation helper methods

func (s *UserService) validateUser(u *user.User) error {
	if u.Email == "" {
		return user.ErrInvalidEmail
	}

	if u.FirstName == "" {
		return user.ErrInvalidName
	}

	if u.LastName == "" {
		return user.ErrInvalidName
	}

	// Add more validation as needed
	return nil
}

func (s *UserService) validateProfile(profile *user.UserProfile) error {
	// Add profile validation logic
	return nil
}

func (s *UserService) validatePreferences(preferences *user.UserPreferences) error {
	// Validate theme
	if preferences.Theme != "" {
		validThemes := []string{"light", "dark", "system"}
		valid := false
		for _, theme := range validThemes {
			if preferences.Theme == theme {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid theme: %s", preferences.Theme)
		}
	}

	return nil
}

func (s *UserService) validatePassword(password string) error {
	if len(password) < 8 {
		return user.ErrWeakPassword
	}

	// Add more password strength validation as needed
	return nil
}

// Cache helper methods

func (s *UserService) cacheUser(ctx context.Context, u *user.User) error {
	key := s.userCacheKey(u.ID)
	expiration := 30 * time.Minute
	return s.cache.Set(ctx, key, u, expiration)
}

func (s *UserService) userCacheKey(userID interface{}) string {
	return fmt.Sprintf("user:%s", userID)
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, req *user.CreateUserRequest) (*user.User, error) {
	s.logger.WithField("email", req.Email).Info("Registering new user")

	// Create user from request
	u := user.NewUser(req.Email, req.FirstName, req.LastName)

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	u.Password = string(hashedPassword)

	// Create user
	if err := s.Create(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID ulid.ULID) (*user.User, error) {
	return s.GetByID(ctx, userID.String())
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email string) (*user.User, error) {
	return s.GetByEmail(ctx, email)
}

// GetUserByEmailWithPassword retrieves a user by email with password hash
func (s *UserService) GetUserByEmailWithPassword(ctx context.Context, email string) (*user.User, error) {
	return s.repository.GetByEmailWithPassword(ctx, email)
}

// UpdateUser updates user information
func (s *UserService) UpdateUser(ctx context.Context, userID ulid.ULID, req *user.UpdateUserRequest) (*user.User, error) {
	// Get existing user
	u, err := s.GetByID(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.FirstName != nil {
		u.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		u.LastName = *req.LastName
	}
	if req.AvatarURL != nil {
		u.AvatarURL = *req.AvatarURL
	}
	if req.Phone != nil {
		u.Phone = *req.Phone
	}
	if req.Timezone != nil {
		u.Timezone = *req.Timezone
	}
	if req.Language != nil {
		u.Language = *req.Language
	}

	// Update user
	if err := s.Update(ctx, u); err != nil {
		return nil, err
	}

	return u, nil
}

// DeactivateUser deactivates a user account
func (s *UserService) DeactivateUser(ctx context.Context, userID ulid.ULID) error {
	u, err := s.GetByID(ctx, userID.String())
	if err != nil {
		return err
	}

	u.IsActive = false
	return s.Update(ctx, u)
}

// ReactivateUser reactivates a user account
func (s *UserService) ReactivateUser(ctx context.Context, userID ulid.ULID) error {
	u, err := s.GetByID(ctx, userID.String())
	if err != nil {
		return err
	}

	u.IsActive = true
	return s.Update(ctx, u)
}

// DeleteUser deletes a user account
func (s *UserService) DeleteUser(ctx context.Context, userID ulid.ULID) error {
	return s.Delete(ctx, userID.String())
}

// GetProfile retrieves user profile
func (s *UserService) GetProfile(ctx context.Context, userID ulid.ULID) (*user.UserProfile, error) {
	return s.repository.GetProfile(ctx, userID)
}

// UpdateProfile updates user profile
func (s *UserService) UpdateProfile(ctx context.Context, userID ulid.ULID, req *user.UpdateProfileRequest) (*user.UserProfile, error) {
	// Get existing profile or create new one
	profile, err := s.repository.GetProfile(ctx, userID)
	if err != nil {
		if err == user.ErrUserNotFound {
			// Create new profile
			profile = user.NewUserProfile(userID)
		} else {
			return nil, err
		}
	}

	// Update fields if provided
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

	// Update or create profile
	if err := s.repository.UpdateProfile(ctx, profile); err != nil {
		return nil, err
	}

	return profile, nil
}

// GetPreferences retrieves user preferences
func (s *UserService) GetPreferences(ctx context.Context, userID ulid.ULID) (*user.UserPreferences, error) {
	return s.repository.GetPreferences(ctx, userID)
}

// UpdatePreferences updates user preferences
func (s *UserService) UpdatePreferences(ctx context.Context, userID ulid.ULID, req *user.UpdatePreferencesRequest) (*user.UserPreferences, error) {
	// Get existing preferences or create new ones
	preferences, err := s.repository.GetPreferences(ctx, userID)
	if err != nil {
		if err == user.ErrUserNotFound {
			// Create new preferences
			preferences = user.NewUserPreferences(userID)
		} else {
			return nil, err
		}
	}

	// Update fields if provided
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

	// Update preferences
	if err := s.repository.UpdatePreferences(ctx, preferences); err != nil {
		return nil, err
	}

	return preferences, nil
}

// SearchUsers searches users by query
func (s *UserService) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*user.User, int, error) {
	return s.repository.Search(ctx, query, limit, offset)
}

// VerifyEmail verifies user email
func (s *UserService) VerifyEmail(ctx context.Context, userID ulid.ULID, token string) error {
	// TODO: Implement token validation
	return s.repository.VerifyEmail(ctx, userID, token)
}

// MarkEmailAsVerified marks user email as verified
func (s *UserService) MarkEmailAsVerified(ctx context.Context, userID ulid.ULID) error {
	return s.repository.MarkEmailAsVerified(ctx, userID)
}

// SendVerificationEmail sends verification email
func (s *UserService) SendVerificationEmail(ctx context.Context, userID ulid.ULID) error {
	// TODO: Implement email sending
	s.logger.WithField("user_id", userID).Info("Sending verification email")
	return nil
}

// RequestPasswordReset requests password reset
func (s *UserService) RequestPasswordReset(ctx context.Context, email string) error {
	// TODO: Implement password reset token generation and email sending
	s.logger.WithField("email", email).Info("Password reset requested")
	return nil
}

// UpdateLastLogin updates user's last login timestamp
func (s *UserService) UpdateLastLogin(ctx context.Context, userID ulid.ULID) error {
	return s.repository.UpdateLastLogin(ctx, userID)
}

// GetUserActivity gets user activity metrics
func (s *UserService) GetUserActivity(ctx context.Context, userID ulid.ULID) (*user.UserActivity, error) {
	// TODO: Implement activity tracking
	return &user.UserActivity{
		UserID: userID,
	}, nil
}

// GetUsersByIDs gets multiple users by their IDs
func (s *UserService) GetUsersByIDs(ctx context.Context, userIDs []ulid.ULID) ([]*user.User, error) {
	return s.repository.GetByIDs(ctx, userIDs)
}

// GetPublicUsers gets public user information by IDs
func (s *UserService) GetPublicUsers(ctx context.Context, userIDs []ulid.ULID) ([]*user.PublicUser, error) {
	users, err := s.repository.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	publicUsers := make([]*user.PublicUser, len(users))
	for i, u := range users {
		publicUsers[i] = u.ToPublic()
	}

	return publicUsers, nil
}

// SetDefaultOrganization sets user's default organization
func (s *UserService) SetDefaultOrganization(ctx context.Context, userID, orgID ulid.ULID) error {
	return s.repository.SetDefaultOrganization(ctx, userID, orgID)
}

// GetDefaultOrganization gets user's default organization
func (s *UserService) GetDefaultOrganization(ctx context.Context, userID ulid.ULID) (*ulid.ULID, error) {
	return s.repository.GetDefaultOrganization(ctx, userID)
}

// CompleteOnboarding marks user onboarding as completed
func (s *UserService) CompleteOnboarding(ctx context.Context, userID ulid.ULID) error {
	return s.repository.CompleteOnboarding(ctx, userID)
}

// IsOnboardingCompleted checks if user onboarding is completed
func (s *UserService) IsOnboardingCompleted(ctx context.Context, userID ulid.ULID) (bool, error) {
	return s.repository.IsOnboardingCompleted(ctx, userID)
}

// GetUserStats gets aggregate user statistics
func (s *UserService) GetUserStats(ctx context.Context) (*user.UserStats, error) {
	return s.repository.GetUserStats(ctx)
}