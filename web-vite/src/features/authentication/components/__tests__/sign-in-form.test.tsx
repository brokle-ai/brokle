import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import { fireEvent, screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '@/mocks/server'
import { SignInForm } from '../sign-in-form'
import { useAuthStore } from '@/stores/auth-store'
import { renderWithProviders } from '@/test/auth-test-utils'

// Smoke tests for the SignInForm. Asserts:
//   - the form renders email + password inputs
//   - submit POSTs to /api/v1/auth/login with the typed payload
//   - error envelope ({error: {message}}) surfaces in the UI

describe('SignInForm', () => {
  let originalHref: string

  beforeEach(() => {
    useAuthStore.getState().setUser(null)
    originalHref = window.location.href
    // jsdom throws on real assignment to location.href; replace with a
    // configurable proxy so the form's `window.location.href = '/'`
    // becomes a no-op observable assignment.
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...window.location, href: originalHref, assign: vi.fn() },
    })
  })

  afterEach(() => {
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: { ...window.location, href: originalHref },
    })
  })

  it('renders the email + password inputs and the submit button', async () => {
    await renderWithProviders(<SignInForm />)
    expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/^password$/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument()
  })

  it('submits the login payload and updates the auth store on success', async () => {
    let capturedBody: { email?: string; password?: string } | undefined
    server.use(
      http.post('*/api/v1/auth/login', async ({ request }) => {
        capturedBody = (await request.json()) as {
          email?: string
          password?: string
        }
        return HttpResponse.json(
          {
            user: {
              id: 'usr_1',
              email: capturedBody.email,
              first_name: 'A',
              last_name: 'B',
              default_organization_id: 'org_1',
              is_email_verified: true,
              created_at: new Date().toISOString(),
            },
            expires_at: Date.now() + 60_000,
            expires_in: 60_000,
          },
          { status: 200 },
        )
      }),
    )

    await renderWithProviders(<SignInForm />)
    fireEvent.change(screen.getByLabelText(/^email$/i), {
      target: { value: 'a@b.test' },
    })
    fireEvent.change(screen.getByLabelText(/^password$/i), {
      target: { value: 'sup3rsecret' },
    })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(capturedBody).toEqual({ email: 'a@b.test', password: 'sup3rsecret' })
    })
    await waitFor(() => {
      expect(useAuthStore.getState().user?.email).toBe('a@b.test')
    })
  })

  it('renders the error envelope message on validation failure', async () => {
    // Use 422 (not 401) — 401 is intercepted by the rawFetch refresh
    // path and only surfaces after a refresh attempt also fails. The
    // assertion target here is "the form's error UI renders the
    // backend message", which works on any non-2xx the client doesn't
    // special-case.
    server.use(
      http.post('*/api/v1/auth/login', () =>
        HttpResponse.json(
          {
            error: {
              type: 'validation',
              message: 'Invalid credentials.',
            },
          },
          { status: 422 },
        ),
      ),
    )

    await renderWithProviders(<SignInForm />)
    fireEvent.change(screen.getByLabelText(/^email$/i), {
      target: { value: 'a@b.test' },
    })
    fireEvent.change(screen.getByLabelText(/^password$/i), {
      target: { value: 'wrongpass' },
    })
    fireEvent.click(screen.getByRole('button', { name: /sign in/i }))

    await waitFor(() => {
      expect(screen.getByText(/invalid credentials/i)).toBeInTheDocument()
    })
  })
})
