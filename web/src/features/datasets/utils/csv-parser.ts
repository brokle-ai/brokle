/**
 * CSV Parser Utility
 * Client-side CSV parsing with column type inference
 * Based on Langfuse patterns for robust CSV handling
 */

export type ColumnType = 'string' | 'number' | 'boolean' | 'json' | 'array' | 'null' | 'mixed'

// Column detection patterns for auto-mapping
export const COLUMN_DETECTION_PATTERNS = {
  input: ['input', 'prompt', 'question', 'query', 'text', 'message', 'content'],
  expected: ['expected', 'output', 'answer', 'response', 'completion', 'result', 'ground_truth', 'label'],
  metadata: ['metadata', 'meta', 'tags', 'category', 'type'],
} as const

// Type detection constants
const BOOLEAN_TRUE = 'true'
const BOOLEAN_FALSE = 'false'
const NUMBER_PATTERN = /^-?\d+\.?\d*$|^-?\d*\.\d+$/

// Column type configuration for labels and colors
export const COLUMN_TYPE_CONFIG: Record<ColumnType, { label: string; color: string }> = {
  string: { label: 'Text', color: 'bg-green-100 text-green-800' },
  number: { label: 'Number', color: 'bg-blue-100 text-blue-800' },
  boolean: { label: 'Boolean', color: 'bg-purple-100 text-purple-800' },
  json: { label: 'JSON Object', color: 'bg-orange-100 text-orange-800' },
  array: { label: 'Array', color: 'bg-orange-100 text-orange-800' },
  null: { label: 'Empty', color: 'bg-gray-100 text-gray-800' },
  mixed: { label: 'Mixed', color: 'bg-yellow-100 text-yellow-800' },
}

export interface ParsedColumn {
  name: string
  type: ColumnType
  sampleValues: string[]
  nullCount: number
  uniqueCount: number
}

export interface CsvParseResult {
  headers: string[]
  rows: string[][]
  columns: ParsedColumn[]
  rowCount: number
  estimatedSize: number
}

/**
 * Parse CSV content with proper handling of:
 * - Quoted fields with commas
 * - Escaped quotes (doubled)
 * - Multiline fields within quotes
 * - Various line endings (CR, LF, CRLF)
 */
export function parseCSV(content: string, hasHeader: boolean): CsvParseResult | null {
  const lines = content.split(/\r?\n/)
  if (lines.length === 0) return null

  const rows: string[][] = []
  let currentRow: string[] = []
  let inQuotes = false
  let currentField = ''

  for (const line of lines) {
    for (let i = 0; i < line.length; i++) {
      const char = line[i]
      const nextChar = line[i + 1]

      if (char === '"' && !inQuotes) {
        inQuotes = true
      } else if (char === '"' && inQuotes) {
        if (nextChar === '"') {
          // Escaped quote
          currentField += '"'
          i++
        } else {
          inQuotes = false
        }
      } else if (char === ',' && !inQuotes) {
        currentRow.push(currentField.trim())
        currentField = ''
      } else {
        currentField += char
      }
    }

    if (!inQuotes) {
      currentRow.push(currentField.trim())
      if (currentRow.some(cell => cell !== '')) {
        rows.push(currentRow)
      }
      currentRow = []
      currentField = ''
    } else {
      // Multiline field - preserve newline
      currentField += '\n'
    }
  }

  // Handle final row
  if (currentRow.length > 0 || currentField) {
    currentRow.push(currentField.trim())
    if (currentRow.some(cell => cell !== '')) {
      rows.push(currentRow)
    }
  }

  if (rows.length === 0) return null

  let headers: string[]
  let dataRows: string[][]

  if (hasHeader && rows.length > 0) {
    headers = rows[0].map((h, i) => h || `col_${i}`)
    dataRows = rows.slice(1)
  } else {
    const numCols = rows[0]?.length || 0
    headers = Array.from({ length: numCols }, (_, i) => `col_${i}`)
    dataRows = rows
  }

  // Normalize rows to have consistent column count
  const maxCols = headers.length
  const normalizedRows = dataRows.map(row => {
    const normalized = [...row]
    while (normalized.length < maxCols) {
      normalized.push('')
    }
    return normalized.slice(0, maxCols)
  })

  // Analyze columns
  const columns = analyzeColumns(headers, normalizedRows)

  return {
    headers,
    rows: normalizedRows,
    columns,
    rowCount: normalizedRows.length,
    estimatedSize: new Blob([content]).size,
  }
}

/**
 * Analyze column types by sampling values
 */
function analyzeColumns(headers: string[], rows: string[][]): ParsedColumn[] {
  return headers.map((name, colIndex) => {
    const values = rows.map(row => row[colIndex])
    const nonEmptyValues = values.filter(v => v !== '' && v !== null && v !== undefined)
    const sampleValues = nonEmptyValues.slice(0, 5)
    const uniqueValues = new Set(values)

    return {
      name,
      type: inferColumnType(nonEmptyValues),
      sampleValues,
      nullCount: values.length - nonEmptyValues.length,
      uniqueCount: uniqueValues.size,
    }
  })
}

/**
 * Infer column type from sample values
 */
function inferColumnType(values: string[]): ColumnType {
  if (values.length === 0) return 'null'

  const types = new Set<ColumnType>()

  for (const value of values.slice(0, 100)) {
    types.add(detectValueType(value))
  }

  // Remove null from consideration
  types.delete('null')

  if (types.size === 0) return 'null'
  if (types.size === 1) return types.values().next().value

  // Check for JSON/array mixed (still JSON-ish)
  if (types.has('json') && types.has('array') && types.size === 2) {
    return 'json'
  }

  return 'mixed'
}

/**
 * Detect the type of a single value
 */
function detectValueType(value: string): ColumnType {
  const trimmed = value.trim()

  if (trimmed === '') return 'null'

  // Check for boolean
  const lower = trimmed.toLowerCase()
  if (lower === BOOLEAN_TRUE || lower === BOOLEAN_FALSE) {
    return 'boolean'
  }

  // Check for number
  if (NUMBER_PATTERN.test(trimmed)) {
    return 'number'
  }

  // Check for JSON object or array
  if ((trimmed.startsWith('{') && trimmed.endsWith('}')) ||
      (trimmed.startsWith('[') && trimmed.endsWith(']'))) {
    try {
      JSON.parse(trimmed)
      return trimmed.startsWith('[') ? 'array' : 'json'
    } catch {
      // Not valid JSON
    }
  }

  return 'string'
}

/**
 * Auto-detect column mapping based on column names
 * Returns suggested mappings for input, expected, and metadata
 */
export function autoDetectColumnMapping(columns: ParsedColumn[]): {
  inputColumn: string | null
  expectedColumn: string | null
  metadataColumns: string[]
} {
  let inputColumn: string | null = null
  let expectedColumn: string | null = null
  const metadataColumns: string[] = []

  for (const col of columns) {
    const lower = col.name.toLowerCase()

    if (!inputColumn && COLUMN_DETECTION_PATTERNS.input.some(p => lower.includes(p))) {
      inputColumn = col.name
    } else if (!expectedColumn && COLUMN_DETECTION_PATTERNS.expected.some(p => lower.includes(p))) {
      expectedColumn = col.name
    } else if (COLUMN_DETECTION_PATTERNS.metadata.some(p => lower.includes(p))) {
      metadataColumns.push(col.name)
    }
  }

  // If no input found, use first column
  if (!inputColumn && columns.length > 0) {
    inputColumn = columns[0].name
  }

  return { inputColumn, expectedColumn, metadataColumns }
}

/**
 * Split CSV content into chunks for upload
 * Uses binary search to find optimal chunk size within payload limit
 */
export function chunkCSVContent(
  content: string,
  hasHeader: boolean,
  maxPayloadSize: number = 500 * 1024 // 500KB default
): string[] {
  const parsed = parseCSV(content, hasHeader)
  if (!parsed) return []

  // If small enough, return as single chunk
  if (parsed.estimatedSize <= maxPayloadSize) {
    return [content]
  }

  const headerRow = hasHeader ? parsed.rows[0] : null
  const dataRows = parsed.rows
  const headerLine = headerRow ? rowToCSV(headerRow) + '\n' : ''
  const headerSize = new Blob([headerLine]).size

  const chunks: string[] = []
  let currentChunk: string[] = []
  let currentSize = headerSize

  for (const row of dataRows) {
    const rowLine = rowToCSV(row) + '\n'
    const rowSize = new Blob([rowLine]).size

    if (currentSize + rowSize > maxPayloadSize && currentChunk.length > 0) {
      // Start new chunk
      chunks.push(headerLine + currentChunk.join('\n'))
      currentChunk = [rowLine.slice(0, -1)]
      currentSize = headerSize + rowSize
    } else {
      currentChunk.push(rowLine.slice(0, -1))
      currentSize += rowSize
    }
  }

  // Add final chunk
  if (currentChunk.length > 0) {
    chunks.push(headerLine + currentChunk.join('\n'))
  }

  return chunks
}

/**
 * Convert row array to CSV string with proper escaping
 */
function rowToCSV(row: string[]): string {
  return row.map(cell => {
    if (cell.includes(',') || cell.includes('"') || cell.includes('\n')) {
      return `"${cell.replace(/"/g, '""')}"`
    }
    return cell
  }).join(',')
}

/**
 * Get a human-readable label for column type
 */
export function getColumnTypeLabel(type: ColumnType): string {
  return COLUMN_TYPE_CONFIG[type].label
}

/**
 * Get suggested icon/badge color for column type
 */
export function getColumnTypeColor(type: ColumnType): string {
  return COLUMN_TYPE_CONFIG[type].color
}
