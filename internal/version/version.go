package version

// Version is the current version of the Brokle platform.
// This value is set via ldflags during the build process.
// Default value is "dev" for local development builds.
//
// To set the version during build:
//
//	go build -ldflags="-X brokle/internal/version.Version=v1.2.3" ./cmd/server
var Version = "dev"

// Get returns the current version of the application.
func Get() string {
	if Version == "" {
		return "dev"
	}
	return Version
}
