import { http, HttpResponse } from 'msw'

export const handlers = [
  // Auth endpoints
  http.post('/api/v1/auth/login', () => {
    return HttpResponse.json({
      user: {
        id: '1',
        email: 'test@example.com',
        name: 'Test User',
        role: 'user',
      },
      token: 'mock-token',
    })
  }),

  http.post('/api/v1/auth/logout', () => {
    return HttpResponse.json({ message: 'Logged out successfully' })
  }),

  http.post('/api/v1/auth/refresh', () => {
    return HttpResponse.json({ token: 'new-mock-token' })
  }),

  // Organizations
  http.get('/api/v1/organizations', () => {
    return HttpResponse.json({
      data: [
        {
          id: '1',
          name: 'Test Organization',
          slug: 'test-org',
        },
      ],
    })
  }),

  // Projects
  http.get('/api/v1/projects', () => {
    return HttpResponse.json({
      data: [
        {
          id: '1',
          name: 'Test Project',
          slug: 'test-project',
        },
      ],
    })
  }),
]
