'use client'

import { useTraces } from '../context/traces-context'
import { TracesMultiDeleteDialog } from './traces-multi-delete-dialog'

export function TracesDialogs() {
  const { open, setOpen } = useTraces()

  // Only delete dialog for traces (read-only feature)
  return (
    <TracesMultiDeleteDialog
      open={open === 'delete'}
      onOpenChange={(isOpen) => setOpen(isOpen ? 'delete' : null)}
      table={null as any} // Will be passed from parent
    />
  )
}
