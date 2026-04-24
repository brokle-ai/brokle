// Top-level barrel for the auth feature. Cross-feature callers (traces
// annotation + comments drawers) import the current-user query hook
// from here. The web/ convention places this under
// `@/features/authentication` — web-vite collapses that to `auth/` but
// exports the same public surface.

export { useCurrentUser } from './hooks'
export { currentUserQueryOptions } from './queries'
