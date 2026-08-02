import { describe, it, expect, vi } from 'vitest'
import { parseCsv, csvToJson, generateCsvTemplate, exportToCsv } from './csv'
import type { TableColumn } from '@/types'

// jsdom does not implement URL.createObjectURL/revokeObjectURL; provide stubs
if (!('createObjectURL' in URL)) {
  Object.defineProperty(URL, 'createObjectURL', { value: () => 'blob:mock', writable: true })
}
if (!('revokeObjectURL' in URL)) {
  Object.defineProperty(URL, 'revokeObjectURL', { value: () => {}, writable: true })
}

describe('parseCsv', () => {
  it('parses a simple CSV', () => {
    const result = parseCsv('a,b,c\n1,2,3\n4,5,6')
    expect(result).toEqual([
      ['a', 'b', 'c'],
      ['1', '2', '3'],
      ['4', '5', '6'],
    ])
  })

  it('handles quoted fields with commas', () => {
    const result = parseCsv('name,desc\n"Site A","Main, datacenter"\n"Site B","Backup"')
    expect(result[1]).toEqual(['Site A', 'Main, datacenter'])
    expect(result[2]).toEqual(['Site B', 'Backup'])
  })

  it('handles escaped quotes inside quoted fields', () => {
    const result = parseCsv('text\n"He said ""hello"""')
    expect(result[1]).toEqual(['He said "hello"'])
  })

  it('handles newlines inside quoted fields', () => {
    const result = parseCsv('text\n"Line 1\nLine 2"')
    expect(result[1]).toEqual(['Line 1\nLine 2'])
  })

  it('handles Windows line endings (\\r\\n)', () => {
    const result = parseCsv('a,b\r\n1,2\r\n3,4')
    expect(result).toEqual([
      ['a', 'b'],
      ['1', '2'],
      ['3', '4'],
    ])
  })

  it('trims whitespace in headers when using csvToJson', () => {
    const result = csvToJson(' name , status \nAlpha, active ')
    expect(result[0]).toEqual({ name: 'Alpha', status: 'active' })
  })

  it('returns empty array for empty input', () => {
    expect(parseCsv('')).toEqual([])
    expect(csvToJson('')).toEqual([])
  })

  it('returns empty array for header-only CSV', () => {
    expect(csvToJson('name,status')).toEqual([])
  })

  it('handles missing trailing fields', () => {
    const result = csvToJson('a,b,c\n1,2\n3,4,5')
    expect(result[0]).toEqual({ a: '1', b: '2', c: '' })
    expect(result[1]).toEqual({ a: '3', b: '4', c: '5' })
  })
})

describe('generateCsvTemplate', () => {
  it('generates a template with headers and example row', () => {
    const fields = [{ key: 'name', required: true }, { key: 'status' }, { key: 'description' }]
    const template = generateCsvTemplate(fields)
    expect(template).toContain('name,status,description')
    expect(template).toContain('required_name_value')
  })
})

describe('exportToCsv', () => {
  it('exports data to CSV (mocked download)', () => {
    const columns: TableColumn[] = [
      { key: 'name', label: 'Name' },
      { key: 'count', label: 'Count' },
    ]
    const data = [
      { name: 'Site A', count: 5 },
      { name: 'Site B', count: 10 },
    ]

    // Mock DOM APIs
    const clickMock = vi.fn()
    const createElementSpy = vi.spyOn(document, 'createElement').mockReturnValue({
      setAttribute: vi.fn(),
      style: { visibility: '' },
      click: clickMock,
    } as unknown as HTMLAnchorElement)
    const appendChildSpy = vi.spyOn(document.body, 'appendChild').mockReturnValue({} as Node)
    const removeChildSpy = vi.spyOn(document.body, 'removeChild').mockReturnValue({} as Node)
    const createObjectURLSpy = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:url')
    const revokeObjectURLSpy = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})

    exportToCsv(data, columns, 'test.csv')

    expect(clickMock).toHaveBeenCalledTimes(1)
    expect(createObjectURLSpy).toHaveBeenCalledTimes(1)

    createElementSpy.mockRestore()
    appendChildSpy.mockRestore()
    removeChildSpy.mockRestore()
    createObjectURLSpy.mockRestore()
    revokeObjectURLSpy.mockRestore()
  })
})
