'use client'

import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { toast } from 'sonner'
import { BulkActionsToolbar } from '@/components/shared/tables/bulk-actions-toolbar'
import { type User } from '../data/schema'
import { UsersMultiDeleteDialog } from './users-multi-delete-dialog'

type DataTableBulkActionsProps<TData> = {
  table: Table<TData>
}

export function DataTableBulkActions<TData>({
  table,
}: DataTableBulkActionsProps<TData>) {
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const selectedRows = table.getFilteredSelectedRowModel().rows

  const handleBulkAction = (action: string) => {
    const selectedUsers = selectedRows.map((row) => row.original as User)
    
    switch (action) {
      case 'activate':
        const activatePromise = new Promise((resolve) => {
          setTimeout(() => {
            resolve(selectedUsers)
          }, 2000)
        })
        
        toast.promise(activatePromise, {
          loading: 'Activating users...',
          success: () => {
            table.resetRowSelection()
            return `Activated ${selectedUsers.length} user${selectedUsers.length > 1 ? 's' : ''}`
          },
          error: 'Error activating users',
        })
        break
        
      case 'deactivate':
        const deactivatePromise = new Promise((resolve) => {
          setTimeout(() => {
            resolve(selectedUsers)
          }, 2000)
        })
        
        toast.promise(deactivatePromise, {
          loading: 'Deactivating users...',
          success: () => {
            table.resetRowSelection()
            return `Deactivated ${selectedUsers.length} user${selectedUsers.length > 1 ? 's' : ''}`
          },
          error: 'Error deactivating users',
        })
        break
        
      case 'email':
        const emailPromise = new Promise((resolve) => {
          setTimeout(() => {
            resolve(selectedUsers)
          }, 2000)
        })
        
        toast.promise(emailPromise, {
          loading: 'Sending emails...',
          success: () => {
            table.resetRowSelection()
            return `Sent emails to ${selectedUsers.length} user${selectedUsers.length > 1 ? 's' : ''}`
          },
          error: 'Error sending emails',
        })
        break
        
      default:
        break
    }
  }

  const handleExport = () => {
    const selectedUsers = selectedRows.map((row) => row.original as User)
    
    // Simulate export
    const exportPromise = new Promise((resolve) => {
      setTimeout(() => {
        resolve(selectedUsers)
      }, 1000)
    })
    
    toast.promise(exportPromise, {
      loading: 'Exporting users...',
      success: `Exported ${selectedUsers.length} user${selectedUsers.length > 1 ? 's' : ''}`,
      error: 'Error exporting users',
    })
  }

  return (
    <>
      <BulkActionsToolbar
        table={table}
        onDelete={() => setShowDeleteConfirm(true)}
        onExport={handleExport}
        onBulkAction={handleBulkAction}
      />

      <UsersMultiDeleteDialog
        table={table}
        open={showDeleteConfirm}
        onOpenChange={setShowDeleteConfirm}
      />
    </>
  )
}