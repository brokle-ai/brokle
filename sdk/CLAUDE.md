# SDK Rules (`sdk/`)

Loaded automatically when Claude reads files under `sdk/javascript/` or `sdk/python/`.
Universal repo rules in `/CLAUDE.md`. Backend wire contract in `internal/CLAUDE.md`.

## Submodule workflow

`sdk/javascript/` and `sdk/python/` are independent git repos mounted as submodules.
- Commits inside them don't appear in the main repo's `git diff` — only the submodule pointer changes.
- Use `cd sdk/javascript && git status` to inspect SDK changes.
- Run `git submodule update --init` after cloning.

## Wire contract

The backend emits raw resource bodies on success (Stripe / OpenAI shape) and a single `error` envelope on failure. Both SDKs centralise REST transport in one module:

- Python: `sdk/python/brokle/_http/client.py` (`SyncHTTPClient` / `AsyncHTTPClient`)
- JS: `sdk/javascript/src/_http/client.ts` (`BrokleHttpClient`)

Resource managers (scores, datasets, prompts, experiments, query, annotations) call `http.get/post/patch/delete` and get the raw resource back — no per-manager envelope unwrap. Do NOT reintroduce `unwrap_response` / `extractData` helpers.

Error paths raise typed exceptions matched to HTTP status: `AuthenticationError` (401/403), `NotFoundError` (404), `ValidationError` (422), `RateLimitError` (429), `ServerError` (5xx), `BrokleError` fallback.

## Error taxonomy — two-axis model

One shared error hierarchy per SDK (`BrokleError` + the typed family above), shipped from `_http/errors.py` / `src/errors.ts`. Module-local error types exist ONLY for axis-2 semantics HTTP status cannot express, and they **extend** the shared family:
- Python: `InvalidFilterError extends ValidationError`.
- JS: `InvalidFilterError`, `QueueNotFoundError`, `ItemNotFoundError`, `ItemLockedError`, `NoItemsAvailableError`, `ScorerError`.

NEVER introduce catch-all wrappers like `QueryAPIError` / `ScoreError` / `AnnotationError(error.message)` that swallow the shared subclass — they break `isinstance(e, AuthenticationError)` / `e instanceof AuthenticationError` at the call site (the N×M class explosion industry explicitly avoids).

Parse failures on 2xx bodies wrap as `BrokleError("Failed to parse …", original_error=e)` — typed family, preserves the chain, still one catch clause.

`classifyError` helpers (see `annotations/manager.ts:77-98`) translate ONLY for axis-2 distinctions (queue vs item 404, lock contention, empty-queue state) — never for blanket re-typing.

## Singleton & override

First-write-wins:
- JS: `Symbol.for('brokle')` on `globalThis`.
- Python: module-level `_client` variable.

First `BrokleClient()` / `Brokle()` call wins — subsequent calls with different configs are ignored. Use `setClient()` / `set_client()` to explicitly override.

## Provider wrappers

Optional peer deps. JS: `wrapOpenAI`, `wrapAnthropic`, etc. require the provider SDK installed but won't fail at import — only at wrapper call. Python: `brokle.wrappers`. Never bundle provider SDKs as direct dependencies.

JS: wrappers use a recursive `Proxy` that intercepts method calls without modifying the original client. Don't extend or subclass provider clients; wrap them.

## JS specifics

- Multi-entry `tsup` build: 11 entry points (core + 10 integrations) with separate `.d.ts` files. Import from sub-paths: `import { wrapOpenAI } from 'brokle/openai'`, not from root `'brokle'`.
- Node ≥ 20 required (relies on `AsyncLocalStorage`). No browser or Node 18 support without polyfills.

## Python specifics

- Lazy module loading: `brokle/__init__.py` uses `__getattr__` + `importlib` for 150+ exports. `from brokle import *` won't work. Import specific names.
- Env var prefix: `BROKLE_*` (uppercase). `BROKLE_API_KEY`, `BROKLE_BASE_URL`, `BROKLE_ENABLED`. Lowercase variants don't work.
- No atexit flush. Serverless / CLI apps must call `brokle.flush()` before process exit or traces are lost.
- mypy selectively disabled: `pyproject.toml` has `ignore_errors = true` overrides for wrappers, config, and client modules. Don't assume full type safety in those areas.

For prior context on the wire-contract migration and the SDK error-hierarchy decision, see `docs/decisions/`.
