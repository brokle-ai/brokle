import { BrokleAPIClient } from '@/lib/api/core/client'
import type { Task } from '../data/schema'

const client = new BrokleAPIClient('/api')

export interface GetTasksParams {
  projectSlug: string
  page?: number
  pageSize?: number
  status?: string[]
  priority?: string[]
  search?: string
}

export interface CreateTaskData {
  title: string
  status: string
  label: string
  priority: string
  description?: string
  assignee?: string
  dueDate?: Date
}

export interface UpdateTaskData extends Partial<CreateTaskData> {}

/**
 * Get all tasks for a project
 */
export const getProjectTasks = async (params: GetTasksParams): Promise<Task[]> => {
  const { projectSlug, ...queryParams } = params
  return client.get(`/v1/projects/${projectSlug}/tasks`, queryParams)
}

/**
 * Create a new task
 */
export const createTask = async (
  projectSlug: string,
  data: CreateTaskData
): Promise<Task> => {
  return client.post(`/v1/projects/${projectSlug}/tasks`, data)
}

/**
 * Update an existing task
 */
export const updateTask = async (
  projectSlug: string,
  taskId: string,
  data: UpdateTaskData
): Promise<Task> => {
  return client.patch(`/v1/projects/${projectSlug}/tasks/${taskId}`, data)
}

/**
 * Delete a task
 */
export const deleteTask = async (
  projectSlug: string,
  taskId: string
): Promise<void> => {
  return client.delete(`/v1/projects/${projectSlug}/tasks/${taskId}`)
}

/**
 * Delete multiple tasks
 */
export const deleteMultipleTasks = async (
  projectSlug: string,
  taskIds: string[]
): Promise<void> => {
  return client.post(`/v1/projects/${projectSlug}/tasks/bulk-delete`, { taskIds })
}

/**
 * Update status for multiple tasks
 */
export const updateMultipleTasksStatus = async (
  projectSlug: string,
  taskIds: string[],
  status: string
): Promise<void> => {
  return client.post(`/v1/projects/${projectSlug}/tasks/bulk-update-status`, {
    taskIds,
    status,
  })
}

/**
 * Update priority for multiple tasks
 */
export const updateMultipleTasksPriority = async (
  projectSlug: string,
  taskIds: string[],
  priority: string
): Promise<void> => {
  return client.post(`/v1/projects/${projectSlug}/tasks/bulk-update-priority`, {
    taskIds,
    priority,
  })
}

/**
 * Import tasks from CSV file
 */
export const importTasks = async (
  projectSlug: string,
  file: File
): Promise<{ imported: number; failed: number }> => {
  const formData = new FormData()
  formData.append('file', file)
  return client.post(`/v1/projects/${projectSlug}/tasks/import`, formData)
}

/**
 * Export tasks to CSV
 */
export const exportTasks = async (
  projectSlug: string,
  taskIds?: string[]
): Promise<Blob> => {
  const params = taskIds ? { taskIds } : {}
  return client.get(`/v1/projects/${projectSlug}/tasks/export`, params)
}
