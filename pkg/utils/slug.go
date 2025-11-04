package utils

import (
	"errors"
	"regexp"
	"strings"

	"brokle/pkg/ulid"
)

var slugRegex = regexp.MustCompile(`[^a-z0-9]+`)

// GenerateCompositeSlug creates a URL-friendly slug from name and ID
// Format: "{name-slug}-{ulid}"
// Example: "acme-corp-01K4FHGHT3XX9WFM293QPZ5G9V"
func GenerateCompositeSlug(name string, id ulid.ULID) string {
	// Convert name to slug
	slug := strings.ToLower(name)
	slug = slugRegex.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	// Combine with ID
	return slug + "-" + id.String()
}

// ExtractIDFromCompositeSlug extracts ULID from composite slug
// Input: "acme-corp-01K4FHGHT3XX9WFM293QPZ5G9V"
// Output: ULID
func ExtractIDFromCompositeSlug(compositeSlug string) (ulid.ULID, error) {
	// ULID is always last 26 characters
	if len(compositeSlug) < 26 {
		return ulid.ULID{}, errors.New("invalid composite slug: too short")
	}

	idStr := compositeSlug[len(compositeSlug)-26:]
	return ulid.Parse(idStr)
}
