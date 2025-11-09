import { CheckCircle, XCircle, Circle, HelpCircle } from 'lucide-react'

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

export const observationTypes = [
  { value: 'generation', label: 'Generation' },
  { value: 'tool', label: 'Tool' },
  { value: 'retrieval', label: 'Retrieval' },
  { value: 'embedding', label: 'Embedding' },
  { value: 'span', label: 'Span' },
]
