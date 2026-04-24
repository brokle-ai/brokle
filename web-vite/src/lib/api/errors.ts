// Typed error hierarchy for all HTTP failures from the Brokle backend.
// Mirrors sdk/javascript/src/errors.ts and sdk/python/brokle/_http/errors.py
// — same base + per-type subclasses, so `instanceof ValidationError`
// works identically in the dashboard and in the SDKs. CLAUDE.md gotcha
// #30b documents the two-axis model at length: one shared hierarchy,
// module-local subclasses only when semantics exceed HTTP status.
//
// Backend envelope is Stripe/OpenAI shape (CLAUDE.md gotcha #23):
//   { "error": { type, code?, message, param?, details?, errors? } }
// HTTP status is the primary dispatch signal; the `type` field inside
// the body disambiguates in edge cases (e.g. 422 can be validation OR
// invalid_request — check the body when status alone is ambiguous).

export interface ErrorFieldIssue {
  location?: string
  message: string
  value?: unknown
}

export interface ErrorBody {
  type: string
  code?: string
  message: string
  details?: string
  param?: string
  errors?: ErrorFieldIssue[]
}

export interface ErrorResponse {
  error: ErrorBody
}

export class BrokleError extends Error {
  readonly status: number
  readonly body: ErrorResponse
  readonly requestId: string | undefined

  constructor(status: number, body: ErrorResponse, requestId?: string) {
    super(body.error.message)
    this.name = 'BrokleError'
    this.status = status
    this.body = body
    this.requestId = requestId
    Object.setPrototypeOf(this, new.target.prototype)
  }

  get type(): string {
    return this.body.error.type
  }

  get code(): string | undefined {
    return this.body.error.code
  }

  get param(): string | undefined {
    return this.body.error.param
  }

  get fieldIssues(): ErrorFieldIssue[] | undefined {
    return this.body.error.errors
  }
}

export class AuthenticationError extends BrokleError {
  constructor(status: number, body: ErrorResponse, requestId?: string) {
    super(status, body, requestId)
    this.name = 'AuthenticationError'
  }
}

export class NotFoundError extends BrokleError {
  constructor(status: number, body: ErrorResponse, requestId?: string) {
    super(status, body, requestId)
    this.name = 'NotFoundError'
  }
}

export class ValidationError extends BrokleError {
  constructor(status: number, body: ErrorResponse, requestId?: string) {
    super(status, body, requestId)
    this.name = 'ValidationError'
  }
}

export class RateLimitError extends BrokleError {
  readonly retryAfter: number | undefined
  constructor(
    status: number,
    body: ErrorResponse,
    requestId?: string,
    retryAfter?: number,
  ) {
    super(status, body, requestId)
    this.name = 'RateLimitError'
    this.retryAfter = retryAfter
  }
}

export class ServerError extends BrokleError {
  constructor(status: number, body: ErrorResponse, requestId?: string) {
    super(status, body, requestId)
    this.name = 'ServerError'
  }
}

// Called by the HTTP middleware on any non-2xx response. Parses the
// body once, classifies by status, and throws the matching subclass.
// A parse failure on a non-2xx status still throws — a generic
// BrokleError with a synthetic body — because silent 2xx-ification
// would corrupt every downstream handler.
export async function throwTypedError(response: Response): Promise<never> {
  const requestId = response.headers.get('x-request-id') ?? undefined
  const retryAfterHeader = response.headers.get('retry-after')
  const retryAfter = retryAfterHeader ? Number.parseInt(retryAfterHeader, 10) : undefined

  let body: ErrorResponse
  try {
    body = (await response.clone().json()) as ErrorResponse
    if (!body || typeof body !== 'object' || !('error' in body)) {
      body = synthBody(response)
    }
  } catch {
    body = synthBody(response)
  }

  const Ctor =
    response.status === 401 || response.status === 403
      ? AuthenticationError
      : response.status === 404
        ? NotFoundError
        : response.status === 422 || response.status === 400
          ? ValidationError
          : response.status === 429
            ? undefined
            : response.status >= 500
              ? ServerError
              : BrokleError

  if (response.status === 429) {
    throw new RateLimitError(response.status, body, requestId, retryAfter)
  }
  if (Ctor) {
    throw new Ctor(response.status, body, requestId)
  }
  throw new BrokleError(response.status, body, requestId)
}

function synthBody(response: Response): ErrorResponse {
  return {
    error: {
      type: response.status >= 500 ? 'api_error' : 'invalid_request',
      message: response.statusText || `HTTP ${response.status}`,
    },
  }
}
