import { useRef } from 'react'
import {
  useMutation,
  type UseMutationOptions,
  type UseMutationResult,
} from '@tanstack/react-query'

// MutationReentryError is the sentinel thrown when a caller invokes
// mutate / mutateAsync while the previous call for the same hook
// instance is still in flight. Callers (form onSubmit catch blocks,
// onError handlers) filter this class and ignore it — a blocked
// duplicate click must not surface as a user-visible error.
export class MutationReentryError extends Error {
  constructor() {
    super('mutation reentry blocked')
    this.name = 'MutationReentryError'
  }
}

// useSafeMutation wraps useMutation with a synchronous in-flight
// guard. React state (isPending) is not visible to a second call
// made in the same microtask — a ref mutation is. This closes the
// window between "first click fires" and "button re-renders with
// disabled={true}" that lets duplicate submit events produce two
// identical POSTs on a single user action.
//
// Guidance for the user-land guard comes from TanStack/query#2191:
// "preventing the call of a function is a check that should run in
// user-land." The button's disabled prop is still the UX surface;
// this hook is the structural backstop.
export function useSafeMutation<
  TData = unknown,
  TError = Error,
  TVariables = void,
  TContext = unknown,
>(
  options: UseMutationOptions<TData, TError, TVariables, TContext>,
): UseMutationResult<TData, TError, TVariables, TContext> {
  const inFlightRef = useRef(false)
  const userMutationFn = options.mutationFn

  return useMutation<TData, TError, TVariables, TContext>({
    ...options,
    mutationFn: async (variables, context) => {
      if (inFlightRef.current) {
        if (process.env.NODE_ENV === 'development') {
          console.warn(
            '[useSafeMutation] Blocked re-entry; a mutation is already in flight.',
          )
        }
        throw new MutationReentryError()
      }
      if (!userMutationFn) {
        throw new Error('useSafeMutation requires a mutationFn')
      }
      inFlightRef.current = true
      try {
        return await userMutationFn(variables, context)
      } finally {
        inFlightRef.current = false
      }
    },
  })
}
