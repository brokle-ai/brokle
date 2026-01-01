'use client'

import { useState, useCallback } from 'react'
import { useParams } from 'next/navigation'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  Bookmark,
  Check,
  Globe,
  Lock,
  MoreHorizontal,
  Pencil,
  Plus,
  Trash2,
  User,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { Badge } from '@/components/ui/badge'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import { toast } from 'sonner'
import {
  getFilterPresets,
  createFilterPreset,
  updateFilterPreset,
  deleteFilterPreset,
  type FilterPreset,
  type FilterCondition,
  type CreateFilterPresetRequest,
} from '../../api/traces-api'

interface FilterPresetsDrawerProps {
  currentFilters: FilterCondition[]
  currentSearchQuery?: string
  currentSearchTypes?: string[]
  onApplyPreset: (preset: FilterPreset) => void
  tableName: 'traces' | 'spans'
}

export function FilterPresetsDrawer({
  currentFilters,
  currentSearchQuery,
  currentSearchTypes,
  onApplyPreset,
  tableName,
}: FilterPresetsDrawerProps) {
  const params = useParams<{ projectId: string }>()
  const projectId = params.projectId
  const queryClient = useQueryClient()

  const [isOpen, setIsOpen] = useState(false)
  const [isSaveDialogOpen, setIsSaveDialogOpen] = useState(false)
  const [editingPreset, setEditingPreset] = useState<FilterPreset | null>(null)
  const [presetToDelete, setPresetToDelete] = useState<FilterPreset | null>(
    null
  )

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [isPublic, setIsPublic] = useState(false)

  const { data: presets = [], isLoading } = useQuery({
    queryKey: ['filterPresets', projectId, tableName],
    queryFn: () => getFilterPresets(projectId, tableName),
    enabled: isOpen,
  })

  const createMutation = useMutation({
    mutationFn: (data: CreateFilterPresetRequest) =>
      createFilterPreset(projectId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['filterPresets', projectId],
      })
      toast.success('Filter preset saved')
      handleCloseSaveDialog()
    },
    onError: () => {
      toast.error('Failed to save preset')
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({
      presetId,
      data,
    }: {
      presetId: string
      data: Partial<CreateFilterPresetRequest>
    }) => updateFilterPreset(projectId, presetId, data),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['filterPresets', projectId],
      })
      toast.success('Filter preset updated')
      handleCloseSaveDialog()
    },
    onError: () => {
      toast.error('Failed to update preset')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (presetId: string) => deleteFilterPreset(projectId, presetId),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: ['filterPresets', projectId],
      })
      toast.success('Filter preset deleted')
      setPresetToDelete(null)
    },
    onError: () => {
      toast.error('Failed to delete preset')
    },
  })

  const handleOpenSaveDialog = useCallback(() => {
    setEditingPreset(null)
    setName('')
    setDescription('')
    setIsPublic(false)
    setIsSaveDialogOpen(true)
  }, [])

  const handleOpenEditDialog = useCallback((preset: FilterPreset) => {
    setEditingPreset(preset)
    setName(preset.name)
    setDescription(preset.description || '')
    setIsPublic(preset.is_public)
    setIsSaveDialogOpen(true)
  }, [])

  const handleCloseSaveDialog = useCallback(() => {
    setIsSaveDialogOpen(false)
    setEditingPreset(null)
    setName('')
    setDescription('')
    setIsPublic(false)
  }, [])

  const handleSave = useCallback(() => {
    if (!name.trim()) {
      toast.error('Please enter a name for the preset')
      return
    }

    if (editingPreset) {
      updateMutation.mutate({
        presetId: editingPreset.id,
        data: {
          name: name.trim(),
          description: description.trim() || undefined,
          is_public: isPublic,
          filters: currentFilters,
          search_query: currentSearchQuery || undefined,
          search_types: currentSearchTypes,
        },
      })
    } else {
      createMutation.mutate({
        name: name.trim(),
        description: description.trim() || undefined,
        table_name: tableName,
        filters: currentFilters,
        search_query: currentSearchQuery || undefined,
        search_types: currentSearchTypes,
        is_public: isPublic,
      })
    }
  }, [
    name,
    description,
    isPublic,
    editingPreset,
    currentFilters,
    currentSearchQuery,
    currentSearchTypes,
    tableName,
    createMutation,
    updateMutation,
  ])

  const handleApplyPreset = useCallback(
    (preset: FilterPreset) => {
      onApplyPreset(preset)
      setIsOpen(false)
      toast.success(`Applied preset "${preset.name}"`)
    },
    [onApplyPreset]
  )

  const isSaving = createMutation.isPending || updateMutation.isPending

  return (
    <>
      <Sheet open={isOpen} onOpenChange={setIsOpen}>
        <SheetTrigger asChild>
          <Button variant="outline" size="sm" className="h-8">
            <Bookmark className="mr-2 h-4 w-4" />
            Presets
          </Button>
        </SheetTrigger>
        <SheetContent className="w-[400px] sm:w-[540px]">
          <SheetHeader>
            <SheetTitle>Filter Presets</SheetTitle>
            <SheetDescription>
              Save and manage your filter configurations for quick access
            </SheetDescription>
          </SheetHeader>

          <div className="mt-6 space-y-4">
            {currentFilters.length > 0 && (
              <>
                <Button
                  variant="outline"
                  className="w-full justify-start"
                  onClick={handleOpenSaveDialog}
                >
                  <Plus className="mr-2 h-4 w-4" />
                  Save current filters as preset
                </Button>
                <Separator />
              </>
            )}

            <ScrollArea className="h-[calc(100vh-280px)]">
              {isLoading ? (
                <div className="flex items-center justify-center py-8 text-muted-foreground">
                  Loading presets...
                </div>
              ) : presets.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
                  <Bookmark className="h-8 w-8 mb-2 opacity-50" />
                  <p className="text-sm">No saved presets</p>
                  <p className="text-xs">
                    Apply some filters and save them as a preset
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  {presets.map((preset) => (
                    <div
                      key={preset.id}
                      className="group flex items-center justify-between rounded-lg border p-3 hover:bg-muted/50"
                    >
                      <div
                        className="flex-1 cursor-pointer"
                        onClick={() => handleApplyPreset(preset)}
                      >
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{preset.name}</span>
                          {preset.is_public ? (
                            <Badge variant="secondary" className="text-xs">
                              <Globe className="mr-1 h-3 w-3" />
                              Public
                            </Badge>
                          ) : (
                            <Badge variant="outline" className="text-xs">
                              <Lock className="mr-1 h-3 w-3" />
                              Private
                            </Badge>
                          )}
                        </div>
                        {preset.description && (
                          <p className="mt-1 text-xs text-muted-foreground line-clamp-1">
                            {preset.description}
                          </p>
                        )}
                        <div className="mt-1 flex items-center gap-2 text-xs text-muted-foreground">
                          <span>
                            {preset.filters?.length || 0} filter
                            {(preset.filters?.length || 0) !== 1 ? 's' : ''}
                          </span>
                          {preset.search_query && (
                            <>
                              <span>•</span>
                              <span>Has search</span>
                            </>
                          )}
                        </div>
                      </div>

                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 opacity-0 group-hover:opacity-100"
                          >
                            <MoreHorizontal className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => handleApplyPreset(preset)}
                          >
                            <Check className="mr-2 h-4 w-4" />
                            Apply
                          </DropdownMenuItem>
                          <DropdownMenuItem
                            onClick={() => handleOpenEditDialog(preset)}
                          >
                            <Pencil className="mr-2 h-4 w-4" />
                            Edit
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem
                            className="text-destructive"
                            onClick={() => setPresetToDelete(preset)}
                          >
                            <Trash2 className="mr-2 h-4 w-4" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>
          </div>
        </SheetContent>
      </Sheet>

      <Dialog open={isSaveDialogOpen} onOpenChange={setIsSaveDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingPreset ? 'Edit Preset' : 'Save Filter Preset'}
            </DialogTitle>
            <DialogDescription>
              {editingPreset
                ? 'Update this filter preset'
                : 'Save your current filters for quick access later'}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="preset-name">Name</Label>
              <Input
                id="preset-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="My filter preset"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="preset-description">Description (optional)</Label>
              <Textarea
                id="preset-description"
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                placeholder="Describe what this preset filters..."
                rows={2}
              />
            </div>

            <div className="flex items-center justify-between rounded-lg border p-4">
              <div className="space-y-0.5">
                <Label htmlFor="preset-public">Share with team</Label>
                <p className="text-xs text-muted-foreground">
                  Make this preset visible to all project members
                </p>
              </div>
              <Switch
                id="preset-public"
                checked={isPublic}
                onCheckedChange={setIsPublic}
              />
            </div>

            <div className="rounded-lg bg-muted p-3">
              <p className="text-xs text-muted-foreground">
                <strong>Saving:</strong> {currentFilters.length} filter
                {currentFilters.length !== 1 ? 's' : ''}
                {currentSearchQuery && ', 1 search query'}
              </p>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={handleCloseSaveDialog}>
              Cancel
            </Button>
            <Button onClick={handleSave} disabled={isSaving}>
              {isSaving ? 'Saving...' : editingPreset ? 'Update' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={!!presetToDelete}
        onOpenChange={() => setPresetToDelete(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Preset</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete &quot;{presetToDelete?.name}
              &quot;? This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPresetToDelete(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() =>
                presetToDelete && deleteMutation.mutate(presetToDelete.id)
              }
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
