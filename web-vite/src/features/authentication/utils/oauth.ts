import { getRuntimeConfig } from '@/lib/config'

// Build an OAuth kick-off URL against the backend. `provider` is
// "google" or "github" — the Huma routes mounted under
// `/api/v1/auth/<provider>` 302-redirect to the provider's consent
// screen. An optional `invitation_token` is forwarded so that the
// post-consent callback lands the user on a signup flow tied to the
// pending invitation.
export function buildOAuthUrl(
  provider: 'google' | 'github',
  invitationToken?: string,
): string {
  const base = getRuntimeConfig().API_URL || ''
  const url = new URL(`${base || window.location.origin}/api/v1/auth/${provider}`)
  if (invitationToken) {
    url.searchParams.set('invitation_token', invitationToken)
  }
  return url.toString()
}
