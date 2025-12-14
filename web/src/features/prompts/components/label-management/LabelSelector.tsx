'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import { cn } from '@/lib/utils'
import { Plus, X, Lock, AlertTriangle } from 'lucide-react'
import { Alert, AlertDescription } from '@/components/ui/alert'

interface LabelSelectorProps {
  labels: string[]
  protectedLabels?: string[]
  availableLabels?: string[]
  onChange: (labels: string[]) => void
  isLoading?: boolean
}

export function LabelSelector({
  labels,
  protectedLabels = [],
  availableLabels = [],
  onChange,
  isLoading,
}: LabelSelectorProps) {
  const [isAddOpen, setIsAddOpen] = useState(false)
  const [newLabel, setNewLabel] = useState('')
  const [removeConfirm, setRemoveConfirm] = useState<string | null>(null)

  const handleAddLabel = () => {
    const label = newLabel.trim().toLowerCase()
    if (label && !labels.includes(label)) {
      onChange([...labels, label])
    }
    setNewLabel('')
    setIsAddOpen(false)
  }

  const handleRemoveLabel = (label: string) => {
    if (protectedLabels.includes(label)) {
      setRemoveConfirm(label)
    } else {
      onChange(labels.filter((l) => l !== label))
    }
  }

  const confirmRemoveProtected = () => {
    if (removeConfirm) {
      onChange(labels.filter((l) => l !== removeConfirm))
      setRemoveConfirm(null)
    }
  }

  // Suggestions: labels used elsewhere but not on this version
  const suggestions = availableLabels.filter(
    (l) => !labels.includes(l) && l !== 'latest'
  )

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap gap-2">
        {labels.map((label) => {
          const isProtected = protectedLabels.includes(label)
          const isLatest = label === 'latest'

          return (
            <Badge
              key={label}
              variant="secondary"
              className={cn(
                'flex items-center gap-1 pr-1',
                isProtected && 'bg-amber-100 dark:bg-amber-900/30',
                isLatest && 'bg-purple-100 dark:bg-purple-900/30'
              )}
            >
              {isProtected && <Lock className="h-3 w-3" />}
              {label}
              {!isLatest && (
                <Button
                  variant="ghost"
                  size="sm"
                  className="h-4 w-4 p-0 hover:bg-transparent"
                  onClick={() => handleRemoveLabel(label)}
                  disabled={isLoading}
                >
                  <X className="h-3 w-3" />
                </Button>
              )}
            </Badge>
          )
        })}

        <Popover open={isAddOpen} onOpenChange={setIsAddOpen}>
          <PopoverTrigger asChild>
            <Button
              variant="outline"
              size="sm"
              className="h-6"
              disabled={isLoading}
            >
              <Plus className="h-3 w-3 mr-1" />
              Add label
            </Button>
          </PopoverTrigger>
          <PopoverContent className="w-64 p-2" align="start">
            <div className="space-y-2">
              <div className="flex gap-2">
                <Input
                  value={newLabel}
                  onChange={(e) => setNewLabel(e.target.value)}
                  placeholder="Label name"
                  className="h-8"
                  onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                      handleAddLabel()
                    }
                  }}
                />
                <Button size="sm" onClick={handleAddLabel}>
                  Add
                </Button>
              </div>
              {suggestions.length > 0 && (
                <div className="space-y-1">
                  <p className="text-xs text-muted-foreground">Suggestions:</p>
                  <div className="flex flex-wrap gap-1">
                    {suggestions.slice(0, 5).map((label) => (
                      <Badge
                        key={label}
                        variant="outline"
                        className="cursor-pointer hover:bg-muted"
                        onClick={() => {
                          onChange([...labels, label])
                          setIsAddOpen(false)
                        }}
                      >
                        {label}
                      </Badge>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </PopoverContent>
        </Popover>
      </div>

      {/* Protected label removal confirmation */}
      <Dialog open={!!removeConfirm} onOpenChange={() => setRemoveConfirm(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangle className="h-5 w-5 text-amber-500" />
              Remove protected label
            </DialogTitle>
            <DialogDescription>
              "{removeConfirm}" is a protected label. Removing it may affect
              production systems that depend on this label for prompt resolution.
            </DialogDescription>
          </DialogHeader>
          <Alert variant="destructive" className="mt-2">
            <AlertDescription>
              This action could break SDK integrations using this label.
            </AlertDescription>
          </Alert>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRemoveConfirm(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={confirmRemoveProtected}
              disabled={isLoading}
            >
              Remove anyway
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
