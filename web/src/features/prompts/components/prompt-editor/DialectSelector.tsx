'use client'

import { Code2Icon, SparklesIcon, BracesIcon } from 'lucide-react'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { TemplateDialect } from '@/features/prompts/types'

interface DialectSelectorProps {
  value: TemplateDialect
  onChange: (dialect: TemplateDialect) => void
  disabled?: boolean
  className?: string
}

interface DialectOption {
  value: TemplateDialect
  label: string
  description: string
  icon: React.ReactNode
}

const DIALECT_OPTIONS: DialectOption[] = [
  {
    value: 'auto',
    label: 'Auto-detect',
    description: 'Automatically detect template syntax',
    icon: <SparklesIcon className="size-4 text-muted-foreground" />,
  },
  {
    value: 'simple',
    label: 'Simple',
    description: '{{variable}} syntax only',
    icon: <BracesIcon className="size-4 text-muted-foreground" />,
  },
  {
    value: 'mustache',
    label: 'Mustache',
    description: 'Sections, loops, conditionals',
    icon: <Code2Icon className="size-4 text-muted-foreground" />,
  },
  {
    value: 'jinja2',
    label: 'Jinja2',
    description: 'Filters, inheritance, macros',
    icon: <Code2Icon className="size-4 text-muted-foreground" />,
  },
]

/**
 * DialectSelector - A dropdown for selecting template dialect.
 *
 * Supports:
 * - auto: Auto-detect from template content
 * - simple: Basic {{variable}} syntax
 * - mustache: Full Mustache syntax with sections
 * - jinja2: Jinja2 syntax with filters and blocks
 */
export function DialectSelector({
  value,
  onChange,
  disabled = false,
  className,
}: DialectSelectorProps) {
  const selectedOption = DIALECT_OPTIONS.find((opt) => opt.value === value)

  return (
    <Select
      value={value}
      onValueChange={(newValue) => onChange(newValue as TemplateDialect)}
      disabled={disabled}
    >
      <SelectTrigger className={className} size="sm">
        <SelectValue>
          {selectedOption && (
            <span className="flex items-center gap-2">
              {selectedOption.icon}
              <span>{selectedOption.label}</span>
            </span>
          )}
        </SelectValue>
      </SelectTrigger>
      <SelectContent>
        {DIALECT_OPTIONS.map((option) => (
          <SelectItem key={option.value} value={option.value}>
            <div className="flex items-center gap-2">
              {option.icon}
              <div className="flex flex-col">
                <span className="font-medium">{option.label}</span>
                <span className="text-xs text-muted-foreground">
                  {option.description}
                </span>
              </div>
            </div>
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  )
}

export function getDialectLabel(dialect: TemplateDialect): string {
  const option = DIALECT_OPTIONS.find((opt) => opt.value === dialect)
  return option?.label ?? dialect
}

export function getDialectDescription(dialect: TemplateDialect): string {
  const option = DIALECT_OPTIONS.find((opt) => opt.value === dialect)
  return option?.description ?? ''
}
