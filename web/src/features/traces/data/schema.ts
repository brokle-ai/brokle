import { z } from 'zod'

export const traceSchema = z.object({
  trace_id: z.string(), // Renamed from 'id'
  name: z.string(),
  startTime: z.date(),
  endTime: z.date().optional(),
  durationMs: z.number().optional(),
  status_code: z.number(), // Now UInt8 (0=UNSET, 1=OK, 2=ERROR)
  resource_attributes: z.string(), // JSON string with OTEL resource attributes
  cost: z.number().optional(), // total_cost
  tokens: z.number().optional(), // total_tokens
  spanCount: z.number().default(0),
  environment: z.string().optional(),
  serviceName: z.string().optional(),
  serviceVersion: z.string().optional(),
  tags: z.array(z.string()).optional(),
  bookmarked: z.boolean().default(false),
})

export const spanSchema = z.object({
  span_id: z.string(), // Renamed from 'id'
  trace_id: z.string(), // Snake case (matches backend)
  parent_span_id: z.string().optional(), // Snake case
  span_name: z.string(), // Renamed from 'name'
  span_kind: z.number(), // UInt8 (0-5)
  startTime: z.date(),
  endTime: z.date().optional(),
  durationMs: z.number().optional(),
  status_code: z.number(), // UInt8 (0-2)

  // JSON attribute fields
  span_attributes: z.string(), // All attributes (gen_ai.*, brokle.*, custom)
  resource_attributes: z.string(), // Resource-level attributes

  // OTEL Events/Links arrays
  events_timestamp: z.array(z.date()).optional(),
  events_name: z.array(z.string()).optional(),
  events_attributes: z.array(z.string()).optional(),
  links_trace_id: z.array(z.string()).optional(),
  links_span_id: z.array(z.string()).optional(),
  links_attributes: z.array(z.string()).optional(),

  // Materialized columns (read-only from backend)
  gen_ai_operation_name: z.string().optional(),
  gen_ai_provider_name: z.string().optional(),
  gen_ai_request_model: z.string().optional(),
  gen_ai_request_max_tokens: z.number().optional(),
  gen_ai_request_temperature: z.number().optional(),
  gen_ai_request_top_p: z.number().optional(),
  gen_ai_usage_input_tokens: z.number().optional(),
  gen_ai_usage_output_tokens: z.number().optional(),
  brokle_span_type: z.string().optional(), // Replaces 'type'
  brokle_span_level: z.string().optional(), // Replaces 'level'
  brokle_cost_input: z.number().optional(),
  brokle_cost_output: z.number().optional(),
  brokle_cost_total: z.number().optional(), // Replaces 'cost'
  brokle_prompt_id: z.string().optional(),
  brokle_prompt_name: z.string().optional(),
  brokle_prompt_version: z.number().optional(),
  brokle_internal_model_id: z.string().optional(),
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
