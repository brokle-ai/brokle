package auth

import (
	"context"
	
	"gorm.io/gorm"

	"brokle/internal/core/domain/auth"
	"brokle/pkg/ulid"
)

// userRoleRepository implements clean auth.UserRoleRepository using GORM
type userRoleRepository struct {
	db *gorm.DB
}

// NewUserRoleRepository creates a new clean user role repository instance
func NewUserRoleRepository(db *gorm.DB) auth.UserRoleRepository {
	return &userRoleRepository{
		db: db,
	}
}

// Core CRUD operations

func (r *userRoleRepository) Create(ctx context.Context, userRole *auth.UserRole) error {
	return r.db.WithContext(ctx).Create(userRole).Error
}

func (r *userRoleRepository) Delete(ctx context.Context, userID, roleID ulid.ULID) error {
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&auth.UserRole{}).Error
}

func (r *userRoleRepository) GetByUser(ctx context.Context, userID ulid.ULID) ([]*auth.UserRole, error) {
	var userRoles []*auth.UserRole
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Preload("Role").
		Find(&userRoles).Error
	return userRoles, err
}

func (r *userRoleRepository) GetByRole(ctx context.Context, roleID ulid.ULID) ([]*auth.UserRole, error) {
	var userRoles []*auth.UserRole
	err := r.db.WithContext(ctx).
		Where("role_id = ?", roleID).
		Find(&userRoles).Error
	return userRoles, err
}

func (r *userRoleRepository) Exists(ctx context.Context, userID, roleID ulid.ULID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&auth.UserRole{}).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Count(&count).Error
	return count > 0, err
}

// Bulk operations

func (r *userRoleRepository) BulkAssign(ctx context.Context, userRoles []*auth.UserRole) error {
	if len(userRoles) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&userRoles).Error
}

func (r *userRoleRepository) BulkRevoke(ctx context.Context, userID ulid.ULID, roleIDs []ulid.ULID) error {
	if len(roleIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("user_id = ? AND role_id IN ?", userID, roleIDs).
		Delete(&auth.UserRole{}).Error
}

// Statistics

func (r *userRoleRepository) GetUserRoleCount(ctx context.Context, userID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&auth.UserRole{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return int(count), err
}

func (r *userRoleRepository) GetRoleUserCount(ctx context.Context, roleID ulid.ULID) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&auth.UserRole{}).
		Where("role_id = ?", roleID).
		Count(&count).Error
	return int(count), err
}