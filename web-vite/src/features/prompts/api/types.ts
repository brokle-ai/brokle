// Wire types for the prompts list endpoint. Narrow shape — only the
// fields the list view renders. Detail/version/editor ports will add
// richer types (template body, variables, dialect, ModelConfig) in a
// follow-up commit.
//
// Shape matches the JSON emitted by `GET /api/v1/projects/{projectId}/prompts`
// (see internal/transport/http/handlers/prompt/types.go `listPromptsResponse`).
// Timestamps arrive as RFC 3339 strings; components format on render.

export type PromptType = 'text' | 'chat'

export interface PromptLabelInfo {
  name: string
  version: number
}

export interface PromptListItem {
  id: string
  name: string
  type: PromptType
  description?: string
  tags: string[]
  latest_version: number
  labels: PromptLabelInfo[]
  created_at: string
  updated_at: string
}

// The dashboard list route returns a flat `{data, total, page, limit}`
// envelope rather than the `{data, pagination}` shape used by the
// observability plane. Kept 1:1 with the backend so conversions stay at
// the render layer.
export interface PromptListResponse {
  data: PromptListItem[]
  total: number
  page: number
  limit: number
}
