import { CheckCircle, XCircle, Circle, HelpCircle } from 'lucide-react'

// OTEL StatusCode enum values (UInt8 from backend)
export const StatusCode = {
  UNSET: 0,
  OK: 1,
  ERROR: 2,
} as const

// OTEL SpanKind enum values (UInt8 from backend)
export const SpanKind = {
  UNSPECIFIED: 0,
  INTERNAL: 1,
  SERVER: 2,
  CLIENT: 3,
  PRODUCER: 4,
  CONSUMER: 5,
} as const

// Convert UInt8 status code to string for display
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

// Convert UInt8 span kind to string for display
export function spanKindToString(kind: number): string {
  switch (kind) {
    case SpanKind.INTERNAL:
      return 'INTERNAL'
    case SpanKind.SERVER:
      return 'SERVER'
    case SpanKind.CLIENT:
      return 'CLIENT'
    case SpanKind.PRODUCER:
      return 'PRODUCER'
    case SpanKind.CONSUMER:
      return 'CONSUMER'
    case SpanKind.UNSPECIFIED:
    default:
      return 'UNSPECIFIED'
  }
}

export const statuses = [
  {
    label: 'OK',
    value: 'ok' as const,
    icon: CheckCircle,
  },
  {
    label: 'Error',
    value: 'error' as const,
    icon: XCircle,
  },
  {
    label: 'Unset',
    value: 'unset' as const,
    icon: HelpCircle,
  },
]

export const environments = [
  { value: 'production', label: 'Production' },
  { value: 'staging', label: 'Staging' },
  { value: 'development', label: 'Development' },
]

export const spanTypes = [
  { value: 'generation', label: 'Generation' },
  { value: 'tool', label: 'Tool' },
  { value: 'retrieval', label: 'Retrieval' },
  { value: 'embedding', label: 'Embedding' },
  { value: 'span', label: 'Span' },
]
