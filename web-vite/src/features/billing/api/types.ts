// Wire types for the org-level billing views. Shapes match the JSON
// emitted by the Go backend (internal/core/domain/billing/entity.go and
// service.go). shopspring/decimal fields serialise as strings on the
// wire — we keep them string-typed and parse at render via
// `parseDecimal` in the components.

export interface Plan {
  id: string
  name: string // "free" | "pro" | "enterprise"
  is_active: boolean
  is_default: boolean
  created_at: string
  updated_at: string

  free_spans: number
  price_per_100k_spans?: string // decimal string; nil = unlimited in free tier

  free_gb: string // decimal string
  price_per_gb?: string

  free_scores: number
  price_per_1k_scores?: string
}

export type ContractStatus = 'draft' | 'active' | 'cancelled' | 'expired'

export interface Contract {
  id: string
  organization_id: string
  contract_name: string
  contract_number: string
  start_date: string
  end_date?: string

  minimum_commit_amount?: string // decimal
  currency: string

  account_owner?: string
  sales_rep_email?: string

  status: ContractStatus

  custom_free_spans?: number
  custom_price_per_100k_spans?: string
  custom_free_gb?: string
  custom_price_per_gb?: string
  custom_free_scores?: number
  custom_price_per_1k_scores?: string

  created_by?: string
  created_at: string
  updated_at: string
  notes?: string
}

// Resolved pricing after contract overrides are applied. Backend
// computes this server-side; the view just renders what it returns.
export interface EffectivePricing {
  organization_id: string
  base_plan: Plan | null
  contract?: Contract

  // Resolved limits (numeric dimensions are JSON numbers, currency
  // dimensions arrive as decimal strings)
  free_spans: number
  price_per_100k_spans: string
  free_gb: string
  price_per_gb: string
  free_scores: number
  price_per_1k_scores: string

  has_volume_tiers: boolean
}

export interface UsageOverview {
  organization_id: string
  period_start: string
  period_end: string

  spans: number
  bytes: number
  scores: number

  free_spans_remaining: number
  free_bytes_remaining: number
  free_scores_remaining: number

  free_spans_total: number
  free_bytes_total: number
  free_scores_total: number

  estimated_cost: string // decimal string
}

// Placeholder shape for the invoices view — the backend has no invoice
// endpoint yet (see billing handler domain). The query stub returns an
// empty list so the route renders cleanly and the shape is ready for
// the real endpoint when it lands.
export interface Invoice {
  id: string
  number: string
  issued_at: string
  period_start: string
  period_end: string
  amount: string // decimal
  currency: string
  status: 'open' | 'paid' | 'void' | 'uncollectible'
  download_url?: string
}

export interface Pagination {
  page: number
  limit: number
  total: number
  total_pages: number
  has_next: boolean
  has_prev: boolean
}

export interface InvoiceListResponse {
  data: Invoice[]
  pagination: Pagination
}
