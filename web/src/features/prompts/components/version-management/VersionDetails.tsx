'use client'

import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { formatDistanceToNow, format } from 'date-fns'
import { Clock, User, GitBranch, MessageSquare } from 'lucide-react'
import { LabelList } from '../label-badge'
import type { PromptVersion } from '../../types'

interface VersionDetailsProps {
  version: PromptVersion
  protectedLabels?: string[]
}

export function VersionDetails({ version, protectedLabels = [] }: VersionDetailsProps) {
  return (
    <Card>
      <CardHeader>
        <div className="flex items-center justify-between">
          <CardTitle className="text-lg">Version {version.version}</CardTitle>
          <div className="flex items-center gap-2">
            <LabelList labels={version.labels.map(name => ({ name }))} protectedLabels={protectedLabels} />
          </div>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {/* Commit Message */}
        {version.commit_message && (
          <div className="space-y-1">
            <div className="flex items-center gap-1 text-sm text-muted-foreground">
              <MessageSquare className="h-3 w-3" />
              <span className="font-medium">Commit Message</span>
            </div>
            <p className="text-sm">{version.commit_message}</p>
          </div>
        )}

        {/* Metadata */}
        <div className="space-y-2">
          <div className="flex items-center gap-2 text-sm">
            <Clock className="h-3 w-3 text-muted-foreground" />
            <span className="text-muted-foreground">Created:</span>
            <span title={format(new Date(version.created_at), 'PPpp')}>
              {formatDistanceToNow(new Date(version.created_at), { addSuffix: true })}
            </span>
          </div>

          {version.created_by && (
            <div className="flex items-center gap-2 text-sm">
              <User className="h-3 w-3 text-muted-foreground" />
              <span className="text-muted-foreground">Created by:</span>
              <span>{version.created_by}</span>
            </div>
          )}

          <div className="flex items-center gap-2 text-sm">
            <GitBranch className="h-3 w-3 text-muted-foreground" />
            <span className="text-muted-foreground">Variables:</span>
            <span className="font-mono">{version.variables.length}</span>
          </div>
        </div>

        {/* Variables List */}
        {version.variables.length > 0 && (
          <div className="space-y-1">
            <p className="text-sm font-medium text-muted-foreground">Variables:</p>
            <div className="flex flex-wrap gap-1">
              {version.variables.map((variable) => (
                <Badge key={variable} variant="outline" className="font-mono text-xs">
                  {`{{${variable}}}`}
                </Badge>
              ))}
            </div>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
