import type { NextConfig } from "next";

// apiProxyTarget is the upstream the Next.js dev server forwards
// /api/v1/* and /v1/* to via rewrites below.
//
// Precedence — mirrors src/lib/env/backend-url.ts exactly (inlined
// here because next.config.ts runs before the app bundle is
// available):
//
//   1. BROKLE_API_PROXY_TARGET — explicit server-only override for
//      split-host deployments (browser sees a public URL via load
//      balancer; the Next.js process reaches a private VPC URL).
//   2. NEXT_PUBLIC_API_URL — the documented public API URL from
//      .env.example. Reused here so single-host deployments work
//      with exactly one env var.
//   3. http://localhost:8080 — local `make dev` default.
//
// Keep these three in lockstep with backend-url.ts. Drift here
// silently breaks server-side fetches in deployments configured
// only via NEXT_PUBLIC_API_URL (the .env.example path).
const apiProxyTarget =
  process.env.BROKLE_API_PROXY_TARGET ||
  process.env.NEXT_PUBLIC_API_URL ||
  'http://localhost:8080';

const nextConfig: NextConfig = {
  /* config options here */
  output: 'standalone',
  typescript: {
    // !! WARN !!
    // Dangerously allow production builds to successfully complete even if
    // your project has type errors.
    ignoreBuildErrors: true,
  },

  // Same-origin proxy: the browser always sees /api/v1/* and /v1/*
  // as paths on the dashboard origin (e.g. http://localhost:3000).
  // Next.js forwards them to the Go backend transparently.
  //
  // Why: eliminates CORS preflight, CSRF allowlist gymnastics,
  // SameSite cookie surprises, and cookie-domain workarounds in dev.
  // Matches production's same-origin posture (CLAUDE.md). httpOnly
  // cookies set by the backend round-trip cleanly because the browser
  // treats the response as same-origin.
  //
  // Cross-origin override: if NEXT_PUBLIC_API_URL is set (advanced
  // case — pointing the local dashboard at a remote API), the API
  // client's baseURL falls back to absolute URLs and these rewrites
  // are bypassed. That path then requires the target origin to be
  // listed in the backend's CORS_ALLOWED_ORIGINS env var so CORS +
  // CSRF let it through.
  async rewrites() {
    return [
      { source: '/api/v1/:path*', destination: `${apiProxyTarget}/api/v1/:path*` },
      { source: '/v1/:path*',     destination: `${apiProxyTarget}/v1/:path*` },
    ];
  },

  async redirects() {
    return [
      {
        source: '/terms',
        destination: 'https://brokle.com/terms',
        permanent: false,
      },
      {
        source: '/privacy',
        destination: 'https://brokle.com/privacy',
        permanent: false,
      },
      {
        source: '/login',
        destination: '/signin',
        permanent: false,
      },
    ];
  },
};

export default nextConfig;
