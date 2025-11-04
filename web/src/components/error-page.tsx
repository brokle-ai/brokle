'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/ui/button'
import { ArrowLeft, Home } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

interface ErrorPageProps {
  /** HTTP status code */
  statusCode: number
  /** Error title */
  title: string
  /** Error description */
  description: string
  /** Icon component to display */
  icon: LucideIcon
  /** Show back button (default: true) */
  showBackButton?: boolean
  /** Show home button (default: true) */
  showHomeButton?: boolean
  /** Custom action button */
  customAction?: {
    label: string
    href: string
  }
}

export function ErrorPage({
  statusCode,
  title,
  description,
  icon: Icon,
  showBackButton = true,
  showHomeButton = true,
  customAction,
}: ErrorPageProps) {
  const router = useRouter()

  const handleGoBack = () => {
    // Check if user has history to go back to
    if (window.history.length > 1) {
      router.back()
    } else {
      // Fallback to home if no history
      router.push('/')
    }
  }

  return (
    <div className="flex min-h-svh flex-col items-center justify-center px-4 py-12">
      <div className="mx-auto max-w-md text-center">
        {/* Icon */}
        <div className="mb-6 flex justify-center">
          <div className="rounded-full bg-muted p-6">
            <Icon className="h-12 w-12 text-muted-foreground" />
          </div>
        </div>

        {/* Status Code */}
        <h1 className="mb-2 text-6xl font-bold tracking-tight text-foreground">
          {statusCode}
        </h1>

        {/* Title */}
        <h2 className="mb-4 text-2xl font-semibold tracking-tight text-foreground">
          {title}
        </h2>

        {/* Description */}
        <p className="mb-8 text-muted-foreground">
          {description}
        </p>

        {/* Action Buttons */}
        <div className="flex flex-col gap-3 sm:flex-row sm:justify-center">
          {showBackButton && (
            <Button
              variant="outline"
              onClick={handleGoBack}
              className="gap-2"
            >
              <ArrowLeft className="h-4 w-4" />
              Go Back
            </Button>
          )}

          {showHomeButton && (
            <Button asChild>
              <Link href="/" className="gap-2">
                <Home className="h-4 w-4" />
                Go to Dashboard
              </Link>
            </Button>
          )}

          {customAction && (
            <Button asChild variant="secondary">
              <Link href={customAction.href}>
                {customAction.label}
              </Link>
            </Button>
          )}
        </div>
      </div>
    </div>
  )
}
