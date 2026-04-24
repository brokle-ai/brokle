// CSRF double-submit: backend sets a non-httpOnly `csrf_token` cookie
// that the client reads at request time and echoes back in the
// `X-CSRF-Token` header on mutating methods (CLAUDE.md gotcha #23).
// Read per-request, NOT cached — the cookie rotates on refresh.

const COOKIE_NAME = 'csrf_token'

export function readCsrfCookie(): string | null {
  if (typeof document === 'undefined') return null
  const needle = `${COOKIE_NAME}=`
  for (const raw of document.cookie.split(';')) {
    const trimmed = raw.trim()
    if (trimmed.startsWith(needle)) {
      return decodeURIComponent(trimmed.slice(needle.length)) || null
    }
  }
  return null
}

export const MUTATION_METHODS = new Set(['POST', 'PUT', 'PATCH', 'DELETE'])
