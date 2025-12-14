'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Lock, X } from 'lucide-react'

interface ProtectedLabelsConfigProps {
  labels: string[]
  onChange: (labels: string[]) => void
  isLoading?: boolean
}

export function ProtectedLabelsConfig({
  labels,
  onChange,
  isLoading,
}: ProtectedLabelsConfigProps) {
  const [newLabel, setNewLabel] = useState('')

  const handleAdd = () => {
    const label = newLabel.trim().toLowerCase()
    if (label && !labels.includes(label)) {
      onChange([...labels, label])
    }
    setNewLabel('')
  }

  const handleRemove = (label: string) => {
    onChange(labels.filter((l) => l !== label))
  }

  return (
    <div className="space-y-4">
      <p className="text-sm text-muted-foreground">
        Protected labels require additional confirmation before they can be
        modified. Use this for critical labels like "production" that shouldn't
        be accidentally changed.
      </p>

      <div className="flex gap-2">
        <Input
          value={newLabel}
          onChange={(e) => setNewLabel(e.target.value)}
          placeholder="Add protected label..."
          className="max-w-xs"
          onKeyDown={(e) => {
            if (e.key === 'Enter') {
              handleAdd()
            }
          }}
        />
        <Button onClick={handleAdd} disabled={isLoading}>
          Add
        </Button>
      </div>

      <div className="flex flex-wrap gap-2">
        {labels.length === 0 ? (
          <p className="text-sm text-muted-foreground italic">
            No protected labels configured
          </p>
        ) : (
          labels.map((label) => (
            <Badge
              key={label}
              variant="secondary"
              className="flex items-center gap-1 pr-1 bg-amber-100 dark:bg-amber-900/30"
            >
              <Lock className="h-3 w-3" />
              {label}
              <Button
                variant="ghost"
                size="sm"
                className="h-4 w-4 p-0 hover:bg-transparent"
                onClick={() => handleRemove(label)}
                disabled={isLoading}
              >
                <X className="h-3 w-3" />
              </Button>
            </Badge>
          ))
        )}
      </div>
    </div>
  )
}
