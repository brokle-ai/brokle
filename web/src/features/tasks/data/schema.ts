import { z } from 'zod'

// We're keeping a simple non-relational schema here.
// IRL, you will have a schema for your data models.
export const taskSchema = z.object({
  id: z.string(),
  title: z.string(),
  status: z.string(),
  label: z.string(),
  priority: z.string(),
  createdAt: z.date().optional(),
  updatedAt: z.date().optional(),
  assignee: z.string().optional(),
  description: z.string().optional(),
  dueDate: z.date().optional(),
})

export type Task = z.infer<typeof taskSchema>