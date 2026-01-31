'use client'

import { useState } from 'react'
import { Plus, Pencil, Trash2, Target, AlertCircle, Loader2, AlertTriangle, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
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
  useScoreConfigsQuery,
  useCreateScoreConfigMutation,
  useUpdateScoreConfigMutation,
  useDeleteScoreConfigMutation,
} from '../hooks/use-score-configs'
import { ScoreConfigForm } from './score-config-form'
import type { ScoreConfig, CreateScoreConfigRequest, UpdateScoreConfigRequest } from '../types'

interface ScoreConfigsSectionProps {
  projectId: string
}

export function ScoreConfigsSection({ projectId }: ScoreConfigsSectionProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false)
  const [editingConfig, setEditingConfig] = useState<ScoreConfig | null>(null)
  const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false)
  const [deletingConfig, setDeletingConfig] = useState<ScoreConfig | null>(null)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)

  const { data: configsResponse, isLoading, isFetching, error, refetch } = useScoreConfigsQuery(projectId, { page, limit: pageSize })
  const configs = configsResponse?.configs
  const totalCount = configsResponse?.totalCount ?? 0
  const totalPages = Math.ceil(totalCount / pageSize)
  const createMutation = useCreateScoreConfigMutation(projectId)
  const updateMutation = useUpdateScoreConfigMutation(projectId, editingConfig?.id ?? '')
  const deleteMutation = useDeleteScoreConfigMutation(projectId)

  const handleAddClick = () => {
    setEditingConfig(null)
    setIsDialogOpen(true)
  }

  const handleEditClick = (config: ScoreConfig) => {
    setEditingConfig(config)
    setIsDialogOpen(true)
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
    setIsDialogOpen(false)
    setEditingConfig(null)
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

  const getConstraintDisplay = (config: ScoreConfig) => {
    if (config.type === 'NUMERIC') {
      if (config.min_value !== undefined || config.max_value !== undefined) {
        return `${config.min_value ?? '−∞'} to ${config.max_value ?? '∞'}`
      }
      return 'Any number'
    }
    if (config.type === 'CATEGORICAL' && config.categories?.length) {
      return config.categories.join(', ')
    }
    if (config.type === 'BOOLEAN') {
      return 'true / false'
    }
    return '—'
  }

  return (
    <div className="space-y-8">
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <div className="flex flex-col items-center gap-2">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
            <p className="text-sm text-muted-foreground">Loading score configs...</p>
          </div>
        </div>
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
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-medium">
                Score Configs {configs && `(${configs.length})`}
              </h3>
              <p className="text-sm text-muted-foreground">
                Define validation rules for evaluation scores
              </p>
            </div>
            <Button onClick={handleAddClick}>
              <Plus className="mr-2 h-4 w-4" />
              Add Config
            </Button>
          </div>

          <div className="rounded-md border">
            <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Data Type</TableHead>
              <TableHead>Constraints</TableHead>
              <TableHead>Created</TableHead>
              <TableHead className="w-[100px] text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {configs?.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-center py-8 text-muted-foreground">
                  <div className="flex flex-col items-center gap-2">
                    <Target className="h-8 w-8 text-muted-foreground/50" />
                    <p>No score configs yet.</p>
                    <p className="text-xs">
                      Add a config to define validation rules for your scores.
                    </p>
                  </div>
                </TableCell>
              </TableRow>
            ) : (
              configs?.map((config) => (
                <TableRow key={config.id}>
                  <TableCell>
                    <div>
                      <div className="font-medium">{config.name}</div>
                      {config.description && (
                        <div className="text-xs text-muted-foreground truncate max-w-[200px]">
                          {config.description}
                        </div>
                      )}
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline">{config.type}</Badge>
                  </TableCell>
                  <TableCell>
                    <code className="text-xs bg-muted px-2 py-1 rounded font-mono">
                      {getConstraintDisplay(config)}
                    </code>
                  </TableCell>
                  <TableCell>
                    <div className="text-sm">
                      {new Date(config.created_at).toLocaleDateString()}
                    </div>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleEditClick(config)}
                      >
                        <Pencil className="h-4 w-4" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => handleDeleteClick(config)}
                        disabled={deleteMutation.isPending}
                        className="text-red-600 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-950"
                      >
                        {deleteMutation.isPending ? (
                          <Loader2 className="h-4 w-4 animate-spin" />
                        ) : (
                          <Trash2 className="h-4 w-4" />
                        )}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
          </div>

          {/* Pagination */}
          {totalCount > 0 && (
            <div className="flex items-center justify-end space-x-6 py-4">
              <div className="flex items-center space-x-2">
                <p className="text-sm font-medium">Rows per page</p>
                <Select
                  value={`${pageSize}`}
                  onValueChange={(value) => {
                    setPageSize(Number(value))
                    setPage(1)
                  }}
                >
                  <SelectTrigger className="h-8 w-[70px]">
                    <SelectValue placeholder={pageSize} />
                  </SelectTrigger>
                  <SelectContent side="top">
                    {[10, 25, 50, 100].map((size) => (
                      <SelectItem key={size} value={`${size}`}>
                        {size}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="flex w-[100px] items-center justify-center text-sm font-medium">
                Page {page} of {totalPages}
              </div>
              <div className="flex items-center space-x-2">
                <Button
                  variant="outline"
                  className="h-8 w-8 p-0"
                  onClick={() => setPage(page - 1)}
                  disabled={page <= 1 || isFetching}
                >
                  <span className="sr-only">Go to previous page</span>
                  <ChevronLeft className="h-4 w-4" />
                </Button>
                <Button
                  variant="outline"
                  className="h-8 w-8 p-0"
                  onClick={() => setPage(page + 1)}
                  disabled={page >= totalPages || isFetching}
                >
                  <span className="sr-only">Go to next page</span>
                  <ChevronRight className="h-4 w-4" />
                </Button>
              </div>
            </div>
          )}
        </div>
      )}

      <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
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
            onCancel={() => setIsDialogOpen(false)}
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
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete Config'
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
