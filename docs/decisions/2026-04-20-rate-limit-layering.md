---
date: 2026-04-20
status: enacted
tags: [middleware, rate-limit, multi-tenant]
---

# Rate-limit layering audit (per-IP for unauth, per-principal for authed)

The `/api/v1/*` stack was carrying both a mux-level `LimitByIP` and a Huma-group-level `LimitByUser` on authenticated traffic. Phase-2 DAL SSR fetches exposed the bug: all users' `/api/v1/users/me` hits from the shared Next.js Node pod collapsed into one IP bucket, cross-throttling legitimate users at normal dashboard concurrency.

The fix was structural, not a forward-XFF workaround — authenticated buckets are now per-principal only, matching GitHub (per-user post-auth, per-IP pre-auth), Stripe (per-account), OpenAI/Anthropic (per-org), and Cloudflare's WAF best-practice guidance.

`LimitByIP` moved from the mux onto the pre-auth Huma groups (`dashPublic`, `sdkPublic`) where it belongs; authed groups run with `LimitByUser` / `LimitByAPIKey` alone.

**Generalisable rule**: never layer an IP counter on top of a principal counter on the same request — NAT, CGNAT, corporate egress, and BFF/SSR all collapse the IP axis, and the stack fails on legitimate traffic before it fails on abusers. Forwarding `X-Forwarded-For` from the SSR hop is now a logging/audit correctness fix (populates `httpctx.ClientIP` with the real user IP), not a rate-limit requirement.
