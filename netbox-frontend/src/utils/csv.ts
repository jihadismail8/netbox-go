/**
 * CSV parsing and export utilities
 */

import type { TableColumn } from '@/types'

/**
 * Parse CSV text into rows of strings.
 * Handles quoted fields, escaped quotes, and newlines within quotes.
 */
export function parseCsv(text: string): string[][] {
  const rows: string[][] = []
  let currentRow: string[] = []
  let currentField = ''
  let inQuotes = false
  let i = 0

  while (i < text.length) {
    const char = text[i]

    if (inQuotes) {
      if (char === '"') {
        if (text[i + 1] === '"') {
          // Escaped quote
          currentField += '"'
          i += 2
        } else {
          // End of quoted field
          inQuotes = false
          i++
        }
      } else {
        currentField += char
        i++
      }
    } else {
      if (char === '"') {
        inQuotes = true
        i++
      } else if (char === ',') {
        currentRow.push(currentField)
        currentField = ''
        i++
      } else if (char === '\n' || char === '\r') {
        currentRow.push(currentField)
        currentField = ''
        rows.push(currentRow)
        currentRow = []
        // Handle \r\n
        if (char === '\r' && text[i + 1] === '\n') i += 2
        else i++
      } else {
        currentField += char
        i++
      }
    }
  }

  // Push last field/row if any content remains
  if (currentField !== '' || currentRow.length > 0) {
    currentRow.push(currentField)
    rows.push(currentRow)
  }

  // Remove empty trailing rows
  return rows.filter((r) => r.length > 0 && !(r.length === 1 && r[0] === ''))
}

/**
 * Parse CSV text into an array of JSON objects using the first row as headers.
 */
export function csvToJson(text: string): Record<string, string>[] {
  const rows = parseCsv(text.trim())
  if (rows.length < 2) return []

  const headers = rows[0].map((h) => h.trim())
  const result: Record<string, string>[] = []

  for (let i = 1; i < rows.length; i++) {
    const obj: Record<string, string> = {}
    for (let j = 0; j < headers.length; j++) {
      obj[headers[j]] = (rows[i][j] || '').trim()
    }
    result.push(obj)
  }

  return result
}

/**
 * Format a cell value for CSV output.
 */
function formatCsvValue(value: unknown): string {
  if (value === null || value === undefined) return ''
  if (typeof value === 'object') {
    if (Array.isArray(value)) return value.length.toString()
    return String(
      Reflect.get(value, 'display') || Reflect.get(value, 'name') || Reflect.get(value, 'id') || '',
    )
  }
  if (typeof value === 'boolean') return value ? 'True' : 'False'
  let str = String(value)
  // Escape quotes and wrap in quotes if needed
  if (str.includes('"') || str.includes(',') || str.includes('\n') || str.includes('\r')) {
    str = '"' + str.replace(/"/g, '""') + '"'
  }
  return str
}

/**
 * Export table data to a CSV file (triggers download).
 */
export function exportToCsv(
  data: object[],
  columns: TableColumn[],
  filename: string = 'export.csv',
): void {
  const headers = columns.map((c) => c.label)
  const lines: string[] = [headers.join(',')]

  for (const row of data) {
    const values = columns.map((col) => {
      const raw = Reflect.get(row, col.key) as unknown
      if (col.formatter) {
        return formatCsvValue(col.formatter(raw, row))
      }
      return formatCsvValue(raw)
    })
    lines.push(values.join(','))
  }

  const csvContent = lines.join('\n')
  const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.setAttribute('href', url)
  link.setAttribute('download', filename)
  link.style.visibility = 'hidden'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
}

/**
 * Generate a CSV template string from form fields (for import templates).
 */
export function generateCsvTemplate(fields: { key: string; required?: boolean }[]): string {
  const headers = fields.map((f) => f.key)
  const example = fields.map((f) => {
    if (f.required) return `required_${f.key}_value`
    return ''
  })
  return [headers.join(','), example.join(',')].join('\n')
}
