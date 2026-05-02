#!/usr/bin/env bash
# Generate TypeScript types from the backend's runtime OpenAPI 3.1 specs.
#
# Brokle's backend (Huma) emits two surfaces:
#   - Dashboard plane: /api/v1/openapi.json  (cookie + JWT auth)
#   - SDK plane:       /v1/openapi.json      (X-API-Key auth)
#
# We emit one TS file per surface so the frontend's openapi-fetch client
# can pin the right `paths` namespace per request group. Re-run any time
# the backend OpenAPI shape changes; CI diffs the output to surface drift.
#
# Requires: a running backend (default: http://localhost:8080). Override
# with BROKLE_API_URL. Requires pnpm + `openapi-typescript` (added as a
# devDependency of web/).

set -euo pipefail

API_URL="${BROKLE_API_URL:-http://localhost:8080}"
OUT_DIR="${OPENAPI_OUT_DIR:-web/src/lib/api/generated}"

mkdir -p "$OUT_DIR"

fetch() {
  local path="$1"
  local out="$2"
  echo "↓ $API_URL$path → $out"
  if ! curl -sSf --max-time 10 "$API_URL$path" -o "$out.json"; then
    echo "✗ failed to fetch $API_URL$path — is the backend running?" >&2
    exit 1
  fi
}

fetch /api/v1/openapi.json "$OUT_DIR/dashboard"
fetch /v1/openapi.json     "$OUT_DIR/sdk"

cd web
pnpm exec openapi-typescript src/lib/api/generated/dashboard.json \
  -o src/lib/api/generated/dashboard.d.ts
pnpm exec openapi-typescript src/lib/api/generated/sdk.json \
  -o src/lib/api/generated/sdk.d.ts

echo "✓ types regenerated in web/src/lib/api/generated/"
