'use client'

import { useCallback, useState } from 'react'
import ReactFlow, {
  Background,
  Controls,
  MiniMap,
  type Node,
  type NodeTypes,
  BackgroundVariant,
} from 'reactflow'
import 'reactflow/dist/style.css'
import { TooltipProvider } from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import type { Span } from '../../data/schema'
import { SpanNode, type SpanNodeData } from './span-node'
import { SystemNode, type SystemNodeData } from './system-node'
import { GraphControls, type LayoutMode } from './graph-controls'
import { useGraphLayout } from './use-graph-layout'

/**
 * Custom node types for React Flow
 */
const nodeTypes: NodeTypes = {
  span: SpanNode,
  system: SystemNode,
}

interface SpanGraphProps {
  spans: Span[]
  selectedSpanId?: string
  onSpanSelect: (span: Span) => void
  className?: string
}

/**
 * SpanGraph - Interactive graph visualization of span execution flow
 *
 * Features:
 * - React Flow canvas with custom nodes
 * - Zoom/pan controls (built-in)
 * - Mini-map for navigation (for large graphs)
 * - Fit view on initial render
 * - Node selection synced with span detail panel
 * - Multiple layout options (dagre/physics)
 * - System nodes (__start__, __end__)
 * - Step grouping for parallel execution visualization
 */
export function SpanGraph({
  spans,
  selectedSpanId,
  onSpanSelect,
  className,
}: SpanGraphProps) {
  // Layout options state
  const [layoutMode, setLayoutMode] = useState<LayoutMode>('dagre')
  const [showSystemNodes, setShowSystemNodes] = useState(true)
  const [groupByStep, setGroupByStep] = useState(true)

  // Get layout data
  const { nodes, edges, steps } = useGraphLayout(spans, selectedSpanId, {
    layoutMode,
    showSystemNodes,
    groupByStep,
  })

  // Handle node click
  const handleNodeClick = useCallback(
    (_: React.MouseEvent, node: Node<SpanNodeData | SystemNodeData>) => {
      // Only handle span nodes, not system nodes
      if (node.type === 'span') {
        const spanData = node.data as SpanNodeData
        onSpanSelect(spanData.span)
      }
    },
    [onSpanSelect]
  )

  // Empty state
  if (!spans || spans.length === 0) {
    return (
      <div
        className={cn(
          'flex items-center justify-center h-full text-sm text-muted-foreground',
          className
        )}
      >
        No spans to visualize
      </div>
    )
  }

  // Calculate if we should show the mini-map
  const showMiniMap = nodes.length > 15

  return (
    <TooltipProvider>
      <div className={cn('h-full w-full relative', className)}>
        {/* Graph Controls */}
        <GraphControls
          layoutMode={layoutMode}
          onLayoutModeChange={setLayoutMode}
          showSystemNodes={showSystemNodes}
          onShowSystemNodesChange={setShowSystemNodes}
          groupByStep={groupByStep}
          onGroupByStepChange={setGroupByStep}
        />

        {/* React Flow Canvas */}
        <ReactFlow
          nodes={nodes}
          edges={edges}
          nodeTypes={nodeTypes}
          onNodeClick={handleNodeClick}
          fitView
          fitViewOptions={{
            padding: 0.2,
            includeHiddenNodes: false,
          }}
          minZoom={0.1}
          maxZoom={2}
          proOptions={{ hideAttribution: true }}
          defaultEdgeOptions={{
            type: 'smoothstep',
            animated: false,
          }}
          className="bg-background"
        >
          <Background
            variant={BackgroundVariant.Dots}
            gap={16}
            size={1}
            color="hsl(var(--muted-foreground) / 0.2)"
          />
          <Controls
            className="!bg-background !border-border !shadow-sm"
            showInteractive={false}
          />
          {showMiniMap && (
            <MiniMap
              className="!bg-muted/50 !border-border"
              nodeColor={(node) => {
                if (node.type === 'system') return 'hsl(var(--muted-foreground))'
                const data = node.data as SpanNodeData
                if (data.hasError) return 'hsl(var(--destructive))'
                if (data.isSelected) return 'hsl(var(--primary))'
                return 'hsl(var(--muted-foreground))'
              }}
              maskColor="hsl(var(--background) / 0.8)"
              pannable
              zoomable
            />
          )}
        </ReactFlow>

        {/* Step count indicator */}
        {groupByStep && steps.length > 0 && (
          <div className="absolute bottom-2 left-2 z-10 text-xs text-muted-foreground bg-background/80 backdrop-blur-sm px-2 py-1 rounded border">
            {steps.length} execution step{steps.length !== 1 ? 's' : ''}
          </div>
        )}
      </div>
    </TooltipProvider>
  )
}
