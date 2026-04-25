import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { MoreHorizontal, Pencil, Plus, Trash2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
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
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  billingKeys,
  budgetsQueryOptions,
  deleteBudget,
} from '../api/queries'
import { BudgetFormDialog } from './budget-form-dialog'
import type { UsageBudget } from '../api/types'

interface BudgetsListProps {
  orgId: string
}

function formatMoney(decimalStr: string | undefined): string {
  if (!decimalStr) return '—'
  const n = Number(decimalStr)
  if (!Number.isFinite(n)) return decimalStr
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(n)
}

// Full budget management surface — list + create + edit + delete.
// Used as the dedicated /budgets tab content. The lighter `BudgetCard`
// stays for the billing landing summary.
export function BudgetsList({ orgId }: BudgetsListProps) {
  const queryClient = useQueryClient()
  const { data, isLoading, isError, error } = useQuery(budgetsQueryOptions(orgId))

  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<UsageBudget | null>(null)
  const [deleting, setDeleting] = useState<UsageBudget | null>(null)

  const deleteMut = useMutation({
    mutationFn: async (budgetId: string) => deleteBudget(orgId, budgetId),
    onSuccess: async () => {
      await queryClient.invalidateQueries({
        queryKey: billingKeys.budgetList(orgId),
      })
      toast.success('Budget deleted')
      setDeleting(null)
    },
    onError: (err) => {
      toast.error(
        err instanceof Error ? err.message : 'Failed to delete budget',
      )
    },
  })

  const openCreate = () => {
    setEditing(null)
    setFormOpen(true)
  }
  const openEdit = (b: UsageBudget) => {
    setEditing(b)
    setFormOpen(true)
  }

  const budgets = data ?? []

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-lg font-semibold">Budgets</h2>
          <p className="text-sm text-muted-foreground">
            Cap monthly or weekly spend and trigger alerts at thresholds.
          </p>
        </div>
        <Button onClick={openCreate} size="sm">
          <Plus className="mr-2 h-4 w-4" />
          Create budget
        </Button>
      </div>

      {isError ? (
        <div className="rounded-lg border border-destructive/40 bg-destructive/5 p-6">
          <p className="text-sm font-medium text-destructive">
            Unable to load budgets
          </p>
          <p className="mt-1 text-sm text-muted-foreground">
            {error instanceof Error ? error.message : 'Unknown error'}
          </p>
        </div>
      ) : isLoading ? (
        <div className="rounded-lg border p-6">
          <p className="text-sm text-muted-foreground">Loading budgets…</p>
        </div>
      ) : budgets.length === 0 ? (
        <div className="rounded-lg border border-dashed p-12 text-center">
          <h3 className="text-base font-semibold">No budgets configured</h3>
          <p className="mt-1 text-sm text-muted-foreground">
            Create a budget to set a spend cap and get alerts before you hit it.
          </p>
          <Button className="mt-4" onClick={openCreate}>
            <Plus className="mr-2 h-4 w-4" />
            Create your first budget
          </Button>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {budgets.map((b) => (
            <BudgetRow
              key={b.id}
              budget={b}
              onEdit={() => openEdit(b)}
              onDelete={() => setDeleting(b)}
            />
          ))}
        </div>
      )}

      <BudgetFormDialog
        orgId={orgId}
        budget={editing}
        open={formOpen}
        onOpenChange={(open) => {
          setFormOpen(open)
          if (!open) setEditing(null)
        }}
      />

      <AlertDialog
        open={deleting !== null}
        onOpenChange={(open) => {
          if (!open) setDeleting(null)
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete budget</AlertDialogTitle>
            <AlertDialogDescription>
              Delete <span className="font-medium">{deleting?.name}</span>? This
              also removes any alerts associated with the budget. This action
              can&apos;t be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMut.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={() => deleting && deleteMut.mutate(deleting.id)}
              disabled={deleteMut.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMut.isPending ? 'Deleting…' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}

function BudgetRow({
  budget,
  onEdit,
  onDelete,
}: {
  budget: UsageBudget
  onEdit: () => void
  onDelete: () => void
}) {
  const limit = budget.cost_limit ? Number(budget.cost_limit) : 0
  const current = Number(budget.current_cost)
  const pct =
    limit > 0 && Number.isFinite(current)
      ? Math.min(100, Math.max(0, (current / limit) * 100))
      : 0

  return (
    <div className="rounded-lg border p-4">
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-1">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold">{budget.name}</h3>
            <Badge variant="outline" className="capitalize">
              {budget.budget_type}
            </Badge>
            {budget.project_id && (
              <Badge variant="secondary">Project-scoped</Badge>
            )}
            {!budget.is_active && <Badge variant="destructive">Inactive</Badge>}
          </div>
          <p className="text-sm text-muted-foreground">
            {formatMoney(budget.current_cost)} of{' '}
            {formatMoney(budget.cost_limit)} this period
          </p>
        </div>
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="h-8 w-8">
              <MoreHorizontal className="h-4 w-4" />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={onEdit}>
              <Pencil className="mr-2 h-4 w-4" />
              Edit
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={onDelete} className="text-destructive">
              <Trash2 className="mr-2 h-4 w-4" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      <div className="mt-3 space-y-1">
        <Progress value={pct} />
        <p className="text-xs text-muted-foreground">
          {pct.toFixed(1)}% used
          {budget.alert_thresholds.length > 0 &&
            ` · alerts at ${[...budget.alert_thresholds]
              .sort((a, b) => a - b)
              .map((t) => `${t}%`)
              .join(', ')}`}
        </p>
      </div>
    </div>
  )
}
