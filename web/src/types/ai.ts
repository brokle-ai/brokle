export interface AIProvider {
  id: string
  name: string
  displayName: string
  status: ProviderStatus
  health: HealthScore
  models: AIModel[]
  pricing: ProviderPricing
  latency: LatencyMetrics
  supportedFeatures: ProviderFeature[]
}

export interface AIModel {
  id: string
  name: string
  displayName: string
  providerId: string
  type: ModelType
  capabilities: ModelCapability[]
  pricing: ModelPricing
  performance: ModelPerformance
  context: ContextLimits
  modality: ModelModality[]
}

export interface AIRequest {
  id: string
  organizationId: string
  projectId: string
  environment: string
  modelId: string
  providerId: string
  input: AIRequestInput
  output?: AIRequestOutput
  metadata: RequestMetadata
  metrics: RequestMetrics
  cost: CostBreakdown
  status: RequestStatus
  createdAt: string
  completedAt?: string
}

export interface AIRequestInput {
  prompt?: string
  messages?: ChatMessage[]
  image?: string
  audio?: string
  parameters: Record<string, any>
}

export interface AIRequestOutput {
  content: string
  usage: TokenUsage
  finishReason: string
  quality?: QualityScore
}

export interface ChatMessage {
  role: 'system' | 'user' | 'assistant'
  content: string
  timestamp?: string
}

export interface RequestMetadata {
  userAgent?: string
  ipAddress?: string
  apiKeyId: string
  routingDecision: RoutingDecision
  cacheHit: boolean
  cached?: boolean
}

export interface RequestMetrics {
  latency: number
  tokensIn: number
  tokensOut: number
  totalTokens: number
  costUsd: number
  qualityScore?: number
}

export interface RoutingDecision {
  selectedProvider: string
  reason: RoutingReason
  alternatives: string[]
  score: number
  factors: RoutingFactor[]
}

export interface CostBreakdown {
  inputCost: number
  outputCost: number
  totalCost: number
  currency: string
  provider: string
  optimization: CostOptimization
}

export interface QualityScore {
  overall: number
  relevance: number
  coherence: number
  accuracy: number
  evaluatedAt: string
}

export type ProviderStatus = 'active' | 'degraded' | 'down' | 'maintenance'

export type ModelType = 'text' | 'chat' | 'embedding' | 'image' | 'audio' | 'multimodal'

export type ModelCapability = 
  | 'text-generation' 
  | 'chat' 
  | 'embedding' 
  | 'image-generation' 
  | 'image-analysis'
  | 'audio-generation' 
  | 'audio-transcription'
  | 'function-calling'
  | 'json-mode'

export type ModelModality = 'text' | 'image' | 'audio' | 'video'

export type RequestStatus = 'pending' | 'completed' | 'failed' | 'timeout' | 'rate_limited'

export type RoutingReason = 'latency' | 'cost' | 'quality' | 'availability' | 'preference'

export type RoutingFactor = 'provider_health' | 'model_performance' | 'cost_efficiency' | 'latency_optimization'

interface HealthScore {
  score: number
  uptime: number
  errorRate: number
  avgLatency: number
  lastChecked: string
}

interface ProviderPricing {
  inputTokenPrice: number
  outputTokenPrice: number
  currency: string
  billingModel: 'token' | 'request' | 'minute'
}

interface LatencyMetrics {
  p50: number
  p95: number
  p99: number
  avg: number
  region: string
}

interface ProviderFeature {
  name: string
  supported: boolean
  description?: string
}

interface ModelPricing {
  inputPrice: number
  outputPrice: number
  currency: string
  per1kTokens: boolean
}

interface ModelPerformance {
  latency: LatencyMetrics
  qualityScore: number
  reliability: number
  lastEvaluated: string
}

interface ContextLimits {
  maxTokens: number
  supports32k: boolean
  supports128k: boolean
}

interface TokenUsage {
  promptTokens: number
  completionTokens: number
  totalTokens: number
}

interface CostOptimization {
  potentialSavings: number
  alternativeProviders: string[]
  recommendedAction?: string
}