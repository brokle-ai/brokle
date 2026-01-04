'use client'

import { Download, Edit, Lock, Save, Unlock, Upload, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { cn } from '@/lib/utils'
import { WidgetPalette } from './widget-palette'
import { AutoSaveIndicator } from './auto-save-indicator'
import type { WidgetType } from '../types'
import type { AutoSaveStatus } from '../hooks/use-auto-save'

interface DashboardEditorToolbarProps {
  isEditMode: boolean
  isLocked: boolean
  hasChanges: boolean
  isLockPending?: boolean
  isUnlockPending?: boolean
  isSavePending?: boolean
  autoSaveStatus?: AutoSaveStatus
  autoSaveError?: string | null
  onEditModeToggle: () => void
  onSave: () => void
  onCancel: () => void
  onLock: () => void
  onUnlock: () => void
  onExport: () => void
  onImport: () => void
  onAddWidget: (type: WidgetType, defaultSize: { w: number; h: number }) => void
  className?: string
}

export function DashboardEditorToolbar({
  isEditMode,
  isLocked,
  hasChanges,
  isLockPending,
  isUnlockPending,
  isSavePending,
  autoSaveStatus,
  autoSaveError,
  onEditModeToggle,
  onSave,
  onCancel,
  onLock,
  onUnlock,
  onExport,
  onImport,
  onAddWidget,
  className,
}: DashboardEditorToolbarProps) {
  return (
    <TooltipProvider>
      <div
        className={cn(
          'flex items-center justify-between gap-2 rounded-lg border bg-background p-2',
          className
        )}
      >
        <div className="flex items-center gap-2">
          {isEditMode ? (
            <>
              <WidgetPalette onSelectWidget={onAddWidget} disabled={isLocked} />

              <Separator orientation="vertical" className="h-6" />

              <Button
                variant="default"
                size="sm"
                onClick={onSave}
                disabled={!hasChanges || isSavePending}
                className="gap-1.5"
              >
                <Save className="h-4 w-4" />
                {isSavePending ? 'Saving...' : 'Save'}
              </Button>

              <Button
                variant="outline"
                size="sm"
                onClick={onCancel}
                className="gap-1.5"
              >
                <X className="h-4 w-4" />
                Cancel
              </Button>

              {autoSaveStatus && (
                <>
                  <Separator orientation="vertical" className="h-6" />
                  <AutoSaveIndicator status={autoSaveStatus} error={autoSaveError} />
                </>
              )}
            </>
          ) : (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={onEditModeToggle}
                  disabled={isLocked}
                  className="gap-1.5"
                >
                  <Edit className="h-4 w-4" />
                  Edit Dashboard
                </Button>
              </TooltipTrigger>
              {isLocked && (
                <TooltipContent>
                  <p>Dashboard is locked. Unlock to edit.</p>
                </TooltipContent>
              )}
            </Tooltip>
          )}
        </div>

        <div className="flex items-center gap-2">
          <Separator orientation="vertical" className="h-6" />

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                onClick={isLocked ? onUnlock : onLock}
                disabled={isLockPending || isUnlockPending}
                className={cn(
                  'h-8 w-8',
                  isLocked && 'text-yellow-600 hover:text-yellow-600'
                )}
              >
                {isLocked ? (
                  <Lock className="h-4 w-4" />
                ) : (
                  <Unlock className="h-4 w-4" />
                )}
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>{isLocked ? 'Unlock dashboard' : 'Lock dashboard'}</p>
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                onClick={onExport}
                className="h-8 w-8"
              >
                <Download className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>Export dashboard as JSON</p>
            </TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                onClick={onImport}
                className="h-8 w-8"
              >
                <Upload className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>Import dashboard from JSON</p>
            </TooltipContent>
          </Tooltip>
        </div>
      </div>
    </TooltipProvider>
  )
}
