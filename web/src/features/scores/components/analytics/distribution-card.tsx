'use client'

import { useMemo } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Cell,
} from 'recharts'
import type { DistributionBin } from '../../types'
import { CHART_COLORS } from '../../lib/color-scales'

interface DistributionCardProps {
  distribution: DistributionBin[]
  compareDistribution?: DistributionBin[]
  scoreName: string
  compareScoreName?: string
}

interface ChartDataPoint {
  label: string
  value: number
  percentage: number
  binStart: number
  binEnd: number
}

function formatBinLabel(binStart: number, binEnd: number): string {
  if (binStart === binEnd) {
    return binStart.toFixed(2)
  }
  return `${binStart.toFixed(2)} - ${binEnd.toFixed(2)}`
}

function processDistribution(bins: DistributionBin[]): ChartDataPoint[] {
  const total = bins.reduce((sum, bin) => sum + bin.count, 0)

  return bins.map((bin) => ({
    label: formatBinLabel(bin.bin_start, bin.bin_end),
    value: bin.count,
    percentage: total > 0 ? (bin.count / total) * 100 : 0,
    binStart: bin.bin_start,
    binEnd: bin.bin_end,
  }))
}

function DistributionChart({
  data,
  color,
  scoreName,
}: {
  data: ChartDataPoint[]
  color: string
  scoreName: string
}) {
  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center h-[250px] text-muted-foreground">
        No distribution data available
      </div>
    )
  }

  return (
    <div className="h-[250px]">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={data}
          margin={{ top: 10, right: 20, left: 10, bottom: 20 }}
        >
          <CartesianGrid strokeDasharray="3 3" className="stroke-muted" />
          <XAxis
            dataKey="label"
            tick={{ fontSize: 10 }}
            tickLine={false}
            axisLine={false}
            angle={-45}
            textAnchor="end"
            height={60}
            className="text-muted-foreground"
          />
          <YAxis
            tick={{ fontSize: 12 }}
            tickLine={false}
            axisLine={false}
            className="text-muted-foreground"
          />
          <Tooltip
            content={({ active, payload }) => {
              if (!active || !payload?.length) return null
              const data = payload[0].payload as ChartDataPoint
              return (
                <div className="bg-popover border rounded-lg shadow-lg p-3">
                  <p className="text-sm font-medium mb-1">{data.label}</p>
                  <p className="text-sm text-muted-foreground">
                    Count: <span className="font-medium">{data.value.toLocaleString()}</span>
                  </p>
                  <p className="text-sm text-muted-foreground">
                    Percentage: <span className="font-medium">{data.percentage.toFixed(1)}%</span>
                  </p>
                </div>
              )
            }}
          />
          <Bar
            dataKey="value"
            name={scoreName}
            radius={[4, 4, 0, 0]}
          >
            {data.map((_, index) => (
              <Cell key={index} fill={color} />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}

export function DistributionCard({
  distribution,
  compareDistribution,
  scoreName,
  compareScoreName,
}: DistributionCardProps) {
  const primaryData = useMemo(
    () => processDistribution(distribution),
    [distribution]
  )

  const compareData = useMemo(
    () => (compareDistribution ? processDistribution(compareDistribution) : null),
    [compareDistribution]
  )

  const hasComparison = compareData && compareScoreName

  if (!hasComparison) {
    return (
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-base font-medium">Score Distribution</CardTitle>
        </CardHeader>
        <CardContent>
          <DistributionChart
            data={primaryData}
            color={CHART_COLORS.series1}
            scoreName={scoreName}
          />
        </CardContent>
      </Card>
    )
  }

  return (
    <Card>
      <CardHeader className="pb-3">
        <CardTitle className="text-base font-medium">Score Distribution</CardTitle>
      </CardHeader>
      <CardContent>
        <Tabs defaultValue="primary" className="w-full">
          <TabsList className="mb-4">
            <TabsTrigger value="primary">{scoreName}</TabsTrigger>
            <TabsTrigger value="compare">{compareScoreName}</TabsTrigger>
          </TabsList>
          <TabsContent value="primary">
            <DistributionChart
              data={primaryData}
              color={CHART_COLORS.series1}
              scoreName={scoreName}
            />
          </TabsContent>
          <TabsContent value="compare">
            <DistributionChart
              data={compareData}
              color={CHART_COLORS.series2}
              scoreName={compareScoreName}
            />
          </TabsContent>
        </Tabs>
      </CardContent>
    </Card>
  )
}
