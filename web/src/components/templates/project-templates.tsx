'use client'

import { useState } from 'react'
import { 
  FileText, 
  MessageSquare, 
  Image, 
  Mic, 
  Code, 
  Brain, 
  Search, 
  Star, 
  Check, 
  Zap,
  ArrowRight
} from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'
import type { ProjectEnvironment } from '@/types/organization'

interface ProjectTemplate {
  id: string
  name: string
  description: string
  category: string
  icon: React.ComponentType<{ className?: string }>
  popular: boolean
  features: string[]
  models: string[]
  setup_time: string
  complexity: 'beginner' | 'intermediate' | 'advanced'
  config: {
    default_models: string[]
    routing_strategy: string
    cache_enabled: boolean
    retry_attempts: number
    timeout_ms: number
  }
}

const PROJECT_TEMPLATES: ProjectTemplate[] = [
  {
    id: 'chatbot',
    name: 'AI Chatbot',
    description: 'Build intelligent conversational AI with multi-turn dialogue support',
    category: 'Conversational AI',
    icon: MessageSquare,
    popular: true,
    features: ['Multi-turn conversations', 'Context awareness', 'Function calling', 'Memory management'],
    models: ['GPT-4 Turbo', 'Claude 3 Opus', 'GPT-3.5 Turbo'],
    setup_time: '5 minutes',
    complexity: 'beginner',
    config: {
      default_models: ['gpt-4-turbo', 'claude-3-opus'],
      routing_strategy: 'quality-optimized',
      cache_enabled: true,
      retry_attempts: 3,
      timeout_ms: 30000
    }
  },
  {
    id: 'content-generator',
    name: 'Content Generator',
    description: 'Generate high-quality content for blogs, marketing, and documentation',
    category: 'Content Creation',
    icon: FileText,
    popular: true,
    features: ['SEO optimization', 'Multiple formats', 'Brand voice consistency', 'Batch generation'],
    models: ['GPT-4 Turbo', 'Claude 3 Sonnet', 'Gemini Pro'],
    setup_time: '3 minutes',
    complexity: 'beginner',
    config: {
      default_models: ['gpt-4-turbo', 'claude-3-sonnet'],
      routing_strategy: 'balanced',
      cache_enabled: true,
      retry_attempts: 2,
      timeout_ms: 45000
    }
  },
  {
    id: 'code-assistant',
    name: 'Code Assistant',
    description: 'AI-powered code generation, review, and debugging assistant',
    category: 'Development',
    icon: Code,
    popular: true,
    features: ['Code generation', 'Bug detection', 'Code review', 'Multiple languages'],
    models: ['GPT-4 Turbo', 'Claude 3 Opus', 'CodeLlama'],
    setup_time: '7 minutes',
    complexity: 'intermediate',
    config: {
      default_models: ['gpt-4-turbo', 'claude-3-opus'],
      routing_strategy: 'quality-optimized',
      cache_enabled: false,
      retry_attempts: 2,
      timeout_ms: 60000
    }
  },
  {
    id: 'image-analysis',
    name: 'Vision AI',
    description: 'Advanced image and document analysis with OCR capabilities',
    category: 'Computer Vision',
    icon: Image,
    popular: false,
    features: ['Image recognition', 'OCR', 'Document analysis', 'Object detection'],
    models: ['GPT-4 Vision', 'Claude 3 Opus', 'Gemini Pro Vision'],
    setup_time: '10 minutes',
    complexity: 'advanced',
    config: {
      default_models: ['gpt-4-vision', 'claude-3-opus'],
      routing_strategy: 'quality-optimized',
      cache_enabled: true,
      retry_attempts: 2,
      timeout_ms: 90000
    }
  },
  {
    id: 'voice-assistant',
    name: 'Voice Assistant',
    description: 'Speech-to-text and text-to-speech AI assistant with voice commands',
    category: 'Audio AI',
    icon: Mic,
    popular: false,
    features: ['Speech recognition', 'Voice synthesis', 'Command processing', 'Multi-language'],
    models: ['Whisper', 'GPT-4 Turbo', 'Claude 3 Sonnet'],
    setup_time: '15 minutes',
    complexity: 'advanced',
    config: {
      default_models: ['whisper', 'gpt-4-turbo'],
      routing_strategy: 'latency-optimized',
      cache_enabled: false,
      retry_attempts: 3,
      timeout_ms: 45000
    }
  },
  {
    id: 'data-analyst',
    name: 'Data Analyst',
    description: 'AI-powered data analysis, visualization, and insights generation',
    category: 'Analytics',
    icon: Brain,
    popular: false,
    features: ['Data processing', 'Statistical analysis', 'Chart generation', 'Insights extraction'],
    models: ['GPT-4 Turbo', 'Claude 3 Opus', 'Code Interpreter'],
    setup_time: '12 minutes',
    complexity: 'advanced',
    config: {
      default_models: ['gpt-4-turbo', 'claude-3-opus'],
      routing_strategy: 'quality-optimized',
      cache_enabled: true,
      retry_attempts: 2,
      timeout_ms: 120000
    }
  },
  {
    id: 'search-engine',
    name: 'Semantic Search',
    description: 'Build intelligent search with semantic understanding and RAG',
    category: 'Search & Retrieval',
    icon: Search,
    popular: false,
    features: ['Vector search', 'RAG pipeline', 'Document indexing', 'Semantic ranking'],
    models: ['GPT-4 Turbo', 'text-embedding-3-large', 'Claude 3 Sonnet'],
    setup_time: '20 minutes',
    complexity: 'advanced',
    config: {
      default_models: ['gpt-4-turbo', 'text-embedding-3-large'],
      routing_strategy: 'balanced',
      cache_enabled: true,
      retry_attempts: 2,
      timeout_ms: 60000
    }
  }
]

interface ProjectTemplatesProps {
  onTemplateSelect?: (template: ProjectTemplate, projectData: any) => void
}

export function ProjectTemplates({ onTemplateSelect }: ProjectTemplatesProps) {
  const [searchTerm, setSearchTerm] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [selectedTemplate, setSelectedTemplate] = useState<ProjectTemplate | null>(null)
  const [isCreateOpen, setIsCreateOpen] = useState(false)
  const [projectName, setProjectName] = useState('')
  const [projectDescription, setProjectDescription] = useState('')
  const [environment, setEnvironment] = useState<ProjectEnvironment>('development')
  const [isCreating, setIsCreating] = useState(false)

  const categories = ['all', ...Array.from(new Set(PROJECT_TEMPLATES.map(t => t.category)))]
  
  const filteredTemplates = PROJECT_TEMPLATES.filter(template => {
    const matchesSearch = template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         template.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
                         template.features.some(f => f.toLowerCase().includes(searchTerm.toLowerCase()))
    const matchesCategory = selectedCategory === 'all' || template.category === selectedCategory
    return matchesSearch && matchesCategory
  })

  const popularTemplates = filteredTemplates.filter(t => t.popular)
  const otherTemplates = filteredTemplates.filter(t => !t.popular)

  const handleTemplateSelect = (template: ProjectTemplate) => {
    setSelectedTemplate(template)
    setProjectName(`My ${template.name}`)
    setProjectDescription(template.description)
    setIsCreateOpen(true)
  }

  const handleCreateProject = async () => {
    if (!selectedTemplate || !projectName.trim()) {
      toast.error('Please enter a project name')
      return
    }

    setIsCreating(true)
    
    try {
      const projectData = {
        name: projectName,
        description: projectDescription,
        environment,
        template: selectedTemplate,
        config: selectedTemplate.config
      }

      // TODO: Implement actual project creation API call
      await new Promise(resolve => setTimeout(resolve, 2000))
      
      toast.success(`Project "${projectName}" created successfully using ${selectedTemplate.name} template`)
      
      onTemplateSelect?.(selectedTemplate, projectData)
      
      // Reset form
      setProjectName('')
      setProjectDescription('')
      setEnvironment('development')
      setSelectedTemplate(null)
      setIsCreateOpen(false)
      
    } catch (error) {
      console.error('Failed to create project:', error)
      toast.error('Failed to create project. Please try again.')
    } finally {
      setIsCreating(false)
    }
  }

  const getComplexityColor = (complexity: ProjectTemplate['complexity']) => {
    switch (complexity) {
      case 'beginner':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-300'
      case 'intermediate':
        return 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-300'
      case 'advanced':
        return 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-300'
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-300'
    }
  }

  const TemplateCard = ({ template }: { template: ProjectTemplate }) => (
    <Card 
      className="cursor-pointer hover:shadow-lg transition-all duration-200 group h-full"
      onClick={() => handleTemplateSelect(template)}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-3">
            <div className="bg-primary/10 flex size-10 items-center justify-center rounded-lg group-hover:bg-primary/20 transition-colors">
              <template.icon className="size-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-lg flex items-center gap-2">
                {template.name}
                {template.popular && <Star className="h-4 w-4 text-yellow-500 fill-current" />}
              </CardTitle>
              <div className="flex items-center gap-2 mt-1">
                <Badge variant="secondary" className="text-xs">
                  {template.category}
                </Badge>
                <Badge className={cn("text-xs", getComplexityColor(template.complexity))}>
                  {template.complexity}
                </Badge>
              </div>
            </div>
          </div>
          <div className="text-right text-sm text-muted-foreground">
            <div className="flex items-center gap-1">
              <Zap className="h-3 w-3" />
              {template.setup_time}
            </div>
          </div>
        </div>
      </CardHeader>
      
      <CardContent className="space-y-4">
        <CardDescription className="line-clamp-2">
          {template.description}
        </CardDescription>

        <div>
          <div className="text-sm font-medium mb-2">Key Features:</div>
          <div className="flex flex-wrap gap-1">
            {template.features.slice(0, 3).map((feature) => (
              <Badge key={feature} variant="outline" className="text-xs">
                {feature}
              </Badge>
            ))}
            {template.features.length > 3 && (
              <Badge variant="outline" className="text-xs">
                +{template.features.length - 3} more
              </Badge>
            )}
          </div>
        </div>

        <div>
          <div className="text-sm font-medium mb-2">Supported Models:</div>
          <div className="text-xs text-muted-foreground">
            {template.models.slice(0, 2).join(', ')}
            {template.models.length > 2 && `, +${template.models.length - 2} more`}
          </div>
        </div>

        <div className="flex items-center justify-between pt-2 border-t">
          <div className="text-xs text-muted-foreground">
            Ready to use
          </div>
          <Button size="sm" className="group-hover:bg-primary group-hover:text-primary-foreground">
            Use Template
            <ArrowRight className="ml-1 h-3 w-3" />
          </Button>
        </div>
      </CardContent>
    </Card>
  )

  return (
    <>
      <div className="space-y-6">
        {/* Search and Filters */}
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground h-4 w-4" />
            <Input
              placeholder="Search templates..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          
          <Select value={selectedCategory} onValueChange={setSelectedCategory}>
            <SelectTrigger className="w-full sm:w-48">
              <SelectValue placeholder="All Categories" />
            </SelectTrigger>
            <SelectContent>
              {categories.map((category) => (
                <SelectItem key={category} value={category}>
                  {category === 'all' ? 'All Categories' : category}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Popular Templates */}
        {popularTemplates.length > 0 && (
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Star className="h-5 w-5 text-yellow-500 fill-current" />
              <h3 className="text-lg font-semibold">Popular Templates</h3>
            </div>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {popularTemplates.map((template) => (
                <TemplateCard key={template.id} template={template} />
              ))}
            </div>
          </div>
        )}

        {/* Other Templates */}
        {otherTemplates.length > 0 && (
          <div className="space-y-4">
            {popularTemplates.length > 0 && <Separator />}
            <h3 className="text-lg font-semibold">All Templates</h3>
            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              {otherTemplates.map((template) => (
                <TemplateCard key={template.id} template={template} />
              ))}
            </div>
          </div>
        )}

        {/* No Results */}
        {filteredTemplates.length === 0 && (
          <div className="text-center py-12">
            <Search className="mx-auto h-12 w-12 text-muted-foreground mb-4" />
            <h3 className="text-lg font-medium mb-2">No templates found</h3>
            <p className="text-muted-foreground mb-4">
              Try adjusting your search or category filter
            </p>
            <Button variant="outline" onClick={() => { setSearchTerm(''); setSelectedCategory('all') }}>
              Clear Filters
            </Button>
          </div>
        )}
      </div>

      {/* Create Project Dialog */}
      <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>
              Create Project from Template
            </DialogTitle>
            <DialogDescription>
              {selectedTemplate && (
                <>
                  Set up your new project using the <strong>{selectedTemplate.name}</strong> template.
                  This will configure everything you need to get started quickly.
                </>
              )}
            </DialogDescription>
          </DialogHeader>

          {selectedTemplate && (
            <div className="space-y-6">
              {/* Template Info */}
              <div className="p-4 bg-muted rounded-lg">
                <div className="flex items-center gap-3">
                  <selectedTemplate.icon className="h-8 w-8 text-primary" />
                  <div>
                    <h4 className="font-medium">{selectedTemplate.name}</h4>
                    <p className="text-sm text-muted-foreground">{selectedTemplate.description}</p>
                  </div>
                </div>
                
                <div className="mt-3 space-y-2">
                  <div className="text-sm">
                    <strong>What's included:</strong>
                  </div>
                  <div className="grid grid-cols-2 gap-2 text-xs">
                    {selectedTemplate.features.map((feature) => (
                      <div key={feature} className="flex items-center gap-1">
                        <Check className="h-3 w-3 text-green-500" />
                        {feature}
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Project Details */}
              <div className="space-y-4">
                <div className="space-y-2">
                  <Label htmlFor="projectName">Project Name *</Label>
                  <Input
                    id="projectName"
                    value={projectName}
                    onChange={(e) => setProjectName(e.target.value)}
                    placeholder="Enter project name"
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="projectDescription">Description</Label>
                  <Textarea
                    id="projectDescription"
                    value={projectDescription}
                    onChange={(e) => setProjectDescription(e.target.value)}
                    placeholder="Describe your project..."
                    rows={3}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="environment">Environment</Label>
                  <Select value={environment} onValueChange={(value: ProjectEnvironment) => setEnvironment(value)}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="development">Development</SelectItem>
                      <SelectItem value="staging">Staging</SelectItem>
                      <SelectItem value="production">Production</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Configuration Preview */}
              <div className="p-4 border rounded-lg">
                <h4 className="font-medium text-sm mb-2">Configuration Preview</h4>
                <div className="space-y-2 text-xs text-muted-foreground">
                  <div>Default Models: {selectedTemplate.config.default_models.join(', ')}</div>
                  <div>Routing Strategy: {selectedTemplate.config.routing_strategy}</div>
                  <div>Cache Enabled: {selectedTemplate.config.cache_enabled ? 'Yes' : 'No'}</div>
                  <div>Timeout: {selectedTemplate.config.timeout_ms}ms</div>
                </div>
              </div>
            </div>
          )}

          <DialogFooter>
            <Button variant="outline" onClick={() => setIsCreateOpen(false)} disabled={isCreating}>
              Cancel
            </Button>
            <Button onClick={handleCreateProject} disabled={isCreating || !projectName.trim()}>
              {isCreating ? (
                <>
                  <Zap className="mr-2 h-4 w-4 animate-pulse" />
                  Creating Project...
                </>
              ) : (
                <>
                  <Zap className="mr-2 h-4 w-4" />
                  Create Project
                </>
              )}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}