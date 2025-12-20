'use client'

import { cn } from '@/lib/utils'
import type { AIProvider } from '../types'

interface ProviderIconProps {
  provider: AIProvider
  className?: string
}

/**
 * Provider icon component with SVG logos
 * Falls back to a generic icon if provider is unknown
 */
export function ProviderIcon({ provider, className }: ProviderIconProps) {
  const iconClass = cn('inline-block', className)

  switch (provider) {
    case 'openai':
      return (
        <svg className={iconClass} viewBox="0 0 24 24" fill="currentColor">
          <path d="M22.2819 9.8211a5.9847 5.9847 0 0 0-.5157-4.9108 6.0462 6.0462 0 0 0-6.5098-2.9A6.0651 6.0651 0 0 0 4.9807 4.1818a5.9847 5.9847 0 0 0-3.9977 2.9 6.0462 6.0462 0 0 0 .7427 7.0966 5.98 5.98 0 0 0 .511 4.9107 6.051 6.051 0 0 0 6.5146 2.9001A5.9847 5.9847 0 0 0 13.2599 24a6.0557 6.0557 0 0 0 5.7718-4.2058 5.9894 5.9894 0 0 0 3.9977-2.9001 6.0557 6.0557 0 0 0-.7475-7.0729zm-9.022 12.6081a4.4755 4.4755 0 0 1-2.8764-1.0408l.1419-.0804 4.7783-2.7582a.7948.7948 0 0 0 .3927-.6813v-6.7369l2.02 1.1686a.071.071 0 0 1 .038.052v5.5826a4.504 4.504 0 0 1-4.4945 4.4944zm-9.6607-4.1254a4.4708 4.4708 0 0 1-.5346-3.0137l.142.0852 4.783 2.7582a.7712.7712 0 0 0 .7806 0l5.8428-3.3685v2.3324a.0804.0804 0 0 1-.0332.0615L9.74 19.9502a4.4992 4.4992 0 0 1-6.1408-1.6464zM2.3408 7.8956a4.485 4.485 0 0 1 2.3655-1.9728V11.6a.7664.7664 0 0 0 .3879.6765l5.8144 3.3543-2.0201 1.1685a.0757.0757 0 0 1-.071 0l-4.8303-2.7865A4.504 4.504 0 0 1 2.3408 7.8956zm16.5963 3.8558L13.1038 8.364 15.1192 7.2a.0757.0757 0 0 1 .071 0l4.8303 2.7913a4.4944 4.4944 0 0 1-.6765 8.1042v-5.6772a.79.79 0 0 0-.407-.667zm2.0107-3.0231l-.142-.0852-4.7735-2.7818a.7759.7759 0 0 0-.7854 0L9.409 9.2297V6.8974a.0662.0662 0 0 1 .0284-.0615l4.8303-2.7866a4.4992 4.4992 0 0 1 6.6802 4.66zM8.3065 12.863l-2.02-1.1638a.0804.0804 0 0 1-.038-.0567V6.0742a4.4992 4.4992 0 0 1 7.3757-3.4537l-.142.0805L8.704 5.459a.7948.7948 0 0 0-.3927.6813zm1.0976-2.3654l2.602-1.4998 2.6069 1.4998v2.9994l-2.5974 1.4997-2.6067-1.4997Z" />
        </svg>
      )

    case 'anthropic':
      return (
        <svg className={iconClass} viewBox="0 0 24 24" fill="currentColor">
          <path d="M17.304 3.541h-3.607l6.696 16.918h3.607l-6.696-16.918zm-10.608 0L0 20.459h3.607l1.357-3.559h6.468l1.357 3.559h3.607L9.7 3.541H6.696zm.542 10.398l2.26-5.924 2.26 5.924H7.238z" />
        </svg>
      )

    case 'azure':
      return (
        <svg className={iconClass} viewBox="0 0 24 24" fill="currentColor">
          <path d="M13.05 4.24 6.56 18.05a.5.5 0 0 1-.37.26H2.57a.5.5 0 0 1-.45-.7L9.1 3.55a.5.5 0 0 1 .45-.3h3.05a.5.5 0 0 1 .45.99zm8.52 13.81-4.7-8.4a.5.5 0 0 0-.87 0l-1.51 2.68 2.87 5.12a.5.5 0 0 0 .44.26h3.32a.5.5 0 0 0 .45-.66zm-5.53-1.73-3.14-5.62a.5.5 0 0 0-.87 0l-3.14 5.62a.5.5 0 0 0 .44.74h6.27a.5.5 0 0 0 .44-.74z" />
        </svg>
      )

    case 'gemini':
      return (
        <svg className={iconClass} viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 24A14.304 14.304 0 0 0 12 0a14.304 14.304 0 0 0 0 24zm0-3.885A10.42 10.42 0 0 1 1.58 12 10.42 10.42 0 0 1 12 1.58 10.42 10.42 0 0 1 22.42 12 10.42 10.42 0 0 1 12 20.115z" />
        </svg>
      )

    case 'openrouter':
      return (
        <svg className={iconClass} viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8zm-1-13h2v6h-2zm0 8h2v2h-2z" />
        </svg>
      )

    case 'custom':
    default:
      return (
        <svg className={iconClass} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
          <line x1="12" y1="8" x2="12" y2="16" />
          <line x1="8" y1="12" x2="16" y2="12" />
        </svg>
      )
  }
}
