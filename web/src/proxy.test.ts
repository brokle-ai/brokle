import { readdirSync } from 'node:fs'
import { join } from 'node:path'
import { describe, it, expect } from 'vitest'

import { config } from './proxy'

// Drift-prevention tests for the Next.js 16 proxy route guard.
//
// The matcher in proxy.ts is a hand-maintained negative regex listing
// every URL segment a logged-out user is allowed to reach. Historically
// it drifted out of sync with the app/(auth)/ directory (matcher had
// `accept-invitation` while the real route was `accept-invite`, and
// `callback` was missing entirely), which silently broke OAuth callback
// and invitation acceptance for every logged-out user.
//
// These tests assert the invariant: every directory under app/(auth)/
// is listed in the matcher's exclusion set. Any rename or addition
// without a corresponding matcher edit fails this spec.
describe('proxy matcher', () => {
  const matcher = config.matcher[0]

  it('excludes every route directory under app/(auth)/', () => {
    const authDir = join(process.cwd(), 'src/app/(auth)')
    const routes = readdirSync(authDir, { withFileTypes: true })
      .filter(d => d.isDirectory())
      .map(d => d.name)

    expect(routes.length).toBeGreaterThan(0)
    for (const route of routes) {
      // `\b` treats `-` as a word boundary, so `\baccept-invite\b`
      // matches `accept-invite` as a whole token and rejects the
      // partial-match pitfall (e.g. `accept` alone).
      expect(matcher).toMatch(new RegExp(`\\b${route}\\b`))
    }
  })

  it('distinguishes protected from public paths', () => {
    // The matcher string IS a JS regex — Next.js compiles it with
    // the same runtime. Exercising it directly here catches any
    // syntax regression as well as the route-level logic.
    const pattern = new RegExp(`^${matcher}$`)

    expect('/dashboard').toMatch(pattern)
    expect('/projects/x/traces').toMatch(pattern)
    expect('/settings').toMatch(pattern)

    // Public auth entry points must NOT match — proxy skipped.
    expect('/signin').not.toMatch(pattern)
    expect('/signup').not.toMatch(pattern)
    expect('/callback').not.toMatch(pattern)
    expect('/accept-invite').not.toMatch(pattern)
    expect('/forgot-password').not.toMatch(pattern)
    expect('/verify-email').not.toMatch(pattern)

    // Framework + asset paths must NOT match.
    expect('/api/v1/users/me').not.toMatch(pattern)
    expect('/v1/traces').not.toMatch(pattern)
    expect('/_next/static/chunks/foo.js').not.toMatch(pattern)
    expect('/favicon.ico').not.toMatch(pattern)
    expect('/logo.svg').not.toMatch(pattern)
  })
})
