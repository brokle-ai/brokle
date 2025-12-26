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

export { ScoreBadge } from './components/score-badge'
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
export { ScoresPageContent } from './components/scores-page-content'
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
