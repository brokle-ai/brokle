import { getRuntimeConfig } from '@/lib/config'

// Single-flight refresh coordinator. Combines a tab-local Promise
// singleton with navigator.locks so N concurrent 401s in one tab
// produce ONE refresh call, and concurrent refresh attempts across
// tabs are serialised by the Web Lock — preventing Shopware-class
// token-reuse race conditions.
//
// Contract:
//   - Resolves when the refresh HTTP call succeeds (cookies rotated).
//   - Rejects with an Error if the refresh call fails (callers should
//     treat this as "session expired, redirect to /signin").
//   - Never throws synchronously.
//
// Why navigator.locks (not SharedWorker / BroadcastChannel): native,
// zero deps, ships in every Safari ≥ 15.4, Chrome, Firefox, Edge,
// serialises across tabs of the same origin.

const LOCK_NAME = 'brokle-auth-refresh'

let inflight: Promise<void> | null = null

export async function refreshWithLock(): Promise<void> {
  if (inflight) return inflight
  inflight = doRefresh().finally(() => {
    inflight = null
  })
  return inflight
}

async function doRefresh(): Promise<void> {
  const cfg = getRuntimeConfig()
  const url = `${cfg.API_URL}/api/v1/auth/refresh`

  // Fallback for environments without navigator.locks (tests, older
  // browsers). The in-tab promise singleton still guards against
  // concurrent refresh within the same tab; cross-tab ordering is
  // best-effort in the fallback path.
  if (typeof navigator === 'undefined' || !navigator.locks?.request) {
    await postRefresh(url)
    return
  }

  await navigator.locks.request(LOCK_NAME, { mode: 'exclusive' }, async () => {
    await postRefresh(url)
  })
}

async function postRefresh(url: string): Promise<void> {
  const resp = await fetch(url, {
    method: 'POST',
    credentials: 'include',
    headers: { 'Content-Type': 'application/json' },
  })
  if (!resp.ok) {
    throw new Error(`auth refresh failed: ${resp.status} ${resp.statusText}`)
  }
}
