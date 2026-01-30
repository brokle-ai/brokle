'use client'

import { useMemo, useCallback, useEffect } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Switch } from '@/components/ui/switch'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Badge } from '@/components/ui/badge'
import { ChevronDown, ChevronRight, Link2, RotateCcw, RefreshCw } from 'lucide-react'
import { cn } from '@/lib/utils'
import { usePlaygroundStore } from '../stores/playground-store'
import { extractVariablesFromMessages } from './variable-editor'

interface SharedVariablesPanelProps {
  className?: string
}

export function SharedVariablesPanel({ className }: SharedVariablesPanelProps) {
  const {
    windows,
    sharedVariables,
    useSharedVariables,
    setSharedVariables,
    toggleSharedVariables,
    updateWindow,
  } = usePlaygroundStore()

  // Extract all variables from all windows
  const allVariables = useMemo(() => {
    const variableSet = new Set<string>()
    for (const window of windows) {
      const extracted = extractVariablesFromMessages(window.messages)
      for (const v of extracted) {
        variableSet.add(v)
      }
    }
    return Array.from(variableSet).sort()
  }, [windows])

  // Count variables per window for summary
  const variableCount = allVariables.length

  // Handle variable value change
  const handleVariableChange = useCallback(
    (name: string, value: string) => {
      setSharedVariables({ ...sharedVariables, [name]: value })
    },
    [sharedVariables, setSharedVariables]
  )

  // Sync shared variables to all windows
  const handleSyncToWindows = useCallback(() => {
    // For each window, merge shared variables with window's relevant variables
    windows.forEach((window, index) => {
      const windowVars = extractVariablesFromMessages(window.messages)
      const mergedVars = { ...window.variables }

      // Only set values for variables that exist in this window's template
      for (const v of windowVars) {
        if (sharedVariables[v] !== undefined) {
          mergedVars[v] = sharedVariables[v]
        }
      }

      updateWindow(index, { variables: mergedVars })
    })
  }, [windows, sharedVariables, updateWindow])

  // Collect values from windows into shared variables
  const handleCollectFromWindows = useCallback(() => {
    const collected: Record<string, string> = { ...sharedVariables }

    // For each variable, find the first non-empty value from windows
    for (const v of allVariables) {
      if (!collected[v]) {
        for (const window of windows) {
          if (window.variables[v]) {
            collected[v] = window.variables[v]
            break
          }
        }
      }
    }

    setSharedVariables(collected)
  }, [allVariables, windows, sharedVariables, setSharedVariables])

  // Clear all shared variable values
  const handleClearAll = useCallback(() => {
    const cleared = Object.keys(sharedVariables).reduce(
      (acc, key) => ({ ...acc, [key]: '' }),
      {} as Record<string, string>
    )
    setSharedVariables(cleared)
  }, [sharedVariables, setSharedVariables])

  // Initialize shared variables with empty values for all detected variables
  const initializeVariables = useCallback(() => {
    const initialized: Record<string, string> = { ...sharedVariables }
    for (const v of allVariables) {
      if (initialized[v] === undefined) {
        initialized[v] = ''
      }
    }
    setSharedVariables(initialized)
  }, [allVariables, sharedVariables, setSharedVariables])

  // Auto-initialize when variables are detected
  useEffect(() => {
    if (useSharedVariables && allVariables.length > 0) {
      const needsInit = allVariables.some((v) => sharedVariables[v] === undefined)
      if (needsInit) {
        initializeVariables()
      }
    }
  }, [useSharedVariables, allVariables, sharedVariables, initializeVariables])

  return (
    <Collapsible defaultOpen={useSharedVariables} className={cn('w-full', className)}>
      <Card className="border-dashed">
        <CollapsibleTrigger asChild>
          <CardHeader className="cursor-pointer hover:bg-muted/50 transition-colors py-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ChevronDown className="h-4 w-4 transition-transform duration-200 group-data-[state=closed]:rotate-[-90deg]" />
                <Link2 className="h-4 w-4 text-muted-foreground" />
                <CardTitle className="text-sm font-medium">Shared Variables</CardTitle>
                <Badge variant="secondary" className="ml-2">
                  {variableCount} variable{variableCount !== 1 ? 's' : ''} across {windows.length} window
                  {windows.length !== 1 ? 's' : ''}
                </Badge>
              </div>
              <div className="flex items-center gap-2" onClick={(e) => e.stopPropagation()}>
                <Label htmlFor="shared-toggle" className="text-xs text-muted-foreground">
                  Enable
                </Label>
                <Switch
                  id="shared-toggle"
                  checked={useSharedVariables}
                  onCheckedChange={toggleSharedVariables}
                />
              </div>
            </div>
          </CardHeader>
        </CollapsibleTrigger>

        <CollapsibleContent>
          <CardContent className="pt-0">
            {!useSharedVariables ? (
              <p className="text-sm text-muted-foreground">
                Enable shared variables to use the same variable values across all windows.
                Variables are detected automatically from message templates.
              </p>
            ) : allVariables.length === 0 ? (
              <p className="text-sm text-muted-foreground italic">
                No variables found in any window. Use{' '}
                <code className="text-xs bg-muted px-1 py-0.5 rounded">{'{{ }}'}</code> syntax
                in your messages to add variables.
              </p>
            ) : (
              <div className="space-y-4">
                {/* Actions row */}
                <div className="flex items-center gap-2 flex-wrap">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleSyncToWindows}
                    title="Push shared values to all windows"
                  >
                    <RefreshCw className="mr-2 h-3 w-3" />
                    Sync to Windows
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={handleCollectFromWindows}
                    title="Collect values from windows"
                  >
                    Collect from Windows
                  </Button>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={handleClearAll}
                    title="Clear all values"
                  >
                    <RotateCcw className="mr-2 h-3 w-3" />
                    Clear All
                  </Button>
                </div>

                {/* Variable grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                  {allVariables.map((variable) => {
                    // Count how many windows use this variable
                    const usageCount = windows.filter((w) =>
                      extractVariablesFromMessages(w.messages).includes(variable)
                    ).length

                    return (
                      <div key={variable} className="space-y-2">
                        <div className="flex items-center justify-between">
                          <Label htmlFor={`shared-${variable}`} className="font-mono text-sm">
                            {`{{${variable}}}`}
                          </Label>
                          <Badge variant="outline" className="text-xs">
                            {usageCount} window{usageCount !== 1 ? 's' : ''}
                          </Badge>
                        </div>
                        <Textarea
                          id={`shared-${variable}`}
                          value={sharedVariables[variable] || ''}
                          onChange={(e) => handleVariableChange(variable, e.target.value)}
                          placeholder={`Enter value for ${variable}...`}
                          className="min-h-[60px] font-mono text-sm"
                          rows={2}
                        />
                      </div>
                    )
                  })}
                </div>
              </div>
            )}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}
