/**
 * Usage API
 *
 * API functions for billable usage endpoints
 * Usage-based billing: Spans + GB Processed + Scores
 */

import { BrokleAPIClient } from '@/lib/api/core/client'
import { timeRangeToApiParams } from '@/components/shared/time-range-picker'
import type { TimeRange } from '@/components/shared/time-range-picker'
import type {
  UsageOverview,
  BillableUsage,
  BillableUsageSummary,
  UsageTimeSeriesParams,
} from '../types'

const client = new BrokleAPIClient('/api')

// Default time range
const DEFAULT_TIME_RANGE: TimeRange = { relative: '30d' }

export const getUsageOverview = async (
  organizationId: string
): Promise<UsageOverview> => {
  return client.get<UsageOverview>(
    `/v1/organizations/${organizationId}/usage/overview`
  )
}

export const getUsageTimeSeries = async (
  organizationId: string,
  timeRange: TimeRange = DEFAULT_TIME_RANGE,
  granularity?: 'hourly' | 'daily'
): Promise<BillableUsage[]> => {
  const timeParams = timeRangeToApiParams(timeRange)
  const params: Record<string, string | undefined> = {
    ...timeParams,
    granularity,
  }
  return client.get<BillableUsage[]>(
    `/v1/organizations/${organizationId}/usage/timeseries`,
    params
  )
}

export const getUsageByProject = async (
  organizationId: string,
  timeRange: TimeRange = DEFAULT_TIME_RANGE
): Promise<BillableUsageSummary[]> => {
  const params = timeRangeToApiParams(timeRange)
  return client.get<BillableUsageSummary[]>(
    `/v1/organizations/${organizationId}/usage/by-project`,
    params
  )
}
