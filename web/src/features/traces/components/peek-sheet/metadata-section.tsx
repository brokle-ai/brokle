'use client'

import * as React from 'react'
import { Badge } from '@/components/ui/badge'
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from '@/components/ui/collapsible'
import { ChevronDown, ChevronRight, Braces, Server, Puzzle } from 'lucide-react'
import { cn } from '@/lib/utils'
import { FlatKeyValueList } from './path-value-tree'

interface AttributesSectionProps {
  /** Span attributes (gen_ai.*, user.*, session.*, custom app attributes) */
  spanAttributes?: Record<string, any>
  /** Resource attributes (service.name, deployment.environment, etc.) */
  resourceAttributes?: Record<string, any>
  /** Scope attributes (instrumentation scope info) */
  scopeAttributes?: Record<string, any>
  /** Scope name and version */
  scopeName?: string
  scopeVersion?: string
  className?: string
}

// Legacy interface for backwards compatibility
interface MetadataSectionProps {
  attributes?: Record<string, any>
  resourceAttributes?: Record<string, any>
  metadata?: Record<string, any>
  className?: string
}

/**
 * Count items in a record
 */
function countItems(obj: Record<string, any> | undefined | null): number {
  if (!obj) return 0
  return Object.keys(obj).length
}

/**
 * AttributeGroup - Collapsible group of attributes with PathValueTree
 */
interface AttributeGroupProps {
  title: string
  icon?: React.ReactNode
  data: Record<string, any>
  defaultOpen?: boolean
  subtitle?: string
}

function AttributeGroup({ title, icon, data, defaultOpen = false, subtitle }: AttributeGroupProps) {
  const [isOpen, setIsOpen] = React.useState(defaultOpen)
  const itemCount = countItems(data)

  if (itemCount === 0) return null

  return (
    <Collapsible open={isOpen} onOpenChange={setIsOpen}>
      <CollapsibleTrigger asChild>
        <div className='flex items-center gap-2 py-1.5 cursor-pointer hover:bg-muted/50 rounded-md px-2 -mx-2'>
          {isOpen ? (
            <ChevronDown className='h-4 w-4 text-muted-foreground flex-shrink-0' />
          ) : (
            <ChevronRight className='h-4 w-4 text-muted-foreground flex-shrink-0' />
          )}
          {icon && <span className='text-muted-foreground'>{icon}</span>}
          <span className='text-sm font-medium truncate'>{title}</span>
          {subtitle && (
            <span className='text-xs text-muted-foreground truncate'>{subtitle}</span>
          )}
          <Badge variant='secondary' className='text-xs ml-auto'>
            {itemCount}
          </Badge>
        </div>
      </CollapsibleTrigger>

      <CollapsibleContent>
        <div className='ml-4 py-1 border-l border-border/50 pl-2'>
          <FlatKeyValueList data={data} showCopyButtons={true} />
        </div>
      </CollapsibleContent>
    </Collapsible>
  )
}

/**
 * AttributesSection - OTEL-native attributes display
 * Shows span attributes, resource attributes, and scope info in collapsible groups
 * Uses FlatKeyValueList for displaying dotted OTEL keys (e.g., gen_ai.request.model)
 */
export function AttributesSection({
  spanAttributes,
  resourceAttributes,
  scopeAttributes,
  scopeName,
  scopeVersion,
  className,
}: AttributesSectionProps) {
  const hasSpanAttributes = countItems(spanAttributes) > 0
  const hasResourceAttributes = countItems(resourceAttributes) > 0
  const hasScopeAttributes = countItems(scopeAttributes) > 0
  const hasAnyData = hasSpanAttributes || hasResourceAttributes || hasScopeAttributes

  if (!hasAnyData) {
    return (
      <div className={cn('text-sm text-muted-foreground italic py-2', className)}>
        No attributes available
      </div>
    )
  }

  // Format scope subtitle
  const scopeSubtitle = scopeName
    ? scopeVersion
      ? `${scopeName}@${scopeVersion}`
      : scopeName
    : undefined

  return (
    <div className={cn('space-y-1', className)}>
      {hasSpanAttributes && (
        <AttributeGroup
          title='Span Attributes'
          icon={<Braces className='h-3.5 w-3.5' />}
          data={spanAttributes!}
          defaultOpen={true}
        />
      )}

      {hasResourceAttributes && (
        <AttributeGroup
          title='Resource Attributes'
          icon={<Server className='h-3.5 w-3.5' />}
          data={resourceAttributes!}
          defaultOpen={false}
        />
      )}

      {hasScopeAttributes && (
        <AttributeGroup
          title='Scope Attributes'
          icon={<Puzzle className='h-3.5 w-3.5' />}
          data={scopeAttributes!}
          defaultOpen={false}
          subtitle={scopeSubtitle}
        />
      )}
    </div>
  )
}

/**
 * MetadataSection - Legacy wrapper for backwards compatibility
 * @deprecated Use AttributesSection instead
 */
export function MetadataSection({
  attributes,
  resourceAttributes,
  metadata,
  className,
}: MetadataSectionProps) {
  // Merge metadata into attributes for backwards compatibility
  const mergedAttributes = {
    ...attributes,
    ...metadata,
  }

  return (
    <AttributesSection
      spanAttributes={Object.keys(mergedAttributes).length > 0 ? mergedAttributes : undefined}
      resourceAttributes={resourceAttributes}
      className={className}
    />
  )
}
