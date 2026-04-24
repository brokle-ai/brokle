// Domain-friendly re-exports of the wire types from `../api/types`.
//
// The traces feature (and any other cross-feature caller) imports score
// shapes via `@/features/scores/types` and its alias `Score` — mirroring
// the web/ convention. We keep the single wire definition in
// `api/types.ts` and expose it here under the domain names so callers
// don't need to know the file layout.

export type {
  ScoreDataType,
  ScoreDataType as ScoreType,
  ScoreSource,
  ScoreListItem as Score,
  ScoreConfig,
  ScoreListResponse,
  ScoreConfigListResponse,
} from '../api/types'
