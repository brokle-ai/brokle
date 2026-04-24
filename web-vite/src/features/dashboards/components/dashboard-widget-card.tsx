import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Widget } from '../api/types'

interface DashboardWidgetCardProps {
  widget: Widget
}

// Static viewer for a single widget. Executing widget queries
// (backend: `POST /dashboards/{id}/widgets/{wid}/execute`) is the
// next-port scope. Today we surface:
//   - widget type as a badge (so a dashboard author can tell at a
//     glance which tile is which)
//   - title + description
//   - a TODO panel standing in for the eventual chart/table render
export function DashboardWidgetCard({ widget }: DashboardWidgetCardProps) {
  return (
    <Card className="h-full">
      <CardHeader className="gap-1">
        <div className="flex items-center justify-between gap-2">
          <CardTitle className="text-sm font-medium">
            {widget.title || 'Untitled widget'}
          </CardTitle>
          <Badge variant="outline" className="font-mono text-[10px] uppercase">
            {widget.type}
          </Badge>
        </div>
        {widget.description ? (
          <p className="text-xs text-muted-foreground">{widget.description}</p>
        ) : null}
      </CardHeader>
      <CardContent>
        <div className="flex h-40 items-center justify-center rounded-md border border-dashed text-xs text-muted-foreground">
          Widget rendering: TODO
        </div>
      </CardContent>
    </Card>
  )
}
