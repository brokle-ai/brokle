/**
 * Budget API
 *
 * API functions for usage budget management endpoints
 */

import { BrokleAPIClient } from '@/lib/api/core/client'
import type {
  UsageBudget,
  CreateBudgetRequest,
  UpdateBudgetRequest,
  UsageAlert,
} from '../types'

const client = new BrokleAPIClient('/api')

export const listBudgets = async (
  organizationId: string
): Promise<UsageBudget[]> => {
  return client.get<UsageBudget[]>(
    `/v1/organizations/${organizationId}/budgets`
  )
}

export const getBudget = async (
  organizationId: string,
  budgetId: string
): Promise<UsageBudget> => {
  return client.get<UsageBudget>(
    `/v1/organizations/${organizationId}/budgets/${budgetId}`
  )
}

export const createBudget = async (
  organizationId: string,
  data: CreateBudgetRequest
): Promise<UsageBudget> => {
  return client.post<UsageBudget>(
    `/v1/organizations/${organizationId}/budgets`,
    data
  )
}

export const updateBudget = async (
  organizationId: string,
  budgetId: string,
  data: UpdateBudgetRequest
): Promise<UsageBudget> => {
  return client.put<UsageBudget>(
    `/v1/organizations/${organizationId}/budgets/${budgetId}`,
    data
  )
}

export const deleteBudget = async (
  organizationId: string,
  budgetId: string
): Promise<void> => {
  await client.delete(`/v1/organizations/${organizationId}/budgets/${budgetId}`)
}

export const getAlerts = async (
  organizationId: string,
  limit?: number
): Promise<UsageAlert[]> => {
  const params = limit ? { limit: limit.toString() } : undefined
  return client.get<UsageAlert[]>(
    `/v1/organizations/${organizationId}/budgets/alerts`,
    params
  )
}

export const acknowledgeAlert = async (
  organizationId: string,
  alertId: string
): Promise<void> => {
  await client.post(
    `/v1/organizations/${organizationId}/budgets/alerts/${alertId}/acknowledge`
  )
}
