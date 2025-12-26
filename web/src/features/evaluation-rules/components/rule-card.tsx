'use client'

import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { Scale, MoreVertical, Trash2, Pencil, Play, Pause } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { RuleStatusBadge } from './rule-status-badge'
import { RuleScorerBadge } from './rule-scorer-badge'
import { useRules } from '../context/rules-context'
import {
  useActivateEvaluationRuleMutation,
  useDeactivateEvaluationRuleMutation,
} from '../hooks/use-evaluation-rules'
import type { EvaluationRule } from '../types'

interface RuleCardProps {
  rule: EvaluationRule
}

export function RuleCard({ rule }: RuleCardProps) {
  const { setOpen, setCurrentRow, projectSlug, projectId } = useRules()
  const activateMutation = useActivateEvaluationRuleMutation(projectId ?? '')
  const deactivateMutation = useDeactivateEvaluationRuleMutation(projectId ?? '')

  const handleEdit = () => {
    setCurrentRow(rule)
    setOpen('edit')
  }

  const handleDelete = () => {
    setCurrentRow(rule)
    setOpen('delete')
  }

  const handleToggleStatus = async () => {
    if (rule.status === 'active') {
      await deactivateMutation.mutateAsync({
        ruleId: rule.id,
        ruleName: rule.name,
      })
    } else {
      await activateMutation.mutateAsync({
        ruleId: rule.id,
        ruleName: rule.name,
      })
    }
  }

  const isToggling = activateMutation.isPending || deactivateMutation.isPending

  return (
    <Card>
      <CardHeader className="flex flex-row items-center justify-between pb-2">
        <CardTitle className="text-lg font-medium">
          <Link
            href={`/projects/${projectSlug}/evaluations/rules/${rule.id}`}
            className="hover:underline"
          >
            {rule.name}
          </Link>
        </CardTitle>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon">
              <MoreVertical className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={handleToggleStatus} disabled={isToggling}>
              {rule.status === 'active' ? (
                <>
                  <Pause className="mr-2 h-4 w-4" />
                  Deactivate
                </>
              ) : (
                <>
                  <Play className="mr-2 h-4 w-4" />
                  Activate
                </>
              )}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleEdit}>
              <Pencil className="mr-2 h-4 w-4" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem
              className="text-destructive focus:text-destructive"
              onClick={handleDelete}
            >
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-muted-foreground mb-4 line-clamp-2">
          {rule.description || 'No description'}
        </p>
        <div className="flex items-center justify-between text-sm">
          <div className="flex items-center gap-2">
            <Scale className="h-4 w-4 text-muted-foreground" />
            <RuleStatusBadge status={rule.status} />
            <RuleScorerBadge scorerType={rule.scorer_type} />
          </div>
          <span className="text-muted-foreground">
            {formatDistanceToNow(new Date(rule.created_at), {
              addSuffix: true,
            })}
          </span>
        </div>
        {rule.span_names?.length > 0 && (
          <div className="mt-2 text-xs text-muted-foreground">
            Matches: {rule.span_names.join(', ')}
          </div>
        )}
        {rule.sampling_rate < 1 && (
          <div className="mt-1 text-xs text-muted-foreground">
            Sampling: {Math.round(rule.sampling_rate * 100)}%
          </div>
        )}
      </CardContent>
    </Card>
  )
}
