'use client'

import { useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { ChevronRight } from 'lucide-react'
import { diffWords, diffJson } from 'diff'
import type { PromptVersion, VersionDiff as VersionDiffType } from '../../types'

interface VersionDiffProps {
  diff: VersionDiffType
}

export function VersionDiff({ diff }: VersionDiffProps) {
  const { from_version, to_version, template_from, template_to, config_from, config_to, variables_added, variables_removed } = diff

  // Word-level template diff
  const templateDiff = useMemo(() => {
    const oldTemplate =
      typeof template_from === 'string'
        ? template_from
        : JSON.stringify(template_from, null, 2)
    const newTemplate =
      typeof template_to === 'string'
        ? template_to
        : JSON.stringify(template_to, null, 2)

    return diffWords(oldTemplate, newTemplate)
  }, [template_from, template_to])

  // JSON config diff
  const configDiff = useMemo(() => {
    if (!config_from && !config_to) return null
    return diffJson(config_from || {}, config_to || {})
  }, [config_from, config_to])

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-center gap-4">
        <Badge variant="outline" className="font-mono text-lg px-4 py-2">
          v{from_version}
        </Badge>
        <ChevronRight className="h-6 w-6 text-muted-foreground" />
        <Badge variant="default" className="font-mono text-lg px-4 py-2">
          v{to_version}
        </Badge>
      </div>

      {/* Variable Changes */}
      {(variables_added.length > 0 || variables_removed.length > 0) && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Variable Changes</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2">
            {variables_added.length > 0 && (
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-green-600 dark:text-green-400">
                  Added:
                </span>
                <div className="flex flex-wrap gap-1">
                  {variables_added.map((v) => (
                    <Badge
                      key={v}
                      variant="outline"
                      className="bg-green-100 dark:bg-green-900/30 font-mono"
                    >
                      +{`{{${v}}}`}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
            {variables_removed.length > 0 && (
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-red-600 dark:text-red-400">
                  Removed:
                </span>
                <div className="flex flex-wrap gap-1">
                  {variables_removed.map((v) => (
                    <Badge
                      key={v}
                      variant="outline"
                      className="bg-red-100 dark:bg-red-900/30 font-mono"
                    >
                      -{`{{${v}}}`}
                    </Badge>
                  ))}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      )}

      {/* Template Diff with Word-Level Highlighting */}
      <Card>
        <CardHeader>
          <CardTitle className="text-lg">Template Changes</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="rounded-lg bg-muted p-4 font-mono text-sm overflow-x-auto">
            <pre className="whitespace-pre-wrap">
              {templateDiff.map((part, index) => (
                <span
                  key={index}
                  className={
                    part.added
                      ? 'bg-green-500/20 text-green-700 dark:text-green-300'
                      : part.removed
                      ? 'bg-red-500/20 text-red-700 dark:text-red-300 line-through'
                      : ''
                  }
                >
                  {part.value}
                </span>
              ))}
            </pre>
          </div>
        </CardContent>
      </Card>

      {/* Config Diff */}
      {configDiff && (
        <Card>
          <CardHeader>
            <CardTitle className="text-lg">Configuration Changes</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="rounded-lg bg-muted p-4 font-mono text-sm overflow-x-auto">
              <pre className="whitespace-pre-wrap">
                {configDiff.map((part, index) => (
                  <span
                    key={index}
                    className={
                      part.added
                        ? 'bg-green-500/20 text-green-700 dark:text-green-300'
                        : part.removed
                        ? 'bg-red-500/20 text-red-700 dark:text-red-300'
                        : ''
                    }
                  >
                    {part.value}
                  </span>
                ))}
              </pre>
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  )
}
