// Runtime config, written by the container entrypoint (see
// public/config.js.template + docker-entrypoint.d/). NEVER read per-env
// values from import.meta.env.VITE_* — those are static-replaced at
// build time and would burn production URLs into the bundle, breaking
// the 12-factor "one image, N environments" guarantee.
//
// VITE_* still has a role: true build-time constants like commit SHA
// and app name. Those are read via import.meta.env.VITE_*.

declare global {
  interface Window {
    __CONFIG__?: RuntimeConfig
  }
}

export interface RuntimeConfig {
  readonly API_URL: string
  readonly SENTRY_DSN: string
  readonly POSTHOG_KEY: string
  readonly APP_ENV: 'development' | 'staging' | 'production' | string
  readonly COMMIT_SHA: string
}

const DEFAULTS: RuntimeConfig = Object.freeze({
  API_URL: '',
  SENTRY_DSN: '',
  POSTHOG_KEY: '',
  APP_ENV: 'development',
  COMMIT_SHA: 'dev',
})

let cached: RuntimeConfig | null = null

export function getRuntimeConfig(): RuntimeConfig {
  if (cached) return cached
  const injected = typeof window !== 'undefined' ? window.__CONFIG__ : undefined
  cached = Object.freeze({ ...DEFAULTS, ...(injected ?? {}) })
  return cached
}

// Build-time constants (inlined at `vite build`). Kept separate from
// runtime config on purpose — don't mix the two. Caller reaches for
// these when truly static across all deploys of the same build.
export const BUILD_INFO = Object.freeze({
  APP_NAME: 'Brokle Dashboard',
})
