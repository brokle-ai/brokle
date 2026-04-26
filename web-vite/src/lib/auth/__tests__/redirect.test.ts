import { describe, it, expect, beforeAll } from 'vitest'
import { parseRedirectTo } from '../redirect'

describe('parseRedirectTo', () => {
  beforeAll(() => {
    // jsdom defaults to http://localhost; pin it so the same-origin
    // assertions are stable.
    if (window.location.origin !== 'http://localhost:3000') {
      Object.defineProperty(window, 'location', {
        configurable: true,
        value: {
          ...window.location,
          origin: 'http://localhost:3000',
          hostname: 'localhost',
          port: '3000',
          protocol: 'http:',
        },
      })
    }
  })

  it('returns the fallback for empty / nullish input', () => {
    expect(parseRedirectTo(undefined)).toEqual({ to: '/', search: {} })
    expect(parseRedirectTo(null)).toEqual({ to: '/', search: {} })
    expect(parseRedirectTo('')).toEqual({ to: '/', search: {} })
    expect(parseRedirectTo('   ')).toEqual({ to: '/', search: {} })
  })

  it('parses plain pathnames without search', () => {
    expect(parseRedirectTo('/')).toEqual({ to: '/', search: {} })
    expect(parseRedirectTo('/o/A/p/X')).toEqual({ to: '/o/A/p/X', search: {} })
  })

  it('splits pathname and search params', () => {
    expect(parseRedirectTo('/accept-invite?token=abc')).toEqual({
      to: '/accept-invite',
      search: { token: 'abc' },
    })
  })

  it('decodes URL-encoded params', () => {
    expect(
      parseRedirectTo('/?session=expired&redirect=%2Fo%2FA'),
    ).toEqual({
      to: '/',
      search: { session: 'expired', redirect: '/o/A' },
    })
  })

  it('preserves hash fragments', () => {
    expect(parseRedirectTo('/o/A/p/X#tab=overview')).toEqual({
      to: '/o/A/p/X',
      search: {},
      hash: 'tab=overview',
    })
  })

  it('rejects external absolute URLs (open-redirect guard)', () => {
    expect(parseRedirectTo('https://evil.com/phish')).toEqual({
      to: '/',
      search: {},
    })
  })

  it('rejects protocol-relative URLs', () => {
    expect(parseRedirectTo('//evil.com/phish')).toEqual({
      to: '/',
      search: {},
    })
  })

  it('rejects javascript: scheme', () => {
    // Either the URL parser rejects it or it parses to a non-http
    // origin; either way we fall back.
    expect(parseRedirectTo('javascript:alert(1)')).toEqual({
      to: '/',
      search: {},
    })
  })

  it('falls back on malformed URLs', () => {
    expect(parseRedirectTo('https://%ZZ')).toEqual({
      to: '/',
      search: {},
    })
  })

  it('uses the provided fallback', () => {
    expect(parseRedirectTo(undefined, '/dashboard')).toEqual({
      to: '/dashboard',
      search: {},
    })
    expect(parseRedirectTo('https://evil.com', '/dashboard')).toEqual({
      to: '/dashboard',
      search: {},
    })
  })
})
