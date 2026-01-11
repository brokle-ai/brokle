'use client'

import * as React from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { ChevronDown, Loader2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { cn } from '@/lib/utils'
import type { UsageBudget, CreateBudgetRequest, BudgetType } from '../types'

const budgetFormSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100),
  budget_type: z.enum(['monthly', 'weekly']),
  project_id: z.string().optional(),
  span_limit: z.number().positive().optional().nullable(),
  bytes_limit: z.number().positive().optional().nullable(),
  score_limit: z.number().positive().optional().nullable(),
  cost_limit: z.number().positive().optional().nullable(),
  alert_thresholds: z.array(z.number().min(1).max(100)),
}).refine(
  (data) => data.span_limit != null || data.bytes_limit != null ||
            data.score_limit != null || data.cost_limit != null,
  {
    message: 'Set a cost limit or expand Advanced Limits to set dimension-specific limits',
    path: ['limits'],
  }
)

type BudgetFormValues = z.infer<typeof budgetFormSchema>

interface BudgetFormProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onSubmit: (data: CreateBudgetRequest) => Promise<void>
  budget?: UsageBudget
  projects?: { id: string; name: string }[]
  isLoading?: boolean
}

const PRESET_THRESHOLDS = [25, 50, 75, 80, 90, 100]
const DEFAULT_THRESHOLDS = [50, 80, 100]

export function BudgetForm({
  open,
  onOpenChange,
  onSubmit,
  budget,
  projects = [],
  isLoading,
}: BudgetFormProps) {
  const isEditing = !!budget
  const [customThreshold, setCustomThreshold] = React.useState('')
  const [advancedOpen, setAdvancedOpen] = React.useState(false)

  const form = useForm<BudgetFormValues>({
    resolver: zodResolver(budgetFormSchema),
    defaultValues: {
      name: budget?.name ?? '',
      budget_type: budget?.budget_type ?? 'monthly',
      project_id: budget?.project_id ?? undefined,
      span_limit: budget?.span_limit ?? null,
      bytes_limit: budget?.bytes_limit ?? null,
      score_limit: budget?.score_limit ?? null,
      cost_limit: budget?.cost_limit ?? null,
      alert_thresholds: budget?.alert_thresholds ?? DEFAULT_THRESHOLDS,
    },
  })

  React.useEffect(() => {
    if (open) {
      form.reset({
        name: budget?.name ?? '',
        budget_type: budget?.budget_type ?? 'monthly',
        project_id: budget?.project_id ?? undefined,
        span_limit: budget?.span_limit ?? null,
        bytes_limit: budget?.bytes_limit ?? null,
        score_limit: budget?.score_limit ?? null,
        cost_limit: budget?.cost_limit ?? null,
        alert_thresholds: budget?.alert_thresholds ?? DEFAULT_THRESHOLDS,
      })
      setCustomThreshold('')
      // Auto-expand advanced section if editing a budget with advanced limits
      const hasAdvancedLimits = budget?.span_limit || budget?.bytes_limit || budget?.score_limit
      setAdvancedOpen(!!hasAdvancedLimits)
    }
  }, [open, budget, form])

  const handleSubmit = async (values: BudgetFormValues) => {
    const data: CreateBudgetRequest = {
      name: values.name,
      budget_type: values.budget_type as BudgetType,
      project_id: values.project_id || undefined,
      span_limit: values.span_limit ?? undefined,
      bytes_limit: values.bytes_limit ?? undefined,
      score_limit: values.score_limit ?? undefined,
      cost_limit: values.cost_limit ?? undefined,
      alert_thresholds: values.alert_thresholds,
    }
    await onSubmit(data)
  }

  // Helper for number inputs
  const parseNumber = (value: string): number | null => {
    if (!value || value === '') return null
    const num = parseFloat(value)
    return isNaN(num) ? null : num
  }

  // Threshold management
  const thresholds = form.watch('alert_thresholds')

  const toggleThreshold = (pct: number) => {
    const current = form.getValues('alert_thresholds')
    if (current.includes(pct)) {
      form.setValue('alert_thresholds', current.filter((t) => t !== pct))
    } else {
      form.setValue('alert_thresholds', [...current, pct].sort((a, b) => a - b))
    }
  }

  const addCustomThreshold = () => {
    const value = parseInt(customThreshold, 10)
    if (isNaN(value) || value < 1 || value > 100) return
    const current = form.getValues('alert_thresholds')
    if (!current.includes(value)) {
      form.setValue('alert_thresholds', [...current, value].sort((a, b) => a - b))
    }
    setCustomThreshold('')
  }

  const removeThreshold = (pct: number) => {
    const current = form.getValues('alert_thresholds')
    form.setValue('alert_thresholds', current.filter((t) => t !== pct))
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-lg max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEditing ? 'Edit Budget' : 'Create Budget'}</DialogTitle>
          <DialogDescription>
            Set usage limits and alert thresholds for your organization.
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-6">
          {/* Basic Info */}
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="name">Budget Name</Label>
              <Input
                id="name"
                placeholder="e.g., Production Budget"
                {...form.register('name')}
              />
              {form.formState.errors.name && (
                <p className="text-sm text-destructive">
                  {form.formState.errors.name.message}
                </p>
              )}
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label htmlFor="budget_type">Budget Type</Label>
                <Select
                  value={form.watch('budget_type')}
                  onValueChange={(value) =>
                    form.setValue('budget_type', value as 'monthly' | 'weekly')
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="monthly">Monthly</SelectItem>
                    <SelectItem value="weekly">Weekly</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {projects.length > 0 && (
                <div className="space-y-2">
                  <Label htmlFor="project">Project (Optional)</Label>
                  <Select
                    value={form.watch('project_id') ?? 'all'}
                    onValueChange={(value) =>
                      form.setValue('project_id', value === 'all' ? undefined : value)
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="All projects" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="all">All Projects</SelectItem>
                      {projects.map((project) => (
                        <SelectItem key={project.id} value={project.id}>
                          {project.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              )}
            </div>
          </div>

          {/* Usage Limit */}
          <div className="space-y-4">
            <h4 className="text-sm font-medium">Usage Limit</h4>

            {/* Primary: Cost Limit */}
            <div className="space-y-2">
              <Label htmlFor="cost_limit">Cost Limit ($)</Label>
              <Input
                id="cost_limit"
                type="number"
                step="0.01"
                placeholder="e.g., 100.00"
                value={form.watch('cost_limit') ?? ''}
                onChange={(e) =>
                  form.setValue('cost_limit', parseNumber(e.target.value))
                }
              />
              <p className="text-xs text-muted-foreground">
                Maximum spend for this budget period.
              </p>
            </div>

            {/* Validation error */}
            {(form.formState.errors as Record<string, { message?: string }>).limits && (
              <p className="text-sm text-destructive">
                {(form.formState.errors as Record<string, { message?: string }>).limits?.message}
              </p>
            )}

            {/* Advanced Limits */}
            <Collapsible open={advancedOpen} onOpenChange={setAdvancedOpen}>
              <CollapsibleTrigger asChild>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="flex items-center gap-2 text-muted-foreground hover:text-foreground p-0 h-auto"
                >
                  <ChevronDown
                    className={cn(
                      'h-4 w-4 transition-transform',
                      advancedOpen && 'rotate-180'
                    )}
                  />
                  Advanced Limits (optional)
                </Button>
              </CollapsibleTrigger>
              <CollapsibleContent className="pt-4">
                <p className="text-xs text-muted-foreground mb-4">
                  Set limits on specific usage dimensions. Leave empty for no limit.
                </p>
                <div className="grid grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <Label htmlFor="span_limit">Span Limit</Label>
                    <Input
                      id="span_limit"
                      type="number"
                      placeholder="e.g., 1000000"
                      value={form.watch('span_limit') ?? ''}
                      onChange={(e) =>
                        form.setValue('span_limit', parseNumber(e.target.value))
                      }
                    />
                    <p className="text-xs text-muted-foreground">Number of spans</p>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="bytes_limit">Data Limit (GB)</Label>
                    <Input
                      id="bytes_limit"
                      type="number"
                      step="0.1"
                      placeholder="e.g., 10"
                      value={
                        form.watch('bytes_limit')
                          ? (form.watch('bytes_limit')! / 1073741824).toFixed(2)
                          : ''
                      }
                      onChange={(e) => {
                        const gb = parseNumber(e.target.value)
                        form.setValue(
                          'bytes_limit',
                          gb !== null ? Math.round(gb * 1073741824) : null
                        )
                      }}
                    />
                    <p className="text-xs text-muted-foreground">Gigabytes processed</p>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="score_limit">Score Limit</Label>
                    <Input
                      id="score_limit"
                      type="number"
                      placeholder="e.g., 10000"
                      value={form.watch('score_limit') ?? ''}
                      onChange={(e) =>
                        form.setValue('score_limit', parseNumber(e.target.value))
                      }
                    />
                    <p className="text-xs text-muted-foreground">Number of scores</p>
                  </div>
                </div>
              </CollapsibleContent>
            </Collapsible>
          </div>

          {/* Alert Thresholds */}
          <div className="space-y-4">
            <h4 className="text-sm font-medium">Alert Thresholds</h4>
            <p className="text-xs text-muted-foreground">
              Get notified when usage reaches these percentages.
            </p>

            {/* Preset buttons */}
            <div className="flex flex-wrap gap-2">
              {PRESET_THRESHOLDS.map((pct) => (
                <Button
                  key={pct}
                  type="button"
                  size="sm"
                  variant={thresholds.includes(pct) ? 'default' : 'outline'}
                  onClick={() => toggleThreshold(pct)}
                  className="h-8"
                >
                  {pct}%
                </Button>
              ))}
            </div>

            {/* Custom input */}
            <div className="flex gap-2">
              <Input
                type="number"
                min="1"
                max="100"
                placeholder="Custom %"
                value={customThreshold}
                onChange={(e) => setCustomThreshold(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault()
                    addCustomThreshold()
                  }
                }}
                className="w-28"
              />
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addCustomThreshold}
              >
                Add
              </Button>
            </div>

            {/* Active thresholds display */}
            {thresholds.length > 0 && (
              <div className="flex flex-wrap gap-1.5">
                {thresholds.sort((a, b) => a - b).map((t) => (
                  <Badge
                    key={t}
                    variant="secondary"
                    className={cn(
                      'flex items-center gap-1 pr-1',
                      t >= 100 && 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
                      t >= 80 && t < 100 && 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
                      t < 80 && 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-300'
                    )}
                  >
                    {t}%
                    <button
                      type="button"
                      onClick={() => removeThreshold(t)}
                      className="ml-1 rounded-full p-0.5 hover:bg-black/10 dark:hover:bg-white/10"
                    >
                      <X className="h-3 w-3" />
                    </button>
                  </Badge>
                ))}
              </div>
            )}

            {thresholds.length === 0 && (
              <p className="text-xs text-muted-foreground italic">
                No thresholds set. You won&apos;t receive any alerts.
              </p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              {isEditing ? 'Save Changes' : 'Create Budget'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
