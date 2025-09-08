package seeder

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// DataLoader handles loading seed data from YAML files
type DataLoader struct{}

// NewDataLoader creates a new DataLoader instance
func NewDataLoader() *DataLoader {
	return &DataLoader{}
}

// LoadSeedData loads seed data for the specified environment mode
func (dl *DataLoader) LoadSeedData(mode string) (*SeedData, error) {
	// Handle common aliases
	aliases := map[string]string{
		"development": "dev",
		"dev":         "dev",
		"demo":        "demo",
		"test":        "test",
	}
	
	actualMode, ok := aliases[mode]
	if !ok {
		actualMode = mode // Use the mode as-is if no alias found
	}
	
	// Get the seed file path
	seedFile := fmt.Sprintf("seeds/%s.yaml", actualMode)
	
	// Check if file exists in current directory first
	if _, err := os.Stat(seedFile); os.IsNotExist(err) {
		// Try relative path from brokle directory
		broklePath := filepath.Join("brokle", seedFile)
		if _, err := os.Stat(broklePath); os.IsNotExist(err) {
			return nil, fmt.Errorf("seed file not found: %s (also tried: %s)", seedFile, broklePath)
		}
		seedFile = broklePath
	}

	// Read the file
	data, err := os.ReadFile(seedFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read seed file %s: %w", seedFile, err)
	}

	// Parse YAML
	var seedData SeedData
	if err := yaml.Unmarshal(data, &seedData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML from %s: %w", seedFile, err)
	}

	// Validate required fields
	if err := dl.validateSeedData(&seedData); err != nil {
		return nil, fmt.Errorf("invalid seed data in %s: %w", seedFile, err)
	}

	return &seedData, nil
}

// validateSeedData validates the seed data for consistency and required fields
func (dl *DataLoader) validateSeedData(data *SeedData) error {
	// Validate organizations
	orgSlugs := make(map[string]bool)
	for _, org := range data.Organizations {
		if org.Name == "" || org.Slug == "" {
			return fmt.Errorf("organization missing required fields (name, slug)")
		}
		if orgSlugs[org.Slug] {
			return fmt.Errorf("duplicate organization slug: %s", org.Slug)
		}
		orgSlugs[org.Slug] = true
	}

	// Validate users
	userEmails := make(map[string]bool)
	for _, user := range data.Users {
		if user.Email == "" || user.FirstName == "" || user.LastName == "" {
			return fmt.Errorf("user missing required fields (email, first_name, last_name)")
		}
		if userEmails[user.Email] {
			return fmt.Errorf("duplicate user email: %s", user.Email)
		}
		userEmails[user.Email] = true
	}

	// Validate permissions have required fields
	permissionNames := make(map[string]bool)
	for _, permission := range data.RBAC.Permissions {
		if permission.Name == "" {
			return fmt.Errorf("permission missing required field: name")
		}
		if permissionNames[permission.Name] {
			return fmt.Errorf("duplicate permission name: %s", permission.Name)
		}
		permissionNames[permission.Name] = true
	}

	// Validate template roles 
	for _, role := range data.RBAC.Roles {
		if role.Name == "" || role.ScopeType == "" {
			return fmt.Errorf("role missing required fields (name, scope_type)")
		}
		// Validate role permissions reference valid permissions
		for _, permName := range role.Permissions {
			if !permissionNames[permName] {
				return fmt.Errorf("role %s references unknown permission: %s", role.Name, permName)
			}
		}
	}

	// Validate memberships reference valid users and organizations
	for _, membership := range data.RBAC.Memberships {
		if !userEmails[membership.UserEmail] {
			return fmt.Errorf("membership references unknown user: %s", membership.UserEmail)
		}
		if !orgSlugs[membership.OrganizationSlug] {
			return fmt.Errorf("membership references unknown organization: %s", membership.OrganizationSlug)
		}
	}

	// Validate projects reference valid organizations
	projectKeys := make(map[string]bool) // org_slug:project_name
	for _, project := range data.Projects {
		if !orgSlugs[project.OrganizationSlug] {
			return fmt.Errorf("project references unknown organization: %s", project.OrganizationSlug)
		}
		if project.Name == "" {
			return fmt.Errorf("project missing required field: name")
		}
		projectKey := fmt.Sprintf("%s:%s", project.OrganizationSlug, project.Name)
		if projectKeys[projectKey] {
			return fmt.Errorf("duplicate project: %s in organization %s", project.Name, project.OrganizationSlug)
		}
		projectKeys[projectKey] = true
	}

	// Validate onboarding questions
	steps := make(map[int]bool)
	for _, question := range data.OnboardingQuestions {
		if question.Title == "" || question.QuestionType == "" {
			return fmt.Errorf("onboarding question missing required fields (title, question_type)")
		}
		if steps[question.Step] {
			return fmt.Errorf("duplicate onboarding question step: %d", question.Step)
		}
		steps[question.Step] = true
		
		// Validate question type
		validTypes := map[string]bool{
			"single_choice":   true,
			"multiple_choice": true,
			"text":           true,
			"skip_optional":  true,
		}
		if !validTypes[question.QuestionType] {
			return fmt.Errorf("invalid question type: %s", question.QuestionType)
		}
		
		// Validate choice questions have options
		if (question.QuestionType == "single_choice" || question.QuestionType == "multiple_choice") && len(question.Options) == 0 {
			return fmt.Errorf("choice question '%s' missing options", question.Title)
		}
	}

	return nil
}