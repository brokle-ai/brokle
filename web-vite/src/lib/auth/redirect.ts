// Parse a redirect URL string (typically from a `?redirect=...` query
// param) into the shape TanStack Router's `navigate` expects.
//
// Why a helper:
//   - TanStack Router's `to` option is a pathname only; embedding a
//     query string ("/path?token=…") makes the router treat the
//     whole literal as a pathname and no route matches.
//   - User-controlled redirect strings are an open-redirect vector
//     (OWASP). We reject any URL that resolves to a different origin
//     and fall back to a safe default.
//
// Usage:
//   await navigate({ ...parseRedirectTo(redirect), replace: true })
export function parseRedirectTo(
  redirectTo: string | undefined | null,
  fallback: string = '/',
): { to: string; search: Record<string, unknown>; hash?: string } {
  const raw = redirectTo?.trim()
  if (!raw) return { to: fallback, search: {} }

  let url: URL
  try {
    url = new URL(raw, window.location.origin)
  } catch {
    return { to: fallback, search: {} }
  }

  // Open-redirect guard: only accept same-origin URLs. External
  // origins fall back to the safe default.
  if (url.origin !== window.location.origin) {
    return { to: fallback, search: {} }
  }

  const search: Record<string, unknown> = {}
  url.searchParams.forEach((value, key) => {
    search[key] = value
  })

  const hash = url.hash ? url.hash.slice(1) : undefined
  return hash
    ? { to: url.pathname, search, hash }
    : { to: url.pathname, search }
}
