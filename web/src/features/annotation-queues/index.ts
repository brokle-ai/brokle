// Annotation Queues Feature
// Human-in-the-Loop (HITL) evaluation workflows

// Main page component (aliased for consistency with other features)
export { QueuesContent as AnnotationQueues } from './components/queues-content'

// Types
export * from './types'

// API
export { annotationQueuesApi } from './api/annotation-queues-api'

// Hooks
export {
  // Query Keys
  annotationQueueQueryKeys,
  // Queue Queries
  useAnnotationQueuesQuery,
  useAnnotationQueueQuery,
  useQueueStatsQuery,
  // Queue Mutations
  useCreateQueueMutation,
  useUpdateQueueMutation,
  useDeleteQueueMutation,
  // Item Queries
  useQueueItemsQuery,
  // Item Mutations
  useAddItemsMutation,
  useAddItemsToQueueMutation,
  useClaimNextItemMutation,
  useCompleteItemMutation,
  useSkipItemMutation,
  useReleaseItemMutation,
  // Assignment Queries
  useQueueAssignmentsQuery,
  // Assignment Mutations
  useAssignUserMutation,
  useUnassignUserMutation,
} from './hooks/use-annotation-queues'

export { useQueuesTableState } from './hooks/use-queues-table-state'
export type { QueueSortField, UseQueuesTableStateReturn } from './hooks/use-queues-table-state'

export {
  useAnnotationKeyboard,
  ANNOTATION_HOTKEYS,
  getSubmitShortcutDisplay,
  isMacPlatform,
} from './hooks/use-annotation-keyboard'

// Context
export {
  AnnotationQueuesProvider,
  useAnnotationQueues,
} from './context/annotation-queues-context'
export type { AnnotationQueuesDialogType } from './context/annotation-queues-context'

// Components
export { QueuesTable } from './components/queues-table'
export { QueuesToolbar } from './components/queues-toolbar'
export { QueueCard } from './components/queue-card'
export { QueueList } from './components/queue-list'
export { QueueForm } from './components/queue-form'
export { QueueDialogs } from './components/queue-dialogs'
export { CreateQueueDialog } from './components/create-queue-dialog'
export { AddItemsForm } from './components/add-items-form'
export { QueuesContent } from './components/queues-content'
export { QueueDetail } from './components/queue-detail'
export { QueueItemTable } from './components/queue-item-table'
export { AnnotationPanel, AnnotationPanelSkeleton } from './components/annotation-panel'
export { ScoreInputForm } from './components/score-input-form'
export { ItemCard } from './components/item-card'
export { AssignmentDialog } from './components/assignment-dialog'
export { SettingsDialog } from './components/settings-dialog'
export { AddItemsDialogStandalone } from './components/add-items-dialog-standalone'
export { StatsCard, QueueStatsCards } from './components/stats-card'
export { AddToQueueButton } from './components/add-to-queue-button'
export { SkipReasonDialog } from './components/skip-reason-dialog'
export { ProgressIndicator, ProgressBadge } from './components/progress-indicator'
export { KeyboardShortcutHint, Kbd } from './components/keyboard-shortcut-hint'
