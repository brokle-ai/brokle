import { authHandlers } from './auth'
import { organizationHandlers } from './organizations'

// Single flat array consumed by setupServer (Vitest) and setupWorker
// (dev + Playwright). Feature modules own their handlers; this index
// re-exports the flat array the runners want.
export const handlers = [...authHandlers, ...organizationHandlers]
