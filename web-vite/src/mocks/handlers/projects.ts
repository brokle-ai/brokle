import { http, HttpResponse } from 'msw'

const MOCK_PROJECT = {
  id: 'prj_test_0000000000000000000000',
  organization_id: 'org_test_0000000000000000000000',
  name: 'Test Project',
  slug: 'test-project',
  created_at: new Date().toISOString(),
  updated_at: new Date().toISOString(),
}

export const projectHandlers = [
  http.get('*/v1/projects/:projectId', ({ params }) => {
    return HttpResponse.json({ ...MOCK_PROJECT, id: params.projectId as string })
  }),
]
