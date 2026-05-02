import { useEffect, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  billingKeys,
  createBudget,
  updateBudget,
} from '../api/queries'
import type {
  BudgetType,
  CreateBudgetRequest,
  UpdateBudgetRequest,
  UsageBudget,
} from '../api/types'

interface BudgetFormDialogProps {
  orgId: string
  budget: UsageBudget | null
  open: boolean
  onOpenChange: (open: boolean) => void
}

function parseThresholds(csv: string): number[] | null {
  const parts = csv
    .split(',')
    .map((s) => s.trim())
    .filter((s) => s.length > 0)
  if (parts.length === 0) return []
  const nums: number[] = []
  for (const p of parts) {
    const n = Number(p)
    if (!Number.isFinite(n) || n < 1 || n > 100 || !Number.isInteger(n)) {
      return null
    }
    nums.push(n)
  }
  return nums
}

// v1 budget form: name + budget_type + cost_limit + alert_thresholds
// (comma-separated percentages). Advanced dimension-specific limits
// (spans / bytes / scores) are deferred — the backend accepts them but
// the first-pass UI targets the common case.
export function BudgetFormDialog({
  orgId,
  budget,
  open,
  onOpenChange,
}: BudgetFormDialogProps) {
  const queryClient = useQueryClient()
  const isEditing = !!budget

  const [name, setName] = useState('')
  const [budgetType, setBudgetType] = useState<BudgetType>('monthly')
  const [costLimit, setCostLimit] = useState('')
  const [thresholdsCsv, setThresholdsCsv] = useState('50, 80, 100')
  const [error, setError] = useState<string | null>(null)

  // Re-seed the form state each time the dialog opens so editing picks
  // up the current budget and creating starts clean.
  useEffect(() => {
    if (!open) return
    if (budget) {
      setName(budget.name)
      setBudgetType(budget.budget_type)
      setCostLimit(budget.cost_limit ?? '')
      setThresholdsCsv(budget.alert_thresholds.join(', '))
    } else {
      setName('')
      setBudgetType('monthly')
      setCostLimit('')
      setThresholdsCsv('50, 80, 100')
    }
    setError(null)
  }, [open, budget])

  const createMut = useMutation({
    mutationFn: async (data: CreateBudgetRequest) => createBudget(orgId, data),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: billingKeys.budgetList(orgId),
      })
      toast.success('Budget created')
      onOpenChange(false)
    },
    onError: (err) => {
      const msg =
        err instanceof Error ? err.message : 'Failed to create budget'
      setError(msg)
      toast.error(msg)
    },
  })

  const updateMut = useMutation({
    mutationFn: async (data: UpdateBudgetRequest) => {
      if (!budget) throw new Error('No budget to update')
      return updateBudget(orgId, budget.id, data)
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: billingKeys.budgetList(orgId),
      })
      toast.success('Budget updated')
      onOpenChange(false)
    },
    onError: (err) => {
      const msg =
        err instanceof Error ? err.message : 'Failed to update budget'
      setError(msg)
      toast.error(msg)
    },
  })

  const pending = createMut.isPending || updateMut.isPending

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    if (name.trim().length === 0) {
      setError('Name is required')
      return
    }
    const cost = Number(costLimit)
    if (!Number.isFinite(cost) || cost <= 0) {
      setError('Cost limit must be a positive number')
      return
    }
    const thresholds = parseThresholds(thresholdsCsv)
    if (thresholds === null) {
      setError('Alert thresholds must be integers 1–100, comma separated')
      return
    }

    if (isEditing) {
      updateMut.mutate({
        name: name.trim(),
        cost_limit: cost,
        alert_thresholds: thresholds,
      })
    } else {
      createMut.mutate({
        name: name.trim(),
        budget_type: budgetType,
        cost_limit: cost,
        alert_thresholds: thresholds,
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>
              {isEditing ? 'Edit budget' : 'Create budget'}
            </DialogTitle>
            <DialogDescription>
              Cap spend for this organization and get alerted at key
              thresholds.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="budget-name">Name</Label>
              <Input
                id="budget-name"
                required
                autoFocus
                maxLength={100}
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="e.g. Monthly cap"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="budget-type">Period</Label>
              <Select
                value={budgetType}
                onValueChange={(v) => setBudgetType(v as BudgetType)}
                disabled={isEditing}
              >
                <SelectTrigger id="budget-type">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="monthly">Monthly</SelectItem>
                  <SelectItem value="weekly">Weekly</SelectItem>
                </SelectContent>
              </Select>
              {isEditing && (
                <p className="text-xs text-muted-foreground">
                  Period can&apos;t be changed after a budget is created.
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="budget-cost">Cost limit (USD)</Label>
              <Input
                id="budget-cost"
                type="number"
                required
                min="0"
                step="0.01"
                value={costLimit}
                onChange={(e) => setCostLimit(e.target.value)}
                placeholder="500"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="budget-thresholds">
                Alert thresholds (%)
              </Label>
              <Input
                id="budget-thresholds"
                value={thresholdsCsv}
                onChange={(e) => setThresholdsCsv(e.target.value)}
                placeholder="50, 80, 100"
              />
              <p className="text-xs text-muted-foreground">
                Comma-separated integers between 1 and 100. Leave empty to
                disable alerts.
              </p>
            </div>

            {error && (
              <p className="text-sm text-destructive" role="alert">
                {error}
              </p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
              disabled={pending}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={pending}>
              {pending
                ? 'Saving…'
                : isEditing
                  ? 'Save changes'
                  : 'Create budget'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}
