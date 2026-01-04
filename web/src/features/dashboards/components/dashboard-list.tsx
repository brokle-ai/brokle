'use client'

import { DashboardCard } from './dashboard-card'
import type { Dashboard } from '../types'

interface DashboardListProps {
  data: Dashboard[]
}

export function DashboardList({ data }: DashboardListProps) {
  return (
    <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      {data.map((dashboard) => (
        <DashboardCard key={dashboard.id} dashboard={dashboard} />
      ))}
    </div>
  )
}
