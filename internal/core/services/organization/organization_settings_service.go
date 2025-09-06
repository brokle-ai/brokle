package organization

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"brokle/internal/core/domain/auth"
	"brokle/internal/core/domain/organization"
	"brokle/pkg/ulid"
)

// organizationSettingsService implements organization.OrganizationSettingsService
type organizationSettingsService struct {
	settingsRepo organization.OrganizationSettingsRepository
	memberRepo   organization.MemberRepository
	auditRepo    auth.AuditLogRepository
}

// NewOrganizationSettingsService creates a new organization settings service instance
func NewOrganizationSettingsService(
	settingsRepo organization.OrganizationSettingsRepository,
	memberRepo organization.MemberRepository,
	auditRepo auth.AuditLogRepository,
) organization.OrganizationSettingsService {
	return &organizationSettingsService{
		settingsRepo: settingsRepo,
		memberRepo:   memberRepo,
		auditRepo:    auditRepo,
	}
}

// CreateSetting creates a new organization setting
func (s *organizationSettingsService) CreateSetting(ctx context.Context, orgID ulid.ULID, userID ulid.ULID, req *organization.CreateOrganizationSettingRequest) (*organization.OrganizationSettings, error) {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "create"); err != nil {
		return nil, err
	}

	// Check if setting already exists
	existing, err := s.settingsRepo.GetByKey(ctx, orgID, req.Key)
	if err != nil && err.Error() != "organization setting not found" {
		return nil, fmt.Errorf("failed to check existing setting: %w", err)
	}
	if existing != nil {
		return nil, errors.New("setting with this key already exists")
	}

	// Create new setting
	setting, err := organization.NewOrganizationSettings(orgID, req.Key, req.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to create setting: %w", err)
	}

	if err := s.settingsRepo.Create(ctx, setting); err != nil {
		return nil, fmt.Errorf("failed to save setting: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.create", "setting", setting.ID.String(), map[string]interface{}{
		"key": req.Key,
	})

	return setting, nil
}

// GetSetting retrieves a specific organization setting
func (s *organizationSettingsService) GetSetting(ctx context.Context, orgID ulid.ULID, key string) (*organization.OrganizationSettings, error) {
	return s.settingsRepo.GetByKey(ctx, orgID, key)
}

// GetAllSettings retrieves all settings for an organization as a map
func (s *organizationSettingsService) GetAllSettings(ctx context.Context, orgID ulid.ULID) (map[string]interface{}, error) {
	return s.settingsRepo.GetSettingsMap(ctx, orgID)
}

// UpdateSetting updates an existing organization setting
func (s *organizationSettingsService) UpdateSetting(ctx context.Context, orgID ulid.ULID, key string, userID ulid.ULID, req *organization.UpdateOrganizationSettingRequest) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "update"); err != nil {
		return err
	}

	// Get existing setting
	setting, err := s.settingsRepo.GetByKey(ctx, orgID, key)
	if err != nil {
		return fmt.Errorf("setting not found: %w", err)
	}

	// Update setting value
	oldValue, _ := setting.GetValue()
	if err := setting.SetValue(req.Value); err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}

	if err := s.settingsRepo.Update(ctx, setting); err != nil {
		return fmt.Errorf("failed to update setting: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.update", "setting", setting.ID.String(), map[string]interface{}{
		"key":       key,
		"old_value": oldValue,
		"new_value": req.Value,
	})

	return nil
}

// DeleteSetting deletes an organization setting
func (s *organizationSettingsService) DeleteSetting(ctx context.Context, orgID ulid.ULID, key string, userID ulid.ULID) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "delete"); err != nil {
		return err
	}

	// Get setting for audit purposes
	setting, err := s.settingsRepo.GetByKey(ctx, orgID, key)
	if err != nil {
		return fmt.Errorf("setting not found: %w", err)
	}

	if err := s.settingsRepo.DeleteByKey(ctx, orgID, key); err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.delete", "setting", setting.ID.String(), map[string]interface{}{
		"key": key,
	})

	return nil
}

// UpsertSetting creates or updates a setting
func (s *organizationSettingsService) UpsertSetting(ctx context.Context, orgID ulid.ULID, key string, value interface{}, userID ulid.ULID) (*organization.OrganizationSettings, error) {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "upsert"); err != nil {
		return nil, err
	}

	setting, err := s.settingsRepo.UpsertSetting(ctx, orgID, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert setting: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.upsert", "setting", setting.ID.String(), map[string]interface{}{
		"key":   key,
		"value": value,
	})

	return setting, nil
}

// CreateMultipleSettings creates multiple settings in bulk
func (s *organizationSettingsService) CreateMultipleSettings(ctx context.Context, orgID ulid.ULID, userID ulid.ULID, settings map[string]interface{}) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "bulk_create"); err != nil {
		return err
	}

	settingEntities := make([]*organization.OrganizationSettings, 0, len(settings))
	for key, value := range settings {
		setting, err := organization.NewOrganizationSettings(orgID, key, value)
		if err != nil {
			return fmt.Errorf("failed to create setting for key %s: %w", key, err)
		}
		settingEntities = append(settingEntities, setting)
	}

	if err := s.settingsRepo.CreateMultiple(ctx, settingEntities); err != nil {
		return fmt.Errorf("failed to create multiple settings: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.bulk_create", "settings", "", map[string]interface{}{
		"count": len(settings),
		"keys":  getMapKeys(settings),
	})

	return nil
}

// GetSettingsByKeys retrieves specific settings by keys
func (s *organizationSettingsService) GetSettingsByKeys(ctx context.Context, orgID ulid.ULID, keys []string) (map[string]interface{}, error) {
	settings, err := s.settingsRepo.GetByKeys(ctx, orgID, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings by keys: %w", err)
	}

	result := make(map[string]interface{})
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
func (s *organizationSettingsService) DeleteMultipleSettings(ctx context.Context, orgID ulid.ULID, keys []string, userID ulid.ULID) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "bulk_delete"); err != nil {
		return err
	}

	if err := s.settingsRepo.DeleteMultiple(ctx, orgID, keys); err != nil {
		return fmt.Errorf("failed to delete multiple settings: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.bulk_delete", "settings", "", map[string]interface{}{
		"count": len(keys),
		"keys":  keys,
	})

	return nil
}

// ValidateSettingsAccess validates if user can perform settings operations
func (s *organizationSettingsService) ValidateSettingsAccess(ctx context.Context, userID, orgID ulid.ULID, operation string) error {
	// Check if user is a member of the organization
	isMember, err := s.memberRepo.IsMember(ctx, userID, orgID)
	if err != nil {
		return fmt.Errorf("failed to check membership: %w", err)
	}
	if !isMember {
		return errors.New("user is not a member of this organization")
	}

	// For now, allow any member to manage settings
	// This could be enhanced with role-based permissions
	return nil
}

// CanUserManageSettings checks if user can manage organization settings
func (s *organizationSettingsService) CanUserManageSettings(ctx context.Context, userID, orgID ulid.ULID) (bool, error) {
	err := s.ValidateSettingsAccess(ctx, userID, orgID, "manage")
	return err == nil, nil
}

// ResetToDefaults resets organization settings to default values
func (s *organizationSettingsService) ResetToDefaults(ctx context.Context, orgID ulid.ULID, userID ulid.ULID) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "reset"); err != nil {
		return err
	}

	// Get all current settings
	currentSettings, err := s.settingsRepo.GetAllByOrganizationID(ctx, orgID)
	if err != nil {
		return fmt.Errorf("failed to get current settings: %w", err)
	}

	// Delete all current settings
	if len(currentSettings) > 0 {
		keys := make([]string, len(currentSettings))
		for i, setting := range currentSettings {
			keys[i] = setting.Key
		}
		if err := s.settingsRepo.DeleteMultiple(ctx, orgID, keys); err != nil {
			return fmt.Errorf("failed to clear current settings: %w", err)
		}
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.reset_defaults", "settings", "", map[string]interface{}{
		"cleared_count": len(currentSettings),
	})

	return nil
}

// ExportSettings exports all organization settings
func (s *organizationSettingsService) ExportSettings(ctx context.Context, orgID ulid.ULID, userID ulid.ULID) (map[string]interface{}, error) {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "export"); err != nil {
		return nil, err
	}

	settings, err := s.settingsRepo.GetSettingsMap(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to export settings: %w", err)
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.export", "settings", "", map[string]interface{}{
		"exported_count": len(settings),
	})

	return settings, nil
}

// ImportSettings imports organization settings
func (s *organizationSettingsService) ImportSettings(ctx context.Context, orgID ulid.ULID, userID ulid.ULID, settings map[string]interface{}) error {
	// Validate user access
	if err := s.ValidateSettingsAccess(ctx, userID, orgID, "import"); err != nil {
		return err
	}

	// Create or update each setting
	for key, value := range settings {
		_, err := s.settingsRepo.UpsertSetting(ctx, orgID, key, value)
		if err != nil {
			return fmt.Errorf("failed to import setting %s: %w", key, err)
		}
	}

	// Create audit log
	s.createAuditLog(ctx, &userID, &orgID, "setting.import", "settings", "", map[string]interface{}{
		"imported_count": len(settings),
		"keys":          getMapKeys(settings),
	})

	return nil
}

// Helper methods

func (s *organizationSettingsService) createAuditLog(ctx context.Context, userID, orgID *ulid.ULID, action, resource, resourceID string, metadata map[string]interface{}) {
	var metadataStr string
	if metadata != nil {
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			metadataStr = string(metadataBytes)
		}
	}
	
	auditLog := auth.NewAuditLog(userID, orgID, action, resource, resourceID, metadataStr, "", "")
	// Note: IP address and user agent would come from context in a real implementation
	s.auditRepo.Create(ctx, auditLog)
}

func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}