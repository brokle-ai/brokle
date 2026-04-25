import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  BillableUsageSummary,
  CreateBudgetRequest,
  EffectivePricing,
  InvoiceListResponse,
  UpdateBudgetRequest,
  UsageAlert,
  UsageBudget,
  UsageOverview,
} from './types'

// TkDodo-style hierarchical keys. Invalidate by level — e.g. a future
// plan-change mutation busts `planKey` only without churning usage.
export const billingKeys = {
  all: ['billing'] as const,
  plan: (orgId: string) => [...billingKeys.all, 'plan', orgId] as const,
  usage: (orgId: string) => [...billingKeys.all, 'usage', orgId] as const,
  invoices: () => [...billingKeys.all, 'invoices'] as const,
  invoiceList: (orgId: string, params: InvoiceListParams) =>
    [...billingKeys.invoices(), orgId, params] as const,
  budgets: () => [...billingKeys.all, 'budgets'] as const,
  budgetList: (orgId: string) => [...billingKeys.budgets(), orgId] as const,
  alerts: () => [...billingKeys.all, 'alerts'] as const,
  alertList: (orgId: string, limit?: number) =>
    [...billingKeys.alerts(), orgId, limit ?? 'all'] as const,
  usageByProject: (orgId: string, range: UsageRangeKey) =>
    [...billingKeys.all, 'usage-by-project', orgId, range] as const,
} as const

// Encodes the period selection used by the usage-by-project query.
// `current` and `previous` map to server-side relative ranges (30d
// behind / 30d before that). `custom` carries explicit RFC 3339
// boundaries forwarded to the backend. Other relative presets reuse
// the analytics enum on the backend (`time_range` query param).
export type UsagePeriodRelative =
  | 'current'
  | 'previous'
  | '7d'
  | '14d'
  | '30d'

export interface UsagePeriodCustom {
  kind: 'custom'
  from: string
  to: string
}

export interface UsagePeriodPreset {
  kind: 'relative'
  relative: UsagePeriodRelative
}

export type UsagePeriod = UsagePeriodPreset | UsagePeriodCustom

// Cache-key shape for usage-by-project — flat strings keep React
// Query's structural-equality check stable across renders.
type UsageRangeKey =
  | { kind: 'relative'; relative: UsagePeriodRelative }
  | { kind: 'custom'; from: string; to: string }

// "Plan" surface is served by the effective-pricing endpoint — it
// resolves contract overrides over the base plan server-side, which is
// what the card actually needs to display.
export const planQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: billingKeys.plan(orgId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/billing/organizations/${orgId}/effective-pricing`,
        { method: 'GET' },
      )
      return (await resp.json()) as EffectivePricing
    },
    // Plans rarely change in a given session. 5 min keeps in-tab nav
    // snappy; a real plan-change flow will invalidate explicitly.
    staleTime: 5 * 60 * 1000,
  })

export const usageQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: billingKeys.usage(orgId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/organizations/${orgId}/usage/overview`,
        { method: 'GET' },
      )
      return (await resp.json()) as UsageOverview
    },
    // Usage counters update continuously; a 30s window matches the
    // traces-list cadence and avoids hammering the aggregation query.
    staleTime: 30 * 1000,
  })

export interface InvoiceListParams {
  page: number
  limit: number
}

// No backend invoice endpoint exists yet (see billing handler domain).
// The query returns an empty page so the route renders a clean empty
// state and the shape is ready for the real endpoint. Swap the stub
// body for `rawFetch(...)` when `/invoices` lands.
export const invoiceListQueryOptions = (
  orgId: string,
  params: InvoiceListParams,
) =>
  queryOptions({
    queryKey: billingKeys.invoiceList(orgId, params),
    queryFn: async (): Promise<InvoiceListResponse> => ({
      data: [],
      pagination: {
        page: params.page,
        limit: params.limit,
        total: 0,
        total_pages: 0,
        has_next: false,
        has_prev: false,
      },
    }),
    staleTime: 60 * 1000,
  })

// ---- Budgets -------------------------------------------------------------
// Backend returns `List` directly (no envelope) — a bare JSON array of
// `UsageBudget`. An org can have one org-level budget plus N project
// budgets; the billing card today uses the first org-level row.

export const budgetsQueryOptions = (orgId: string) =>
  queryOptions({
    queryKey: billingKeys.budgetList(orgId),
    queryFn: async () => {
      const resp = await rawFetch(
        `/api/v1/organizations/${orgId}/budgets`,
        { method: 'GET' },
      )
      return (await resp.json()) as UsageBudget[]
    },
    // Usage counters embedded in the budget update continuously, but
    // the card is a landing-page summary — 30s keeps it lively without
    // hammering the billing service.
    staleTime: 30 * 1000,
  })

export async function createBudget(
  orgId: string,
  data: CreateBudgetRequest,
): Promise<UsageBudget> {
  const resp = await rawFetch(`/api/v1/organizations/${orgId}/budgets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data),
  })
  return (await resp.json()) as UsageBudget
}

// Backend uses PUT (not PATCH) for budget updates — see
// `handlers/billing/handlers.go` operation "update-budget".
export async function updateBudget(
  orgId: string,
  budgetId: string,
  data: UpdateBudgetRequest,
): Promise<UsageBudget> {
  const resp = await rawFetch(
    `/api/v1/organizations/${orgId}/budgets/${budgetId}`,
    {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data),
    },
  )
  return (await resp.json()) as UsageBudget
}

// DELETE /budgets/{id} returns 204 No Content. Backend operation
// "delete-budget" — see `handlers/billing/handlers.go`.
export async function deleteBudget(
  orgId: string,
  budgetId: string,
): Promise<void> {
  await rawFetch(`/api/v1/organizations/${orgId}/budgets/${budgetId}`, {
    method: 'DELETE',
  })
}

// ---- Alerts -------------------------------------------------------------
// Budget alerts surface threshold breaches. Backend operation
// "list-budget-alerts" — `/budgets/alerts?limit=N`. Returns an array
// of `UsageAlert` (no envelope).

export const alertsQueryOptions = (orgId: string, limit?: number) =>
  queryOptions({
    queryKey: billingKeys.alertList(orgId, limit),
    queryFn: async () => {
      const search = new URLSearchParams()
      if (limit !== undefined) search.set('limit', String(limit))
      const qs = search.toString()
      const resp = await rawFetch(
        `/api/v1/organizations/${orgId}/budgets/alerts${qs ? `?${qs}` : ''}`,
        { method: 'GET' },
      )
      return (await resp.json()) as UsageAlert[]
    },
    // Alerts should refresh more frequently than the budget summary —
    // they're the actionable signal on the page.
    staleTime: 60 * 1000,
  })

export async function acknowledgeAlert(
  orgId: string,
  alertId: string,
): Promise<void> {
  await rawFetch(
    `/api/v1/organizations/${orgId}/budgets/alerts/${alertId}/acknowledge`,
    { method: 'POST' },
  )
}

// ---- Usage by project ---------------------------------------------------
// Backed by `/usage/by-project` — accepts either a `time_range` enum
// preset (30d max) or a custom `from`/`to` window. The `current` and
// `previous` selections are computed client-side because the backend
// doesn't have a billing-period concept yet — we pin "current" to the
// last 30d and "previous" to the 30d before that. Swap to a server
// `period` param when the billing service grows real periods.

function periodToQueryString(period: UsagePeriod): string {
  const search = new URLSearchParams()
  if (period.kind === 'custom') {
    search.set('from', period.from)
    search.set('to', period.to)
    return search.toString()
  }
  const rel = period.relative
  if (rel === 'current') {
    // Last 30 days — backend `time_range=30d`.
    search.set('time_range', '30d')
  } else if (rel === 'previous') {
    // 30 days before "current". Compute custom from/to.
    const now = Date.now()
    const day = 24 * 60 * 60 * 1000
    const from = new Date(now - 60 * day).toISOString()
    const to = new Date(now - 30 * day).toISOString()
    search.set('from', from)
    search.set('to', to)
  } else {
    search.set('time_range', rel)
  }
  return search.toString()
}

function periodToKey(period: UsagePeriod): UsageRangeKey {
  if (period.kind === 'custom') {
    return { kind: 'custom', from: period.from, to: period.to }
  }
  return { kind: 'relative', relative: period.relative }
}

export const usageByProjectQueryOptions = (
  orgId: string,
  period: UsagePeriod,
) =>
  queryOptions({
    queryKey: billingKeys.usageByProject(orgId, periodToKey(period)),
    queryFn: async () => {
      const qs = periodToQueryString(period)
      const resp = await rawFetch(
        `/api/v1/organizations/${orgId}/usage/by-project${qs ? `?${qs}` : ''}`,
        { method: 'GET' },
      )
      return (await resp.json()) as BillableUsageSummary[]
    },
    staleTime: 60 * 1000,
  })
