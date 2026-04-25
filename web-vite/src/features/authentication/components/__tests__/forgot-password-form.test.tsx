import { describe, it, expect, vi } from 'vitest'
import { fireEvent, screen, waitFor } from '@testing-library/react'
import { http, HttpResponse } from 'msw'
import { server } from '@/mocks/server'
import { ForgotPasswordForm } from '../forgot-password-form'
import { renderWithProviders } from '@/test/auth-test-utils'

describe('ForgotPasswordForm', () => {
  it('renders the email input and continue button', async () => {
    await renderWithProviders(<ForgotPasswordForm />)
    expect(screen.getByLabelText(/^email$/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /continue/i })).toBeInTheDocument()
  })

  it('submits the email to the forgot-password endpoint and calls onSuccess', async () => {
    let capturedBody: { email?: string } | undefined
    server.use(
      http.post('*/api/v1/auth/forgot-password', async ({ request }) => {
        capturedBody = (await request.json()) as { email?: string }
        return new HttpResponse(null, { status: 204 })
      }),
    )

    const onSuccess = vi.fn()
    await renderWithProviders(<ForgotPasswordForm onSuccess={onSuccess} />)
    fireEvent.change(screen.getByLabelText(/^email$/i), {
      target: { value: 'reset@brokle.test' },
    })
    fireEvent.click(screen.getByRole('button', { name: /continue/i }))

    await waitFor(() => {
      expect(capturedBody).toEqual({ email: 'reset@brokle.test' })
    })
    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith('reset@brokle.test')
    })
  })

  it('still calls onSuccess on backend error (anti-enumeration)', async () => {
    server.use(
      http.post('*/api/v1/auth/forgot-password', () =>
        HttpResponse.json(
          { error: { type: 'api_error', message: 'boom' } },
          { status: 500 },
        ),
      ),
    )

    const onSuccess = vi.fn()
    await renderWithProviders(<ForgotPasswordForm onSuccess={onSuccess} />)
    fireEvent.change(screen.getByLabelText(/^email$/i), {
      target: { value: 'reset@brokle.test' },
    })
    fireEvent.click(screen.getByRole('button', { name: /continue/i }))

    await waitFor(() => {
      expect(onSuccess).toHaveBeenCalledWith('reset@brokle.test')
    })
  })
})
