'use client'

import * as React from 'react'
import { Code, Link2, Terminal, Copy, Check } from 'lucide-react'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { Prompt, PromptVersion } from '../../types'
import type { PromptDetailTab } from '../../hooks/use-prompt-detail-state'
import { PromptEditor } from '../prompt-editor'
import { LabelSelector } from '../label-management/LabelSelector'
import { JsonConfigViewer } from './json-config-viewer'

// ============================================================================
// Types
// ============================================================================

interface PromptViewerPanelProps {
  prompt: Prompt
  selectedVersion: PromptVersion | null
  protectedLabels: string[]
  availableLabels: string[]
  projectSlug: string // Used for SDK code snippets
  onLabelsChange: (labels: string[]) => void
  currentVariables: string[]
  activeTab: PromptDetailTab
  onTabChange: (tab: PromptDetailTab) => void
  isLabelsLoading?: boolean
  isSidebarCollapsed?: boolean
}

// ============================================================================
// CopyButton Component
// ============================================================================

function CopyButton({ text, className }: { text: string; className?: string }) {
  const [copied, setCopied] = React.useState(false)

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Button
      variant="ghost"
      size="icon"
      className={cn('h-7 w-7', className)}
      onClick={handleCopy}
    >
      {copied ? (
        <Check className="h-3.5 w-3.5 text-green-500" />
      ) : (
        <Copy className="h-3.5 w-3.5" />
      )}
    </Button>
  )
}

// ============================================================================
// VariablesBar Component
// ============================================================================

interface VariablesBarProps {
  variables: string[]
}

function VariablesBar({ variables }: VariablesBarProps) {
  if (variables.length === 0) return null

  return (
    <div className="flex items-center gap-2 px-4 py-2 border-b bg-muted/30">
      <span className="text-xs text-muted-foreground">Variables:</span>
      <div className="flex flex-wrap gap-1.5">
        {variables.map((variable) => (
          <Badge
            key={variable}
            variant="secondary"
            className="font-mono text-xs h-5 px-1.5 cursor-pointer hover:bg-secondary/80"
            onClick={() => {
              navigator.clipboard.writeText(`{{${variable}}}`)
            }}
          >
            {`{{${variable}}}`}
          </Badge>
        ))}
      </div>
    </div>
  )
}

// ============================================================================
// SDKCodeSnippets Component
// ============================================================================

interface SDKCodeSnippetsProps {
  promptName: string
  versionId: string
  labels: string[]
  projectSlug: string
}

function SDKCodeSnippets({
  promptName,
  versionId,
  labels,
  projectSlug,
}: SDKCodeSnippetsProps) {
  const [activeLanguage, setActiveLanguage] = React.useState<'python' | 'typescript' | 'curl'>('python')

  const productionLabel = labels.find((l) =>
    ['production', 'prod'].includes(l.toLowerCase())
  )
  const labelForExample = productionLabel || labels[0] || 'production'

  const pythonCode = `from brokle import Brokle

client = Brokle()

# Get prompt by label (recommended for production)
prompt = client.prompts.get("${promptName}", label="${labelForExample}")

# Or get prompt by specific version ID
prompt = client.prompts.get("${promptName}", version_id="${versionId}")

# Use the prompt template
print(prompt.template)
print(prompt.variables)  # List of required variables`

  const typescriptCode = `import { Brokle } from '@brokle/sdk';

const client = new Brokle();

// Get prompt by label (recommended for production)
const prompt = await client.prompts.get("${promptName}", {
  label: "${labelForExample}"
});

// Or get prompt by specific version ID
const prompt = await client.prompts.get("${promptName}", {
  versionId: "${versionId}"
});

// Use the prompt template
console.log(prompt.template);
console.log(prompt.variables);  // List of required variables`

  const curlCode = `# Get prompt by label
curl -X GET "https://api.brokle.dev/v1/projects/${projectSlug}/prompts/${promptName}?label=${labelForExample}" \\
  -H "Authorization: Bearer $BROKLE_API_KEY"

# Get prompt by version ID
curl -X GET "https://api.brokle.dev/v1/projects/${projectSlug}/prompts/${promptName}?version_id=${versionId}" \\
  -H "Authorization: Bearer $BROKLE_API_KEY"`

  const codeMap = {
    python: pythonCode,
    typescript: typescriptCode,
    curl: curlCode,
  }

  return (
    <div className="space-y-4">
      {/* Language tabs */}
      <div className="flex items-center gap-1 border-b">
        {(['python', 'typescript', 'curl'] as const).map((lang) => (
          <button
            key={lang}
            className={cn(
              'px-3 py-2 text-sm font-medium border-b-2 -mb-px transition-colors',
              activeLanguage === lang
                ? 'border-primary text-foreground'
                : 'border-transparent text-muted-foreground hover:text-foreground'
            )}
            onClick={() => setActiveLanguage(lang)}
          >
            {lang === 'python' ? 'Python' : lang === 'typescript' ? 'TypeScript' : 'cURL'}
          </button>
        ))}
      </div>

      {/* Code snippet */}
      <div className="relative">
        <pre className="rounded-lg bg-muted p-4 font-mono text-sm overflow-x-auto whitespace-pre">
          {codeMap[activeLanguage]}
        </pre>
        <CopyButton
          text={codeMap[activeLanguage]}
          className="absolute top-2 right-2"
        />
      </div>

      {/* Info */}
      <div className="text-xs text-muted-foreground space-y-1">
        <p>
          <strong>Tip:</strong> Use labels like `production` or `staging` to reference prompts
          without hardcoding version IDs.
        </p>
      </div>
    </div>
  )
}

// ============================================================================
// Main PromptViewerPanel Component
// ============================================================================

export function PromptViewerPanel({
  prompt,
  selectedVersion,
  protectedLabels,
  availableLabels,
  projectSlug,
  onLabelsChange,
  currentVariables,
  activeTab,
  onTabChange,
  isLabelsLoading,
  isSidebarCollapsed,
}: PromptViewerPanelProps) {
  // Use selected version data if available, otherwise fall back to prompt
  const version = selectedVersion || {
    id: prompt.version_id,
    version: prompt.version,
    template: prompt.template,
    config: prompt.config,
    variables: prompt.variables,
    commit_message: prompt.commit_message,
    labels: prompt.labels,
    created_at: prompt.created_at,
    created_by: prompt.created_by,
  }

  return (
    <div className="flex h-full flex-col bg-background">
      {/* Header row with labels */}
      <div className={cn(
        "flex items-center border-b py-2 pr-4",
        isSidebarCollapsed ? "pl-12" : "pl-4"
      )}>
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted-foreground">Labels:</span>
          <LabelSelector
            labels={version.labels}
            protectedLabels={protectedLabels}
            availableLabels={availableLabels}
            onChange={onLabelsChange}
            isLoading={isLabelsLoading}
          />
        </div>
      </div>

      {/* Tabs */}
      <Tabs
        value={activeTab}
        onValueChange={(v) => onTabChange(v as PromptDetailTab)}
        className="flex-1 flex flex-col"
      >
        <TabsList className="mx-4 mt-2 w-fit">
          <TabsTrigger value="prompt" className="gap-1.5">
            <Code className="h-3.5 w-3.5" />
            Prompt
          </TabsTrigger>
          <TabsTrigger value="traces" className="gap-1.5">
            <Link2 className="h-3.5 w-3.5" />
            Linked Traces
          </TabsTrigger>
          <TabsTrigger value="sdk" className="gap-1.5">
            <Terminal className="h-3.5 w-3.5" />
            SDK
          </TabsTrigger>
        </TabsList>

        <div className="flex-1 overflow-hidden">
          {/* Prompt Tab - Read Only */}
          <TabsContent value="prompt" className="h-full m-0 flex flex-col">
            <VariablesBar variables={currentVariables} />
            <ScrollArea className="flex-1">
              <div className="p-4 space-y-6">
                {/* Read-only prompt editor */}
                <PromptEditor
                  type={prompt.type}
                  template={version.template}
                  onChange={() => {}} // No-op for read-only
                  variables={currentVariables}
                  readOnly
                />

                {/* Divider */}
                <div className="border-t" />

                {/* Read-only JSON Config Viewer */}
                <JsonConfigViewer
                  config={(version.config as Record<string, unknown>) || null}
                />
              </div>
            </ScrollArea>
          </TabsContent>

          {/* Linked Traces Tab */}
          <TabsContent value="traces" className="h-full m-0">
            <div className="flex flex-col items-center justify-center h-full py-12 text-center">
              <Link2 className="h-12 w-12 text-muted-foreground/30 mb-4" />
              <p className="text-muted-foreground">Coming soon</p>
              <p className="text-xs text-muted-foreground/70 mt-1">
                View traces that used this prompt version
              </p>
            </div>
          </TabsContent>

          {/* SDK Tab */}
          <TabsContent value="sdk" className="h-full m-0">
            <ScrollArea className="h-full">
              <div className="p-4">
                <SDKCodeSnippets
                  promptName={prompt.name}
                  versionId={version.id}
                  labels={version.labels}
                  projectSlug={projectSlug}
                />
              </div>
            </ScrollArea>
          </TabsContent>
        </div>
      </Tabs>
    </div>
  )
}
