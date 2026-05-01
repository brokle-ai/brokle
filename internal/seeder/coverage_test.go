package seeder_test

// Build-time drift guard: every permission string referenced by an
// `r.With(middleware.RequirePermission(authD, "..."))` decoration in
// `internal/server/routes.go` MUST exist in `seeds/permissions.yaml`.
// Closes the regression class where a new route gates on a permission
// that no role holds — every caller would 403 in dev/prod alike.
//
// This is a pure-text test: it parses the routes file with a regex and
// the YAML with go-yaml. No DB, no service wiring. Runs in <50ms as
// part of `go test ./internal/seeder/...`.

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"testing"

	"gopkg.in/yaml.v3"
)

// repoRoot walks up from the test file's directory until it finds the
// go.mod, so the test runs correctly regardless of where `go test` is
// invoked from.
func repoRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := cwd
	for i := 0; i < 8; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("could not locate repo root from %s", cwd)
	return ""
}

// permissionLiteralRE matches the exact pattern used in routes.go:
//
//	middleware.RequirePermission(authD, "<perm>")
//	middleware.RequireAnyPermission(authD, "<a>", "<b>", ...)
//	middleware.RequireAllPermissions(authD, "<a>", "<b>", ...)
//
// We extract every quoted string that appears inside one of these calls.
// The regex is deliberately broad on the call site (matches all three
// constructors) and tight on the argument shape ("<lowercase-with-hyphens>:<lowercase>").
var permissionLiteralRE = regexp.MustCompile(`middleware\.Require(?:Any|All)?Permissions?\(authD,\s*((?:"[^"]+"(?:,\s*)?)+)\)`)
var quotedRE = regexp.MustCompile(`"([a-z][a-z0-9_-]*:[a-z][a-z_]*)"`)

func extractRouteFilePermissions(t *testing.T, routesPath string) []string {
	t.Helper()
	body, err := os.ReadFile(routesPath)
	if err != nil {
		t.Fatalf("read %s: %v", routesPath, err)
	}
	seen := make(map[string]struct{})
	for _, m := range permissionLiteralRE.FindAllSubmatch(body, -1) {
		// m[1] is the comma-separated string list inside the call.
		for _, q := range quotedRE.FindAllSubmatch(m[1], -1) {
			seen[string(q[1])] = struct{}{}
		}
	}
	out := make([]string, 0, len(seen))
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func loadSeededPermissions(t *testing.T, yamlPath string) map[string]struct{} {
	t.Helper()
	body, err := os.ReadFile(yamlPath)
	if err != nil {
		t.Fatalf("read %s: %v", yamlPath, err)
	}
	var doc struct {
		Permissions []struct {
			Name string `yaml:"name"`
		} `yaml:"permissions"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		t.Fatalf("parse %s: %v", yamlPath, err)
	}
	out := make(map[string]struct{}, len(doc.Permissions))
	for _, p := range doc.Permissions {
		out[p.Name] = struct{}{}
	}
	return out
}

// TestRoutePermissionsSeeded asserts every permission referenced by a
// `RequirePermission`/`Any`/`All` decoration in `routes.go` is present
// in `seeds/permissions.yaml`. Catches the regression class where a new
// route is added with a perm string that the seeder doesn't know about
// — the route would 403 in dev/prod for every caller.
func TestRoutePermissionsSeeded(t *testing.T) {
	root := repoRoot(t)
	used := extractRouteFilePermissions(t, filepath.Join(root, "internal", "server", "routes.go"))
	if len(used) == 0 {
		t.Fatalf("regex matched zero RequirePermission literals — routes.go format may have changed")
	}

	seeded := loadSeededPermissions(t, filepath.Join(root, "seeds", "permissions.yaml"))
	if len(seeded) == 0 {
		t.Fatalf("loaded zero permissions from seeds/permissions.yaml")
	}

	var missing []string
	for _, p := range used {
		if _, ok := seeded[p]; !ok {
			missing = append(missing, p)
		}
	}
	if len(missing) > 0 {
		t.Errorf("routes.go references %d permission(s) not in seeds/permissions.yaml: %v", len(missing), missing)
	}
}
