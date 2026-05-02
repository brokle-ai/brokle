import { CheckCircle, XCircle, HelpCircle } from 'lucide-react'

// OTEL StatusCode enum values (UInt8 from backend)
export const StatusCode = {
  UNSET: 0,
  OK: 1,
  ERROR: 2,
} as const

// Convert UInt8 status code to string for display.
export function statusCodeToString(code: number): 'ok' | 'error' | 'unset' {
  switch (code) {
    case StatusCode.OK:
      return 'ok'
    case StatusCode.ERROR:
      return 'error'
    case StatusCode.UNSET:
    default:
      return 'unset'
  }
}

export const statuses = [
  { label: 'OK', value: 'ok' as const, icon: CheckCircle },
  { label: 'Error', value: 'error' as const, icon: XCircle },
  { label: 'Unset', value: 'unset' as const, icon: HelpCircle },
]
