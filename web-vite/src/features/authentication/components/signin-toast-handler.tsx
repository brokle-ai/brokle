import { useEffect, useRef } from 'react'
import { toast } from 'sonner'

interface SignInToastHandlerProps {
  logout?: string
  session?: string
}

// One-shot toast surface for the signin page. web/ reads these from
// `useSearchParams()`; here the caller extracts them from the TanStack
// Router search object and passes them in. After firing, we scrub
// `logout` / `session` out of the URL so a refresh doesn't re-toast.
export function SignInToastHandler({ logout, session }: SignInToastHandlerProps) {
  const hasShownToast = useRef(false)

  useEffect(() => {
    if (hasShownToast.current) return
    if (!logout && !session) return
    hasShownToast.current = true

    if (logout === 'success') {
      toast.success('Logged out successfully', {
        description: 'You have been securely logged out.',
      })
    } else if (session === 'ended') {
      toast.info('Session ended', {
        description: 'You have been logged out from another tab.',
      })
    } else if (session === 'expired') {
      toast.error('Session expired', {
        description: 'Your session has expired. Please log in again.',
      })
    } else if (logout === 'error') {
      toast.warning('Logged out locally', {
        description: 'Session cleared locally.',
      })
    }

    const url = new URL(window.location.href)
    url.searchParams.delete('logout')
    url.searchParams.delete('session')
    window.history.replaceState({}, '', url.pathname + url.search)
  }, [logout, session])

  return null
}
