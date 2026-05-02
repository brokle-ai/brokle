#!/usr/bin/env bash
# lint-conventions.sh — enforce project conventions that golangci-lint
# cannot express (DDL canonical forms, sqlc named params, migration
# hygiene). Every rule here has a documented reason in CLAUDE.md.
#
# Run locally: make lint-conventions
# Run in CI:   part of `make lint`

set -euo pipefail

fail=0

header() { printf '\n━━ %s ━━\n' "$1"; }

check() {
  local desc="$1" pattern="$2" paths="$3" exclude="${4:-}"
  local cmd=(grep -rEn --color=always "$pattern" $paths)
  if [ -n "$exclude" ]; then
    cmd+=(--exclude-dir="$exclude")
  fi
  local hits
  hits="$("${cmd[@]}" 2>/dev/null || true)"
  if [ -n "$hits" ]; then
    printf '\033[31m✗\033[0m %s\n%s\n\n' "$desc" "$hits"
    fail=1
  else
    printf '\033[32m✓\033[0m %s\n' "$desc"
  fi
}

header "Postgres DDL canonical forms (CLAUDE.md Mandatory Rules)"
# TIMESTAMPTZ / TIMESTAMP shorthand → sqlc override fails silently.
check "no TIMESTAMPTZ (use TIMESTAMP WITH TIME ZONE)" \
  '\bTIMESTAMPTZ\b' \
  'migrations/postgres/'
# DECIMAL without pg_catalog match → sqlc emits pgtype.Numeric.
check "no bare DECIMAL (use NUMERIC)" \
  '\bDECIMAL\s*\(' \
  'migrations/postgres/'
# INT shorthand → inconsistent with INTEGER canonical form.
check "no bare INT column type (use INTEGER)" \
  '^\s*[a-z_]+\s+INT(\s|,|$)' \
  'migrations/postgres/'
# BOOL shorthand.
check "no bare BOOL column type (use BOOLEAN)" \
  '^\s*[a-z_]+\s+BOOL(\s|,|$)' \
  'migrations/postgres/'

header "chi Mount invariant (one r.Route per prefix per tree)"
# chi's Router.Route internally calls Mount(pattern, subRouter). Mount
# panics when the same exact pattern is Mounted twice on the same
# routing tree. Two sibling chi.Group sub-routers share their parent
# mux's routing tree, so two RegisterRoutes functions each calling
# r.Route("/api/v1/auth", ...) inside separate groups collide at boot.
#
# This check flags any /api/v1 or /v1 prefix that appears as the first
# argument of r.Route in more than one handler file. The only
# legitimate duplicate across files is when comments reference the
# pattern — those are stripped by checking for `r.Route(` syntactically.
#
# Root cause of the 2026-04-24 boot panic; see CLAUDE.md Transport
# gotcha on duplicate Mounts.
# Only flag actual code — skip lines that start with // (single-line
# comment) or * (continuation of block comment), plus blank-prefixed
# comment lines. An exact match on the r.Route code form after any
# amount of whitespace means the call is real, not a reference in docs.
mount_hits="$(grep -rEn --color=never '^\s*r\.Route\("(/api/v1|/v1)[^"]*"' \
  internal/transport/http/handlers/ --include='*.go' --exclude='*_test.go' 2>/dev/null || true)"
if [ -n "$mount_hits" ]; then
  # Extract the pattern from each line, count duplicates across files.
  dup_patterns="$(printf '%s\n' "$mount_hits" \
    | grep -oE 'r\.Route\("[^"]+"' \
    | sed 's/^r\.Route("//; s/"$//' \
    | sort | uniq -c \
    | awk '$1 > 1 { print $2 }')"
  if [ -n "$dup_patterns" ]; then
    printf '\033[31m✗\033[0m duplicate r.Route prefix across handler files:\n'
    while IFS= read -r p; do
      printf '  %s\n' "$p"
      printf '%s\n' "$mount_hits" | grep -F "r.Route(\"$p\"" | sed 's/^/    /'
    done <<< "$dup_patterns"
    printf '  chi.Mount will panic at boot — consolidate into one r.Route\n'
    printf '  with inner r.Group blocks for middleware posture scoping.\n\n'
    fail=1
  else
    printf '\033[32m✓\033[0m chi.Route prefix uniqueness\n'
  fi
else
  printf '\033[32m✓\033[0m chi.Route prefix uniqueness (no handler mounts found)\n'
fi

header "slog two-layer message convention (CLAUDE.md)"
# Service layer: lowercase-first (mirrors error strings — Go convention
# that error.Error() begins lowercase). Transport, workers, bootstrap:
# Capital-first (operational events, not error strings).
#
# Acronyms exempt from both rules: gRPC (canonical lowercase-g), plus
# the usual HTTP/JWT/OTLP/SSE/CSRF — but those start uppercase anyway,
# so only gRPC needs the explicit allowlist below.
#
# Two checks: (1) services must NOT start Capital, (2) non-service
# layers must NOT start lowercase (excluding gRPC + sse acronyms).

# (1) Service-layer Capital-first violations.
svc_caps="$(grep -rEn --include='*.go' \
  'logger\.(Error|Warn|Info|Debug)(Context)?\("[A-Z]' \
  internal/core/services/ 2>/dev/null || true)"
if [ -n "$svc_caps" ]; then
  printf '\033[31m✗\033[0m service-layer slog messages must start lowercase\n%s\n\n' "$svc_caps"
  fail=1
else
  printf '\033[32m✓\033[0m service-layer slog lowercase-first\n'
fi

# (2) Transport / worker / bootstrap lowercase-first violations.
# Allowlist: gRPC (lowercase-g acronym) and sse/SSE.
non_svc_lc="$(grep -rEn --include='*.go' \
  'logger\.(Error|Warn|Info|Debug)(Context)?\("[a-z]' \
  internal/transport/ internal/workers/ internal/app/ internal/server/ 2>/dev/null \
  | grep -vE 'logger\.(Error|Warn|Info|Debug)(Context)?\("(gRPC|sse) ' || true)"
if [ -n "$non_svc_lc" ]; then
  printf '\033[31m✗\033[0m transport/worker/bootstrap slog messages must start Capital (acronyms exempt)\n%s\n\n' "$non_svc_lc"
  fail=1
else
  printf '\033[32m✓\033[0m transport/worker/bootstrap slog Capital-first\n'
fi

header "Domain packages must not ship errors.go (centralisation per pkg/errors)"
# All failure classification flows through pkg/errors constructors +
# predicates (`appErrors.NotFound("xxx")` + `appErrors.IsNotFound(err)`).
# Domain packages ship ZERO errors.go files. The 2026-04-30 sweep
# deleted all 11 historical files and migrated their sentinels to
# typed errors at the repository boundary. The k8s / Cockroach /
# Boundary precedent says "no per-domain sentinels"; this lint
# enforces the rule structurally so the files can't be re-created.
domain_errors_files=$(find internal/core/domain -name 'errors.go' 2>/dev/null || true)
if [ -n "$domain_errors_files" ]; then
  printf '\033[31m✗\033[0m domain errors.go files present (centralise via pkg/errors NotFound/IsNotFound)\n%s\n\n' "$domain_errors_files"
  fail=1
else
  printf '\033[32m✓\033[0m no per-domain errors.go files\n'
fi

header "Repository nil-slice trap (CLAUDE.md gotcha #23)"
# Go's encoding/json serialises nil slices as `null`. List-bearing
# repository methods that declare `var x []T` and conditionally
# append produce `data: null` for empty result sets, violating
# the wire contract (`data: []` always) and crashing typed-as-array
# frontend consumers. Use `make([]T, 0[, cap])` at construction so
# empty result sets serialise as `[]`. Same canonical Go pattern
# as SigNoz's `pkg/transition/v5_to_v4.go:18-24`.
check "no \`var x []T\` in repository files (use \`make([]T, 0)\`)" \
  '^[[:space:]]*var[[:space:]]+[a-zA-Z_][a-zA-Z0-9_]*[[:space:]]+\[\]' \
  'internal/infrastructure/repository/'

header "NotFound / AlreadyExists / Conflict resource-name discipline"
# The first arg of NotFound("X") / AlreadyExists("X") / Conflict("X", ...)
# is an entity NAME — the PublicMessage derives "X not found" / "X already
# exists" / "X conflict" from it. Sentence-style first-args produce
# "X not found not found" on the wire. Use NotFound("user") not
# NotFound("user not found"); when the original phrasing is load-bearing
# attach it via WithMessage(...).
sentence_resource=$(grep -rnE 'appErrors\.(NotFound|AlreadyExists)\("[^"]*( not found| already exists| does not exist| missing)"\)' --include='*.go' internal pkg cmd 2>/dev/null || true)
if [ -n "$sentence_resource" ]; then
  printf '\033[31m✗\033[0m sentence-style resource arg (use single entity name; attach prose via WithMessage)\n%s\n\n' "$sentence_resource"
  fail=1
else
  printf '\033[32m✓\033[0m no sentence-style resource args\n'
fi

header "InvalidParam first-arg field-name discipline"
# InvalidParam(field, message) puts the FIELD NAME first (Stripe
# error.param). Adverbs/adjectives ("invalid", "missing", "required",
# "empty") and type names ("numeric", "categorical", "boolean") are
# NOT field names — they're message fragments accidentally swapped
# into the field slot during the bulk migration. Multi-word first
# args ("too many tags", "score[%d]: project_id is required") are the
# same regression in a different shape.
swap_adverb=$(grep -rnE 'appErrors\.InvalidParam\("(invalid|missing|too|last|numeric|categorical|boolean|required|empty|error|linkage)",' --include='*.go' internal pkg cmd 2>/dev/null || true)
swap_sentence=$(grep -rnE 'appErrors\.InvalidParam\("[^"]* [^"]*"' --include='*.go' internal pkg cmd 2>/dev/null || true)
if [ -n "$swap_adverb" ] || [ -n "$swap_sentence" ]; then
  printf '\033[31m✗\033[0m InvalidParam first arg must be a field name (snake_case identifier), not an adverb/adjective/sentence\n'
  [ -n "$swap_adverb" ] && printf '%s\n' "$swap_adverb"
  [ -n "$swap_sentence" ] && printf '%s\n' "$swap_sentence"
  printf '\n'
  fail=1
else
  printf '\033[32m✓\033[0m InvalidParam first arg discipline\n'
fi

header "InvalidParam state-word smell"
# InvalidParam's first arg is a FIELD name (Stripe error.param).
# State-machine verbs like "cannot" / "already" / "archived" / "locked"
# are not field names — they're signals that the wrong constructor was
# picked. State-machine rejections belong in Conflict() (HTTP 409),
# not InvalidParam (HTTP 422). The 2026-04-29 review found 4 sites
# using `InvalidParam("cannot", ...)` for archived/inactive/locked
# state errors; the wire response was a 422 with `error.param=cannot`,
# pointing at a non-existent form field.
state_word_param=$(grep -rnE 'appErrors\.InvalidParam\("(cannot|already|archived|locked|deleted|inactive|expired|disabled)"' --include='*.go' internal pkg cmd 2>/dev/null || true)
if [ -n "$state_word_param" ]; then
  printf '\033[31m✗\033[0m InvalidParam used for state-machine rejection (use Conflict instead — HTTP 409, no error.param)\n%s\n\n' "$state_word_param"
  fail=1
else
  printf '\033[32m✓\033[0m no state-word InvalidParam misuse\n'
fi

header "Error constructor field-vs-multi-field discipline"
# pkg/errors exposes TWO HTTP-422 constructors:
#   InvalidParam(field, message, opts...)         → single field, populates Param
#   InvalidFields(details []ErrorDetail, opts...) → multi-field validator output
# A bare `InvalidInput(...)` constructor was deleted on 2026-04-28 — it
# conflated entity-level vs field-level cases. Reintroducing it would
# regress the cleanup. The Reason enum value `ReasonInvalidInput` is
# still valid (Reason isn't a constructor); we only ban the constructor
# call shape `appErrors.InvalidInput(`.
invalid_input=$(grep -rn 'appErrors\.InvalidInput(\|^\s*InvalidInput(' --include='*.go' internal pkg cmd 2>/dev/null || true)
if [ -n "$invalid_input" ]; then
  printf '\033[31m✗\033[0m InvalidInput constructor reintroduced (use InvalidParam for single-field or InvalidFields for validator output)\n%s\n\n' "$invalid_input"
  fail=1
else
  printf '\033[32m✓\033[0m no InvalidInput constructor reintroductions\n'
fi

header "Migration symmetry honesty (down ADD vs up DROP)"
# A `.down.sql` that re-adds a column dropped by its `.up.sql` is
# DISHONESTLY symmetric: PostgreSQL's `ALTER TABLE ADD COLUMN` appends
# at the end, while the original column likely sat in a different
# position — sqlc positional `SELECT *` scans from a pre-change
# binary mis-decode the rolled-back schema. Down migrations for
# column drops should be no-ops with a documented WHY (Stripe
# "Online migrations at scale", Rails strong_migrations precedent:
# "roll forward beats schema rollback" for column drops). See
# 2026-04-30 (rollback-compat / SELECT *) Lessons Learned.
dishonest_down=""
# Forward-only: the lint applies to migrations authored on or after the
# 2026-04-30 (rollback-compat / SELECT *) Lessons Learned cutover.
# Pre-existing dishonestly-symmetric down migrations (17 sites) are
# grandfathered; pre-prod stage means rolling them back has no live
# binary to break, and rewriting them is scope creep beyond the
# immediate reviewer concern. Bump this cutoff if/when those get swept.
migration_lint_cutoff="20260427171335"
for down_file in migrations/postgres/*.down.sql; do
  [ -f "$down_file" ] || continue
  base=$(basename "$down_file")
  ts="${base%%_*}"
  [ "$ts" \< "$migration_lint_cutoff" ] && continue
  up_file="${down_file%.down.sql}.up.sql"
  [ -f "$up_file" ] || continue
  # Strip optional `IF NOT EXISTS` / `IF EXISTS` clauses so the column
  # name is the captured token, not the literal word "IF".
  added=$({ grep -oiE 'ADD COLUMN([[:space:]]+IF[[:space:]]+NOT[[:space:]]+EXISTS)?[[:space:]]+[a-zA-Z_][a-zA-Z0-9_]*' "$down_file" 2>/dev/null || true; } | awk '{print tolower($NF)}' | sort -u)
  dropped=$({ grep -oiE 'DROP COLUMN([[:space:]]+IF[[:space:]]+EXISTS)?[[:space:]]+[a-zA-Z_][a-zA-Z0-9_]*' "$up_file" 2>/dev/null || true; } | awk '{print tolower($NF)}' | sort -u)
  if [ -n "$added" ] && [ -n "$dropped" ]; then
    overlap=$(comm -12 <(printf '%s\n' "$added") <(printf '%s\n' "$dropped") || true)
    if [ -n "$overlap" ]; then
      dishonest_down+="$down_file → re-adds column(s) dropped in up: $(echo "$overlap" | tr '\n' ' ')\n"
    fi
  fi
done
if [ -n "$dishonest_down" ]; then
  printf '\033[31m✗\033[0m down migration re-adds a column dropped by its up migration (column position cannot be restored — make the down a documented no-op)\n%b\n' "$dishonest_down"
  fail=1
else
  printf '\033[32m✓\033[0m no dishonestly-symmetric down migrations\n'
fi

header "sqlc query hygiene"
# `SELECT *` + sqlc's positional row.Scan() is a column-order trap:
# any ALTER TABLE ADD/DROP/REORDER between codegen-time and runtime
# corrupts scans (Postgres ADD COLUMN always appends; the Scan target
# order is frozen at codegen). Industry consensus (Stripe online-
# migrations, sqlc docs, Rails strong_migrations, PlanetScale): use
# explicit column lists. See 2026-04-30 (rollback-compat / SELECT *)
# Lessons Learned for the full survey + decision.
#
# Scoped to the files migrated in 2026-04-30. Codebase-wide sweep is
# tracked as a follow-up; until it lands, the lint enforces no
# regression on the cleaned files. Add files here as they're migrated.
star_lint_files=(
  internal/infrastructure/db/queries/member.sql
  internal/infrastructure/db/queries/project_member.sql
)
star_queries=$(grep -En '^\s*SELECT\s+([a-zA-Z_][a-zA-Z0-9_]*\.)?\*(\s|$)' "${star_lint_files[@]}" 2>/dev/null || true)
if [ -n "$star_queries" ]; then
  printf '\033[31m✗\033[0m SELECT * found in sqlc query (use explicit column list — positional Scan corrupts on column drift)\n%s\n\n' "$star_queries"
  fail=1
else
  printf '\033[32m✓\033[0m no SELECT * in member/project_member sqlc queries\n'
fi

header "Frontend URL-tenancy invariant (gotcha #41 at consumer side)"
# The backend mounts tenant-child resources under the parent URL tree —
# /v1/organizations/{orgId}/projects, /v1/projects/{projectId}/members,
# etc. — and rejects org/project IDs in query/body (CLAUDE.md gotcha
# #41). When an endpoint moves to the org-nested shape and a stale
# frontend helper still calls /v1/<resource> at path root, the request
# silently 404s or 422s. This lint catches that class structurally.
#
# Tenant-child resources whose ROOT-PATH usage is forbidden:
#   - projects (always under /organizations/{orgId}/)
#   - members, invitations (always under /organizations/{orgId}/ or /projects/{projectId}/)
#   - api-keys (always under /projects/{projectId}/)
#
# Other resources (traces, scores, comments, dashboards, datasets,
# evaluations, etc.) have their own URL conventions and aren't covered
# by this guard. Extend the resource list as new endpoints adopt the
# org/project-nested shape.
#
# Pattern matched: any quoted URL fragment of the form
# `'/v1/<resource>` or `'/api/v1/<resource>`, followed by `/`, `'`,
# `"`, `?`, or end-of-string, EXCLUDING comment lines (// or  *) and
# excluding strings that already contain an `organizations/` or
# `projects/<id>/` parent segment.
# True bug shape: list/create at flat /v1/<resource> instead of nested
# /v1/<parent>/<parentId>/<resource>. Single-resource detail at
# /v1/<resource>/<id> is legitimate (the ID disambiguates), so we only
# flag patterns where /v1/<resource> is immediately closed by `'`, `"`,
# `\`, `?`, `&`, or `}` (template-literal close) — i.e., no resource
# ID follows.
tenancy_offenders=$(grep -rEn "(\`|['\"])/?(api/)?v1/(projects|api-keys)(\\?|\`|['\"])" \
  web/src/ sdk/javascript/src/ sdk/python/brokle/ 2>/dev/null \
  | grep -v -E "^[^:]*:[0-9]+: *(\*|//)" \
  || true)
if [ -n "$tenancy_offenders" ]; then
  printf '\033[31m✗\033[0m frontend/SDK calls a tenant-child resource at path root without an organizations/{orgId}/ or projects/{projectId}/ parent — see CLAUDE.md gotcha #41\n%s\n\n' "$tenancy_offenders"
  fail=1
else
  printf '\033[32m✓\033[0m frontend/SDK URL-tenancy invariant\n'
fi

header "Go conventions (complementing forbidigo)"
# Migrations must go through CLI — hand-written files get silently ignored.
# This catches files named without the framework's timestamp prefix.
if [ -d migrations/postgres ]; then
  bad_files="$(find migrations/postgres -maxdepth 1 -type f -name '*.sql' -regextype posix-extended ! -regex '.*/[0-9]{14}_[a-z0-9_]+\.(up|down)\.sql$' 2>/dev/null || true)"
  if [ -n "$bad_files" ]; then
    printf '\033[31m✗\033[0m migration files without CLI-generated timestamp prefix\n%s\n\n' "$bad_files"
    fail=1
  else
    printf '\033[32m✓\033[0m migration filename convention\n'
  fi
fi

if [ "$fail" -ne 0 ]; then
  printf '\n\033[31mconvention lint failed\033[0m — fix violations above.\n'
  exit 1
fi
printf '\n\033[32mall conventions satisfied\033[0m\n'
