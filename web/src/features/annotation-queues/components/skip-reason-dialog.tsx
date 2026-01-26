'use client'

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Textarea } from '@/components/ui/textarea'
import { Loader2 } from 'lucide-react'

/**
 * Predefined skip reasons for structured data collection
 * Helps improve queue quality by understanding why items are skipped
 */
const SKIP_REASONS = [
  { value: 'missing_context', label: 'Missing context', description: 'Not enough information to evaluate' },
  { value: 'unclear_output', label: 'Unclear output', description: 'Output is ambiguous or hard to interpret' },
  { value: 'technical_issue', label: 'Technical issue', description: 'Error loading data or display problems' },
  { value: 'not_applicable', label: 'Not applicable', description: 'Item doesn\'t match evaluation criteria' },
  { value: 'other', label: 'Other', description: 'Specify a custom reason' },
] as const

type SkipReasonValue = (typeof SKIP_REASONS)[number]['value']

interface SkipReasonDialogProps {
  /** Whether the dialog is open */
  open: boolean
  /** Callback when dialog open state changes */
  onOpenChange: (open: boolean) => void
  /** Callback when user confirms skip with a reason */
  onConfirm: (reason: string) => void
  /** Whether the skip operation is in progress */
  isLoading?: boolean
}

/**
 * Dialog for capturing structured skip reasons
 *
 * Follows competitive analysis best practices:
 * - Predefined reasons for structured data
 * - "Other" option with free-text input
 * - Clear labels and descriptions
 * - Improves queue quality analytics
 */
export function SkipReasonDialog({
  open,
  onOpenChange,
  onConfirm,
  isLoading = false,
}: SkipReasonDialogProps) {
  const [selectedReason, setSelectedReason] = useState<SkipReasonValue | ''>('')
  const [customReason, setCustomReason] = useState('')

  const handleConfirm = () => {
    if (!selectedReason) return

    // For "other", use the custom reason; otherwise use the predefined label
    const reason =
      selectedReason === 'other'
        ? customReason.trim() || 'Other (no details provided)'
        : SKIP_REASONS.find((r) => r.value === selectedReason)?.label ?? selectedReason

    onConfirm(reason)
  }

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      // Reset state when closing
      setSelectedReason('')
      setCustomReason('')
    }
    onOpenChange(newOpen)
  }

  // Can submit if a reason is selected, and if "other" is selected, custom text is provided
  const canSubmit =
    selectedReason !== '' &&
    (selectedReason !== 'other' || customReason.trim().length > 0)

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Skip this item?</DialogTitle>
          <DialogDescription>
            Please select a reason for skipping. This helps improve queue quality.
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <RadioGroup
            value={selectedReason}
            onValueChange={(value) => setSelectedReason(value as SkipReasonValue)}
            className="space-y-3"
          >
            {SKIP_REASONS.map((reason) => (
              <div key={reason.value} className="flex items-start space-x-3">
                <RadioGroupItem
                  value={reason.value}
                  id={reason.value}
                  className="mt-1"
                />
                <div className="flex-1">
                  <Label
                    htmlFor={reason.value}
                    className="font-medium cursor-pointer"
                  >
                    {reason.label}
                  </Label>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {reason.description}
                  </p>
                </div>
              </div>
            ))}
          </RadioGroup>

          {/* Custom reason text area - only shown when "Other" is selected */}
          {selectedReason === 'other' && (
            <div className="space-y-2 pt-2">
              <Label htmlFor="custom-reason" className="text-sm font-medium">
                Please describe the reason
              </Label>
              <Textarea
                id="custom-reason"
                placeholder="Enter your reason for skipping..."
                value={customReason}
                onChange={(e) => setCustomReason(e.target.value)}
                rows={3}
                className="resize-none"
                autoFocus
              />
            </div>
          )}
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button
            variant="outline"
            onClick={() => handleOpenChange(false)}
            disabled={isLoading}
          >
            Cancel
          </Button>
          <Button
            onClick={handleConfirm}
            disabled={!canSubmit || isLoading}
            variant="destructive"
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Skipping...
              </>
            ) : (
              'Skip Item'
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
