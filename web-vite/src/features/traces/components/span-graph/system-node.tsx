import { memo } from 'react'
import { Handle, Position, type NodeProps } from 'reactflow'
import { Play, Square } from 'lucide-react'
import { cn } from '@/lib/utils'

/**
 * Data structure for system nodes (__start__, __end__).
 */
export interface SystemNodeData {
  type: 'start' | 'end'
  label: string
}

/**
 * SystemNode - Special node for __start__ and __end__ markers.
 */
function SystemNodeComponent({ data }: NodeProps<SystemNodeData>) {
  const isStart = data.type === 'start'

  return (
    <>
      {isStart && (
        <Handle
          type="source"
          position={Position.Bottom}
          className="!w-2 !h-2 !bg-muted-foreground/50"
        />
      )}

      {!isStart && (
        <Handle
          type="target"
          position={Position.Top}
          className="!w-2 !h-2 !bg-muted-foreground/50"
        />
      )}

      <div
        className={cn(
          'px-4 py-2 rounded-full border-2 border-dashed',
          'bg-muted/50 text-muted-foreground',
          'min-w-[100px] flex items-center justify-center',
          isStart
            ? 'border-green-500/50 dark:border-green-400/50'
            : 'border-gray-400/50 dark:border-gray-500/50',
        )}
      >
        <div className="flex items-center gap-2">
          {isStart ? (
            <Play className="h-3.5 w-3.5 text-green-600 dark:text-green-400" />
          ) : (
            <Square className="h-3.5 w-3.5 text-gray-500 dark:text-gray-400" />
          )}
          <span className="text-xs font-mono">{data.label}</span>
        </div>
      </div>
    </>
  )
}

export const SystemNode = memo(SystemNodeComponent)
