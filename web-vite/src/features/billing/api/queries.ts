import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  CreateBudgetRequest,
  EffectivePricing,
  InvoiceListResponse,
  UpdateBudgetRequest,
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
} as const

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
