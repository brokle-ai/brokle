import { BrokleAPIClient } from '@/lib/api/core/client'
import { timeRangeToApiParams } from '@/components/shared/time-range-picker'
import type { TimeRange } from '@/components/shared/time-range-picker'
import type { OverviewResponse } from '../types'

const client = new BrokleAPIClient('/api')

// Default time range
const DEFAULT_TIME_RANGE: TimeRange = { relative: '24h' }

/**
 * Fetch project overview data
 */
export const getProjectOverview = async (
  projectId: string,
  timeRange: TimeRange = DEFAULT_TIME_RANGE
): Promise<OverviewResponse> => {
  const params = timeRangeToApiParams(timeRange)
  return client.get<OverviewResponse>(
    `/v1/projects/${projectId}/overview`,
    params
  )
}
