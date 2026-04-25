import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query'
import {
  Bookmark,
  Check,
  Globe,
  Lock,
  MoreHorizontal,
  Pencil,
  Plus,
  Trash2,
} from 'lucide-react'
import { toast } from 'sonner'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
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
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Separator } from '@/components/ui/separator'
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import { useLocalStorage } from '@/hooks/use-local-storage'
import { BrokleError } from '@/lib/api/errors'
import type { FilterCondition } from '@/components/shared/filter-builder'
import {
  createFilterPreset,
  deleteFilterPreset,
  filterPresetsKeys,
  filterPresetsQueryOptions,
  updateFilterPreset,
  type CreateFilterPresetRequest,
  type FilterPreset,
} from '../../api/queries'

interface FilterPresetsDrawerProps {
  projectId: string
  currentFilters: FilterCondition[]
  currentSearchQuery?: string | null
  onApplyPreset: (preset: FilterPreset) => void
  tableName: 'traces' | 'spans'
}

/**
 * Synthesise a stable id for a localStorage-backed preset. Backend
 * presets use UUIDs, so we prefix to avoid an accidental collision
 * when the same drawer flips between local and remote modes.
 */
function localPresetId(): string {
  return `local-${crypto.randomUUID()}`
}

function buildLocalPreset(
  projectId: string,
  data: CreateFilterPresetRequest,
): FilterPreset {
  const now = new Date().toISOString()
  return {
    id: localPresetId(),
    project_id: projectId,
    name: data.name,
    description: data.description,
    table_name: data.table_name,
    filters: data.filters,
    search_query: data.search_query,
    search_types: data.search_types,
    is_public: data.is_public ?? false,
    created_at: now,
    updated_at: now,
  }
}

/**
 * FilterPresetsDrawer — manages saved filter configurations for the
 * traces / spans table. Backed by `/api/v1/projects/{id}/filter-presets`
 * (Huma `filter-presets` ops). When the backend endpoint returns a
 * non-2xx (older deployment, RBAC denial, etc.), the drawer falls
 * back to a localStorage cache keyed by project + table so users still
 * get persistence on a single browser. Local-only presets are tagged
 * with a `local-` id prefix and surfaced with the same UX.
 */
export function FilterPresetsDrawer({
  projectId,
  currentFilters,
  currentSearchQuery,
  onApplyPreset,
  tableName,
}: FilterPresetsDrawerProps) {
  const queryClient = useQueryClient()
  const [isOpen, setIsOpen] = useState(false)
  const [isSaveDialogOpen, setIsSaveDialogOpen] = useState(false)
  const [editingPreset, setEditingPreset] = useState<FilterPreset | null>(null)
  const [presetToDelete, setPresetToDelete] = useState<FilterPreset | null>(
    null,
  )

  const [name, setName] = useState('')
  const [description, setDescription] = useState('')
  const [isPublic, setIsPublic] = useState(false)

  const localStorageKey = `traces.filter-presets.${projectId}.${tableName}`
  const [localPresets, setLocalPresets] = useLocalStorage<FilterPreset[]>(
    localStorageKey,
    [],
  )

  const remoteQuery = useQuery({
    ...filterPresetsQueryOptions(projectId, tableName),
    enabled: isOpen && projectId.length > 0,
  })

  // Decide which source feeds the list. Default = remote; on a 4xx /
  // network failure we silently degrade to localStorage so the drawer
  // remains usable. 5xx still surfaces an error toast on save.
  const usingLocalFallback = useMemo(() => {
    if (!remoteQuery.error) return false
    if (remoteQuery.error instanceof BrokleError) {
      return remoteQuery.error.status >= 400 && remoteQuery.error.status < 500
    }
    return true
  }, [remoteQuery.error])

  const presets = usingLocalFallback
    ? localPresets
    : remoteQuery.data ?? []
  const isLoading = !usingLocalFallback && remoteQuery.isLoading

  const createMutation = useMutation({
    mutationFn: async (data: CreateFilterPresetRequest) => {
      if (usingLocalFallback) return buildLocalPreset(projectId, data)
      return createFilterPreset(projectId, data)
    },
    onSuccess: (preset) => {
      if (usingLocalFallback) {
        setLocalPresets([preset, ...localPresets])
      } else {
        queryClient.invalidateQueries({
          queryKey: filterPresetsKeys.list(projectId, tableName),
        })
      }
      toast.success('Filter preset saved')
      handleCloseSaveDialog()
    },
    onError: () => {
      toast.error('Failed to save preset')
    },
  })

  const updateMutation = useMutation({
    mutationFn: async ({
      preset,
      data,
    }: {
      preset: FilterPreset
      data: Partial<CreateFilterPresetRequest>
    }) => {
      if (preset.id.startsWith('local-')) {
        return {
          ...preset,
          ...data,
          updated_at: new Date().toISOString(),
        } as FilterPreset
      }
      return updateFilterPreset(projectId, preset.id, data)
    },
    onSuccess: (preset, variables) => {
      if (variables.preset.id.startsWith('local-')) {
        setLocalPresets(
          localPresets.map((p) => (p.id === preset.id ? preset : p)),
        )
      } else {
        queryClient.invalidateQueries({
          queryKey: filterPresetsKeys.list(projectId, tableName),
        })
      }
      toast.success('Filter preset updated')
      handleCloseSaveDialog()
    },
    onError: () => {
      toast.error('Failed to update preset')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: async (preset: FilterPreset) => {
      if (preset.id.startsWith('local-')) return preset
      await deleteFilterPreset(projectId, preset.id)
      return preset
    },
    onSuccess: (preset) => {
      if (preset.id.startsWith('local-')) {
        setLocalPresets(localPresets.filter((p) => p.id !== preset.id))
      } else {
        queryClient.invalidateQueries({
          queryKey: filterPresetsKeys.list(projectId, tableName),
        })
      }
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
    const payload: CreateFilterPresetRequest = {
      name: name.trim(),
      description: description.trim() || undefined,
      table_name: tableName,
      filters: currentFilters,
      search_query: currentSearchQuery || undefined,
      is_public: isPublic,
    }
    if (editingPreset) {
      updateMutation.mutate({ preset: editingPreset, data: payload })
    } else {
      createMutation.mutate(payload)
    }
  }, [
    name,
    description,
    isPublic,
    editingPreset,
    currentFilters,
    currentSearchQuery,
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
    [onApplyPreset],
  )

  // Surface a one-time toast when degrading to local storage so users
  // know writes won't sync across devices.
  useEffect(() => {
    if (isOpen && usingLocalFallback) {
      toast.info('Filter presets are saved locally', {
        description:
          'Backend filter-preset API not reachable; falling back to this browser only.',
        id: 'filter-presets-local-fallback',
      })
    }
  }, [isOpen, usingLocalFallback])

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
              Save and manage filter configurations for quick access.
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
                  Loading presets…
                </div>
              ) : presets.length === 0 ? (
                <div className="flex flex-col items-center justify-center py-8 text-center text-muted-foreground">
                  <Bookmark className="mb-2 h-8 w-8 opacity-50" />
                  <p className="text-sm">No saved presets</p>
                  <p className="text-xs">
                    Apply some filters and save them as a preset.
                  </p>
                </div>
              ) : (
                <div className="space-y-2">
                  {presets.map((preset) => (
                    <div
                      key={preset.id}
                      className="group flex items-center justify-between rounded-lg border p-3 hover:bg-muted/50"
                    >
                      <button
                        type="button"
                        className="flex-1 text-left"
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
                          <p className="mt-1 line-clamp-1 text-xs text-muted-foreground">
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
                      </button>

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
                ? 'Update this filter preset.'
                : 'Save your current filters for quick access later.'}
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
                placeholder="Describe what this preset filters…"
                rows={2}
              />
            </div>

            <div className="flex items-center justify-between rounded-lg border p-4">
              <div className="space-y-0.5">
                <Label htmlFor="preset-public">Share with team</Label>
                <p className="text-xs text-muted-foreground">
                  Make this preset visible to all project members.
                </p>
              </div>
              <Switch
                id="preset-public"
                checked={isPublic}
                onCheckedChange={setIsPublic}
                disabled={usingLocalFallback}
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
              {isSaving ? 'Saving…' : editingPreset ? 'Update' : 'Save'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog
        open={presetToDelete !== null}
        onOpenChange={(open) => {
          if (!open) setPresetToDelete(null)
        }}
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
                presetToDelete && deleteMutation.mutate(presetToDelete)
              }
              disabled={deleteMutation.isPending}
            >
              {deleteMutation.isPending ? 'Deleting…' : 'Delete'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
