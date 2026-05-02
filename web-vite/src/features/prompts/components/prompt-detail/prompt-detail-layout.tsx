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
import type { Prompt, PromptVersion } from '../../types'
import { extractVariables } from '../../utils/variable-extraction'
import { usePromptDetailState } from '../../hooks/use-prompt-detail-state'
import { useVersionQuery } from '../../api/queries'
import { VersionSidebar } from './version-sidebar'
import { PromptViewerPanel } from './prompt-viewer-panel'

interface PromptDetailLayoutProps {
  prompt: Prompt
  versions: PromptVersion[]
  versionsLoading: boolean
  protectedLabels: string[]
  availableLabels: string[]
  projectId: string
  projectSlug: string
  selectedVersionId: string | null
  onVersionChange: (versionId: string | null) => void
  onLabelsChange: (labels: string[]) => void
  onCompare?: (fromVersion: number, toVersion: number) => void
  onRestore?: (version: PromptVersion) => void
  onCreateVersion?: () => void
  isLabelsLoading?: boolean
}

const PANEL_SIZES = {
  left: { default: 22, min: 18, max: 30 },
  center: { default: 78, min: 60 },
}

function ExpandButton({ onClick }: { onClick: () => void }) {
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

export function PromptDetailLayout({
  prompt,
  versions,
  versionsLoading,
  protectedLabels,
  availableLabels,
  projectId,
  projectSlug,
  selectedVersionId,
  onVersionChange,
  onLabelsChange,
  onRestore,
  onCreateVersion,
  isLabelsLoading,
}: PromptDetailLayoutProps) {
  const { activeTab, setActiveTab } = usePromptDetailState()
  const [isLeftCollapsed, setIsLeftCollapsed] = React.useState(false)

  const shouldFetchVersion =
    selectedVersionId && selectedVersionId !== prompt.version_id

  const { data: fetchedVersion } = useVersionQuery(
    projectId,
    prompt.id,
    selectedVersionId || '',
    { enabled: !!shouldFetchVersion },
  )

  const selectedVersion = React.useMemo(() => {
    if (!selectedVersionId) return null
    const fromList = versions.find((v) => v.id === selectedVersionId)
    if (fromList) return fromList
    if (fetchedVersion) return fetchedVersion
    return null
  }, [selectedVersionId, versions, fetchedVersion])

  React.useEffect(() => {
    if (!selectedVersionId && versions.length > 0 && !versionsLoading) {
      onVersionChange(versions[0].id)
    }
  }, [selectedVersionId, versions, versionsLoading, onVersionChange])

  const handleVersionSelect = React.useCallback(
    (version: PromptVersion) => {
      onVersionChange(version.id)
    },
    [onVersionChange],
  )

  const currentTemplate = selectedVersion?.template || prompt.template
  const currentVariables = React.useMemo(
    () => extractVariables(currentTemplate, prompt.type),
    [currentTemplate, prompt.type],
  )

  return (
    <div className="flex h-full flex-col">
      <ResizablePanelGroup direction="horizontal" className="flex-1">
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
                selectedVersionId={selectedVersionId}
                protectedLabels={protectedLabels}
                promptType={prompt.type}
                promptName={prompt.name}
                onVersionSelect={handleVersionSelect}
                onCreateVersion={onCreateVersion}
                onCollapse={() => setIsLeftCollapsed(true)}
                onRestore={onRestore}
              />
            </ResizablePanel>
            <ResizableHandle withHandle />
          </>
        )}

        <ResizablePanel
          defaultSize={isLeftCollapsed ? 100 : PANEL_SIZES.center.default}
          minSize={PANEL_SIZES.center.min}
          className="relative min-w-0"
        >
          {isLeftCollapsed && (
            <ExpandButton onClick={() => setIsLeftCollapsed(false)} />
          )}
          <PromptViewerPanel
            prompt={prompt}
            selectedVersion={selectedVersion}
            protectedLabels={protectedLabels}
            availableLabels={availableLabels}
            projectSlug={projectSlug}
            onLabelsChange={onLabelsChange}
            currentVariables={currentVariables}
            activeTab={activeTab}
            onTabChange={setActiveTab}
            isLabelsLoading={isLabelsLoading}
            isSidebarCollapsed={isLeftCollapsed}
          />
        </ResizablePanel>
      </ResizablePanelGroup>
    </div>
  )
}
