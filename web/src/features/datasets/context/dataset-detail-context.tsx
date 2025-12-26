'use client'

import React, { useMemo, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import useDialogState from '@/hooks/use-dialog-state'
import { useProjectOnly } from '@/features/projects'
import {
  useDatasetQuery,
  useUpdateDatasetMutation,
  useDeleteDatasetMutation,
} from '../hooks/use-datasets'
import type { Dataset, UpdateDatasetRequest } from '../types'

export type DatasetDetailDialogType = 'edit' | 'delete'

interface DatasetDetailContextType {
  dataset: Dataset | null
  isLoading: boolean
  error: string | null
  refetch: () => void
  open: DatasetDetailDialogType | null
  setOpen: (dialog: DatasetDetailDialogType | null) => void
  handleUpdate: (data: UpdateDatasetRequest) => Promise<void>
  handleDelete: () => Promise<void>
  isUpdating: boolean
  isDeleting: boolean
  projectSlug: string
  projectId: string
  datasetId: string
}

const DatasetDetailContext = React.createContext<DatasetDetailContextType | null>(null)

interface DatasetDetailProviderProps {
  children: React.ReactNode
  projectSlug: string
  datasetId: string
}

export function DatasetDetailProvider({
  children,
  projectSlug,
  datasetId,
}: DatasetDetailProviderProps) {
  const router = useRouter()
  const [open, setOpen] = useDialogState<DatasetDetailDialogType>(null)
  const { currentProject, isLoading: projectLoading } = useProjectOnly()

  const projectId = currentProject?.id ?? ''

  const {
    data: dataset,
    isLoading: datasetLoading,
    error,
    refetch,
  } = useDatasetQuery(projectId, datasetId)

  const updateMutation = useUpdateDatasetMutation(projectId, datasetId)
  const deleteMutation = useDeleteDatasetMutation(projectId)

  const isLoading = projectLoading || datasetLoading

  const handleUpdate = useCallback(
    async (data: UpdateDatasetRequest) => {
      await updateMutation.mutateAsync(data)
      setOpen(null)
    },
    [updateMutation, setOpen]
  )

  const handleDelete = useCallback(async () => {
    if (dataset) {
      await deleteMutation.mutateAsync({
        datasetId: dataset.id,
        datasetName: dataset.name,
      })
      router.push(`/projects/${projectSlug}/datasets`)
    }
  }, [dataset, deleteMutation, router, projectSlug])

  const contextValue = useMemo(
    () => ({
      dataset: dataset ?? null,
      isLoading,
      error: error?.message ?? null,
      refetch,
      open,
      setOpen,
      handleUpdate,
      handleDelete,
      isUpdating: updateMutation.isPending,
      isDeleting: deleteMutation.isPending,
      projectSlug,
      projectId,
      datasetId,
    }),
    [
      dataset,
      isLoading,
      error,
      refetch,
      open,
      setOpen,
      handleUpdate,
      handleDelete,
      updateMutation.isPending,
      deleteMutation.isPending,
      projectSlug,
      projectId,
      datasetId,
    ]
  )

  return (
    <DatasetDetailContext value={contextValue}>
      {children}
    </DatasetDetailContext>
  )
}

export function useDatasetDetail() {
  const context = React.useContext(DatasetDetailContext)

  if (!context) {
    throw new Error('useDatasetDetail must be used within <DatasetDetailProvider>')
  }

  return context
}
