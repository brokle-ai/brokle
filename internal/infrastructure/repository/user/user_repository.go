package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	userDomain "brokle/internal/core/domain/user"
	"brokle/pkg/ulid"
)

// userRepository implements the userDomain.Repository interface using GORM
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository instance
func NewUserRepository(db *gorm.DB) userDomain.Repository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, u *userDomain.User) error {
	return r.db.WithContext(ctx).Create(u).Error
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id ulid.ULID) (*userDomain.User, error) {
	var u userDomain.User
	err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get user by ID %s: %w", id, userDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database query failed for user ID %s: %w", id, err)
	}
	return &u, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	var u userDomain.User
	err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get user by email %s: %w", email, userDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database query failed for email %s: %w", email, err)
	}
	return &u, nil
}

// GetByEmailWithPassword retrieves a user by email with password included
func (r *userRepository) GetByEmailWithPassword(ctx context.Context, email string) (*userDomain.User, error) {
	var u userDomain.User
	err := r.db.WithContext(ctx).Select("*").Where("email = ? AND deleted_at IS NULL", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get user by email with password %s: %w", email, userDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database query failed for email with password %s: %w", email, err)
	}
	return &u, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, u *userDomain.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&userDomain.User{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

// List retrieves users with filters
func (r *userRepository) List(ctx context.Context, filters *userDomain.ListFilters) ([]*userDomain.User, int, error) {
	// Convert ListFilters to UserFilters for compatibility
	userFilters := (*userDomain.UserFilters)(filters)
	users, err := r.GetByFilters(ctx, userFilters)
	if err != nil {
		return nil, 0, err
	}
	
	// Get total count for the same filters - for now just return length
	// TODO: Implement proper count query with the same filters
	totalCount := len(users)
	return users, totalCount, nil
}

// Count returns the total number of active users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userDomain.User{}).Where("deleted_at IS NULL").Count(&count).Error
	return count, err
}

// UpdatePassword updates a user's password
func (r *userRepository) UpdatePassword(ctx context.Context, userID ulid.ULID, hashedPassword string) error {
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Update("password", hashedPassword).Error
}

// UpdateLastLogin updates the user's last login timestamp
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Update("last_login_at", time.Now()).Error
}

// MarkEmailAsVerified marks the user's email as verified
func (r *userRepository) MarkEmailAsVerified(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"is_email_verified": true,
			"email_verified_at": time.Now(),
		}).Error
}

// SetDefaultOrganization sets the user's default organization
func (r *userRepository) SetDefaultOrganization(ctx context.Context, userID ulid.ULID, orgID ulid.ULID) error {
	var orgIDPtr *ulid.ULID
	if orgID != (ulid.ULID{}) { // Check if not zero value
		orgIDPtr = &orgID
	}
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Update("default_organization_id", orgIDPtr).Error
}

// CompleteOnboarding marks the user's onboarding as completed
func (r *userRepository) CompleteOnboarding(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Update("onboarding_completed_at", time.Now()).Error
}

// GetActiveUsers returns active users (those who have logged in recently)
func (r *userRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]*userDomain.User, int, error) {
	var users []*userDomain.User
	var count int64
	
	// Get count first
	err := r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("deleted_at IS NULL AND last_login_at IS NOT NULL").
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get users
	err = r.db.WithContext(ctx).
		Where("deleted_at IS NULL AND last_login_at IS NOT NULL").
		Limit(limit).
		Offset(offset).
		Order("last_login_at DESC").
		Find(&users).Error
	return users, int(count), err
}

// GetUsersByIDs retrieves multiple users by their IDs
func (r *userRepository) GetUsersByIDs(ctx context.Context, ids []ulid.ULID) ([]*userDomain.User, error) {
	var users []*userDomain.User
	err := r.db.WithContext(ctx).
		Where("id IN ? AND deleted_at IS NULL", ids).
		Find(&users).Error
	return users, err
}

// SearchUsers searches users by email, first name, or last name
func (r *userRepository) SearchUsers(ctx context.Context, query string, limit, offset int) ([]*userDomain.User, int, error) {
	var users []*userDomain.User
	var count int64
	
	searchPattern := "%" + query + "%"
	whereClause := "deleted_at IS NULL AND (email ILIKE ? OR first_name ILIKE ? OR last_name ILIKE ?)"
	
	// Get count first
	err := r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where(whereClause, searchPattern, searchPattern, searchPattern).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get users
	err = r.db.WithContext(ctx).
		Where(whereClause, searchPattern, searchPattern, searchPattern).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	return users, int(count), err
}

// GetUserStats returns user statistics
func (r *userRepository) GetUserStats(ctx context.Context) (*userDomain.UserStats, error) {
	stats := &userDomain.UserStats{}

	// Total users
	err := r.db.WithContext(ctx).Model(&userDomain.User{}).Where("deleted_at IS NULL").Count(&stats.TotalUsers).Error
	if err != nil {
		return nil, err
	}

	// Active users (logged in within last 30 days)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	err = r.db.WithContext(ctx).Model(&userDomain.User{}).
		Where("deleted_at IS NULL AND last_login_at > ?", thirtyDaysAgo).
		Count(&stats.ActiveUsers).Error
	if err != nil {
		return nil, err
	}

	// Verified users
	err = r.db.WithContext(ctx).Model(&userDomain.User{}).
		Where("deleted_at IS NULL AND is_email_verified = true").
		Count(&stats.VerifiedUsers).Error
	if err != nil {
		return nil, err
	}

	// Users created today
	today := time.Now().Truncate(24 * time.Hour)
	err = r.db.WithContext(ctx).Model(&userDomain.User{}).
		Where("deleted_at IS NULL AND created_at >= ?", today).
		Count(&stats.NewUsersToday).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// UpdateUserActivity updates user activity timestamp
func (r *userRepository) UpdateUserActivity(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Update("last_activity_at", time.Now()).Error
}

// Deactivate deactivates a user account
func (r *userRepository) Deactivate(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Update("is_active", false).Error
}

// Activate activates a user account
func (r *userRepository) Activate(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where("id = ?", userID).
		Update("is_active", true).Error
}

// GetByFilters retrieves users based on filters
func (r *userRepository) GetByFilters(ctx context.Context, filters *userDomain.UserFilters) ([]*userDomain.User, error) {
	var users []*userDomain.User
	query := r.db.WithContext(ctx).Where("deleted_at IS NULL")

	// Apply filters
	if filters.IsActive != nil {
		query = query.Where("is_active = ?", *filters.IsActive)
	}
	if filters.IsEmailVerified != nil {
		query = query.Where("is_email_verified = ?", *filters.IsEmailVerified)
	}
	if filters.OnboardingCompleted != nil {
		if *filters.OnboardingCompleted {
			query = query.Where("onboarding_completed_at IS NOT NULL")
		} else {
			query = query.Where("onboarding_completed_at IS NULL")
		}
	}
	if filters.CreatedAfter != nil {
		query = query.Where("created_at > ?", *filters.CreatedAfter)
	}
	if filters.CreatedBefore != nil {
		query = query.Where("created_at < ?", *filters.CreatedBefore)
	}
	if filters.LastLoginAfter != nil {
		query = query.Where("last_login_at > ?", *filters.LastLoginAfter)
	}

	// Apply sorting
	switch filters.SortBy {
	case "email":
		if filters.SortOrder == "desc" {
			query = query.Order("email DESC")
		} else {
			query = query.Order("email ASC")
		}
	case "created_at":
		if filters.SortOrder == "desc" {
			query = query.Order("created_at DESC")
		} else {
			query = query.Order("created_at ASC")
		}
	case "last_login_at":
		if filters.SortOrder == "desc" {
			query = query.Order("last_login_at DESC")
		} else {
			query = query.Order("last_login_at ASC")
		}
	default:
		query = query.Order("created_at DESC")
	}

	// Apply pagination
	if filters.Limit > 0 {
		query = query.Limit(filters.Limit)
	}
	if filters.Offset > 0 {
		query = query.Offset(filters.Offset)
	}

	err := query.Find(&users).Error
	return users, err
}

// Profile operations
func (r *userRepository) CreateProfile(ctx context.Context, profile *userDomain.UserProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *userRepository) GetProfile(ctx context.Context, userID ulid.ULID) (*userDomain.UserProfile, error) {
	var profile userDomain.UserProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get profile for user %s: %w", userID, userDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database query failed for profile %s: %w", userID, err)
	}
	return &profile, nil
}

func (r *userRepository) UpdateProfile(ctx context.Context, profile *userDomain.UserProfile) error {
	return r.db.WithContext(ctx).Save(profile).Error
}


// Additional missing interface methods
func (r *userRepository) VerifyEmail(ctx context.Context, userID ulid.ULID, token string) error {
	// TODO: Implement token validation logic
	return r.MarkEmailAsVerified(ctx, userID)
}

func (r *userRepository) GetDefaultOrganization(ctx context.Context, userID ulid.ULID) (*ulid.ULID, error) {
	var u userDomain.User
	err := r.db.WithContext(ctx).Select("default_organization_id").Where("id = ?", userID).First(&u).Error
	if err != nil {
		return nil, err
	}
	return u.DefaultOrganizationID, nil
}

func (r *userRepository) DeactivateUser(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&userDomain.User{}).Where("id = ?", userID).Update("is_active", false).Error
}

func (r *userRepository) ReactivateUser(ctx context.Context, userID ulid.ULID) error {
	return r.db.WithContext(ctx).Model(&userDomain.User{}).Where("id = ?", userID).Update("is_active", true).Error
}

func (r *userRepository) IsOnboardingCompleted(ctx context.Context, userID ulid.ULID) (bool, error) {
	var u userDomain.User
	err := r.db.WithContext(ctx).Select("onboarding_completed_at").Where("id = ?", userID).First(&u).Error
	if err != nil {
		return false, err
	}
	return u.OnboardingCompletedAt != nil, nil
}

func (r *userRepository) GetNewUsersCount(ctx context.Context, since time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userDomain.User{}).Where("created_at > ? AND deleted_at IS NULL", since).Count(&count).Error
	return count, err
}

// GetUsersByOrganization returns users who belong to an organization
func (r *userRepository) GetUsersByOrganization(ctx context.Context, organizationID ulid.ULID) ([]*userDomain.User, error) {
	var users []*userDomain.User
	// This would require a join with the organization_members table
	// For now, return empty slice as this requires cross-domain queries
	// TODO: Implement proper join or separate query to get organization members
	return users, nil
}

// GetVerifiedUsers returns verified users
func (r *userRepository) GetVerifiedUsers(ctx context.Context, limit, offset int) ([]*userDomain.User, int, error) {
	var users []*userDomain.User
	var count int64
	
	whereClause := "deleted_at IS NULL AND is_email_verified = true"
	
	// Get count first
	err := r.db.WithContext(ctx).
		Model(&userDomain.User{}).
		Where(whereClause).
		Count(&count).Error
	if err != nil {
		return nil, 0, err
	}
	
	// Get users
	err = r.db.WithContext(ctx).
		Where(whereClause).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error
	return users, int(count), err
}

func (r *userRepository) Search(ctx context.Context, query string, limit, offset int) ([]*userDomain.User, int, error) {
	return r.SearchUsers(ctx, query, limit, offset)
}

func (r *userRepository) GetByIDs(ctx context.Context, ids []ulid.ULID) ([]*userDomain.User, error) {
	return r.GetUsersByIDs(ctx, ids)
}

// Transaction executes a function within a database transaction
func (r *userRepository) Transaction(fn func(userDomain.Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepo := &userRepository{db: tx}
		return fn(txRepo)
	})
}

// Onboarding question operations

func (r *userRepository) CreateOnboardingQuestion(ctx context.Context, question *userDomain.OnboardingQuestion) error {
	return r.db.WithContext(ctx).Create(question).Error
}

func (r *userRepository) GetActiveOnboardingQuestions(ctx context.Context) ([]*userDomain.OnboardingQuestion, error) {
	var questions []*userDomain.OnboardingQuestion
	err := r.db.WithContext(ctx).Where("is_active = ?", true).Order("display_order ASC, step ASC").Find(&questions).Error
	return questions, err
}

func (r *userRepository) GetOnboardingQuestionByID(ctx context.Context, id ulid.ULID) (*userDomain.OnboardingQuestion, error) {
	var question userDomain.OnboardingQuestion
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&question).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get onboarding question %s: %w", id, userDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database query failed for onboarding question %s: %w", id, err)
	}
	return &question, nil
}

func (r *userRepository) GetActiveOnboardingQuestionCount(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&userDomain.OnboardingQuestion{}).Where("is_active = ?", true).Count(&count).Error
	return int(count), err
}

func (r *userRepository) GetNextUnansweredQuestion(ctx context.Context, userID ulid.ULID) (*userDomain.OnboardingQuestion, error) {
	var question userDomain.OnboardingQuestion
	
	// Find the first question that the user hasn't answered
	subquery := r.db.Model(&userDomain.UserOnboardingResponse{}).
		Select("question_id").
		Where("user_id = ?", userID)
	
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND id NOT IN (?)", true, subquery).
		Order("display_order ASC, step ASC").
		First(&question).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get next unanswered question for user %s: %w", userID, userDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database query failed for next unanswered question %s: %w", userID, err)
	}
	
	return &question, nil
}

// Onboarding response operations

func (r *userRepository) GetUserOnboardingResponses(ctx context.Context, userID ulid.ULID) ([]*userDomain.UserOnboardingResponse, error) {
	var responses []*userDomain.UserOnboardingResponse
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&responses).Error
	return responses, err
}

func (r *userRepository) UpsertUserOnboardingResponse(ctx context.Context, response *userDomain.UserOnboardingResponse) error {
	// Try to update existing response first
	result := r.db.WithContext(ctx).Model(&userDomain.UserOnboardingResponse{}).
		Where("user_id = ? AND question_id = ?", response.UserID, response.QuestionID).
		Updates(map[string]interface{}{
			"response_value": response.ResponseValue,
			"skipped":        response.Skipped,
		})

	if result.Error != nil {
		return result.Error
	}

	// If no rows were affected, create new response
	if result.RowsAffected == 0 {
		return r.db.WithContext(ctx).Create(response).Error
	}

	return nil
}

func (r *userRepository) GetUserOnboardingResponse(ctx context.Context, userID, questionID ulid.ULID) (*userDomain.UserOnboardingResponse, error) {
	var response userDomain.UserOnboardingResponse
	err := r.db.WithContext(ctx).Where("user_id = ? AND question_id = ?", userID, questionID).First(&response).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("get onboarding response for user %s question %s: %w", userID, questionID, userDomain.ErrNotFound)
		}
		return nil, fmt.Errorf("database query failed for onboarding response %s %s: %w", userID, questionID, err)
	}
	return &response, nil
}