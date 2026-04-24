import { Badge } from '@/components/ui/badge'
import { DashboardWidgetCard } from './dashboard-widget-card'
import type { DashboardDetail as DashboardDetailType } from '../api/types'

interface DashboardDetailProps {
  dashboard: DashboardDetailType
}

// Static dashboard viewer. Renders widgets in `config.widgets` order
// as a 2-column responsive grid. Layout-driven placement (x/y/w/h
// from `dashboard.layout`) and drag/drop editing are second-port
// concerns — we intentionally ignore `dashboard.layout` here so the
// viewer stays simple and additive. When the grid editor lands, it
// will read `layout` and place widgets by widget_id.
export function DashboardDetail({ dashboard }: DashboardDetailProps) {
  const widgets = dashboard.config?.widgets ?? []

  return (
    <div className="space-y-4">
      <header className="space-y-1">
        <div className="flex items-center gap-2">
          <h1 className="text-2xl font-semibold">{dashboard.name}</h1>
          {dashboard.is_locked ? (
            <Badge variant="secondary">Locked</Badge>
          ) : null}
        </div>
        {dashboard.description ? (
          <p className="text-sm text-muted-foreground">
            {dashboard.description}
          </p>
        ) : null}
      </header>

      {widgets.length === 0 ? (
        <div className="rounded-lg border p-12 text-center">
          <p className="text-sm text-muted-foreground">
            This dashboard has no widgets yet.
          </p>
          <p className="mt-1 text-xs text-muted-foreground">
            The widget editor will land in the next iteration.
          </p>
        </div>
      ) : (
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
          {widgets.map((w) => (
            <DashboardWidgetCard key={w.id} widget={w} />
          ))}
        </div>
      )}
    </div>
  )
}
