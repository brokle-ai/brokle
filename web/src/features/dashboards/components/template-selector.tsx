'use client'

import { useState } from 'react'
import { LayoutDashboard, Check, Loader2 } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { cn } from '@/lib/utils'
import { useTemplatesQuery } from '../hooks/use-templates'
import type { DashboardTemplate, TemplateCategory } from '../types'

/**
 * Category display configuration
 */
const categoryConfig: Record<TemplateCategory, { label: string; color: string }> = {
  'llm-overview': { label: 'LLM Overview', color: 'bg-blue-500/10 text-blue-500' },
  'cost-analytics': { label: 'Cost Analytics', color: 'bg-green-500/10 text-green-500' },
  'quality-scores': { label: 'Quality Scores', color: 'bg-purple-500/10 text-purple-500' },
}

interface TemplateCardProps {
  template: DashboardTemplate
  isSelected: boolean
  onSelect: () => void
}

function TemplateCard({ template, isSelected, onSelect }: TemplateCardProps) {
  const config = categoryConfig[template.category]
  const widgetCount = template.config?.widgets?.length ?? 0

  return (
    <Card
      className={cn(
        'cursor-pointer transition-all hover:border-primary/50',
        isSelected && 'border-primary ring-2 ring-primary/20'
      )}
      onClick={onSelect}
    >
      <CardHeader className="pb-3">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <CardTitle className="text-base font-medium flex items-center gap-2">
              {template.name}
              {isSelected && <Check className="h-4 w-4 text-primary" />}
            </CardTitle>
            <Badge variant="secondary" className={cn('mt-2 text-xs', config.color)}>
              {config.label}
            </Badge>
          </div>
        </div>
      </CardHeader>
      <CardContent className="pt-0">
        <CardDescription className="line-clamp-2 mb-3">
          {template.description}
        </CardDescription>
        <div className="flex items-center gap-1 text-xs text-muted-foreground">
          <LayoutDashboard className="h-3 w-3" />
          {widgetCount} widget{widgetCount !== 1 ? 's' : ''}
        </div>
      </CardContent>
    </Card>
  )
}

interface TemplateSelectorProps {
  onSelect: (template: DashboardTemplate, dashboardName: string) => void
  onCancel: () => void
  isLoading?: boolean
}

export function TemplateSelector({ onSelect, onCancel, isLoading }: TemplateSelectorProps) {
  const [selectedTemplate, setSelectedTemplate] = useState<DashboardTemplate | null>(null)
  const [dashboardName, setDashboardName] = useState('')

  const { data: templates, isLoading: templatesLoading, error } = useTemplatesQuery()

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (selectedTemplate && dashboardName.trim()) {
      onSelect(selectedTemplate, dashboardName.trim())
    }
  }

  const handleTemplateSelect = (template: DashboardTemplate) => {
    setSelectedTemplate(template)
    // Pre-fill with template name if dashboard name is empty
    if (!dashboardName) {
      setDashboardName(template.name)
    }
  }

  if (templatesLoading) {
    return (
      <div className="flex items-center justify-center py-8">
        <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <p>Failed to load templates. Please try again.</p>
      </div>
    )
  }

  if (!templates || templates.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        <LayoutDashboard className="h-10 w-10 mx-auto mb-3 opacity-50" />
        <p>No templates available.</p>
      </div>
    )
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-6">
      {/* Template Grid */}
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {templates.map((template) => (
          <TemplateCard
            key={template.id}
            template={template}
            isSelected={selectedTemplate?.id === template.id}
            onSelect={() => handleTemplateSelect(template)}
          />
        ))}
      </div>

      {/* Dashboard Name Input */}
      {selectedTemplate && (
        <div className="space-y-2 pt-2 border-t">
          <Label htmlFor="dashboard-name">Dashboard Name</Label>
          <Input
            id="dashboard-name"
            placeholder="Enter dashboard name"
            value={dashboardName}
            onChange={(e) => setDashboardName(e.target.value)}
            autoFocus
          />
          <p className="text-xs text-muted-foreground">
            Based on template: {selectedTemplate.name}
          </p>
        </div>
      )}

      {/* Actions */}
      <div className="flex justify-end gap-3 pt-2">
        <Button type="button" variant="outline" onClick={onCancel}>
          Cancel
        </Button>
        <Button
          type="submit"
          disabled={!selectedTemplate || !dashboardName.trim() || isLoading}
        >
          {isLoading && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
          Create from Template
        </Button>
      </div>
    </form>
  )
}
