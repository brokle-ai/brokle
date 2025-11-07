import { useMemo } from 'react'

export enum FeatureFlag {
  REALTIME_ANALYTICS = 'realtime_analytics',
  ADVANCED_ROUTING = 'advanced_routing',
  COST_OPTIMIZATION = 'cost_optimization',
  CUSTOM_MODELS = 'custom_models',
  TEAM_COLLABORATION = 'team_collaboration',
  AUDIT_LOGS = 'audit_logs',
}

export function evaluateFeatureFlag(flag: FeatureFlag): boolean {
  const envValue = process.env[`NEXT_PUBLIC_FF_${flag.toUpperCase()}`]
  return envValue === 'true'
}

export function useFeatureFlagMap(): Record<FeatureFlag, boolean> {
  return useMemo(() => {
    const flags: Partial<Record<FeatureFlag, boolean>> = {}
    Object.values(FeatureFlag).forEach(flag => {
      flags[flag] = evaluateFeatureFlag(flag)
    })
    return flags as Record<FeatureFlag, boolean>
  }, [])
}
