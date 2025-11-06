// Test data factories for creating mock objects

export function createMockUser(overrides = {}) {
  return {
    id: '1',
    email: 'test@example.com',
    name: 'Test User',
    firstName: 'Test',
    lastName: 'User',
    role: 'user' as const,
    organizationId: '1',
    defaultOrganizationId: '1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    isEmailVerified: true,
    onboardingCompletedAt: new Date().toISOString(),
    ...overrides,
  }
}

export function createMockOrganization(overrides = {}) {
  return {
    id: '1',
    name: 'Test Organization',
    slug: 'test-org',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    ...overrides,
  }
}

export function createMockProject(overrides = {}) {
  return {
    id: '1',
    name: 'Test Project',
    slug: 'test-project',
    organizationId: '1',
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString(),
    ...overrides,
  }
}
