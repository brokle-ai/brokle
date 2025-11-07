'use client'

import { Overview } from './overview'
import { RecentSales } from './recent-sales'

export function DashboardView() {
  return (
    <div className="space-y-6">
      <Overview />
      <RecentSales />
    </div>
  )
}
