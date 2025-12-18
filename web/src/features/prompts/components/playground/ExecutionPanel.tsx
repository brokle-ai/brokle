'use client'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Loader2, Play } from 'lucide-react'

interface ExecutionPanelProps {
  onExecute: () => void
  isExecuting: boolean
  missingVariables: string[]
  disabled?: boolean
}

export function ExecutionPanel({
  onExecute,
  isExecuting,
  missingVariables,
  disabled,
}: ExecutionPanelProps) {
  const canExecute = missingVariables.length === 0 && !disabled

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-lg">Execute Prompt</CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        <Button
          onClick={onExecute}
          disabled={!canExecute || isExecuting}
          className="w-full"
          size="lg"
        >
          {isExecuting ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Executing...
            </>
          ) : (
            <>
              <Play className="mr-2 h-4 w-4" />
              Execute
            </>
          )}
        </Button>

        {missingVariables.length > 0 && (
          <div className="rounded-md bg-amber-50 dark:bg-amber-950/30 p-3">
            <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
              Missing variables:
            </p>
            <p className="text-sm text-amber-700 dark:text-amber-300 font-mono mt-1">
              {missingVariables.join(', ')}
            </p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
