package organization

import (
	"context"

	"github.com/google/uuid"

	orgDomain "brokle/internal/core/domain/organization"
	appErrors "brokle/pkg/errors"
)

// OrganizationSettingsService implements *OrganizationSettingsService
type OrganizationSettingsService struct {
	settingsRepo orgDomain.OrganizationSettingsRepository
	memberRepo   orgDomain.MemberRepository
}

// NewOrganizationSettingsService creates a new organization settings service instance
func NewOrganizationSettingsService(
	settingsRepo orgDomain.OrganizationSettingsRepository,
	memberRepo orgDomain.MemberRepository,
) *OrganizationSettingsService {
	return &OrganizationSettingsService{
		settingsRepo: settingsRepo,
		memberRepo:   memberRepo,
	}
}

// CreateSetting creates a new organization setting
func (s *OrganizationSettingsService) CreateSetting(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, req *orgDomain.CreateOrganizationSettingRequest) (*orgDomain.OrganizationSettings, error) {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "create"); err != nil {
		return nil, err
	}

	// Check if setting already exists. The repo returns *Error
	// (NotFound on no-rows, Internal on infra failure); use the
	// canonical predicate instead of stringly-typed comparison so a
	// rename of the repo's error message can never silently flip this
	// branch into a 500.
	existing, err := s.settingsRepo.GetByKey(ctx, orgID, req.Key)
	if err != nil && !appErrors.IsNotFound(err) {
		return nil, appErrors.Internal("failed to check existing setting", err)
	}
	if existing != nil {
		return nil, appErrors.AlreadyExists("organization_setting")
	}

	// Create new setting
	setting, err := orgDomain.NewOrganizationSettings(orgID, req.Key, req.Value)
	if err != nil {
		return nil, appErrors.Internal("failed to create setting", err)
	}

	if err := s.settingsRepo.Create(ctx, setting); err != nil {
		return nil, appErrors.Internal("failed to save setting", err)
	}

	return setting, nil
}

// GetSetting retrieves a specific organization setting
func (s *OrganizationSettingsService) GetSetting(ctx context.Context, orgID uuid.UUID, key string) (*orgDomain.OrganizationSettings, error) {
	return s.settingsRepo.GetByKey(ctx, orgID, key)
}

// ListSettings retrieves all settings for an organization as a map
func (s *OrganizationSettingsService) ListSettings(ctx context.Context, orgID uuid.UUID) (map[string]any, error) {
	return s.settingsRepo.GetSettingsMap(ctx, orgID)
}

// UpdateSetting updates an existing organization setting
func (s *OrganizationSettingsService) UpdateSetting(ctx context.Context, orgID uuid.UUID, key string, userID uuid.UUID, req *orgDomain.UpdateOrganizationSettingRequest) (*orgDomain.OrganizationSettings, error) {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "update"); err != nil {
		return nil, err
	}

	// Get existing setting. Pass the repo's self-describing *Error
	// through (NotFound vs Internal) — the previous blanket
	// translation to NotFound("setting") was hiding 500s as 404s.
	setting, err := s.settingsRepo.GetByKey(ctx, orgID, key)
	if err != nil {
		return nil, err
	}

	// Update setting value
	if err := setting.SetValue(req.Value); err != nil {
		return nil, appErrors.Internal("failed to set value", err)
	}

	if err := s.settingsRepo.Update(ctx, setting); err != nil {
		return nil, err
	}

	return setting, nil
}

// DeleteSetting deletes an organization setting
func (s *OrganizationSettingsService) DeleteSetting(ctx context.Context, orgID uuid.UUID, key string, userID uuid.UUID) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "delete"); err != nil {
		return err
	}

	// Verify setting exists before deletion. Pass the repo's
	// self-describing *Error through (NotFound vs Internal) — the
	// previous blanket translation hid 500s as 404s.
	if _, err := s.settingsRepo.GetByKey(ctx, orgID, key); err != nil {
		return err
	}

	if err := s.settingsRepo.DeleteByKey(ctx, orgID, key); err != nil {
		return err
	}

	return nil
}

// UpsertSetting creates or updates a setting
func (s *OrganizationSettingsService) UpsertSetting(ctx context.Context, orgID uuid.UUID, key string, value any, userID uuid.UUID) (*orgDomain.OrganizationSettings, error) {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "upsert"); err != nil {
		return nil, err
	}

	setting, err := s.settingsRepo.UpsertSetting(ctx, orgID, key, value)
	if err != nil {
		return nil, err
	}

	return setting, nil
}

// CreateMultipleSettings creates multiple settings in bulk
func (s *OrganizationSettingsService) CreateMultipleSettings(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, settings map[string]any) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "bulk_create"); err != nil {
		return err
	}

	settingEntities := make([]*orgDomain.OrganizationSettings, 0, len(settings))
	for key, value := range settings {
		setting, err := orgDomain.NewOrganizationSettings(orgID, key, value)
		if err != nil {
			return appErrors.Internal("failed to create setting for key "+key, err)
		}
		settingEntities = append(settingEntities, setting)
	}

	if err := s.settingsRepo.CreateMultiple(ctx, settingEntities); err != nil {
		return appErrors.Internal("failed to create multiple settings", err)
	}

	return nil
}

// GetSettingsByKeys retrieves specific settings by keys
func (s *OrganizationSettingsService) GetSettingsByKeys(ctx context.Context, orgID uuid.UUID, keys []string) (map[string]any, error) {
	settings, err := s.settingsRepo.GetByKeys(ctx, orgID, keys)
	if err != nil {
		return nil, appErrors.Internal("failed to get settings by keys", err)
	}

	result := make(map[string]any)
	for _, setting := range settings {
		value, err := setting.GetValue()
		if err != nil {
			// If unmarshaling fails, store as string
			value = setting.Value
		}
		result[setting.Key] = value
	}

	return result, nil
}

// DeleteMultipleSettings deletes multiple settings by keys
func (s *OrganizationSettingsService) DeleteMultipleSettings(ctx context.Context, orgID uuid.UUID, keys []string, userID uuid.UUID) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "bulk_delete"); err != nil {
		return err
	}

	if err := s.settingsRepo.DeleteMultiple(ctx, orgID, keys); err != nil {
		return appErrors.Internal("failed to delete multiple settings", err)
	}

	return nil
}

// ValidateSettingsAccess validates if user can perform settings operations
func (s *OrganizationSettingsService) ValidateSettingsAccess(ctx context.Context, userID, orgID uuid.UUID, operation string) error {
	// Check if user is a member of the organization
	isMember, err := s.memberRepo.IsMember(ctx, userID, orgID)
	if err != nil {
		return appErrors.Internal("failed to check membership", err)
	}
	if !isMember {
		return appErrors.PermissionDenied("organization", "user is not a member of this organization")
	}

	// For now, allow any member to manage settings
	// This could be enhanced with role-based permissions
	return nil
}

// CanUserManageSettings checks if user can manage organization settings
func (s *OrganizationSettingsService) CanUserManageSettings(ctx context.Context, userID, orgID uuid.UUID) (bool, error) {
	err := s.ValidateSettingsAccess(ctx, userID, orgID, "manage")
	return err == nil, nil
}

// ResetToDefaults resets organization settings to default values
func (s *OrganizationSettingsService) ResetToDefaults(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "reset"); err != nil {
		return err
	}

	// Get all current settings
	currentSettings, err := s.settingsRepo.GetAllByOrganizationID(ctx, orgID)
	if err != nil {
		return appErrors.Internal("failed to get current settings", err)
	}

	// Delete all current settings
	if len(currentSettings) > 0 {
		keys := make([]string, len(currentSettings))
		for i, setting := range currentSettings {
			keys[i] = setting.Key
		}
		if err := s.settingsRepo.DeleteMultiple(ctx, orgID, keys); err != nil {
			return appErrors.Internal("failed to clear current settings", err)
		}
	}

	return nil
}

// ExportSettings exports all organization settings
func (s *OrganizationSettingsService) ExportSettings(ctx context.Context, orgID uuid.UUID, userID uuid.UUID) (map[string]any, error) {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "export"); err != nil {
		return nil, err
	}

	settings, err := s.settingsRepo.GetSettingsMap(ctx, orgID)
	if err != nil {
		return nil, appErrors.Internal("failed to export settings", err)
	}

	return settings, nil
}

// ImportSettings imports organization settings
func (s *OrganizationSettingsService) ImportSettings(ctx context.Context, orgID uuid.UUID, userID uuid.UUID, settings map[string]any) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "import"); err != nil {
		return err
	}

	// Create or update each setting
	for key, value := range settings {
		_, err := s.settingsRepo.UpsertSetting(ctx, orgID, key, value)
		if err != nil {
			return appErrors.Internal("failed to import setting "+key, err)
		}
	}

	return nil
}
