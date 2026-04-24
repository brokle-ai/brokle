import { describe, expect, it } from 'vitest'
import {
  AuthenticationError,
  BrokleError,
  NotFoundError,
  RateLimitError,
  ServerError,
  ValidationError,
  throwTypedError,
} from '../errors'

function resp(status: number, body: object | null, headers: Record<string, string> = {}) {
  return new Response(body ? JSON.stringify(body) : null, {
    status,
    statusText: statusText(status),
    headers: { 'content-type': 'application/json', ...headers },
  })
}

function statusText(s: number): string {
  const map: Record<number, string> = {
    400: 'Bad Request',
    401: 'Unauthorized',
    403: 'Forbidden',
    404: 'Not Found',
    422: 'Unprocessable Entity',
    429: 'Too Many Requests',
    500: 'Internal Server Error',
    503: 'Service Unavailable',
  }
  return map[s] ?? ''
}

const envelope = (type: string, message = 'oops') => ({ error: { type, message } })

describe('throwTypedError', () => {
  it.each([
    [401, AuthenticationError, 'authentication'],
    [403, AuthenticationError, 'authentication'],
    [404, NotFoundError, 'not_found'],
    [400, ValidationError, 'validation'],
    [422, ValidationError, 'validation'],
    [500, ServerError, 'api_error'],
    [503, ServerError, 'api_error'],
  ])('maps status %s to %s', async (status, Ctor, type) => {
    await expect(throwTypedError(resp(status, envelope(type)))).rejects.toBeInstanceOf(Ctor)
  })

  it('maps 429 to RateLimitError with retry-after header', async () => {
    try {
      await throwTypedError(resp(429, envelope('rate_limit'), { 'retry-after': '42' }))
      expect.fail('should have thrown')
    } catch (err) {
      expect(err).toBeInstanceOf(RateLimitError)
      expect((err as RateLimitError).retryAfter).toBe(42)
    }
  })

  it('falls back to synthetic body on non-JSON response', async () => {
    try {
      await throwTypedError(resp(500, null))
      expect.fail('should have thrown')
    } catch (err) {
      expect(err).toBeInstanceOf(ServerError)
      expect((err as ServerError).body.error.type).toBe('api_error')
      expect((err as ServerError).body.error.message).toMatch(/Internal Server Error|HTTP 500/)
    }
  })

  it('preserves request-id header on the thrown error', async () => {
    try {
      await throwTypedError(resp(500, envelope('api_error'), { 'x-request-id': 'req_abc' }))
      expect.fail('should have thrown')
    } catch (err) {
      expect((err as BrokleError).requestId).toBe('req_abc')
    }
  })

  it('reads field-level issues from the envelope', async () => {
    const body = {
      error: {
        type: 'validation',
        message: 'bad input',
        errors: [{ location: 'email', message: 'must be an email' }],
      },
    }
    try {
      await throwTypedError(resp(422, body))
      expect.fail('should have thrown')
    } catch (err) {
      expect(err).toBeInstanceOf(ValidationError)
      expect((err as ValidationError).fieldIssues).toEqual([
        { location: 'email', message: 'must be an email' },
      ])
    }
  })
})
