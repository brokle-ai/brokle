// Server-side backend URL resolution — single source of truth.
//
// Precedence:
//
//   1. BROKLE_API_PROXY_TARGET — explicit server-only override for
//      deployments where the browser and the Next.js process see
//      different backend origins. Example: dashboard at
//      https://app.brokle.com fronted by a load balancer, backend
//      only reachable inside a VPC at http://backend.internal:8080.
//      Server code uses the internal URL; the browser continues to
//      use the public URL through same-origin rewrites.
//
//   2. NEXT_PUBLIC_API_URL — the documented public API URL from
//      .env.example. Reused server-side because in the common
//      single-host deployment the browser and the Next.js process
//      share one reachable backend origin. Must be a full origin
//      (e.g. https://api.brokle.com) — no prefix stripping.
//
//   3. http://localhost:8080 — local `make dev` default.
//
// Exported as a function rather than a module-level constant so:
//   - tests can use vi.stubEnv() and see the change per call.
//   - Next.js runtime env changes (standalone builds that load env
//     after module evaluation) are observed correctly.
export function getBackendBaseURL(): string {
  // The `||` chain also collapses empty strings to "unset", which
  // is important because the web Dockerfile bakes
  // `ENV NEXT_PUBLIC_API_URL=${NEXT_PUBLIC_API_URL}` and an absent
  // build ARG materialises as `""` at runtime.
  return (
    process.env.BROKLE_API_PROXY_TARGET ||
    process.env.NEXT_PUBLIC_API_URL ||
    'http://localhost:8080'
  )
}
