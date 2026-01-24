export { Scores } from './components/scores-content'

export { ScoresProvider, useScores } from './context/scores-context'

export type {
  ScoreConfig,
  Score,
  ScoreDataType,
  ScoreSource,
  ScoreListParams,
  CreateScoreConfigRequest,
  UpdateScoreConfigRequest,
  ScoreAnalyticsParams,
  ScoreAnalyticsData,
  ScoreStatistics,
  TimeSeriesPoint,
  DistributionBin,
  HeatmapCell,
  ComparisonMetrics,
  InterpretationResult,
} from './types'

export { scoresApi } from './api/scores-api'

export {
  useScoreConfigsQuery,
  useScoreConfigQuery,
  useScoreConfigsByIdsQuery,
  useCreateScoreConfigMutation,
  useUpdateScoreConfigMutation,
  useDeleteScoreConfigMutation,
  scoreConfigQueryKeys,
} from './hooks/use-score-configs'

export {
  useScoresQuery,
  useScoreQuery,
  scoreQueryKeys,
} from './hooks/use-scores'

export {
  useScoreAnalyticsQuery,
  useScoreNamesQuery,
  scoreAnalyticsQueryKeys,
} from './hooks/use-score-analytics'

export {
  useScoresTableState,
  type UseScoresTableStateReturn,
  type ScoreSortField,
} from './hooks/use-scores-table-state'

export { ScoreBadge } from './components/score-badge'
export { ScoreTag, ScoreTagList } from './components/score-tag'
export { ScoreList } from './components/score-list'
export { ScoreConfigForm } from './components/score-config-form'
export { ScoreConfigsSection } from './components/score-configs-section'
export {
  ScoreValueCell,
  formatValue as formatScoreValue,
  getValueVariant as getScoreValueVariant,
  getSourceStyles as getScoreSourceStyles,
  getSourceLabel as getScoreSourceLabel,
} from './components/score-value-cell'
export { ScoresTable } from './components/scores-table'
export { ScoresToolbar } from './components/scores-toolbar'
export { ScoresPageContent } from './components/scores-page-content'
export {
  ScoreInputField,
  NumericInput,
  CategoricalInput,
  BooleanInput,
} from './components/annotation/score-input-field'
export { ReasonEditor, ReasonDisplay } from './components/annotation/reason-editor'
export { AnnotationFormDialog } from './components/annotation/annotation-form-dialog'
export {
  ScoreAnalyticsDashboard,
  StatisticsCard,
  TimelineChartCard,
  DistributionCard,
  HeatmapCard,
  Heatmap,
} from './components/analytics'

export {
  interpretCorrelation,
  interpretCohensKappa,
  interpretMAE,
  interpretRMSE,
  interpretAgreement,
  formatNumber,
  formatPercent,
  getColorClass,
  getBadgeColor,
} from './lib/statistics-utils'

export {
  SCORE_COLORS,
  CHART_COLORS,
  oklchToCss,
  getHeatmapCellColor,
  getCorrelationColor,
  getChartColor,
  getScoreValueColor,
  getScoreColorClass,
} from './lib/color-scales'

export {
  getScoreTagColor,
  getScoreTagClasses,
  getDataTypeIndicator,
  getSourceIndicator,
} from './lib/score-colors'
