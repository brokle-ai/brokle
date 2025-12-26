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
import { EditRuleDialog } from './edit-rule-dialog'
import { useRuleDetail } from '../context/rule-detail-context'

export function RuleDetailDialogs() {
  const {
    rule,
    open,
    setOpen,
    handleDelete,
    isDeleting,
    projectId,
  } = useRuleDetail()

  if (!rule) return null

  return (
    <>
      <EditRuleDialog
        projectId={projectId}
        rule={rule}
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
              Are you sure you want to delete &quot;{rule.name}&quot;? This action cannot be undone.
              The rule will stop evaluating spans immediately.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={isDeleting}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              disabled={isDeleting}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {isDeleting ? (
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
