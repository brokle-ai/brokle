package pagination

// Package pagination provides offset-based pagination utilities.
// The Pagination struct and response helpers are in pkg/response for cleaner architecture.

// Constants for valid page sizes across the platform
const (
	DefaultPageSize = 50    // Default items per page
	MaxPageSize     = 100   // Maximum allowed items per page
	MinPageSize     = 10    // Minimum allowed items per page
	MaxOffset       = 10000 // Maximum safe offset to prevent performance issues
)

// ValidPageSizes are the allowed page sizes: 10, 25, 50, 100
var ValidPageSizes = []int{10, 25, 50, 100}

// IsValidPageSize checks if a page size is valid (10, 25, 50, 100)
func IsValidPageSize(size int) bool {
	for _, validSize := range ValidPageSizes {
		if size == validSize {
			return true
		}
	}
	return false
}

// CalculateTotalPages calculates total number of pages based on total count and limit
func CalculateTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}
	return totalPages
}
