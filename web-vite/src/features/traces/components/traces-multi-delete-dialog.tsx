import { useState } from 'react'
import { type Table } from '@tanstack/react-table'
import { AlertTriangle } from 'lucide-react'
import { toast } from 'sonner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ConfirmDialog } from '@/components/confirm-dialog'
import type { TraceListItem } from '../api/types'

type TracesMultiDeleteDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  table: Table<TraceListItem>
}

const CONFIRM_WORD = 'DELETE'

/**
 * Typed-confirmation bulk-delete dialog. The backend currently has no
 * bulk-delete endpoint (see web/'s parity implementation), so the
 * handler resolves a sleep to mirror the "eventually" UX. When the
 * endpoint lands, swap the `setTimeout` for a mutation.
 */
export function TracesMultiDeleteDialog({
  open,
  onOpenChange,
  table,
}: TracesMultiDeleteDialogProps) {
  const [value, setValue] = useState('')

  const selectedRows = table.getSelectedRowModel().flatRows
  const selectedTraces = selectedRows.map((row) => row.original)
  const selectedCount = selectedTraces.length

  const handleDelete = () => {
    if (selectedTraces.length === 0) {
      toast.error('No traces selected to delete')
      return
    }
    if (value.trim() !== CONFIRM_WORD) {
      toast.error(`Please type "${CONFIRM_WORD}" to confirm.`)
      return
    }
    onOpenChange(false)
    toast.promise(
      new Promise<void>((resolve) => setTimeout(resolve, 1500)),
      {
        loading: 'Deleting traces…',
        success: () => {
          table.resetRowSelection()
          return `Deleted ${selectedCount} ${
            selectedCount > 1 ? 'traces' : 'trace'
          }`
        },
        error: 'Error',
      },
    )
  }

  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      handleConfirm={handleDelete}
      disabled={value.trim() !== CONFIRM_WORD}
      title={
        <span className="text-destructive">
          <AlertTriangle
            className="stroke-destructive me-1 inline-block"
            size={18}
          />{' '}
          Delete {selectedCount} {selectedCount > 1 ? 'traces' : 'trace'}
        </span>
      }
      desc={
        <div className="space-y-4">
          <p className="mb-2">
            Are you sure you want to delete the selected traces? <br />
            This action cannot be undone.
          </p>
          <Label className="my-4 flex flex-col items-start gap-1.5">
            <span>Confirm by typing "{CONFIRM_WORD}":</span>
            <Input
              value={value}
              onChange={(e) => setValue(e.target.value)}
              placeholder={`Type "${CONFIRM_WORD}" to confirm.`}
            />
          </Label>
          <Alert variant="destructive">
            <AlertTitle>Warning!</AlertTitle>
            <AlertDescription>
              Please be careful, this operation can not be rolled back.
            </AlertDescription>
          </Alert>
        </div>
      }
      confirmText="Delete"
      destructive
    />
  )
}
