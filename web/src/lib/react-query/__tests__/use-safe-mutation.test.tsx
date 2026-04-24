/**
 * useSafeMutation tests — structural re-entry guard on top of React
 * Query's useMutation. The hook exists to close the window between
 * "click fires" and "button re-renders disabled={true}" where a
 * duplicate event would otherwise trigger a second HTTP request for
 * a single user action (the signup "Email already registered" bug).
 */

import { describe, it, expect, vi } from 'vitest'
import { renderHook, act, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import type { ReactNode } from 'react'

import {
  MutationReentryError,
  useSafeMutation,
} from '../use-safe-mutation'

function wrapper() {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false },
      mutations: { retry: false },
    },
  })
  return function Wrapper({ children }: { children: ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  }
}

// deferred returns a controllable Promise so a test can hold the
// first mutation in flight while dispatching a duplicate call.
function deferred<T>() {
  let resolve!: (value: T) => void
  let reject!: (err: unknown) => void
  const promise = new Promise<T>((res, rej) => {
    resolve = res
    reject = rej
  })
  return { promise, resolve, reject }
}

describe('useSafeMutation', () => {
  it('rejects a second mutateAsync while the first is in flight', async () => {
    const d = deferred<string>()
    const mutationFn = vi.fn().mockImplementation(() => d.promise)

    const { result } = renderHook(() => useSafeMutation({ mutationFn }), {
      wrapper: wrapper(),
    })

    // Kick off the first call without awaiting — it stays in flight.
    let firstPromise!: Promise<string>
    act(() => {
      firstPromise = result.current.mutateAsync()
    })

    // Second call while first is pending must reject with the sentinel.
    let secondError: unknown
    await act(async () => {
      try {
        await result.current.mutateAsync()
      } catch (err) {
        secondError = err
      }
    })

    expect(secondError).toBeInstanceOf(MutationReentryError)
    expect(mutationFn).toHaveBeenCalledTimes(1)

    // Resolve the first call so the test can finish cleanly.
    await act(async () => {
      d.resolve('ok')
      await firstPromise
    })
  })

  it('releases the lock after a failure so the user can retry', async () => {
    const mutationFn = vi
      .fn()
      .mockRejectedValueOnce(new Error('boom'))
      .mockResolvedValueOnce('second-ok')

    const { result } = renderHook(() => useSafeMutation({ mutationFn }), {
      wrapper: wrapper(),
    })

    await act(async () => {
      await expect(result.current.mutateAsync()).rejects.toThrow('boom')
    })

    await act(async () => {
      await expect(result.current.mutateAsync()).resolves.toBe('second-ok')
    })

    expect(mutationFn).toHaveBeenCalledTimes(2)
  })

  it('releases the lock after a successful mutation', async () => {
    const mutationFn = vi.fn().mockResolvedValue('ok')
    const { result } = renderHook(() => useSafeMutation({ mutationFn }), {
      wrapper: wrapper(),
    })

    await act(async () => {
      await result.current.mutateAsync()
    })
    await act(async () => {
      await result.current.mutateAsync()
    })

    expect(mutationFn).toHaveBeenCalledTimes(2)
  })

  it('forwards the mutationFn result on success', async () => {
    const mutationFn = vi.fn().mockResolvedValue({ id: 'abc' })
    const { result } = renderHook(() => useSafeMutation({ mutationFn }), {
      wrapper: wrapper(),
    })

    let data: unknown
    await act(async () => {
      data = await result.current.mutateAsync()
    })
    expect(data).toEqual({ id: 'abc' })
  })

  it('isPending is true while the mutation runs and false after settle', async () => {
    const d = deferred<string>()
    const { result } = renderHook(
      () => useSafeMutation({ mutationFn: () => d.promise }),
      { wrapper: wrapper() },
    )

    let p!: Promise<string>
    act(() => {
      p = result.current.mutateAsync()
    })
    await waitFor(() => expect(result.current.isPending).toBe(true))

    await act(async () => {
      d.resolve('done')
      await p
    })
    await waitFor(() => expect(result.current.isPending).toBe(false))
  })
})
