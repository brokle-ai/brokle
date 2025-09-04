//go:build enterprise
// +build enterprise

package compliance

// Enterprise build uses real implementation
// This file is only included in enterprise builds
// Real implementation would import from private ee-modules

// import "brokle/internal/ee-real/compliance"

// func New() Compliance {
//     return compliance.NewEnterpriseCompliance()
// }

// Note: This file would be replaced/overwritten in enterprise builds
// with real implementation imports