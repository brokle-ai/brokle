'use client'

import * as React from 'react'
import { PanelLeftOpen } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle,
} from '@/components/ui/resizable'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import type { Prompt, PromptVersion, CreateVersionRequest } from '../../types'
import { extractVariables } from '../../utils/variable-extraction'
import { VersionSidebar } from '../prompt-detail/version-sidebar'
import { PromptEditPanel } from './prompt-edit-panel'

// ============================================================================
// Types
// ============================================================================

interface PromptEditLayoutProps {
  prompt: Prompt
  versions: PromptVersion[]
  versionsLoading: boolean
  sourceVersion: PromptVersion | null
  isRestoreFlow: boolean
  onSave: (data: CreateVersionRequest) => Promise<void>
  onCancel: () => void
  onVersionSelect: (version: PromptVersion) => void
  isSaving?: boolean
}

// ============================================================================
// Panel Configuration
// ============================================================================

const PANEL_SIZES = {
  left: { default: 22, min: 18, max: 30 },
  center: { default: 78, min: 60 },
}

// ============================================================================
// ExpandButton Component
// ============================================================================

interface ExpandButtonProps {
  onClick: () => void
}

function ExpandButton({ onClick }: ExpandButtonProps) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className="absolute top-2 left-2 z-10 h-7 w-7"
          onClick={onClick}
        >
          <PanelLeftOpen className="h-4 w-4" />
          <span className="sr-only">Show versions</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent side="right">
        <p className="text-xs">Show versions</p>
      </TooltipContent>
    </Tooltip>
  )
}

// ============================================================================
// Main PromptEditLayout Component
// ============================================================================

export function PromptEditLayout({
  prompt,
  versions,
  versionsLoading,
  sourceVersion,
  isRestoreFlow,
  onSave,
  onCancel,
  onVersionSelect,
  isSaving,
}: PromptEditLayoutProps) {
  // Local UI state - sidebar collapsed by default on edit page
  const [isLeftCollapsed, setIsLeftCollapsed] = React.useState(true)

  // Calculate current variables from source version template
  const currentTemplate = sourceVersion?.template || prompt.template
  const currentVariables = React.useMemo(
    () => extractVariables(currentTemplate, prompt.type),
    [currentTemplate, prompt.type]
  )

  // Handle version selection - load that version's content
  const handleVersionSelect = React.useCallback(
    (version: PromptVersion) => {
      onVersionSelect(version)
    },
    [onVersionSelect]
  )

  return (
    <div className="flex h-full flex-col">
      <ResizablePanelGroup direction="horizontal" className="flex-1">
        {/* Left Panel: Version Sidebar (reference only, collapsed by default) */}
        {!isLeftCollapsed && (
          <>
            <ResizablePanel
              defaultSize={PANEL_SIZES.left.default}
              minSize={PANEL_SIZES.left.min}
              maxSize={PANEL_SIZES.left.max}
              className="min-w-0"
            >
              <VersionSidebar
                versions={versions}
                selectedVersionId={sourceVersion?.id || null}
                protectedLabels={[]}
                promptType={prompt.type}
                promptName={prompt.name}
                onVersionSelect={handleVersionSelect}
                onCollapse={() => setIsLeftCollapsed(true)}
                // No create or restore on edit page sidebar - omitting onCreateVersion hides the button
              />
            </ResizablePanel>
            <ResizableHandle withHandle />
          </>
        )}

        {/* Center Panel: Editor */}
        <ResizablePanel
          defaultSize={isLeftCollapsed ? 100 : PANEL_SIZES.center.default}
          minSize={PANEL_SIZES.center.min}
          className="relative min-w-0"
        >
          {isLeftCollapsed && (
            <ExpandButton onClick={() => setIsLeftCollapsed(false)} />
          )}
          <PromptEditPanel
            prompt={prompt}
            sourceVersion={sourceVersion}
            currentVariables={currentVariables}
            isRestoreFlow={isRestoreFlow}
            onSave={onSave}
            onCancel={onCancel}
            isSaving={isSaving}
          />
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  )
}
