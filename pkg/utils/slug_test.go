package utils

import (
	"testing"

	"brokle/pkg/ulid"
)

func TestGenerateCompositeSlug(t *testing.T) {
	tests := []struct {
		name     string
		orgName  string
		id       ulid.ULID
		expected string
	}{
		{
			name:     "simple name",
			orgName:  "Acme Corp",
			id:       ulid.MustParse("01K4FHGHT3XX9WFM293QPZ5G9V"),
			expected: "acme-corp-01K4FHGHT3XX9WFM293QPZ5G9V",
		},
		{
			name:     "name with special characters",
			orgName:  "Acme & Co. Inc!",
			id:       ulid.MustParse("01K4FHGHT3XX9WFM293QPZ5G9V"),
			expected: "acme-co-inc-01K4FHGHT3XX9WFM293QPZ5G9V",
		},
		{
			name:     "name with multiple spaces",
			orgName:  "The   Big   Company",
			id:       ulid.MustParse("01K4FHGHT3XX9WFM293QPZ5G9V"),
			expected: "the-big-company-01K4FHGHT3XX9WFM293QPZ5G9V",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateCompositeSlug(tt.orgName, tt.id)
			if result != tt.expected {
				t.Errorf("GenerateCompositeSlug() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractIDFromCompositeSlug(t *testing.T) {
	tests := []struct {
		name          string
		compositeSlug string
		expectedID    string
		wantErr       bool
	}{
		{
			name:          "valid composite slug",
			compositeSlug: "acme-corp-01K4FHGHT3XX9WFM293QPZ5G9V",
			expectedID:    "01K4FHGHT3XX9WFM293QPZ5G9V",
			wantErr:       false,
		},
		{
			name:          "too short",
			compositeSlug: "short",
			wantErr:       true,
		},
		{
			name:          "just id",
			compositeSlug: "01K4FHGHT3XX9WFM293QPZ5G9V",
			expectedID:    "01K4FHGHT3XX9WFM293QPZ5G9V",
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractIDFromCompositeSlug(tt.compositeSlug)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractIDFromCompositeSlug() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result.String() != tt.expectedID {
				t.Errorf("ExtractIDFromCompositeSlug() = %v, want %v", result.String(), tt.expectedID)
			}
		})
	}
}
