'use client'

import { useMemo } from 'react'
import dagre from 'dagre'
import { forceSimulation, forceLink, forceManyBody, forceCenter, forceX, forceY } from 'd3-force'
import type { Node, Edge } from 'reactflow'
import type { Span } from '../../data/schema'
import type { SpanNodeData } from './span-node'
import type { SystemNodeData } from './system-node'
import type { LayoutMode } from './graph-controls'
import { buildStepGroups, buildStepEdges, type StepGroup } from './step-grouping'
import { detectSpanCategory } from '../../utils/span-type-detector'

/**
 * Configuration for the graph layout hook
 */
export interface UseGraphLayoutOptions {
  layoutMode: LayoutMode
  showSystemNodes: boolean
  groupByStep: boolean
}

/**
 * Result of the graph layout hook
 */
export interface UseGraphLayoutResult {
  nodes: Node<SpanNodeData | SystemNodeData>[]
  edges: Edge[]
  isLoading: boolean
  steps: StepGroup[]
}

/**
 * Node dimensions for layout calculation
 */
const NODE_WIDTH = 180
const NODE_HEIGHT = 60
const SYSTEM_NODE_WIDTH = 120
const SYSTEM_NODE_HEIGHT = 40

/**
 * Get status code label
 */
function getStatusLabel(statusCode: number): string {
  switch (statusCode) {
    case 0:
      return 'UNSET'
    case 1:
      return 'OK'
    case 2:
      return 'ERROR'
    default:
      return 'UNKNOWN'
  }
}

/**
 * Flatten nested spans into a flat array
 */
function flattenSpans(spans: Span[]): Span[] {
  const result: Span[] = []

  function traverse(span: Span) {
    result.push(span)
    if (span.child_spans) {
      span.child_spans.forEach(traverse)
    }
  }

  spans.forEach(traverse)
  return result
}

/**
 * Build parent-child edges from span relationships
 */
function buildHierarchyEdges(spans: Span[]): Edge[] {
  const edges: Edge[] = []
  const spanIdSet = new Set(spans.map((s) => s.span_id))

  for (const span of spans) {
    if (span.parent_span_id && spanIdSet.has(span.parent_span_id)) {
      edges.push({
        id: `hierarchy-${span.parent_span_id}-${span.span_id}`,
        source: span.parent_span_id,
        target: span.span_id,
        type: 'smoothstep',
        animated: false,
        style: { strokeWidth: 1.5 },
      })
    }
  }

  return edges
}

/**
 * Create span nodes from spans array
 */
function createSpanNodes(
  spans: Span[],
  selectedSpanId?: string
): Node<SpanNodeData>[] {
  return spans.map((span) => {
    const category = detectSpanCategory(span.span_name, span.attributes)

    return {
      id: span.span_id,
      type: 'span',
      position: { x: 0, y: 0 }, // Will be set by layout
      data: {
        span,
        category,
        label: span.span_name,
        duration: span.duration ? `${Math.round(span.duration / 1_000_000)}ms` : undefined,
        tokens:
          (span.gen_ai_usage_input_tokens || 0) + (span.gen_ai_usage_output_tokens || 0) ||
          undefined,
        cost: span.total_cost,
        hasError: span.has_error || span.status_code === 2,
        isSelected: span.span_id === selectedSpanId,
        model: span.model_name || span.gen_ai_request_model,
        statusCode: getStatusLabel(span.status_code),
      },
    }
  })
}

/**
 * Create system nodes (__start__, __end__)
 */
function createSystemNodes(): {
  startNode: Node<SystemNodeData>
  endNode: Node<SystemNodeData>
} {
  return {
    startNode: {
      id: '__start__',
      type: 'system',
      position: { x: 0, y: 0 },
      data: { type: 'start', label: '__start__' },
    },
    endNode: {
      id: '__end__',
      type: 'system',
      position: { x: 0, y: 0 },
      data: { type: 'end', label: '__end__' },
    },
  }
}

/**
 * Connect system nodes to span nodes
 */
function connectSystemNodes(
  spans: Span[],
  groups: StepGroup[],
  groupByStep: boolean
): Edge[] {
  const edges: Edge[] = []

  if (groupByStep && groups.length > 0) {
    // Connect start to first step spans
    const firstStep = groups[0]
    for (const span of firstStep.spans) {
      edges.push({
        id: `__start__-${span.span_id}`,
        source: '__start__',
        target: span.span_id,
        type: 'smoothstep',
        style: { strokeWidth: 1.5, strokeDasharray: '5,5' },
      })
    }

    // Connect last step spans to end
    const lastStep = groups[groups.length - 1]
    for (const span of lastStep.spans) {
      edges.push({
        id: `${span.span_id}-__end__`,
        source: span.span_id,
        target: '__end__',
        type: 'smoothstep',
        style: { strokeWidth: 1.5, strokeDasharray: '5,5' },
      })
    }
  } else {
    // Find root spans (no parent or parent not in set)
    const spanIdSet = new Set(spans.map((s) => s.span_id))
    const rootSpans = spans.filter(
      (s) => !s.parent_span_id || !spanIdSet.has(s.parent_span_id)
    )

    // Find leaf spans (no children)
    const parentIds = new Set(spans.filter((s) => s.parent_span_id).map((s) => s.parent_span_id))
    const leafSpans = spans.filter((s) => !parentIds.has(s.span_id))

    // Connect start to roots
    for (const span of rootSpans) {
      edges.push({
        id: `__start__-${span.span_id}`,
        source: '__start__',
        target: span.span_id,
        type: 'smoothstep',
        style: { strokeWidth: 1.5, strokeDasharray: '5,5' },
      })
    }

    // Connect leaves to end
    for (const span of leafSpans) {
      edges.push({
        id: `${span.span_id}-__end__`,
        source: span.span_id,
        target: '__end__',
        type: 'smoothstep',
        style: { strokeWidth: 1.5, strokeDasharray: '5,5' },
      })
    }
  }

  return edges
}

/**
 * Apply dagre (hierarchical) layout to nodes
 */
function applyDagreLayout(
  nodes: Node[],
  edges: Edge[]
): Node[] {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))
  dagreGraph.setGraph({
    rankdir: 'TB', // Top to bottom
    nodesep: 60, // Horizontal spacing
    ranksep: 80, // Vertical spacing
    marginx: 40,
    marginy: 40,
  })

  // Add nodes to dagre graph
  nodes.forEach((node) => {
    const isSystem = node.type === 'system'
    dagreGraph.setNode(node.id, {
      width: isSystem ? SYSTEM_NODE_WIDTH : NODE_WIDTH,
      height: isSystem ? SYSTEM_NODE_HEIGHT : NODE_HEIGHT,
    })
  })

  // Add edges to dagre graph
  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target)
  })

  // Run dagre layout
  dagre.layout(dagreGraph)

  // Update node positions
  return nodes.map((node) => {
    const nodeWithPosition = dagreGraph.node(node.id)
    const isSystem = node.type === 'system'
    const width = isSystem ? SYSTEM_NODE_WIDTH : NODE_WIDTH
    const height = isSystem ? SYSTEM_NODE_HEIGHT : NODE_HEIGHT

    return {
      ...node,
      position: {
        x: nodeWithPosition.x - width / 2,
        y: nodeWithPosition.y - height / 2,
      },
    }
  })
}

/**
 * D3 force simulation node interface
 */
interface SimulationNode {
  id: string
  x?: number
  y?: number
  vx?: number
  vy?: number
  fx?: number | null
  fy?: number | null
}

/**
 * Apply physics (force-directed) layout to nodes
 */
function applyPhysicsLayout(
  nodes: Node[],
  edges: Edge[]
): Node[] {
  // Create simulation nodes
  const simNodes: SimulationNode[] = nodes.map((n) => ({
    id: n.id,
    x: Math.random() * 800,
    y: Math.random() * 600,
  }))

  // Create simulation links
  const simLinks = edges.map((e) => ({
    source: e.source,
    target: e.target,
  }))

  // Create node index map for link resolution
  const nodeById = new Map(simNodes.map((n) => [n.id, n]))

  // Run simulation
  const simulation = forceSimulation(simNodes)
    .force(
      'link',
      forceLink<SimulationNode, typeof simLinks[0]>(simLinks)
        .id((d) => d.id)
        .distance(150)
        .strength(0.5)
    )
    .force('charge', forceManyBody().strength(-400))
    .force('center', forceCenter(400, 300))
    .force('x', forceX(400).strength(0.05))
    .force('y', forceY(300).strength(0.05))
    .stop()

  // Run 300 iterations for stabilization
  for (let i = 0; i < 300; i++) {
    simulation.tick()
  }

  // Update node positions
  return nodes.map((node) => {
    const simNode = nodeById.get(node.id)
    return {
      ...node,
      position: {
        x: simNode?.x ?? 0,
        y: simNode?.y ?? 0,
      },
    }
  })
}

/**
 * Hook to convert spans into React Flow nodes and edges with layout
 *
 * @param spans - Array of spans to visualize
 * @param selectedSpanId - Currently selected span ID
 * @param options - Layout options
 * @returns Nodes and edges for React Flow
 */
export function useGraphLayout(
  spans: Span[],
  selectedSpanId?: string,
  options: UseGraphLayoutOptions = {
    layoutMode: 'dagre',
    showSystemNodes: true,
    groupByStep: true,
  }
): UseGraphLayoutResult {
  return useMemo(() => {
    if (!spans || spans.length === 0) {
      return {
        nodes: [],
        edges: [],
        isLoading: false,
        steps: [],
      }
    }

    // Flatten nested spans
    const flatSpans = flattenSpans(spans)

    // Build step groups for temporal analysis
    const steps = buildStepGroups(flatSpans)

    // Create span nodes
    let nodes: Node<SpanNodeData | SystemNodeData>[] = createSpanNodes(
      flatSpans,
      selectedSpanId
    )

    // Build edges based on groupByStep option
    let edges: Edge[] = []

    if (options.groupByStep && steps.length > 1) {
      // Use step-based edges for parallel execution visualization
      const stepEdges = buildStepEdges(steps)
      edges = stepEdges.map((e) => ({
        ...e,
        type: 'smoothstep',
        animated: false,
        style: { strokeWidth: 1.5 },
      }))
    } else {
      // Use hierarchy edges from parent-child relationships
      edges = buildHierarchyEdges(flatSpans)
    }

    // Add system nodes if enabled
    if (options.showSystemNodes) {
      const { startNode, endNode } = createSystemNodes()
      const systemEdges = connectSystemNodes(flatSpans, steps, options.groupByStep)
      nodes = [startNode, ...nodes, endNode]
      edges = [...edges, ...systemEdges]
    }

    // Apply layout
    const layoutedNodes =
      options.layoutMode === 'dagre'
        ? applyDagreLayout(nodes, edges)
        : applyPhysicsLayout(nodes, edges)

    return {
      nodes: layoutedNodes,
      edges,
      isLoading: false,
      steps,
    }
  }, [spans, selectedSpanId, options.layoutMode, options.showSystemNodes, options.groupByStep])
}
