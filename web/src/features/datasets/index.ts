export { Datasets } from './components/datasets-content'
export { DatasetDetail } from './components/dataset-detail'
export { DatasetDetailLayout } from './components/dataset-detail/dataset-detail-layout'

export { DatasetsProvider, useDatasets } from './context/datasets-context'
export type { DatasetsDialogType } from './context/datasets-context'

export { DatasetDetailProvider, useDatasetDetail } from './context/dataset-detail-context'
export type { DatasetDetailDialogType } from './context/dataset-detail-context'

export type {
  Dataset,
  DatasetItem,
  DatasetItemSource,
  CreateDatasetRequest,
  UpdateDatasetRequest,
  CreateDatasetItemRequest,
  DatasetItemListResponse,
  KeysMapping,
  BulkImportResult,
  ImportFromJsonRequest,
  ImportFromTracesRequest,
  ImportFromSpansRequest,
  // Version types
  DatasetVersion,
  DatasetVersionResponse,
  DatasetWithVersionInfo,
  CreateDatasetVersionRequest,
  PinDatasetVersionRequest,
} from './types'

export { datasetsApi } from './api/datasets-api'

export {
  useDatasetsQuery,
  useDatasetQuery,
  useDatasetItemsQuery,
  useCreateDatasetMutation,
  useUpdateDatasetMutation,
  useDeleteDatasetMutation,
  useCreateDatasetItemMutation,
  useDeleteDatasetItemMutation,
  useImportFromJsonMutation,
  useImportFromTracesMutation,
  useImportFromSpansMutation,
  useExportDatasetQuery,
  datasetQueryKeys,
  // Version hooks
  useDatasetWithVersionInfoQuery,
  useDatasetVersionsQuery,
  useDatasetVersionQuery,
  useDatasetVersionItemsQuery,
  useCreateDatasetVersionMutation,
  usePinDatasetVersionMutation,
  useUnpinDatasetVersionMutation,
} from './hooks/use-datasets'
export { useProjectDatasets } from './hooks/use-project-datasets'
export type { UseProjectDatasetsReturn } from './hooks/use-project-datasets'

export { DatasetList } from './components/dataset-list'
export { DatasetCard } from './components/dataset-card'
export { DatasetForm } from './components/dataset-form'
export { CreateDatasetDialog } from './components/create-dataset-dialog'
export { DatasetItemTable } from './components/dataset-item-table'
export { AddDatasetItemDialog } from './components/add-dataset-item-dialog'
export { DatasetsDialogs } from './components/datasets-dialogs'

export { DatasetDetailDialogs } from './components/dataset-detail-dialogs'
export { DatasetDetailSkeleton } from './components/dataset-detail-skeleton'

// Import Dialogs
export { ImportJsonDialog } from './components/import-json-dialog'
export { ImportCsvDialog } from './components/import-csv-dialog'
export { AddFromTracesDialog, AddTraceToDatasetDialog } from './components/add-from-traces-dialog'

// Version components
export { DatasetVersionBadge } from './components/dataset-version-badge'
export { DatasetVersionManager } from './components/dataset-version-manager'

// Cell components
export {
  AutodetectCell,
  JsonCell,
  TextCell,
  MediaCell,
  RowHeightSelector,
  ROW_HEIGHT_VALUES,
  ROW_HEIGHT_LABELS,
} from './components/cells'
export type { RowHeight, CellProps } from './components/cells'

// Dataset items components
export { ItemPreviewSidebar } from './components/dataset-items/item-preview-sidebar'

// Row height hook
export { useRowHeight } from './hooks/use-row-height'

// CSV import hooks
export { useChunkedCsvImport, useCsvImportMutation } from './hooks/use-csv-import'

// CSV parser utilities
export {
  parseCSV,
  autoDetectColumnMapping,
  chunkCSVContent,
  getColumnTypeLabel,
  getColumnTypeColor,
} from './utils/csv-parser'
export type { ColumnType, ParsedColumn, CsvParseResult } from './utils/csv-parser'
