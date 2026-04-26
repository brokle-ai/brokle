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
  internal/transport/http/handlers/ --include='*.go' 2>/dev/null || true)"
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

header "Domain sentinel orphan audit (CLAUDE.md producer/consumer parity)"
# Every Err* sentinel declared in internal/core/domain/<x>/errors.go
# must have at least one producer or consumer outside its own
# directory. A sentinel with no external references and no internal
# uses beyond its own declaration line is dead code: errors.Is
# against it will silently miss whatever the producer was supposed
# to wrap. Delete on sight.
orphans=""
# Disable -e for this audit — internal arithmetic comparisons return
# non-zero on false, which would trip the trap.
set +e
for f in internal/core/domain/*/errors.go; do
  [ -f "$f" ] || continue
  domain_dir=$(dirname "$f")
  for sentinel in $(grep -E '^\s+Err[A-Z][A-Za-z]+\s+=\s*errors\.New' "$f" | awk '{print $1}' | sort -u); do
    external=$(grep -rEln --include='*.go' "\b$sentinel\b" internal/ 2>/dev/null \
      | grep -v "^$domain_dir/" | wc -l)
    self=$(grep -cE "\b$sentinel\b" "$f")
    if [ "$external" -eq 0 ] && [ "$self" -le 1 ]; then
      orphans="${orphans}${domain_dir}/errors.go: ${sentinel}"$'\n'
    fi
  done
done
set -e
if [ -n "$orphans" ]; then
  printf '\033[31m✗\033[0m orphan domain sentinels (no producer or consumer)\n%s\n' "$orphans"
  fail=1
else
  printf '\033[32m✓\033[0m no orphan domain sentinels\n'
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
