'use client'

import { Bot, Calculator, Regex, Copy, Trash2, MoreVertical } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useExperimentWizard } from '../../../context/experiment-wizard-context'
import type { WizardEvaluator } from '../../../types'

interface EvaluatorCardProps {
  evaluator: WizardEvaluator
}

export function EvaluatorCard({ evaluator }: EvaluatorCardProps) {
  const { removeEvaluator, duplicateEvaluator } = useExperimentWizard()

  const getScorerIcon = () => {
    switch (evaluator.scorer_type) {
      case 'llm':
        return <Bot className="h-4 w-4" />
      case 'regex':
        return <Regex className="h-4 w-4" />
      case 'builtin':
      default:
        return <Calculator className="h-4 w-4" />
    }
  }

  const getScorerTypeLabel = () => {
    switch (evaluator.scorer_type) {
      case 'llm':
        return 'LLM'
      case 'regex':
        return 'Regex'
      case 'builtin':
        return 'Builtin'
      default:
        return evaluator.scorer_type
    }
  }

  const getScorerDetails = () => {
    const config = evaluator.scorer_config as Record<string, unknown>
    if (!config) return null

    if ('scorer_name' in config) {
      // Builtin scorer
      return (
        <span className="text-xs text-muted-foreground">
          Scorer: {String(config.scorer_name)}
        </span>
      )
    }

    if ('pattern' in config) {
      // Regex scorer
      return (
        <span className="text-xs text-muted-foreground font-mono">
          Pattern: /{String(config.pattern)}/
        </span>
      )
    }

    if ('model' in config) {
      // LLM scorer
      return (
        <span className="text-xs text-muted-foreground">
          Model: {String(config.model)}
        </span>
      )
    }

    return null
  }

  return (
    <Card>
      <CardContent className="p-4">
        <div className="flex items-start justify-between gap-4">
          <div className="flex items-start gap-3 min-w-0">
            <div className="mt-0.5 rounded-md bg-muted p-2">{getScorerIcon()}</div>
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                <h4 className="font-medium truncate">{evaluator.name}</h4>
                <Badge variant="outline" className="shrink-0">
                  {getScorerTypeLabel()}
                </Badge>
              </div>
              {getScorerDetails()}
            </div>
          </div>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon" className="shrink-0">
                <MoreVertical className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => duplicateEvaluator(evaluator.id)}>
                <Copy className="mr-2 h-4 w-4" />
                Duplicate
              </DropdownMenuItem>
              <DropdownMenuItem
                onClick={() => removeEvaluator(evaluator.id)}
                className="text-destructive"
              >
                <Trash2 className="mr-2 h-4 w-4" />
                Remove
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </CardContent>
    </Card>
  )
}
