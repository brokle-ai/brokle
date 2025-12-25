'use client'

import { useState } from 'react'
import Link from 'next/link'
import { formatDistanceToNow } from 'date-fns'
import { Scale, MoreVertical, Trash2, Pencil, Loader2, Play, Pause } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog'
import { RuleStatusBadge } from './rule-status-badge'
import { RuleScorerBadge } from './rule-scorer-badge'
import {
  useDeleteEvaluationRuleMutation,
  useActivateEvaluationRuleMutation,
  useDeactivateEvaluationRuleMutation,
} from '../hooks/use-evaluation-rules'
import type { EvaluationRule } from '../types'

interface RuleCardProps {
  rule: EvaluationRule
  projectId: string
  projectSlug: string
  onEdit?: (rule: EvaluationRule) => void
}

export function RuleCard({ rule, projectId, projectSlug, onEdit }: RuleCardProps) {
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)
  const deleteMutation = useDeleteEvaluationRuleMutation(projectId)
  const activateMutation = useActivateEvaluationRuleMutation(projectId)
  const deactivateMutation = useDeactivateEvaluationRuleMutation(projectId)

  const handleDelete = async () => {
    await deleteMutation.mutateAsync({
      ruleId: rule.id,
      ruleName: rule.name,
    })
    setShowDeleteDialog(false)
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
    <>
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
              <DropdownMenuItem
                onClick={handleToggleStatus}
                disabled={isToggling}
              >
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
              {onEdit && (
                <DropdownMenuItem onClick={() => onEdit(rule)}>
                  <Pencil className="mr-2 h-4 w-4" />
                  Edit
                </DropdownMenuItem>
              )}
              <DropdownMenuSeparator />
              <DropdownMenuItem
                className="text-destructive focus:text-destructive"
                onClick={() => setShowDeleteDialog(true)}
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

      <AlertDialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Evaluation Rule</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{rule.name}&quot;? This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
