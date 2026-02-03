'use client'

import { useState, useMemo } from 'react'
import { Target, AlertCircle, AlertTriangle } from 'lucide-react'
import {
  useReactTable,
  getCoreRowModel,
  flexRender,
} from '@tanstack/react-table'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DataTableSkeleton,
  DataTableEmptyState,
  DataTablePagination,
} from '@/components/shared/tables'
import {
  useScoreConfigsQuery,
  useCreateScoreConfigMutation,
  useUpdateScoreConfigMutation,
  useDeleteScoreConfigMutation,
} from '../hooks/use-score-configs'
import { ScoreConfigForm } from './score-config-form'
import { createScoreConfigsColumns } from './score-configs-columns'
import type { ScoreConfig, CreateScoreConfigRequest, UpdateScoreConfigRequest } from '../types'

interface ScoreConfigsSectionProps {
  projectId: string
  createDialogOpen?: boolean
  onCreateDialogOpenChange?: (open: boolean) => void
}

export function ScoreConfigsSection({
  projectId,
  createDialogOpen,
  onCreateDialogOpenChange,
}: ScoreConfigsSectionProps) {
  // editingConfig drives edit mode, createDialogOpen drives create mode
  const [editingConfig, setEditingConfig] = useState<ScoreConfig | null>(null)

  // Dialog is open when parent requests create OR we're editing
  const isDialogOpen = createDialogOpen || editingConfig !== null
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [deletingConfig, setDeletingConfig] = useState<ScoreConfig | null>(null)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const { data: configsResponse, isLoading, isFetching, error, refetch } = useScoreConfigsQuery(projectId, { page, limit: pageSize })
  const configs = configsResponse?.data || []
  const totalCount = configsResponse?.pagination?.total ?? 0
  const totalPages = Math.ceil(totalCount / pageSize)
  const createMutation = useCreateScoreConfigMutation(projectId)
  const updateMutation = useUpdateScoreConfigMutation(projectId, editingConfig?.id ?? '')
  const deleteMutation = useDeleteScoreConfigMutation(projectId)

  // Handler for dialog open/close - ensures both states are properly reset
  const handleDialogOpenChange = (open: boolean) => {
    if (!open) {
      // Closing: reset both create and edit states
      onCreateDialogOpenChange?.(false)
      setEditingConfig(null)
    }
    // Opening is handled by parent (create) or handleEditClick (edit)
  }

  const handleEditClick = (config: ScoreConfig) => {
    setEditingConfig(config)
    // Dialog opens automatically via: editingConfig !== null
  }

  const handleDeleteClick = (config: ScoreConfig) => {
    setDeletingConfig(config)
    setIsDeleteDialogOpen(true)
  }

  const handleSubmit = async (data: CreateScoreConfigRequest) => {
    if (editingConfig) {
      await updateMutation.mutateAsync(data as UpdateScoreConfigRequest)
    } else {
      await createMutation.mutateAsync(data)
    }
    // Close via the handler to ensure both states reset
    handleDialogOpenChange(false)
  }

  const handleConfirmDelete = async () => {
    if (deletingConfig) {
      await deleteMutation.mutateAsync({
        configId: deletingConfig.id,
        configName: deletingConfig.name,
      })
      setIsDeleteDialogOpen(false)
      setDeletingConfig(null)
    }
  }

  // Create columns
  const columns = useMemo(
    () =>
      createScoreConfigsColumns({
        onEdit: handleEditClick,
        onDelete: handleDeleteClick,
        isDeleting: deleteMutation.isPending,
      }),
    [deleteMutation.isPending]
  )

  // Initialize React Table
  const table = useReactTable({
    data: configs,
    columns,
    getCoreRowModel: getCoreRowModel(),
    manualPagination: true,
    pageCount: totalPages,
    state: {
      pagination: {
        pageIndex: page - 1,
        pageSize,
      },
    },
    onPaginationChange: (updater) => {
      const newState = typeof updater === 'function'
        ? updater({ pageIndex: page - 1, pageSize })
        : updater
      setPage(newState.pageIndex + 1)
      if (newState.pageSize !== pageSize) {
        setPageSize(newState.pageSize)
        setPage(1)
      }
    },
  })

  return (
    <div className="space-y-8">
      {isLoading && (
        <DataTableSkeleton columns={5} rows={5} showToolbar={false} />
      )}

      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription className="flex items-center justify-between">
            <span>Failed to load score configs</span>
            <Button variant="outline" size="sm" onClick={() => refetch()}>
              Try Again
            </Button>
          </AlertDescription>
        </Alert>
      )}

      {!isLoading && !error && (
        <div className="space-y-4">
          <div className="rounded-md border">
            <Table>
              <TableHeader>
                {table.getHeaderGroups().map((headerGroup) => (
                  <TableRow key={headerGroup.id}>
                    {headerGroup.headers.map((header) => (
                      <TableHead key={header.id}>
                        {header.isPlaceholder
                          ? null
                          : flexRender(header.column.columnDef.header, header.getContext())}
                      </TableHead>
                    ))}
                  </TableRow>
                ))}
              </TableHeader>
              <TableBody>
                {table.getRowModel().rows?.length ? (
                  table.getRowModel().rows.map((row) => (
                    <TableRow key={row.id}>
                      {row.getVisibleCells().map((cell) => (
                        <TableCell key={cell.id}>
                          {flexRender(cell.column.columnDef.cell, cell.getContext())}
                        </TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  <TableRow>
                    <TableCell colSpan={columns.length} className="h-24 text-center">
                      <DataTableEmptyState
                        title="No score configs yet"
                        description="Add a config to define validation rules for your scores."
                        icon={<Target className="h-8 w-8 text-muted-foreground/50" />}
                      />
                    </TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </div>

          {/* Pagination */}
          {totalCount > 0 && (
            <DataTablePagination
              table={table}
              pageSizes={[10, 25, 50, 100]}
              isPending={isFetching}
              serverPagination={{
                page,
                pageSize,
                total: totalCount,
                totalPages,
                hasNextPage: page < totalPages,
                hasPreviousPage: page > 1,
              }}
            />
          )}
        </div>
      )}

      <Dialog open={isDialogOpen} onOpenChange={handleDialogOpenChange}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              {editingConfig ? `Edit ${editingConfig.name}` : 'Add Score Config'}
            </DialogTitle>
            <DialogDescription>
              {editingConfig
                ? 'Update the configuration for this score type.'
                : 'Configure a new score type with validation rules.'}
            </DialogDescription>
          </DialogHeader>
          <ScoreConfigForm
            config={editingConfig ?? undefined}
            onSubmit={handleSubmit}
            onCancel={() => handleDialogOpenChange(false)}
            isLoading={createMutation.isPending || updateMutation.isPending}
          />
        </DialogContent>
      </Dialog>

      <Dialog open={isDeleteDialogOpen} onOpenChange={setIsDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[450px]">
          <DialogHeader>
            <DialogTitle className="text-red-600">Delete Score Config</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this score config? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>

          {deletingConfig && (
            <div className="space-y-4">
              <div className="rounded-lg bg-muted p-3 space-y-1">
                <div className="flex items-center gap-2">
                  <Target className="h-4 w-4" />
                  <span className="font-medium">{deletingConfig.name}</span>
                </div>
                <div className="flex items-center gap-2 mt-1 text-xs text-muted-foreground">
                  <Badge variant="outline" className="text-xs">
                    {deletingConfig.type}
                  </Badge>
                  {deletingConfig.description && (
                    <>
                      <span>•</span>
                      <span className="truncate">{deletingConfig.description}</span>
                    </>
                  )}
                </div>
              </div>

              <Alert className="border-red-200 bg-red-50 dark:border-red-900 dark:bg-red-950">
                <AlertTriangle className="h-4 w-4 text-red-500" />
                <AlertDescription className="text-red-600 dark:text-red-400">
                  <strong>Warning:</strong> Existing scores using this config will not be
                  affected, but validation will no longer apply.
                </AlertDescription>
              </Alert>
            </div>
          )}

          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setIsDeleteDialogOpen(false)}
              disabled={deleteMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleConfirmDelete}
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete Config'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
