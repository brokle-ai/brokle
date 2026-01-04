/**
 * Drilldown Utilities Tests
 *
 * Tests for the dashboard drilldown functionality that enables
 * clicking on chart data points to navigate to filtered traces.
 */

import { describe, it, expect, vi } from 'vitest'
import {
  buildFilterFromDataPoint,
  buildFiltersFromDataPoint,
  encodeFiltersForUrl,
  buildDrilldownUrl,
  createDrilldownHandler,
  type DrilldownFilter,
} from '../utils/drilldown'
import type { WidgetQuery } from '../types'

describe('drilldown utilities', () => {
  describe('buildFilterFromDataPoint', () => {
    it('maps known dimension names to column names', () => {
      const filter = buildFilterFromDataPoint('model_name', 'gpt-4')

      expect(filter.column).toBe('model')
      expect(filter.operator).toBe('=')
      expect(filter.value).toBe('gpt-4')
      expect(filter.id).toMatch(/^drilldown_/)
    })

    it('maps time dimension to started_at', () => {
      const filter = buildFilterFromDataPoint('time', '2024-01-15')

      expect(filter.column).toBe('started_at')
      expect(filter.value).toBe('2024-01-15')
    })

    it('maps provider_name to provider', () => {
      const filter = buildFilterFromDataPoint('provider_name', 'openai')

      expect(filter.column).toBe('provider')
      expect(filter.value).toBe('openai')
    })

    it('passes through unknown dimension names', () => {
      const filter = buildFilterFromDataPoint('custom_field', 'value')

      expect(filter.column).toBe('custom_field')
      expect(filter.value).toBe('value')
    })

    it('handles numeric values', () => {
      const filter = buildFilterFromDataPoint('count', 42)

      expect(filter.value).toBe(42)
    })

    it('converts string numbers to strings', () => {
      const filter = buildFilterFromDataPoint('status', 'error')

      expect(filter.value).toBe('error')
      expect(typeof filter.value).toBe('string')
    })

    it('generates unique IDs for each filter', () => {
      const filter1 = buildFilterFromDataPoint('model', 'gpt-4')
      const filter2 = buildFilterFromDataPoint('model', 'gpt-4')

      expect(filter1.id).not.toBe(filter2.id)
    })
  })

  describe('buildFiltersFromDataPoint', () => {
    it('extracts single dimension from name field', () => {
      const data = { name: 'openai', value: 100 }
      const query: WidgetQuery = {
        view: 'traces',
        dimensions: ['provider'],
        measures: ['count'],
      }

      const filters = buildFiltersFromDataPoint(data, query)

      expect(filters).toHaveLength(1)
      expect(filters[0].column).toBe('provider')
      expect(filters[0].value).toBe('openai')
    })

    it('extracts multiple dimensions from data fields', () => {
      const data = { provider: 'openai', model: 'gpt-4', value: 50 }
      const query: WidgetQuery = {
        view: 'traces',
        dimensions: ['provider', 'model'],
        measures: ['count'],
      }

      const filters = buildFiltersFromDataPoint(data, query)

      expect(filters).toHaveLength(2)
      expect(filters.find((f) => f.column === 'provider')?.value).toBe('openai')
      expect(filters.find((f) => f.column === 'model')?.value).toBe('gpt-4')
    })

    it('ignores missing dimension values', () => {
      const data = { provider: 'openai' }
      const query: WidgetQuery = {
        view: 'traces',
        dimensions: ['provider', 'model'],
        measures: ['count'],
      }

      const filters = buildFiltersFromDataPoint(data, query)

      expect(filters).toHaveLength(1)
      expect(filters[0].column).toBe('provider')
    })

    it('ignores null and empty values', () => {
      const data = { provider: null, model: '', status: 'ok' }
      const query: WidgetQuery = {
        view: 'traces',
        dimensions: ['provider', 'model', 'status'],
        measures: ['count'],
      }

      const filters = buildFiltersFromDataPoint(data, query)

      expect(filters).toHaveLength(1)
      expect(filters[0].column).toBe('status')
    })

    it('returns empty array when no dimensions configured', () => {
      const data = { name: 'test', value: 100 }
      const query: WidgetQuery = {
        view: 'traces',
        measures: ['count'],
      }

      const filters = buildFiltersFromDataPoint(data, query)

      expect(filters).toHaveLength(0)
    })

    it('handles time series data with timestamp', () => {
      const data = { timestamp: '2024-01-15T10:00:00Z', value: 42 }
      const query: WidgetQuery = {
        view: 'traces',
        dimensions: ['timestamp'],
        measures: ['count'],
      }

      const filters = buildFiltersFromDataPoint(data, query)

      expect(filters).toHaveLength(1)
      expect(filters[0].column).toBe('started_at')
      expect(filters[0].value).toBe('2024-01-15T10:00:00Z')
    })
  })

  describe('encodeFiltersForUrl', () => {
    it('returns empty string for no filters', () => {
      expect(encodeFiltersForUrl([])).toBe('')
    })

    it('encodes filters as JSON', () => {
      const filters: DrilldownFilter[] = [
        { id: 'test_1', column: 'model', operator: '=', value: 'gpt-4' },
      ]

      const encoded = encodeFiltersForUrl(filters)

      expect(encoded).toBeTruthy()
      expect(decodeURIComponent(encoded)).toBe(JSON.stringify(filters))
    })

    it('handles multiple filters', () => {
      const filters: DrilldownFilter[] = [
        { id: 'test_1', column: 'model', operator: '=', value: 'gpt-4' },
        { id: 'test_2', column: 'provider', operator: '=', value: 'openai' },
      ]

      const encoded = encodeFiltersForUrl(filters)
      const decoded = JSON.parse(decodeURIComponent(encoded))

      expect(decoded).toHaveLength(2)
    })
  })

  describe('buildDrilldownUrl', () => {
    it('builds basic URL without filters', () => {
      const url = buildDrilldownUrl('my-project', [])

      expect(url).toBe('/projects/my-project/traces')
    })

    it('includes filters in URL', () => {
      const filters: DrilldownFilter[] = [
        { id: 'test_1', column: 'model', operator: '=', value: 'gpt-4' },
      ]

      const url = buildDrilldownUrl('my-project', filters)

      expect(url).toContain('/projects/my-project/traces?')
      expect(url).toContain('filters=')
    })

    it('includes relative time range', () => {
      const url = buildDrilldownUrl('my-project', [], { relative: '24h' })

      expect(url).toContain('time_rel=24h')
    })

    it('includes absolute time range', () => {
      const url = buildDrilldownUrl('my-project', [], {
        from: '2024-01-01T00:00:00Z',
        to: '2024-01-02T00:00:00Z',
        relative: 'custom',
      })

      expect(url).toContain('time_from=2024-01-01T00%3A00%3A00Z')
      expect(url).toContain('time_to=2024-01-02T00%3A00%3A00Z')
    })

    it('prefers relative time over absolute when not custom', () => {
      const url = buildDrilldownUrl('my-project', [], {
        from: '2024-01-01T00:00:00Z',
        to: '2024-01-02T00:00:00Z',
        relative: '7d',
      })

      expect(url).toContain('time_rel=7d')
      expect(url).not.toContain('time_from')
    })

    it('combines filters and time range', () => {
      const filters: DrilldownFilter[] = [
        { id: 'test_1', column: 'model', operator: '=', value: 'gpt-4' },
      ]

      const url = buildDrilldownUrl('my-project', filters, { relative: '24h' })

      expect(url).toContain('filters=')
      expect(url).toContain('time_rel=24h')
    })
  })

  describe('createDrilldownHandler', () => {
    it('creates a handler that navigates on click', () => {
      const navigate = vi.fn()
      const context = {
        projectSlug: 'my-project',
        query: {
          view: 'traces' as const,
          dimensions: ['provider'],
          measures: ['count'],
        },
      }

      const handler = createDrilldownHandler(context, navigate)
      handler({ name: 'openai', value: 100 })

      expect(navigate).toHaveBeenCalledTimes(1)
      expect(navigate).toHaveBeenCalledWith(expect.stringContaining('/projects/my-project/traces'))
    })

    it('includes filters from clicked data', () => {
      const navigate = vi.fn()
      const context = {
        projectSlug: 'my-project',
        query: {
          view: 'traces' as const,
          dimensions: ['provider'],
          measures: ['count'],
        },
      }

      const handler = createDrilldownHandler(context, navigate)
      handler({ name: 'openai', value: 100 })

      const navigatedUrl = navigate.mock.calls[0][0] as string
      expect(navigatedUrl).toContain('filters=')

      // Decode and check the filter content
      const url = new URL(`http://localhost${navigatedUrl}`)
      const filtersJson = url.searchParams.get('filters')
      expect(filtersJson).toBeTruthy()

      const filters = JSON.parse(filtersJson!) as DrilldownFilter[]
      expect(filters).toHaveLength(1)
      expect(filters[0].column).toBe('provider')
      expect(filters[0].value).toBe('openai')
    })

    it('preserves time range in navigation', () => {
      const navigate = vi.fn()
      const context = {
        projectSlug: 'my-project',
        query: {
          view: 'traces' as const,
          dimensions: ['provider'],
          measures: ['count'],
        },
        timeRange: { relative: '7d' },
      }

      const handler = createDrilldownHandler(context, navigate)
      handler({ name: 'openai', value: 100 })

      expect(navigate).toHaveBeenCalledWith(expect.stringContaining('time_rel=7d'))
    })
  })
})
