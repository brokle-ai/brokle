// Public exports for authentication feature

// Hooks
export { useAuth } from './hooks/use-auth'
export { useAuthGuard } from './hooks/use-auth-guard'
export {
  useCurrentUser,
  useCurrentOrganization,
  useLoginMutation,
  useSignupMutation,
  useCompleteOAuthSignupMutation,
  useUpdateProfileMutation,
  useChangePasswordMutation,
  useRequestPasswordResetMutation,
  useConfirmPasswordResetMutation,
  useLogoutMutation,
  authQueryKeys,
} from './hooks/use-auth-queries'

// Components
export { SignInForm } from './components/sign-in-form'
export { SignUpForm } from './components/sign-up-form'
export { TwoStepSignUpForm } from './components/two-step-signup-form'
export { ForgotPasswordForm } from './components/forgot-password-form'
export { OtpForm as OTPForm } from './components/otp-form'
export { AuthGuard } from './components/auth-guard'
export { UnauthorizedFallback } from './components/unauthorized-fallback'
export { LogoutButton } from './components/logout-button'
export { AuthStatus } from './components/auth-status'
export { AuthLayout } from './components/auth-layout'
export { AuthFormWrapper } from './components/auth-form-wrapper'
export { SignInToastHandler } from './components/signin-toast-handler'

// Store
export { useAuthStore } from './stores/auth-store'

// API Functions
export { exchangeLoginSession } from './api/auth-api'

// Types
export type {
  User,
  UserRole,
  AuthState,
  LoginCredentials,
  SignUpCredentials,
  AuthResponse,
  Organization,
  OrganizationRole,
  OrganizationMember,
  OrganizationWithProjects,
  Project,
  ProjectEnvironment,
  ApiKey,
  SubscriptionPlan,
  LoginResponse,
  UserResponse,
  RefreshTokenRequest,
  InvitationDetails,
  Permission,
  // Re-exported from organizations/types (canonical source)
  ProjectSummary,
  ProjectMetrics,
  ProjectStatus,
  ProjectSettings,
  RoutingPreferences,
  UsageStats,
  AuthTokens,
} from './types'
