import { z } from 'zod'

export const traceSchema = z.object({
  id: z.string(),
  name: z.string(),
  startTime: z.date(),
  endTime: z.date().optional(),
  durationMs: z.number().optional(),
  status: z.enum(['ok', 'error', 'unset']),
  cost: z.number().optional(),
  tokens: z.number().optional(),
  spanCount: z.number().default(0),
  environment: z.string().optional(),
  serviceName: z.string().optional(),
  serviceVersion: z.string().optional(),
  tags: z.array(z.string()).optional(),
  bookmarked: z.boolean().default(false),
})

export const spanSchema = z.object({
  id: z.string(),
  traceId: z.string(),
  parentSpanId: z.string().optional(),
  name: z.string(),
  type: z.string(),
  startTime: z.date(),
  endTime: z.date().optional(),
  durationMs: z.number().optional(),
  status: z.enum(['ok', 'error', 'unset']),
  level: z.number().default(0), // Hierarchy depth
  model: z.string().optional(),
  cost: z.number().optional(),
  tokens: z.number().optional(),
})

export const scoreSchema = z.object({
  id: z.string(),
  name: z.string(),
  value: z.number(),
  comment: z.string().optional(),
  source: z.string().optional(),
})

export type Trace = z.infer<typeof traceSchema>
export type Span = z.infer<typeof spanSchema>
export type Score = z.infer<typeof scoreSchema>
