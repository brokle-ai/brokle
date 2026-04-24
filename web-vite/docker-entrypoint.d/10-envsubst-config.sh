#!/bin/sh
# Render /config.js from its template at container boot. Every BROKLE_*
# env var supplied by the k8s ConfigMap / Deployment becomes a runtime
# value accessible to the SPA via window.__CONFIG__. See src/lib/config.ts.
set -eu

TEMPLATE=/etc/nginx/templates/config.js.template
OUT=/usr/share/nginx/html/config.js

if [ ! -f "$TEMPLATE" ]; then
  echo "config.js.template missing; skipping runtime config generation" >&2
  exit 0
fi

# Pin the allowed variable set — envsubst with no args would substitute
# every env var in the template, so a stray $PATH inside a template
# comment would leak. Whitelist the ones we expect.
envsubst '${BROKLE_API_URL} ${BROKLE_SENTRY_DSN} ${BROKLE_POSTHOG_KEY} ${BROKLE_APP_ENV} ${BROKLE_COMMIT_SHA}' \
  < "$TEMPLATE" > "$OUT"

echo "config.js rendered from template at $OUT"
