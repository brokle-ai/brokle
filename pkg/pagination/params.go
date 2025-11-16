package pagination

import (
	"errors"
	"fmt"
)

// Params represents offset-based pagination parameters
// Embed this struct in domain filters for DRY pagination
type Params struct {
	SortBy  string `json:"sort_by"`
	SortDir string `json:"sort_dir"`
	Page    int    `json:"page"`
	Limit   int    `json:"limit"`
}

// Validate validates pagination parameters
func (p *Params) Validate() error {
	// Validate page number (must be >= 1)
	if p.Page < 1 {
		return errors.New("page must be >= 1")
	}

	// Validate limit
	if p.Limit != 0 && !IsValidPageSize(p.Limit) {
		return errors.New("limit must be one of: 10, 25, 50, 100")
	}

	// Validate sort direction
	if p.SortDir != "" && p.SortDir != "asc" && p.SortDir != "desc" {
		return errors.New("sort_dir must be 'asc' or 'desc'")
	}

	// Validate maximum offset to prevent performance issues
	offset := p.GetOffset()
	if offset > MaxOffset {
		return fmt.Errorf("offset %d exceeds maximum allowed %d (consider using filters to narrow results)", offset, MaxOffset)
	}

	return nil
}

// SetDefaults sets default values for pagination parameters
func (p *Params) SetDefaults(defaultSortBy string) {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.Limit == 0 || !IsValidPageSize(p.Limit) {
		p.Limit = DefaultPageSize
	}
	if p.SortBy == "" {
		p.SortBy = defaultSortBy
	}
	if p.SortDir == "" {
		p.SortDir = "desc"
	}
}

// GetOffset calculates the OFFSET value for SQL queries
// Formula: (page - 1) * limit
func (p *Params) GetOffset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

// ValidateSortField validates a sort field against a whitelist of allowed columns
// Returns the validated field or an error if the field is not in the whitelist
// This prevents SQL injection attacks via user-supplied sort_by parameters
func ValidateSortField(field string, allowedFields []string) (string, error) {
	if field == "" {
		return "", nil // Empty field will use default
	}

	for _, allowed := range allowedFields {
		if field == allowed {
			return field, nil
		}
	}

	return "", fmt.Errorf("invalid sort field '%s', allowed fields: %v", field, allowedFields)
}

// GetSortOrder returns SQL-formatted sort order
// Always includes secondary sort field for stable ordering
func (p *Params) GetSortOrder(primaryField, secondaryField string) string {
	sortField := p.SortBy
	if sortField == "" {
		sortField = primaryField
	}

	sortDir := "DESC"
	if p.SortDir == "asc" {
		sortDir = "ASC"
	}

	// Always include secondary sort for stable ordering
	return fmt.Sprintf("%s %s, %s %s", sortField, sortDir, secondaryField, sortDir)
}
