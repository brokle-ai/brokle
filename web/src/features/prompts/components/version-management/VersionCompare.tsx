'use client'

import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ArrowLeftRight } from 'lucide-react'
import type { PromptVersion } from '../../types'

interface VersionCompareProps {
  versions: PromptVersion[]
  fromVersion?: number
  toVersion?: number
  onFromChange: (version: number) => void
  onToChange: (version: number) => void
  onCompare: () => void
  disabled?: boolean
}

export function VersionCompare({
  versions,
  fromVersion,
  toVersion,
  onFromChange,
  onToChange,
  onCompare,
  disabled,
}: VersionCompareProps) {
  const canCompare = fromVersion !== undefined && toVersion !== undefined && fromVersion !== toVersion

  return (
    <div className="flex items-center gap-3">
      <Select
        value={fromVersion?.toString()}
        onValueChange={(v) => onFromChange(parseInt(v))}
      >
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="From version" />
        </SelectTrigger>
        <SelectContent>
          {versions.map((v) => (
            <SelectItem key={v.version} value={v.version.toString()}>
              Version {v.version}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <ArrowLeftRight className="h-4 w-4 text-muted-foreground" />

      <Select
        value={toVersion?.toString()}
        onValueChange={(v) => onToChange(parseInt(v))}
      >
        <SelectTrigger className="w-[140px]">
          <SelectValue placeholder="To version" />
        </SelectTrigger>
        <SelectContent>
          {versions.map((v) => (
            <SelectItem key={v.version} value={v.version.toString()}>
              Version {v.version}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>

      <Button onClick={onCompare} disabled={!canCompare || disabled}>
        Compare
      </Button>
    </div>
  )
}
