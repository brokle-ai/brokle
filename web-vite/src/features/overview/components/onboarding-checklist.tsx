import { useState, type ComponentType } from 'react'
import { useNavigate } from '@tanstack/react-router'
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
import type { ChecklistStatus } from '../api/types'

interface ChecklistItem {
  id: keyof ChecklistStatus
  label: string
  description: string
  icon: ComponentType<{ className?: string }>
  externalHref?: string
  internalTo?: string
  actionLabel?: string
}

const checklistItems: ChecklistItem[] = [
  {
    id: 'has_project',
    label: 'Create project',
    description: "You're already here - your project is set up!",
    icon: Rocket,
  },
  {
    id: 'has_traces',
    label: 'Send first trace',
    description:
      'Integrate the Brokle SDK to start sending traces from your AI application.',
    icon: Activity,
    externalHref: 'https://docs.brokle.ai/quickstart',
    actionLabel: 'View Docs',
  },
  {
    id: 'has_ai_provider',
    label: 'Configure AI provider',
    description:
      'Add your AI provider credentials to enable playground and evaluations.',
    icon: BrainCircuit,
    internalTo: 'ai-providers',
    actionLabel: 'Add Provider',
  },
  {
    id: 'has_evaluations',
    label: 'Set up evaluations',
    description: 'Create score configurations to track quality metrics.',
    icon: Target,
    internalTo: 'scores',
    actionLabel: 'Create Score',
  },
]

interface OnboardingChecklistProps {
  checklistStatus: ChecklistStatus | null
  orgId: string
  projectId: string
  onDismiss?: () => void
  className?: string
}

export function OnboardingChecklist({
  checklistStatus,
  orgId,
  projectId,
  onDismiss,
  className,
}: OnboardingChecklistProps) {
  const navigate = useNavigate()
  const [isMinimized, setIsMinimized] = useState(false)

  if (!checklistStatus) return null

  const total = checklistItems.length
  const completed = checklistItems.filter((i) => checklistStatus[i.id]).length
  const percentage = total === 0 ? 0 : Math.round((completed / total) * 100)

  if (completed === total) return null

  const handleItemClick = (item: ChecklistItem) => {
    if (item.externalHref) {
      window.open(item.externalHref, '_blank')
      return
    }
    if (item.internalTo === 'scores') {
      navigate({
        to: '/o/$orgId/p/$projectId/scores',
        params: { orgId, projectId },
        search: {
          page: 1,
          limit: 20,
          name: undefined,
          source: undefined,
          type: undefined,
          traceId: undefined,
          spanId: undefined,
        },
      })
    } else if (item.internalTo === 'ai-providers') {
      // TODO(port): AI providers settings screen not yet ported to web-vite.
      navigate({
        to: '/o/$orgId/p/$projectId/settings',
        params: { orgId, projectId },
      })
    }
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
        Setup: {completed}/{total}
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
          Complete these steps to unlock the full potential of your AI
          observability.
        </CardDescription>
        <div className="flex items-center gap-3 mt-3">
          <Progress value={percentage} className="h-2" />
          <span className="text-sm text-muted-foreground whitespace-nowrap">
            {completed}/{total}
          </span>
        </div>
      </CardHeader>
      <CardContent className="grid gap-3 pt-0">
        {checklistItems.map((item) => {
          const done = checklistStatus[item.id]
          const Icon = item.icon
          return (
            <div
              key={item.id}
              className={cn(
                'flex items-start gap-3 p-3 rounded-lg transition-colors',
                done
                  ? 'bg-muted/50'
                  : 'bg-muted/30 hover:bg-muted/50 cursor-pointer',
              )}
              onClick={() => !done && handleItemClick(item)}
              role={done ? undefined : 'button'}
              tabIndex={done ? undefined : 0}
            >
              <div className="mt-0.5">
                {done ? (
                  <CheckCircle2 className="h-5 w-5 text-green-500" />
                ) : (
                  <Circle className="h-5 w-5 text-muted-foreground" />
                )}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2">
                  <Icon
                    className={cn(
                      'h-4 w-4',
                      done ? 'text-green-500' : 'text-muted-foreground',
                    )}
                  />
                  <span
                    className={cn(
                      'font-medium text-sm',
                      done && 'text-muted-foreground line-through',
                    )}
                  >
                    {item.label}
                  </span>
                </div>
                <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                  {item.description}
                </p>
              </div>
              {!done && item.actionLabel && (
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
