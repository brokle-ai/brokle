'use client'

import { Loader2 } from 'lucide-react'
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
import { useRules } from '../context/rules-context'
import { useDeleteEvaluationRuleMutation } from '../hooks/use-evaluation-rules'
import { EditRuleDialog } from './edit-rule-dialog'

export function RulesDialogs() {
  const { open, setOpen, currentRow, projectId, orgId } = useRules()
  const deleteMutation = useDeleteEvaluationRuleMutation(projectId ?? '')

  const handleDelete = async () => {
    if (!currentRow || !projectId) return
    await deleteMutation.mutateAsync({
      ruleId: currentRow.id,
      ruleName: currentRow.name,
    })
    setOpen(null)
  }

  return (
    <>
      <EditRuleDialog
        projectId={projectId ?? ''}
        orgId={orgId}
        rule={currentRow}
        open={open === 'edit'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'edit' : null)}
      />

      <AlertDialog
        open={open === 'delete'}
        onOpenChange={(isOpen) => setOpen(isOpen ? 'delete' : null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Evaluation Rule</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete &quot;{currentRow?.name}&quot;?
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>
              Cancel
            </AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={deleteMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Deleting...
                </>
              ) : (
                'Delete'
              )}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  )
}
