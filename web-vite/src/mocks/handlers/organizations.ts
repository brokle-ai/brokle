import { http, HttpResponse } from 'msw'

const MOCK_ORG = {
  id: 'org_test_0000000000000000000000',
  name: 'Test Organization',
  slug: 'test-org',
  billing_email: 'billing@brokle.test',
  subscription_plan: 'free' as const,
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

const MOCK_PROJECT = {
  id: 'prj_test_0000000000000000000000',
  organization_id: MOCK_ORG.id,
  name: 'Test Project',
  slug: 'test-project',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

const listShape = <T>(data: T[]) => ({
  data,
  pagination: {
    page: 1,
    limit: 20,
    total: data.length,
    total_pages: 1,
    has_next: false,
    has_prev: false,
  },
})

export const organizationHandlers = [
  http.get('*/v1/organizations', () => HttpResponse.json(listShape([MOCK_ORG]))),
  http.get('*/v1/organizations/:orgId', ({ params }) =>
    HttpResponse.json({ ...MOCK_ORG, id: params.orgId as string }),
  ),
  http.get('*/v1/organizations/:orgId/projects', () =>
    HttpResponse.json(listShape([MOCK_PROJECT])),
  ),
]
