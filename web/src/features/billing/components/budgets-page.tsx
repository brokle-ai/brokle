'use client'

import * as React from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { Plus, MoreHorizontal, Pencil, Trash2, AlertTriangle, Bell } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { PageHeader } from '@/components/layout/page-header'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
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
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

import {
  useBudgetsQuery,
  useCreateBudgetMutation,
  useDeleteBudgetMutation,
  useAlertsQuery,
  useAcknowledgeAlertMutation,
  budgetQueryKeys,
} from '../hooks'
import { updateBudget } from '../api/budget-api'
import type { UsageBudget, CreateBudgetRequest, UsageAlert } from '../types'
import { BudgetProgress } from './budget-progress'
import { BudgetForm } from './budget-form'

interface BudgetsPageProps {
  organizationId: string
  projects?: { id: string; name: string }[]
  className?: string
}

function AlertCard({
  alert,
  onAcknowledge,
  isAcknowledging,
}: {
  alert: UsageAlert
  onAcknowledge: () => void
  isAcknowledging: boolean
}) {
  const severityColors = {
    info: 'bg-blue-50 border-blue-200 dark:bg-blue-950/20 dark:border-blue-900',
    warning: 'bg-yellow-50 border-yellow-200 dark:bg-yellow-950/20 dark:border-yellow-900',
    critical: 'bg-red-50 border-red-200 dark:bg-red-950/20 dark:border-red-900',
  }

  const severityBadge = {
    info: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-300',
    warning: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300',
    critical: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300',
  }

  const dimensionLabels = {
    spans: 'Spans',
    bytes: 'Data',
    scores: 'Scores',
    cost: 'Cost',
  }

  return (
    <div
      className={cn(
        'flex items-center justify-between p-3 rounded-lg border',
        severityColors[alert.severity]
      )}
    >
      <div className="flex items-center gap-3">
        <AlertTriangle className="h-4 w-4 text-muted-foreground" />
        <div>
          <p className="text-sm font-medium">
            {dimensionLabels[alert.dimension]} at {alert.percent_used.toFixed(0)}%
          </p>
          <p className="text-xs text-muted-foreground">
            {new Date(alert.triggered_at).toLocaleString()}
          </p>
        </div>
      </div>
      <div className="flex items-center gap-2">
        <Badge className={severityBadge[alert.severity]}>
          {alert.severity}
        </Badge>
        {alert.status === 'triggered' && (
          <Button
            variant="ghost"
            size="sm"
            onClick={onAcknowledge}
            disabled={isAcknowledging}
          >
            Acknowledge
          </Button>
        )}
      </div>
    </div>
  )
}

function BudgetCard({
  budget,
  onEdit,
  onDelete,
}: {
  budget: UsageBudget
  onEdit: () => void
  onDelete: () => void
}) {
  return (
    <Card>
      <CardHeader className="flex flex-row items-start justify-between space-y-0 pb-2">
        <div className="space-y-1">
          <CardTitle className="text-base">{budget.name}</CardTitle>
          <CardDescription className="flex items-center gap-2">
            <Badge variant="outline" className="capitalize">
              {budget.budget_type}
            </Badge>
            {budget.project_id && (
              <Badge variant="secondary">Project-specific</Badge>
            )}
            {!budget.is_active && (
              <Badge variant="destructive">Inactive</Badge>
            )}
          </CardDescription>
          {budget.alert_thresholds && budget.alert_thresholds.length > 0 && (
            <p className="text-xs text-muted-foreground">
              Alerts at: {[...budget.alert_thresholds].sort((a, b) => a - b).map(t => `${t}%`).join(', ')}
            </p>
          )}
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
      </CardHeader>
      <CardContent>
        <BudgetProgress budget={budget} />
      </CardContent>
    </Card>
  )
}

export function BudgetsPage({
  organizationId,
  projects = [],
  className,
}: BudgetsPageProps) {
  const queryClient = useQueryClient()
  const [formOpen, setFormOpen] = React.useState(false)
  const [editingBudget, setEditingBudget] = React.useState<UsageBudget | undefined>()
  const [deleteDialogOpen, setDeleteDialogOpen] = React.useState(false)
  const [budgetToDelete, setBudgetToDelete] = React.useState<UsageBudget | null>(null)
  const [isUpdating, setIsUpdating] = React.useState(false)

  const {
    data: budgets,
    isLoading: isBudgetsLoading,
    error: budgetsError,
  } = useBudgetsQuery(organizationId)

  const {
    data: alerts,
    isLoading: isAlertsLoading,
  } = useAlertsQuery(organizationId, 10)

  const createMutation = useCreateBudgetMutation(organizationId)
  const deleteMutation = useDeleteBudgetMutation(organizationId)
  const acknowledgeMutation = useAcknowledgeAlertMutation(organizationId)

  // Active (unacknowledged) alerts
  const activeAlerts = React.useMemo(
    () => alerts?.filter((a) => a.status === 'triggered') ?? [],
    [alerts]
  )

  const handleCreateBudget = async (data: CreateBudgetRequest) => {
    try {
      await createMutation.mutateAsync(data)
      setFormOpen(false)
      toast.success('Budget created successfully')
    } catch (error) {
      toast.error('Failed to create budget')
    }
  }

  const handleEditBudget = async (data: CreateBudgetRequest) => {
    if (!editingBudget) return
    setIsUpdating(true)
    try {
      await updateBudget(organizationId, editingBudget.id, data)
      // Invalidate the budgets query to refetch
      await queryClient.invalidateQueries({
        queryKey: budgetQueryKeys.list(organizationId),
      })
      setFormOpen(false)
      setEditingBudget(undefined)
      toast.success('Budget updated successfully')
    } catch (error) {
      toast.error('Failed to update budget')
    } finally {
      setIsUpdating(false)
    }
  }

  const handleDeleteBudget = async () => {
    if (!budgetToDelete) return
    try {
      await deleteMutation.mutateAsync(budgetToDelete.id)
      setDeleteDialogOpen(false)
      setBudgetToDelete(null)
      toast.success('Budget deleted successfully')
    } catch (error) {
      toast.error('Failed to delete budget')
    }
  }

  const handleAcknowledgeAlert = async (alertId: string) => {
    try {
      await acknowledgeMutation.mutateAsync(alertId)
      toast.success('Alert acknowledged')
    } catch (error) {
      toast.error('Failed to acknowledge alert')
    }
  }

  const openEditForm = (budget: UsageBudget) => {
    setEditingBudget(budget)
    setFormOpen(true)
  }

  const openDeleteDialog = (budget: UsageBudget) => {
    setBudgetToDelete(budget)
    setDeleteDialogOpen(true)
  }

  const closeForm = () => {
    setFormOpen(false)
    setEditingBudget(undefined)
  }

  const errorMessage = budgetsError
    ? typeof budgetsError === 'object' && 'message' in budgetsError
      ? (budgetsError.message as string)
      : String(budgetsError)
    : null

  return (
    <div className={cn('space-y-6', className)}>
      <PageHeader title="Budgets">
        <Button onClick={() => setFormOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Budget
        </Button>
      </PageHeader>

      {/* Active Alerts Section */}
      {activeAlerts.length > 0 && (
        <Card>
          <CardHeader className="pb-3">
            <div className="flex items-center gap-2">
              <Bell className="h-4 w-4 text-muted-foreground" />
              <CardTitle className="text-base">Active Alerts</CardTitle>
              <Badge variant="destructive">{activeAlerts.length}</Badge>
            </div>
          </CardHeader>
          <CardContent className="space-y-2">
            {activeAlerts.map((alert) => (
              <AlertCard
                key={alert.id}
                alert={alert}
                onAcknowledge={() => handleAcknowledgeAlert(alert.id)}
                isAcknowledging={acknowledgeMutation.isPending}
              />
            ))}
          </CardContent>
        </Card>
      )}

      {/* Budgets List */}
      {isBudgetsLoading ? (
        <div className="grid gap-4 md:grid-cols-2">
          {[1, 2].map((i) => (
            <Card key={i}>
              <CardHeader>
                <Skeleton className="h-5 w-[150px]" />
                <Skeleton className="h-4 w-[100px]" />
              </CardHeader>
              <CardContent className="space-y-4">
                <Skeleton className="h-2 w-full" />
                <Skeleton className="h-2 w-full" />
                <Skeleton className="h-2 w-full" />
              </CardContent>
            </Card>
          ))}
        </div>
      ) : errorMessage ? (
        <Card>
          <CardContent className="py-8 text-center">
            <p className="text-destructive">{errorMessage}</p>
          </CardContent>
        </Card>
      ) : !budgets || budgets.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <div className="mx-auto max-w-sm space-y-4">
              <div className="rounded-full bg-muted p-4 w-fit mx-auto">
                <AlertTriangle className="h-8 w-8 text-muted-foreground" />
              </div>
              <div className="space-y-2">
                <h3 className="text-lg font-medium">No budgets configured</h3>
                <p className="text-sm text-muted-foreground">
                  Create a budget to set usage limits and receive alerts when approaching thresholds.
                </p>
              </div>
              <Button onClick={() => setFormOpen(true)}>
                <Plus className="mr-2 h-4 w-4" />
                Create Your First Budget
              </Button>
            </div>
          </CardContent>
        </Card>
      ) : (
        <div className="grid gap-4 md:grid-cols-2">
          {budgets.map((budget) => (
            <BudgetCard
              key={budget.id}
              budget={budget}
              onEdit={() => openEditForm(budget)}
              onDelete={() => openDeleteDialog(budget)}
            />
          ))}
        </div>
      )}

      {/* Budget Form Dialog */}
      <BudgetForm
        open={formOpen}
        onOpenChange={closeForm}
        onSubmit={editingBudget ? handleEditBudget : handleCreateBudget}
        budget={editingBudget}
        projects={projects}
        isLoading={createMutation.isPending || isUpdating}
      />

      {/* Delete Confirmation Dialog */}
      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Budget</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &ldquo;{budgetToDelete?.name}&rdquo;? This action
              cannot be undone and will remove all associated alerts.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteBudget}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
