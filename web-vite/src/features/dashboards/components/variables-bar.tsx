import { RotateCcw } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import type { Variable, VariableValues } from '../types'

interface VariablesBarProps {
  variables: Variable[]
  values: VariableValues
  onValueChange: (name: string, value: unknown) => void
  onReset: () => void
  hasActiveVariables: boolean
  getVariableOptions: (variable: Variable) => string[]
  isLoadingOptions: boolean
  className?: string
}

export function VariablesBar({
  variables,
  values,
  onValueChange,
  onReset,
  hasActiveVariables,
  getVariableOptions,
  isLoadingOptions,
  className,
}: VariablesBarProps) {
  if (!variables || variables.length === 0) return null

  return (
    <div
      className={cn(
        'flex flex-wrap items-center gap-4 rounded-lg border bg-card p-3',
        className,
      )}
    >
      {variables.map((variable) => (
        <VariableInput
          key={variable.name}
          variable={variable}
          value={values[variable.name]}
          onChange={(value) => onValueChange(variable.name, value)}
          options={getVariableOptions(variable)}
          isLoadingOptions={isLoadingOptions && variable.type === 'query'}
        />
      ))}

      {hasActiveVariables && (
        <Button
          variant="ghost"
          size="sm"
          onClick={onReset}
          className="ml-auto text-muted-foreground hover:text-foreground"
        >
          <RotateCcw className="mr-1.5 h-3.5 w-3.5" />
          Reset
        </Button>
      )}
    </div>
  )
}

interface VariableInputProps {
  variable: Variable
  value: unknown
  onChange: (value: unknown) => void
  options: string[]
  isLoadingOptions: boolean
}

function VariableInput({
  variable,
  value,
  onChange,
  options,
  isLoadingOptions,
}: VariableInputProps) {
  const label = variable.label ?? variable.name
  const stringValue = value !== undefined && value !== null ? String(value) : ''

  if ((variable.type === 'select' || variable.type === 'query') && variable.multi) {
    const selectedValues = Array.isArray(value) ? (value as string[]) : []

    return (
      <div className="flex flex-col gap-1.5">
        <Label className="text-xs text-muted-foreground">{label}</Label>
        {isLoadingOptions ? (
          <Skeleton className="h-9 w-40" />
        ) : (
          <div className="flex flex-wrap items-center gap-1.5 min-h-9 rounded-md border px-2 py-1.5">
            {options.map((option) => {
              const isSelected = selectedValues.includes(option)
              return (
                <Badge
                  key={option}
                  variant={isSelected ? 'default' : 'outline'}
                  className={cn(
                    'cursor-pointer transition-colors',
                    isSelected
                      ? 'bg-primary text-primary-foreground'
                      : 'hover:bg-muted',
                  )}
                  onClick={() => {
                    if (isSelected) {
                      onChange(selectedValues.filter((v) => v !== option))
                    } else {
                      onChange([...selectedValues, option])
                    }
                  }}
                >
                  {option}
                </Badge>
              )
            })}
            {options.length === 0 && (
              <span className="text-xs text-muted-foreground">No options</span>
            )}
          </div>
        )}
      </div>
    )
  }

  if (variable.type === 'select' || variable.type === 'query') {
    return (
      <div className="flex flex-col gap-1.5">
        <Label className="text-xs text-muted-foreground">{label}</Label>
        {isLoadingOptions ? (
          <Skeleton className="h-9 w-40" />
        ) : (
          <Select value={stringValue} onValueChange={onChange}>
            <SelectTrigger className="h-9 w-40">
              <SelectValue placeholder={`Select ${label}`} />
            </SelectTrigger>
            <SelectContent>
              {options.map((option) => (
                <SelectItem key={option} value={option}>
                  {option}
                </SelectItem>
              ))}
              {options.length === 0 && (
                <SelectItem value="__none__" disabled>
                  No options available
                </SelectItem>
              )}
            </SelectContent>
          </Select>
        )}
      </div>
    )
  }

  if (variable.type === 'number') {
    const numValue =
      typeof value === 'number' ? value : parseFloat(stringValue) || 0

    return (
      <div className="flex flex-col gap-1.5">
        <Label className="text-xs text-muted-foreground">{label}</Label>
        <Input
          type="number"
          value={numValue}
          onChange={(e) => onChange(parseFloat(e.target.value) || 0)}
          className="h-9 w-32"
        />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-1.5">
      <Label className="text-xs text-muted-foreground">{label}</Label>
      <Input
        type="text"
        value={stringValue}
        onChange={(e) => onChange(e.target.value)}
        placeholder={`Enter ${label}`}
        className="h-9 w-40"
      />
    </div>
  )
}

export type { VariablesBarProps }
