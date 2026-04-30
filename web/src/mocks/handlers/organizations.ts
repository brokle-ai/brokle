import { http, HttpResponse } from 'msw'

// List endpoints use the canonical `{data, pagination}` shape (CLAUDE.md
// gotcha #23). Resources are raw on GET-by-id.

const MOCK_ORG = {
  id: 'org_test_0000000000000000000000',
  name: 'Test Organization',
  slug: 'test-org',
  billing_email: 'billing@brokle.test',
  subscription_plan: 'free' as const,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

export const organizationHandlers = [
  http.get('*/v1/organizations', () => {
    return HttpResponse.json({
      data: [MOCK_ORG],
      pagination: {
        page: 1,
        limit: 20,
        total: 1,
        total_pages: 1,
        has_next: false,
        has_prev: false,
      },
    })
  }),

  http.get('*/v1/organizations/:orgId', ({ params }) => {
    return HttpResponse.json({ ...MOCK_ORG, id: params.orgId as string })
  }),
]
