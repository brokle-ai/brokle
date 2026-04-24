# Security header suite sourced by nginx.conf. Kept separate so Phase
# 1.6 owns the CSP + header policy independently of the main server
# block. See plan §Security posture for the reasoning behind each.

add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
add_header X-Content-Type-Options "nosniff" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "camera=(), microphone=(), geolocation=(), payment=()" always;
add_header Cross-Origin-Opener-Policy "same-origin" always;
add_header Cross-Origin-Resource-Policy "same-origin" always;

# CSP: start in Report-Only for 2 weeks post-deploy, fix false positives
# against /api/v1/_csp-report reports, then flip to enforcing by
# swapping to `Content-Security-Policy`. Hash-based — vite-plugin-csp-
# guard emits the script hashes into the built index.html at build time
# (meta http-equiv), and this header is the transport-level duplicate
# that also carries the report-to directive.
#
# The policy list here must stay in lockstep with src/csp-policy.ts.
add_header Content-Security-Policy-Report-Only "default-src 'self'; script-src 'self' 'strict-dynamic' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https://*.sentry.io; frame-ancestors 'none'; base-uri 'none'; form-action 'self'; object-src 'none'; report-to csp-endpoint" always;

# Reporting API endpoint. Browsers POST CSP violation reports to the
# first URL in `endpoints`. The backend handler is at Phase-1.6 TODO
# — add a lightweight Go handler at /api/v1/_csp-report that ingests
# the report body and forwards to Sentry or structured-logs.
add_header Report-To '{"group":"csp-endpoint","max_age":10886400,"endpoints":[{"url":"/api/v1/_csp-report"}]}' always;
