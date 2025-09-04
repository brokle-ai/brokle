'use client'

import { useState } from 'react'
import { Brain, Settings, Save, RotateCcw, Zap, DollarSign, Clock, BarChart } from 'lucide-react'
import { useOrganization } from '@/context/organization-context'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Slider } from '@/components/ui/slider'
import { Badge } from '@/components/ui/badge'
import { TabsContent } from '@/components/ui/tabs'
import { Separator } from '@/components/ui/separator'
import { Progress } from '@/components/ui/progress'
import { toast } from 'sonner'

interface ModelProvider {
  id: string
  name: string
  enabled: boolean
  priority: number
  models: Model[]
}

interface Model {
  id: string
  name: string
  provider: string
  enabled: boolean
  cost_per_1k_tokens: number
  avg_latency_ms: number
  success_rate: number
  capabilities: string[]
}

const MOCK_PROVIDERS: ModelProvider[] = [
  {
    id: 'openai',
    name: 'OpenAI',
    enabled: true,
    priority: 1,
    models: [
      {
        id: 'gpt-4-turbo',
        name: 'GPT-4 Turbo',
        provider: 'openai',
        enabled: true,
        cost_per_1k_tokens: 0.01,
        avg_latency_ms: 1200,
        success_rate: 99.8,
        capabilities: ['chat', 'completion', 'reasoning']
      },
      {
        id: 'gpt-3.5-turbo',
        name: 'GPT-3.5 Turbo',
        provider: 'openai',
        enabled: true,
        cost_per_1k_tokens: 0.002,
        avg_latency_ms: 800,
        success_rate: 99.9,
        capabilities: ['chat', 'completion']
      }
    ]
  },
  {
    id: 'anthropic',
    name: 'Anthropic',
    enabled: true,
    priority: 2,
    models: [
      {
        id: 'claude-3-opus',
        name: 'Claude 3 Opus',
        provider: 'anthropic',
        enabled: true,
        cost_per_1k_tokens: 0.015,
        avg_latency_ms: 1400,
        success_rate: 99.7,
        capabilities: ['chat', 'completion', 'analysis', 'reasoning']
      },
      {
        id: 'claude-3-sonnet',
        name: 'Claude 3 Sonnet',
        provider: 'anthropic',
        enabled: false,
        cost_per_1k_tokens: 0.003,
        avg_latency_ms: 900,
        success_rate: 99.8,
        capabilities: ['chat', 'completion', 'analysis']
      }
    ]
  },
  {
    id: 'google',
    name: 'Google AI',
    enabled: false,
    priority: 3,
    models: [
      {
        id: 'gemini-pro',
        name: 'Gemini Pro',
        provider: 'google',
        enabled: false,
        cost_per_1k_tokens: 0.0025,
        avg_latency_ms: 1000,
        success_rate: 99.5,
        capabilities: ['chat', 'completion', 'multimodal']
      }
    ]
  }
]

export default function ProjectModelsPage() {
  const { currentProject } = useOrganization()
  
  const [providers, setProviders] = useState<ModelProvider[]>(MOCK_PROVIDERS)
  const [routingStrategy, setRoutingStrategy] = useState<'cost-optimized' | 'latency-optimized' | 'quality-optimized' | 'balanced'>('balanced')
  const [fallbackEnabled, setFallbackEnabled] = useState(true)
  const [retryOnFailure, setRetryOnFailure] = useState(true)
  const [loadBalancing, setLoadBalancing] = useState(true)
  const [costBudgetEnabled, setCostBudgetEnabled] = useState(false)
  const [costBudget, setCostBudget] = useState([100])
  const [isLoading, setIsLoading] = useState(false)

  if (!currentProject) {
    return null
  }

  const toggleProvider = (providerId: string) => {
    setProviders(providers.map(p => 
      p.id === providerId ? { ...p, enabled: !p.enabled } : p
    ))
  }

  const toggleModel = (providerId: string, modelId: string) => {
    setProviders(providers.map(p => 
      p.id === providerId 
        ? { 
            ...p, 
            models: p.models.map(m => 
              m.id === modelId ? { ...m, enabled: !m.enabled } : m
            )
          }
        : p
    ))
  }

  const updateProviderPriority = (providerId: string, priority: number) => {
    setProviders(providers.map(p => 
      p.id === providerId ? { ...p, priority } : p
    ))
  }

  const saveConfiguration = async () => {
    setIsLoading(true)
    
    try {
      // TODO: Implement API call to save model configuration
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      toast.success('Model configuration saved successfully')
    } catch (error) {
      console.error('Failed to save configuration:', error)
      toast.error('Failed to save configuration. Please try again.')
    } finally {
      setIsLoading(false)
    }
  }

  const resetToDefaults = () => {
    setProviders(MOCK_PROVIDERS)
    setRoutingStrategy('balanced')
    setFallbackEnabled(true)
    setRetryOnFailure(true)
    setLoadBalancing(true)
    setCostBudgetEnabled(false)
    setCostBudget([100])
    
    toast.success('Configuration reset to defaults')
  }

  const getProviderStatusColor = (provider: ModelProvider) => {
    const enabledModels = provider.models.filter(m => m.enabled).length
    if (!provider.enabled || enabledModels === 0) return 'text-red-500'
    if (enabledModels === provider.models.length) return 'text-green-500'
    return 'text-yellow-500'
  }

  const totalEnabledModels = providers.reduce((acc, p) => acc + p.models.filter(m => m.enabled).length, 0)

  return (
    <TabsContent value="models" className="space-y-6">
      {/* Model Configuration Overview */}
      <Card>
        <CardHeader>
          <div className="flex items-center justify-between">
            <div>
              <CardTitle className="flex items-center gap-2">
                <Brain className="h-5 w-5" />
                Model Configuration
              </CardTitle>
              <CardDescription>
                Configure AI models and routing strategies for optimal performance
              </CardDescription>
            </div>
            <div className="flex items-center gap-2">
              <Button variant="outline" onClick={resetToDefaults}>
                <RotateCcw className="mr-2 h-4 w-4" />
                Reset
              </Button>
              <Button onClick={saveConfiguration} disabled={isLoading}>
                {isLoading ? (
                  <>
                    <Settings className="mr-2 h-4 w-4 animate-spin" />
                    Saving...
                  </>
                ) : (
                  <>
                    <Save className="mr-2 h-4 w-4" />
                    Save Config
                  </>
                )}
              </Button>
            </div>
          </div>
        </CardHeader>
      </Card>

      {/* Quick Stats */}
      <div className="grid gap-4 md:grid-cols-4">
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Active Providers</p>
                <p className="text-2xl font-bold">{providers.filter(p => p.enabled).length}</p>
              </div>
              <Zap className="h-8 w-8 text-blue-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Active Models</p>
                <p className="text-2xl font-bold">{totalEnabledModels}</p>
              </div>
              <Brain className="h-8 w-8 text-green-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Avg Cost/1K</p>
                <p className="text-2xl font-bold">
                  ${providers
                    .flatMap(p => p.models.filter(m => m.enabled))
                    .reduce((acc, m, _, arr) => acc + m.cost_per_1k_tokens / arr.length, 0)
                    .toFixed(3)}
                </p>
              </div>
              <DollarSign className="h-8 w-8 text-yellow-500" />
            </div>
          </CardContent>
        </Card>
        
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-muted-foreground">Avg Latency</p>
                <p className="text-2xl font-bold">
                  {Math.round(providers
                    .flatMap(p => p.models.filter(m => m.enabled))
                    .reduce((acc, m, _, arr) => acc + m.avg_latency_ms / arr.length, 0))}ms
                </p>
              </div>
              <Clock className="h-8 w-8 text-purple-500" />
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Routing Strategy */}
      <Card>
        <CardHeader>
          <CardTitle>Routing Strategy</CardTitle>
          <CardDescription>
            Choose how requests should be routed across your enabled models
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <div className="space-y-2">
            <Label>Strategy</Label>
            <Select value={routingStrategy} onValueChange={(value: string) => setRoutingStrategy(value)}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="cost-optimized">Cost Optimized</SelectItem>
                <SelectItem value="latency-optimized">Latency Optimized</SelectItem>
                <SelectItem value="quality-optimized">Quality Optimized</SelectItem>
                <SelectItem value="balanced">Balanced</SelectItem>
              </SelectContent>
            </Select>
            <p className="text-xs text-muted-foreground">
              {routingStrategy === 'cost-optimized' && 'Route to cheapest available models first'}
              {routingStrategy === 'latency-optimized' && 'Route to fastest responding models first'}
              {routingStrategy === 'quality-optimized' && 'Route to highest quality models first'}
              {routingStrategy === 'balanced' && 'Balance cost, latency, and quality factors'}
            </p>
          </div>

          <div className="grid grid-cols-3 gap-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Fallback on Error</Label>
                <p className="text-xs text-muted-foreground">
                  Try next model if primary fails
                </p>
              </div>
              <Switch
                checked={fallbackEnabled}
                onCheckedChange={setFallbackEnabled}
              />
            </div>
            
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Retry on Failure</Label>
                <p className="text-xs text-muted-foreground">
                  Retry failed requests automatically
                </p>
              </div>
              <Switch
                checked={retryOnFailure}
                onCheckedChange={setRetryOnFailure}
              />
            </div>
            
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Load Balancing</Label>
                <p className="text-xs text-muted-foreground">
                  Distribute load across models
                </p>
              </div>
              <Switch
                checked={loadBalancing}
                onCheckedChange={setLoadBalancing}
              />
            </div>
          </div>

          {/* Cost Budget */}
          <div className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="space-y-0.5">
                <Label>Cost Budget Limit</Label>
                <p className="text-xs text-muted-foreground">
                  Set daily spending limit for this project
                </p>
              </div>
              <Switch
                checked={costBudgetEnabled}
                onCheckedChange={setCostBudgetEnabled}
              />
            </div>
            
            {costBudgetEnabled && (
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Daily Budget: ${costBudget[0]}</span>
                  <span className="text-muted-foreground">Current: $24.50</span>
                </div>
                <Slider
                  value={costBudget}
                  onValueChange={setCostBudget}
                  max={1000}
                  min={10}
                  step={10}
                />
                <Progress value={24.5} className="h-2" />
              </div>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Provider Configuration */}
      <Card>
        <CardHeader>
          <CardTitle>Provider Configuration</CardTitle>
          <CardDescription>
            Enable/disable providers and configure model priorities
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          {providers.map((provider) => (
            <div key={provider.id} className="space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <Switch
                    checked={provider.enabled}
                    onCheckedChange={() => toggleProvider(provider.id)}
                  />
                  <div>
                    <h3 className="font-medium flex items-center gap-2">
                      {provider.name}
                      <div className={`w-2 h-2 rounded-full ${getProviderStatusColor(provider)}`} />
                    </h3>
                    <p className="text-sm text-muted-foreground">
                      {provider.models.filter(m => m.enabled).length} of {provider.models.length} models active
                    </p>
                  </div>
                </div>
                
                <div className="flex items-center gap-2">
                  <Label className="text-xs">Priority</Label>
                  <Select 
                    value={provider.priority.toString()} 
                    onValueChange={(value) => updateProviderPriority(provider.id, parseInt(value))}
                  >
                    <SelectTrigger className="w-20">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">1</SelectItem>
                      <SelectItem value="2">2</SelectItem>
                      <SelectItem value="3">3</SelectItem>
                      <SelectItem value="4">4</SelectItem>
                      <SelectItem value="5">5</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {provider.enabled && (
                <div className="ml-6 space-y-3">
                  {provider.models.map((model) => (
                    <div key={model.id} className="flex items-center justify-between p-3 border rounded-lg">
                      <div className="flex items-center gap-3">
                        <Switch
                          checked={model.enabled}
                          onCheckedChange={() => toggleModel(provider.id, model.id)}
                        />
                        <div>
                          <h4 className="font-medium text-sm">{model.name}</h4>
                          <div className="flex items-center gap-4 mt-1">
                            <div className="flex items-center gap-1 text-xs text-muted-foreground">
                              <DollarSign className="h-3 w-3" />
                              ${model.cost_per_1k_tokens}/1K
                            </div>
                            <div className="flex items-center gap-1 text-xs text-muted-foreground">
                              <Clock className="h-3 w-3" />
                              {model.avg_latency_ms}ms
                            </div>
                            <div className="flex items-center gap-1 text-xs text-muted-foreground">
                              <BarChart className="h-3 w-3" />
                              {model.success_rate}%
                            </div>
                          </div>
                        </div>
                      </div>
                      
                      <div className="flex flex-wrap gap-1">
                        {model.capabilities.map((capability) => (
                          <Badge key={capability} variant="secondary" className="text-xs">
                            {capability}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              )}
              
              {provider.id !== providers[providers.length - 1].id && <Separator />}
            </div>
          ))}
        </CardContent>
      </Card>
    </TabsContent>
  )
}