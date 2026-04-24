import { authHandlers } from './auth'
import { organizationHandlers } from './organizations'

// Single flat array consumed by both setupServer (Vitest) and setupWorker
// (dev + Playwright). Import per-feature handler modules above so the
// feature owns its mocks; the index just re-exports a flat array for the
// test runners.
export const handlers = [...authHandlers, ...organizationHandlers]
