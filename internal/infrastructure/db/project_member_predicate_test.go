package db

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestProjectMemberQueriesEnforceActiveOrgPredicate is a drift guard: every
// :one / :many read query in queries/project_member.sql that touches
// project_members MUST either JOIN organization_members with
// `om.deleted_at IS NULL` (the active-org filter) or carry an explicit
// `-- noqa: orphan-ok` directive on the line above the `-- name:` header.
//
// Rationale (CLAUDE.md 2026-04-30 (project-rbac)): the project_members
// table participates in a soft-delete cascade hierarchy (organization_members
// is the parent). All reads used for authz / identity / membership-check
// must apply the same filter as the listings + permission resolver, or
// orphan rows leak across read paths and create the asymmetry bug class
// (re-add fails with Conflict despite listings hiding the row).
//
// Six query shapes participate:
//
//	ListProjectMembersByProject       — :many — must JOIN
//	ListProjectMembersByUser          — :many — must JOIN
//	GetProjectMemberByUserAndProject  — :one  — must JOIN
//	IsProjectMember                   — :one  — must JOIN (subquery)
//	GetProjectMemberRoleID            — :one  — must JOIN
//	ListUserEffectivePermissionsInScope — :many — uses a CTE form; tagged noqa
//
// The CountProjectMembersByProject query (a simple COUNT(*)) is exempt
// because it returns a row count that callers never use for authz; it's
// tagged `-- noqa: orphan-ok`.
func TestProjectMemberQueriesEnforceActiveOrgPredicate(t *testing.T) {
	t.Parallel()

	path := filepath.Join("..", "..", "..", "internal", "infrastructure", "db", "queries", "project_member.sql")
	// The above path is from the test's working dir (this file's package
	// dir, internal/infrastructure/db). Resolve relative to repo root via
	// a simpler path now that we're in the same parent.
	path = "queries/project_member.sql"
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read project_member.sql: %v", err)
	}

	queries := splitSqlcQueries(string(body))
	if len(queries) == 0 {
		t.Fatal("no queries parsed from project_member.sql — file path or format change?")
	}

	// Active-org JOIN signature (any of these two-line pairs suffices).
	// We require the literal `om.deleted_at IS NULL` token because that's
	// the structural invariant — JOINs without the deleted_at gate would
	// still leak orphans.
	activeOrgJoin := regexp.MustCompile(`(?s)organization_members\s+om\b.*?om\.deleted_at\s+IS\s+NULL`)
	noqa := regexp.MustCompile(`(?m)^--\s*noqa:\s*orphan-ok\b`)

	for _, q := range queries {
		// Only check read queries that touch project_members.
		if q.kind != ":one" && q.kind != ":many" {
			continue
		}
		if !strings.Contains(q.body, "project_members") {
			continue
		}

		hasJoin := activeOrgJoin.MatchString(q.body)
		hasNoqa := noqa.MatchString(q.body)

		if !hasJoin && !hasNoqa {
			t.Errorf("query %q (%s) reads project_members without active-org JOIN and without `-- noqa: orphan-ok` directive\nbody:\n%s",
				q.name, q.kind, q.body)
		}
	}
}

type sqlcQuery struct {
	name string // e.g. "ListProjectMembersByProject"
	kind string // e.g. ":many"
	body string // SQL up to next `-- name:` or EOF
}

// splitSqlcQueries parses a sqlc query file into individual queries.
// The header format is `-- name: <Name> :<kind>`. Comments and blank
// lines between the header and the SQL body are part of that query's
// body for the purpose of `-- noqa:` discovery.
func splitSqlcQueries(s string) []sqlcQuery {
	header := regexp.MustCompile(`(?m)^--\s*name:\s*(\w+)\s+(:\w+)\s*$`)
	matches := header.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		return nil
	}

	out := make([]sqlcQuery, 0, len(matches))
	for i, m := range matches {
		end := len(s)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		// Include the line above the header (so `-- noqa:` directives
		// directly preceding `-- name:` are discoverable).
		start := m[0]
		if start > 0 {
			// back up past the previous newline to include the prior line
			prev := strings.LastIndexByte(s[:start-1], '\n')
			if prev >= 0 {
				start = prev + 1
			}
		}
		out = append(out, sqlcQuery{
			name: s[m[2]:m[3]],
			kind: s[m[4]:m[5]],
			body: s[start:end],
		})
	}
	return out
}
