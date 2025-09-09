import { z } from 'zod'

// Environment variable schema
const envSchema = z.object({
  // API Configuration
  NEXT_PUBLIC_API_URL: z.string().url().default('http://localhost:8080'),
  NEXT_PUBLIC_WS_URL: z.string().url().optional(),
  
  // App Configuration
  NEXT_PUBLIC_APP_NAME: z.string().default('Brokle Dashboard'),
  NEXT_PUBLIC_APP_VERSION: z.string().default('1.0.0'),
  NEXT_PUBLIC_APP_ENV: z.enum(['development', 'staging', 'production']).default('development'),
  
  // Features
  NEXT_PUBLIC_ENABLE_ANALYTICS: z.string().transform(val => val === 'true').default('true'),
  NEXT_PUBLIC_ENABLE_REALTIME: z.string().transform(val => val === 'true').default('true'),
  NEXT_PUBLIC_ENABLE_NOTIFICATIONS: z.string().transform(val => val === 'true').default('true'),
  
  // External Services
  NEXT_PUBLIC_SENTRY_DSN: z.string().optional(),
  NEXT_PUBLIC_POSTHOG_KEY: z.string().optional(),
  NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY: z.string().optional(),
})

// Parse and validate environment variables
const env = envSchema.parse({
  NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL,
  NEXT_PUBLIC_WS_URL: process.env.NEXT_PUBLIC_WS_URL,
  NEXT_PUBLIC_APP_NAME: process.env.NEXT_PUBLIC_APP_NAME,
  NEXT_PUBLIC_APP_VERSION: process.env.NEXT_PUBLIC_APP_VERSION,
  NEXT_PUBLIC_APP_ENV: process.env.NEXT_PUBLIC_APP_ENV,
  NEXT_PUBLIC_ENABLE_ANALYTICS: process.env.NEXT_PUBLIC_ENABLE_ANALYTICS,
  NEXT_PUBLIC_ENABLE_REALTIME: process.env.NEXT_PUBLIC_ENABLE_REALTIME,
  NEXT_PUBLIC_ENABLE_NOTIFICATIONS: process.env.NEXT_PUBLIC_ENABLE_NOTIFICATIONS,
  NEXT_PUBLIC_SENTRY_DSN: process.env.NEXT_PUBLIC_SENTRY_DSN,
  NEXT_PUBLIC_POSTHOG_KEY: process.env.NEXT_PUBLIC_POSTHOG_KEY,
  NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY: process.env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY,
})

// Application configuration
export const config = {
  app: {
    name: env.NEXT_PUBLIC_APP_NAME,
    version: env.NEXT_PUBLIC_APP_VERSION,
    environment: env.NEXT_PUBLIC_APP_ENV,
    isDevelopment: env.NEXT_PUBLIC_APP_ENV === 'development',
    isProduction: env.NEXT_PUBLIC_APP_ENV === 'production',
  },
  
  api: {
    baseUrl: env.NEXT_PUBLIC_API_URL,
    wsUrl: env.NEXT_PUBLIC_WS_URL || env.NEXT_PUBLIC_API_URL.replace('http', 'ws'),
    timeout: 30000, // 30 seconds
    retryAttempts: 3,
    retryDelay: 1000, // 1 second
  },
  
  features: {
    analytics: env.NEXT_PUBLIC_ENABLE_ANALYTICS,
    realtime: env.NEXT_PUBLIC_ENABLE_REALTIME,
    notifications: env.NEXT_PUBLIC_ENABLE_NOTIFICATIONS,
  },
  
  external: {
    sentry: {
      dsn: env.NEXT_PUBLIC_SENTRY_DSN,
      enabled: !!env.NEXT_PUBLIC_SENTRY_DSN,
    },
    posthog: {
      key: env.NEXT_PUBLIC_POSTHOG_KEY,
      enabled: !!env.NEXT_PUBLIC_POSTHOG_KEY,
    },
    stripe: {
      publishableKey: env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY,
      enabled: !!env.NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY,
    },
  },
  
  ai: {
    supportedProviders: [
      'openai',
      'anthropic', 
      'google',
      'cohere',
      'huggingface',
      'azure-openai',
      'aws-bedrock'
    ] as const,
    defaultProvider: 'openai' as const,
    maxRetries: 3,
    timeoutMs: 30000,
  },
  
  dashboard: {
    defaultRefreshInterval: 30000, // 30 seconds
    maxRealtimeRequests: 100,
    defaultTimeRange: '24h' as const,
    supportedTimeRanges: ['1h', '24h', '7d', '30d', '90d'] as const,
  },
  
  ui: {
    themes: ['light', 'dark', 'system'] as const,
    fonts: ['inter', 'manrope'] as const,
    defaultTheme: 'system' as const,
    defaultFont: 'inter' as const,
  },
  
  storage: {
    keys: {
      auth: 'brokle-auth-storage',
      ui: 'brokle-ui-storage',
      preferences: 'brokle-preferences',
    },
  },
  
  limits: {
    fileUpload: {
      maxSize: 10 * 1024 * 1024, // 10MB
      allowedTypes: ['image/jpeg', 'image/png', 'image/gif', 'image/webp'],
    },
    pagination: {
      defaultLimit: 20,
      maxLimit: 100,
    },
  },
} as const

// Type for the configuration
export type Config = typeof config

// Export individual sections for convenience
export const { app, api, features, external, ai, dashboard, ui, storage, limits } = config

// Validation helper
export function validateConfig(): void {
  try {
    envSchema.parse(process.env)
  } catch (error) {
    console.error('Invalid environment configuration:', error)
    throw new Error('Invalid environment configuration')
  }
}

// Get a config value with type safety
export function getConfig<T extends keyof Config>(section: T): Config[T] {
  return config[section]
}

// Check if a feature is enabled
export function isFeatureEnabled(feature: keyof typeof features): boolean {
  return features[feature]
}

// Get environment-specific values
export function getEnvironmentValue<T>(
  development: T,
  staging: T,
  production: T
): T {
  switch (config.app.environment) {
    case 'development':
      return development
    case 'staging':
      return staging
    case 'production':
      return production
    default:
      return development
  }
}