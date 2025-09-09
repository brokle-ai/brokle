// Auth queries
export {
  useCurrentUser,
  useCurrentOrganization,
  useLoginMutation,
  useSignupMutation,
  useLogoutMutation,
  useUpdateProfileMutation,
  useChangePasswordMutation,
  useRequestPasswordResetMutation,
  useConfirmPasswordResetMutation,
  authQueryKeys,
} from './use-auth-queries'

// Protected queries
export {
  useProtectedQuery,
  useProtectedMutation,
  useOptimisticMutation,
  useAutoRefreshQuery,
  usePaginatedQuery,
} from './use-protected-query'