import { useMemo } from 'react'
import dagre from 'dagre'
import {
  forceSimulation,
  forceLink,
  forceManyBody,
  forceCenter,
  forceX,
  forceY,
} from 'd3-force'
import type { Node, Edge } from 'reactflow'
import type { Span } from '../../api/types'
import type { SpanNodeData } from './span-node'
import type { SystemNodeData } from './system-node'
import type { LayoutMode } from './graph-controls'
import { buildStepGroups, buildStepEdges, type StepGroup } from './step-grouping'
import { detectSpanCategory } from '../../utils/span-type-detector'

export interface UseGraphLayoutOptions {
  layoutMode: LayoutMode
  showSystemNodes: boolean
  groupByStep: boolean
}

export interface UseGraphLayoutResult {
  nodes: Node<SpanNodeData | SystemNodeData>[]
  edges: Edge[]
  isLoading: boolean
  steps: StepGroup[]
}

const NODE_WIDTH = 180
const NODE_HEIGHT = 60
const SYSTEM_NODE_WIDTH = 120
const SYSTEM_NODE_HEIGHT = 40

function getStatusLabel(statusCode: number | undefined): string {
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
 * Extract total token count from a Span's usage_details bag. The
 * ingestion worker materialises gen-AI counters under keys like
 * `input_tokens` / `output_tokens` / `total_tokens`; we prefer the
 * explicit `total_tokens` when present and otherwise sum the two
 * halves, so we never double-count.
 */
function extractTokens(span: Span): number | undefined {
  const usage = span.usage_details
  if (!usage) return undefined
  const total = usage.total_tokens
  if (typeof total === 'number' && total > 0) return total
  const input = typeof usage.input_tokens === 'number' ? usage.input_tokens : 0
  const output = typeof usage.output_tokens === 'number' ? usage.output_tokens : 0
  const sum = input + output
  return sum > 0 ? sum : undefined
}

function parseCost(cost: string | undefined): number | undefined {
  if (!cost) return undefined
  const n = Number(cost)
  return Number.isFinite(n) ? n : undefined
}

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

function createSpanNodes(
  spans: Span[],
  selectedSpanId?: string,
): Node<SpanNodeData>[] {
  return spans.map((span) => {
    const category = detectSpanCategory(span.span_name, span.span_attributes)
    const tokens = extractTokens(span)

    return {
      id: span.span_id,
      type: 'span',
      position: { x: 0, y: 0 },
      data: {
        span,
        category,
        label: span.span_name,
        duration: span.duration
          ? `${Math.round(span.duration / 1_000_000)}ms`
          : undefined,
        tokens,
        cost: parseCost(span.total_cost),
        hasError: span.has_error || span.status_code === 2,
        isSelected: span.span_id === selectedSpanId,
        model: span.model_name,
        statusCode: getStatusLabel(span.status_code),
      },
    }
  })
}

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

function connectSystemNodes(
  spans: Span[],
  groups: StepGroup[],
  groupByStep: boolean,
): Edge[] {
  const edges: Edge[] = []

  if (groupByStep && groups.length > 0) {
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
    const spanIdSet = new Set(spans.map((s) => s.span_id))
    const rootSpans = spans.filter(
      (s) => !s.parent_span_id || !spanIdSet.has(s.parent_span_id),
    )

    const parentIds = new Set(
      spans.filter((s) => s.parent_span_id).map((s) => s.parent_span_id),
    )
    const leafSpans = spans.filter((s) => !parentIds.has(s.span_id))

    for (const span of rootSpans) {
      edges.push({
        id: `__start__-${span.span_id}`,
        source: '__start__',
        target: span.span_id,
        type: 'smoothstep',
        style: { strokeWidth: 1.5, strokeDasharray: '5,5' },
      })
    }

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

function applyDagreLayout(nodes: Node[], edges: Edge[]): Node[] {
  const dagreGraph = new dagre.graphlib.Graph()
  dagreGraph.setDefaultEdgeLabel(() => ({}))
  dagreGraph.setGraph({
    rankdir: 'TB',
    nodesep: 60,
    ranksep: 80,
    marginx: 40,
    marginy: 40,
  })

  nodes.forEach((node) => {
    const isSystem = node.type === 'system'
    dagreGraph.setNode(node.id, {
      width: isSystem ? SYSTEM_NODE_WIDTH : NODE_WIDTH,
      height: isSystem ? SYSTEM_NODE_HEIGHT : NODE_HEIGHT,
    })
  })

  edges.forEach((edge) => {
    dagreGraph.setEdge(edge.source, edge.target)
  })

  dagre.layout(dagreGraph)

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

interface SimulationNode {
  id: string
  x?: number
  y?: number
  vx?: number
  vy?: number
  fx?: number | null
  fy?: number | null
}

function applyPhysicsLayout(nodes: Node[], edges: Edge[]): Node[] {
  const simNodes: SimulationNode[] = nodes.map((n) => ({
    id: n.id,
    x: Math.random() * 800,
    y: Math.random() * 600,
  }))

  const simLinks = edges.map((e) => ({
    source: e.source,
    target: e.target,
  }))

  const nodeById = new Map(simNodes.map((n) => [n.id, n]))

  const simulation = forceSimulation(simNodes)
    .force(
      'link',
      forceLink<SimulationNode, (typeof simLinks)[0]>(simLinks)
        .id((d) => d.id)
        .distance(150)
        .strength(0.5),
    )
    .force('charge', forceManyBody().strength(-400))
    .force('center', forceCenter(400, 300))
    .force('x', forceX(400).strength(0.05))
    .force('y', forceY(300).strength(0.05))
    .stop()

  for (let i = 0; i < 300; i++) {
    simulation.tick()
  }

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
 * Convert a flat span array into React Flow nodes and edges with layout.
 */
export function useGraphLayout(
  spans: Span[],
  selectedSpanId?: string,
  options: UseGraphLayoutOptions = {
    layoutMode: 'dagre',
    showSystemNodes: true,
    groupByStep: true,
  },
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

    // web-vite spans arrive flat — no nested child_spans array.
    const flatSpans = spans

    const steps = buildStepGroups(flatSpans)

    let nodes: Node<SpanNodeData | SystemNodeData>[] = createSpanNodes(
      flatSpans,
      selectedSpanId,
    )

    let edges: Edge[] = []

    if (options.groupByStep && steps.length > 1) {
      const stepEdges = buildStepEdges(steps)
      edges = stepEdges.map((e) => ({
        ...e,
        type: 'smoothstep',
        animated: false,
        style: { strokeWidth: 1.5 },
      }))
    } else {
      edges = buildHierarchyEdges(flatSpans)
    }

    if (options.showSystemNodes) {
      const { startNode, endNode } = createSystemNodes()
      const systemEdges = connectSystemNodes(flatSpans, steps, options.groupByStep)
      nodes = [startNode, ...nodes, endNode]
      edges = [...edges, ...systemEdges]
    }

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
  }, [
    spans,
    selectedSpanId,
    options.layoutMode,
    options.showSystemNodes,
    options.groupByStep,
  ])
}
