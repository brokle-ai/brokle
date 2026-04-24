// Content-Security-Policy used by vite-plugin-csp-guard at build time
// to emit a hashed meta tag + an nginx snippet. Starts in Report-Only
// mode for 2 weeks post-deploy; flip to enforcing once the false-
// positive curve flattens.
//
// Principles:
//   - 'strict-dynamic' transitively trusts anything the bootstrap
//     hashes load, which covers Vite's dynamic import() chunks cleanly.
//   - 'unsafe-inline' on script-src is overridden by the presence of
//     'strict-dynamic' + hashes in CSP3-capable browsers.
//   - 'unsafe-inline' on style-src is pragmatic — Tailwind + Radix + CM6
//     inject inline styles that hashes cannot anchor without a per-file
//     plugin refactor. Revisit in Phase 2 if Trusted Types adoption
//     lands.
//   - report-to endpoint is a Go handler on the main backend at
//     /api/v1/_csp-report (to be added in Phase 1.6 follow-up).

// Locally-typed CSP policy — the plugin's preferred `CSPPolicy` type
// lives in the transitive `csp-toolkit` package which isn't re-exported.
// This shape is what the plugin's `policy` option accepts: each
// directive name mapped to an array of sources (strings).
type PolicyDirectives =
  | 'default-src'
  | 'script-src'
  | 'style-src'
  | 'img-src'
  | 'font-src'
  | 'connect-src'
  | 'frame-ancestors'
  | 'base-uri'
  | 'form-action'
  | 'object-src'

export const cspPolicy: Record<PolicyDirectives, string[]> = {
  'default-src': ["'self'"],
  'script-src': ["'self'", "'strict-dynamic'", "'unsafe-inline'"],
  'style-src': ["'self'", "'unsafe-inline'"],
  'img-src': ["'self'", 'data:', 'https:'],
  'font-src': ["'self'", 'data:'],
  'connect-src': ["'self'", 'https://*.sentry.io'],
  'frame-ancestors': ["'none'"],
  'base-uri': ["'none'"],
  'form-action': ["'self'"],
  'object-src': ["'none'"],
}
