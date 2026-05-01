// gen-frontend-permissions reads seeds/permissions.yaml and emits
// web/src/generated/permissions.ts (the canonical Scope union +
// SCOPE_LEVELS map consumed by the frontend useHasAccess hook).
//
// Run: `make gen-frontend-permissions` (also wired into `make generate`).
//
// Render logic lives in internal/seeder/permissions_codegen.go so the
// drift-guard test can call it in-memory; this binary is just the
// I/O wrapper.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"brokle/internal/seeder"
)

const tsFileRel = "web/src/generated/permissions.ts"

func main() {
	root, err := repoRoot()
	if err != nil {
		log.Fatalf("locate repo root: %v", err)
	}

	yamlPath := filepath.Join(root, "seeds", "permissions.yaml")
	body, err := os.ReadFile(yamlPath)
	if err != nil {
		log.Fatalf("read %s: %v", yamlPath, err)
	}

	var doc seeder.PermissionsFile
	if err := yaml.Unmarshal(body, &doc); err != nil {
		log.Fatalf("parse %s: %v", yamlPath, err)
	}

	out := seeder.RenderFrontendPermissions(doc.Permissions)

	target := filepath.Join(root, tsFileRel)
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		log.Fatalf("mkdir %s: %v", filepath.Dir(target), err)
	}
	if err := os.WriteFile(target, out, 0o644); err != nil {
		log.Fatalf("write %s: %v", target, err)
	}
	fmt.Printf("✅ wrote %s (%d permissions)\n", tsFileRel, len(doc.Permissions))
}

// repoRoot walks upward from the working directory until go.mod is found.
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for i := 0; i < 10; i++ {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("go.mod not found from %s", dir)
}
