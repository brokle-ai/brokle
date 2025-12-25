import { BrokleAPIError } from './core/types'

/**
 * Safely extracts an error message from an unknown error type.
 * Handles BrokleAPIError, standard Error, and object-with-message patterns.
 *
 * @param error - The error to extract a message from
 * @param fallback - The fallback message if extraction fails
 * @returns The extracted error message or the fallback
 */
export function extractErrorMessage(error: unknown, fallback: string): string {
  if (error instanceof BrokleAPIError) {
    return error.message
  }

  if (error instanceof Error) {
    return error.message
  }

  if (
    typeof error === 'object' &&
    error !== null &&
    'message' in error &&
    typeof (error as Record<string, unknown>).message === 'string'
  ) {
    return (error as Record<string, unknown>).message as string
  }

  return fallback
}
