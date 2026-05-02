import { useState } from 'react'
import { Loader2 } from 'lucide-react'
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

// Predefined skip reasons feed structured queue-quality analytics. The
// "other" branch lets reviewers type a free-form reason — required to
// capture skips that don't fit the canned list (e.g., bug reports).
const SKIP_REASONS = [
  {
    value: 'missing_context',
    label: 'Missing context',
    description: 'Not enough information to evaluate',
  },
  {
    value: 'unclear_output',
    label: 'Unclear output',
    description: 'Output is ambiguous or hard to interpret',
  },
  {
    value: 'technical_issue',
    label: 'Technical issue',
    description: 'Error loading data or display problems',
  },
  {
    value: 'not_applicable',
    label: 'Not applicable',
    description: "Item doesn't match evaluation criteria",
  },
  {
    value: 'other',
    label: 'Other',
    description: 'Specify a custom reason',
  },
] as const

type SkipReasonValue = (typeof SKIP_REASONS)[number]['value']

interface SkipReasonDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onConfirm: (reason: string) => void
  isLoading?: boolean
}

export function SkipReasonDialog({
  open,
  onOpenChange,
  onConfirm,
  isLoading = false,
}: SkipReasonDialogProps) {
  const [selected, setSelected] = useState<SkipReasonValue | ''>('')
  const [custom, setCustom] = useState('')

  const handleConfirm = () => {
    if (!selected) return
    const reason =
      selected === 'other'
        ? custom.trim() || 'Other (no details provided)'
        : (SKIP_REASONS.find((r) => r.value === selected)?.label ?? selected)
    onConfirm(reason)
  }

  const handleOpenChange = (next: boolean) => {
    if (!next) {
      setSelected('')
      setCustom('')
    }
    onOpenChange(next)
  }

  const canSubmit =
    selected !== '' && (selected !== 'other' || custom.trim().length > 0)

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
            value={selected}
            onValueChange={(v) => setSelected(v as SkipReasonValue)}
            className="space-y-3"
          >
            {SKIP_REASONS.map((r) => (
              <div key={r.value} className="flex items-start space-x-3">
                <RadioGroupItem value={r.value} id={r.value} className="mt-1" />
                <div className="flex-1">
                  <Label
                    htmlFor={r.value}
                    className="font-medium cursor-pointer"
                  >
                    {r.label}
                  </Label>
                  <p className="text-xs text-muted-foreground mt-0.5">
                    {r.description}
                  </p>
                </div>
              </div>
            ))}
          </RadioGroup>

          {selected === 'other' && (
            <div className="space-y-2 pt-2">
              <Label htmlFor="custom-reason" className="text-sm font-medium">
                Please describe the reason
              </Label>
              <Textarea
                id="custom-reason"
                placeholder="Enter your reason for skipping..."
                value={custom}
                onChange={(e) => setCustom(e.target.value)}
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
