'use client'

import * as React from 'react'
import { useRouter } from 'next/navigation'
import {
  CheckCircle2,
  Circle,
  ChevronDown,
  ChevronUp,
  Rocket,
  Activity,
  BrainCircuit,
  Target,
  X,
} from 'lucide-react'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'
import type { ChecklistStatus } from '../types'

interface ChecklistItem {
  id: keyof ChecklistStatus
  label: string
  description: string
  icon: React.ComponentType<{ className?: string }>
  href?: string
  actionLabel?: string
}

const checklistItems: ChecklistItem[] = [
  {
    id: 'has_project',
    label: 'Create project',
    description: 'You\'re already here - your project is set up!',
    icon: Rocket,
  },
  {
    id: 'has_traces',
    label: 'Send first trace',
    description: 'Integrate the Brokle SDK to start sending traces from your AI application.',
    icon: Activity,
    href: 'https://docs.brokle.ai/quickstart',
    actionLabel: 'View Docs',
  },
  {
    id: 'has_ai_provider',
    label: 'Configure AI provider',
    description: 'Add your AI provider credentials to enable playground and evaluations.',
    icon: BrainCircuit,
    href: '/settings/ai-providers',
    actionLabel: 'Add Provider',
  },
  {
    id: 'has_evaluations',
    label: 'Set up evaluations',
    description: 'Create score configurations to track quality metrics.',
    icon: Target,
    href: '/evaluations/scores',
    actionLabel: 'Create Score',
  },
]

interface OnboardingChecklistProps {
  checklistStatus: ChecklistStatus | null
  onboardingProgress: {
    completed: number
    total: number
    percentage: number
  }
  className?: string
  onDismiss?: () => void
  projectSlug?: string
}

export function OnboardingChecklist({
  checklistStatus,
  onboardingProgress,
  className,
  onDismiss,
  projectSlug,
}: OnboardingChecklistProps) {
  const router = useRouter()
  const [isMinimized, setIsMinimized] = React.useState(false)

  // Don't render if all items are complete
  if (onboardingProgress.completed === onboardingProgress.total) {
    return null
  }

  if (!checklistStatus) {
    return null
  }

  const handleItemClick = (item: ChecklistItem) => {
    if (item.href) {
      if (item.href.startsWith('http')) {
        window.open(item.href, '_blank')
      } else if (projectSlug) {
        router.push(`/projects/${projectSlug}${item.href}`)
      }
    }
  }

  const isComplete = (id: keyof ChecklistStatus) => {
    return checklistStatus[id]
  }

  if (isMinimized) {
    return (
      <Button
        variant="outline"
        size="sm"
        className="gap-2"
        onClick={() => setIsMinimized(false)}
      >
        <Rocket className="h-4 w-4" />
        Setup: {onboardingProgress.completed}/{onboardingProgress.total}
        <ChevronDown className="h-4 w-4" />
      </Button>
    )
  }

  return (
    <Card className={cn('relative', className)}>
      <CardHeader className="pb-3">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Rocket className="h-5 w-5 text-primary" />
            <CardTitle className="text-lg">Get Started with Brokle</CardTitle>
          </div>
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => setIsMinimized(true)}
            >
              <ChevronUp className="h-4 w-4" />
            </Button>
            {onDismiss && (
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={onDismiss}
              >
                <X className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
        <CardDescription>
          Complete these steps to unlock the full potential of your AI observability.
        </CardDescription>
        <div className="flex items-center gap-3 mt-3">
          <Progress value={onboardingProgress.percentage} className="h-2" />
          <span className="text-sm text-muted-foreground whitespace-nowrap">
            {onboardingProgress.completed}/{onboardingProgress.total}
          </span>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 pt-0">
        {checklistItems.map((item) => {
          const completed = isComplete(item.id)
          const Icon = item.icon

          return (
            <div
              key={item.id}
              className={cn(
                'flex items-start gap-3 p-3 rounded-lg transition-colors',
                completed
                  ? 'bg-muted/50'
                  : 'bg-muted/30 hover:bg-muted/50 cursor-pointer',
              )}
              onClick={() => !completed && handleItemClick(item)}
              role={completed ? undefined : 'button'}
              tabIndex={completed ? undefined : 0}
            >
              <div className="mt-0.5">
                {completed ? (
                  <CheckCircle2 className="h-5 w-5 text-green-500" />
                ) : (
                  <Circle className="h-5 w-5 text-muted-foreground" />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <Icon className={cn(
                    'h-4 w-4',
                    completed ? 'text-green-500' : 'text-muted-foreground'
                  )} />
                  <span className={cn(
                    'font-medium text-sm',
                    completed && 'text-muted-foreground line-through'
                  )}>
                    {item.label}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                  {item.description}
                </p>
              </div>
              {!completed && item.actionLabel && (
                <Button size="sm" variant="secondary" className="shrink-0">
                  {item.actionLabel}
                </Button>
              )}
            </div>
          )
        })}
      </CardContent>
    </Card>
  )
}
