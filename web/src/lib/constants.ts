// API Constants
export const API_ENDPOINTS = {
  // Authentication
  AUTH: {
    SIGNIN: '/api/v1/auth/signin',
    SIGNUP: '/api/v1/auth/signup',
    SIGNOUT: '/api/v1/auth/signout',
    REFRESH: '/api/v1/sessions/refresh',
    ME: '/api/v1/auth/me',
    PASSWORD_RESET: '/api/v1/auth/password-reset',
    TWO_FACTOR: '/api/v1/auth/2fa',
  },
  
  // Organizations
  ORGANIZATIONS: {
    CURRENT: '/api/v1/organizations/current',
    INVITE: '/api/v1/organizations/invite',
    MEMBERS: '/api/v1/organizations/members',
  },
  
  // Projects
  PROJECTS: {
    LIST: '/api/v1/projects',
    CREATE: '/api/v1/projects',
    DETAIL: (id: string) => `/api/v1/projects/${id}`,
    API_KEYS: (id: string) => `/api/v1/projects/${id}/api-keys`,
  },
  
  // AI Analytics
  ANALYTICS: {
    DASHBOARD: '/api/v1/analytics/dashboard',
    REQUESTS: '/api/v1/analytics/requests',
    COSTS: '/api/v1/analytics/costs',
    PROVIDERS: '/api/v1/analytics/providers',
    MODELS: '/api/v1/analytics/models',
  },
  
  // AI Models and Providers
  AI: {
    PROVIDERS: '/api/v1/ai/providers',
    MODELS: '/api/v1/ai/models',
    REQUESTS: '/api/v1/ai/requests',
    ROUTING: '/api/v1/ai/routing',
  },
} as const

// AI Provider Constants
export const AI_PROVIDERS = {
  OPENAI: {
    id: 'openai',
    name: 'OpenAI',
    models: ['gpt-4', 'gpt-4-turbo', 'gpt-3.5-turbo', 'text-embedding-ada-002'],
  },
  ANTHROPIC: {
    id: 'anthropic',
    name: 'Anthropic',
    models: ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'],
  },
  GOOGLE: {
    id: 'google',
    name: 'Google AI',
    models: ['gemini-pro', 'gemini-pro-vision', 'text-embedding-004'],
  },
  COHERE: {
    id: 'cohere',
    name: 'Cohere',
    models: ['command-r', 'command-r-plus', 'embed-english-v3.0'],
  },
} as const

// Time Range Constants
export const TIME_RANGES = {
  '1h': { label: 'Last Hour', value: '1h', duration: 1000 * 60 * 60 },
  '24h': { label: 'Last 24 Hours', value: '24h', duration: 1000 * 60 * 60 * 24 },
  '7d': { label: 'Last 7 Days', value: '7d', duration: 1000 * 60 * 60 * 24 * 7 },
  '30d': { label: 'Last 30 Days', value: '30d', duration: 1000 * 60 * 60 * 24 * 30 },
  '90d': { label: 'Last 90 Days', value: '90d', duration: 1000 * 60 * 60 * 24 * 90 },
} as const

// User Roles and Permissions
export const ROLES = {
  ORGANIZATION: {
    OWNER: 'owner',
    ADMIN: 'admin',
    DEVELOPER: 'developer',
    VIEWER: 'viewer',
  },
  USER: {
    USER: 'user',
    ADMIN: 'admin',
    SUPER_ADMIN: 'super_admin',
  },
} as const

export const PERMISSIONS = {
  AUTH: {
    READ: 'auth:read',
    WRITE: 'auth:write',
  },
  ANALYTICS: {
    READ: 'analytics:read',
    WRITE: 'analytics:write',
  },
  MODELS: {
    READ: 'models:read',
    WRITE: 'models:write',
  },
  COSTS: {
    READ: 'costs:read',
    WRITE: 'costs:write',
  },
  SETTINGS: {
    READ: 'settings:read',
    WRITE: 'settings:write',
  },
} as const

// Subscription Plans
export const SUBSCRIPTION_PLANS = {
  FREE: {
    id: 'free',
    name: 'Free',
    requestLimit: 10000,
    features: ['basic_analytics', 'email_support'],
  },
  PRO: {
    id: 'pro',
    name: 'Pro',
    requestLimit: 100000,
    features: ['advanced_analytics', 'priority_support', 'intelligent_routing'],
  },
  BUSINESS: {
    id: 'business',
    name: 'Business',
    requestLimit: 1000000,
    features: ['predictive_analytics', 'custom_dashboards', 'team_collaboration'],
  },
  ENTERPRISE: {
    id: 'enterprise',
    name: 'Enterprise',
    requestLimit: -1, // Unlimited
    features: ['custom_integrations', 'compliance', 'dedicated_support'],
  },
} as const

// UI Constants
export const CHART_COLORS = {
  PRIMARY: '#3b82f6',
  SUCCESS: '#10b981',
  WARNING: '#f59e0b',
  DANGER: '#ef4444',
  INFO: '#8b5cf6',
  GRAY: '#6b7280',
} as const

export const CHART_TYPES = {
  LINE: 'line',
  BAR: 'bar',
  PIE: 'pie',
  DONUT: 'donut',
  AREA: 'area',
  SCATTER: 'scatter',
} as const

// Status Constants
export const REQUEST_STATUS = {
  PENDING: 'pending',
  COMPLETED: 'completed',
  FAILED: 'failed',
  TIMEOUT: 'timeout',
  RATE_LIMITED: 'rate_limited',
} as const

export const PROVIDER_STATUS = {
  ACTIVE: 'active',
  DEGRADED: 'degraded',
  DOWN: 'down',
  MAINTENANCE: 'maintenance',
} as const

// Error Codes
export const ERROR_CODES = {
  // Authentication
  INVALID_CREDENTIALS: 'INVALID_CREDENTIALS',
  TOKEN_EXPIRED: 'TOKEN_EXPIRED',
  UNAUTHORIZED: 'UNAUTHORIZED',
  
  // Rate Limiting
  RATE_LIMIT_EXCEEDED: 'RATE_LIMIT_EXCEEDED',
  QUOTA_EXCEEDED: 'QUOTA_EXCEEDED',
  
  // Validation
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  INVALID_REQUEST: 'INVALID_REQUEST',
  
  // Network
  NETWORK_ERROR: 'NETWORK_ERROR',
  TIMEOUT: 'TIMEOUT',
  
  // Server
  INTERNAL_ERROR: 'INTERNAL_ERROR',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
} as const

// Notification Types
export const NOTIFICATION_TYPES = {
  SUCCESS: 'success',
  ERROR: 'error',
  WARNING: 'warning',
  INFO: 'info',
} as const

// Local Storage Keys
export const STORAGE_KEYS = {
  AUTH_TOKEN: 'auth_token',
  REFRESH_TOKEN: 'refresh_token',
  USER_PREFERENCES: 'user_preferences',
  DASHBOARD_LAYOUT: 'dashboard_layout',
  THEME: 'theme',
  SIDEBAR_STATE: 'sidebar_state',
} as const

// Validation Constants
export const VALIDATION = {
  PASSWORD: {
    MIN_LENGTH: 8,
    REQUIRE_UPPERCASE: true,
    REQUIRE_LOWERCASE: true,
    REQUIRE_NUMBERS: true,
    REQUIRE_SYMBOLS: false,
  },
  API_KEY: {
    LENGTH: 64,
    PREFIX: 'bk_',
  },
  ORGANIZATION_SLUG: {
    MIN_LENGTH: 3,
    MAX_LENGTH: 32,
    PATTERN: /^[a-z0-9-]+$/,
  },
  PROJECT_SLUG: {
    MIN_LENGTH: 3,
    MAX_LENGTH: 32,
    PATTERN: /^[a-z0-9-]+$/,
  },
} as const

// Feature Flags
export const FEATURE_FLAGS = {
  REALTIME_ANALYTICS: 'realtime_analytics',
  ADVANCED_ROUTING: 'advanced_routing',
  COST_OPTIMIZATION: 'cost_optimization',
  CUSTOM_MODELS: 'custom_models',
  TEAM_COLLABORATION: 'team_collaboration',
  AUDIT_LOGS: 'audit_logs',
} as const

// Default Values
export const DEFAULTS = {
  PAGINATION_LIMIT: 20,
  REFRESH_INTERVAL: 30000, // 30 seconds
  DEBOUNCE_DELAY: 300, // 300ms
  TOAST_DURATION: 5000, // 5 seconds
  SIDEBAR_WIDTH: 280, // pixels
} as const