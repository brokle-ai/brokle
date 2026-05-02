package seeder_test

// TestFrontendPermissionsCatalogInSync re-runs the YAML→TS render in
// memory and asserts byte-equality with the on-disk
// `web/src/generated/permissions.ts`. Catches stale generated files
// in `go test ./...` (already gating CI) without requiring a separate
// `git diff --exit-code` step.
//
// On failure, the assertion message tells the contributor exactly how
// to fix it: `make gen-frontend-permissions`. The file is
// machine-derived from `seeds/permissions.yaml`; do not hand-edit.
//
// See CLAUDE.md 2026-04-30 (project-rbac).

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v3"

	"brokle/internal/seeder"
)

func TestFrontendPermissionsCatalogInSync(t *testing.T) {
	t.Parallel()
	root := repoRoot(t)

	yamlBody, err := os.ReadFile(filepath.Join(root, "seeds", "permissions.yaml"))
	if err != nil {
		t.Fatalf("read seeds/permissions.yaml: %v", err)
	}
	var doc seeder.PermissionsFile
	if err := yaml.Unmarshal(yamlBody, &doc); err != nil {
		t.Fatalf("parse seeds/permissions.yaml: %v", err)
	}

	want := seeder.RenderFrontendPermissions(doc.Permissions)

	tsPath := filepath.Join(root, "web", "src", "generated", "permissions.ts")
	got, err := os.ReadFile(tsPath)
	if err != nil {
		t.Fatalf("read %s: %v (run `make gen-frontend-permissions` if the file is missing)", tsPath, err)
	}

	if string(got) != string(want) {
		t.Errorf("frontend permission catalog is stale.\n"+
			"file: %s\n"+
			"fix:  run `make gen-frontend-permissions`\n"+
			"\n"+
			"the on-disk TS file does not match what would be generated from the\n"+
			"current seeds/permissions.yaml. seeds/permissions.yaml is the source\n"+
			"of truth; the TS file is mechanically derived. NEVER hand-edit the\n"+
			"generated file — it will be reverted by the next codegen run.\n",
			tsPath)
	}
}
