import { BrokleAPIClient } from '@/lib/api/core/client'
import type { OverviewResponse, OverviewTimeRange } from '../types'

const client = new BrokleAPIClient('/api')

/**
 * Fetch project overview data
 */
export const getProjectOverview = async (
  projectId: string,
  timeRange: OverviewTimeRange = '24h'
): Promise<OverviewResponse> => {
  return client.get<OverviewResponse>(
    `/v1/projects/${projectId}/overview`,
    { time_range: timeRange }
  )
}
