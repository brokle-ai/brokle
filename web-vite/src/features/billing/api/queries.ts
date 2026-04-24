import { queryOptions } from '@tanstack/react-query'
import { rawFetch } from '@/lib/api/client'
import type {
  EffectivePricing,
  InvoiceListResponse,
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
