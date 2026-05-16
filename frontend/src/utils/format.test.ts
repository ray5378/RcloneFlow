import { describe, it, expect } from 'vitest'
import { formatBytes, formatBytesPerSec, formatDuration, formatEta } from './format'

describe('formatBytes', () => {
  it('should format zero bytes as readable text', () => {
    expect(formatBytes(0)).toBe('0 B')
  })

  it('should return - for undefined', () => {
    expect(formatBytes(undefined as any)).toBe('-')
  })

  it('should return - for null/NaN', () => {
    expect(formatBytes(null as any)).toBe('-')
    expect(formatBytes(Number.NaN)).toBe('-')
  })

  it('should format bytes correctly', () => {
    expect(formatBytes(500)).toBe('500 B')
    expect(formatBytes(1024)).toBe('1.0 KB')
    expect(formatBytes(1024 * 1024)).toBe('1.0 MB')
    expect(formatBytes(1024 * 1024 * 1024)).toBe('1.00 GB')
  })

  it('should format GB values with 2 decimals', () => {
    // 1.33 GB
    expect(formatBytes(1429365116)).toBe('1.33 GB')
    // 4.22 GB
    expect(formatBytes(4534483558)).toBe('4.22 GB')
    // 5.4 GB
    expect(formatBytes(5798205850)).toBe('5.40 GB')
  })

  it('should handle TB with 2 decimals', () => {
    expect(formatBytes(1024 * 1024 * 1024 * 1024)).toBe('1.00 TB')
  })
})

describe('formatBytesPerSec', () => {
  it('should return - for zero', () => {
    expect(formatBytesPerSec(0)).toBe('-')
  })

  it('should format speed correctly', () => {
    expect(formatBytesPerSec(1024)).toBe('1.0 KB/s')
    expect(formatBytesPerSec(1024 * 1024)).toBe('1.0 MB/s')
  })
})

describe('formatDuration', () => {
  it('should return - for undefined start time', () => {
    expect(formatDuration(undefined, undefined)).toBe('-')
  })

  it('should format seconds correctly', () => {
    const start = '2024-01-01T00:00:00Z'
    const end = '2024-01-01T00:00:30Z'
    expect(formatDuration(start, end)).toBe('30秒')
  })

  it('should format minutes correctly', () => {
    const start = '2024-01-01T00:00:00Z'
    const end = '2024-01-01T00:05:30Z'
    expect(formatDuration(start, end)).toBe('5分30秒')
  })

  it('should format hours correctly', () => {
    const start = '2024-01-01T00:00:00Z'
    const end = '2024-01-01T02:30:00Z'
    expect(formatDuration(start, end)).toBe('2小时30分')
  })

  it('should format days correctly', () => {
    const start = '2024-01-01T00:00:00Z'
    const end = '2024-01-03T05:00:00Z'
    expect(formatDuration(start, end)).toMatch(/天/)
  })
})

describe('formatEta', () => {
  it('should return - for zero', () => {
    expect(formatEta(0)).toBe('-')
  })

  it('should return - for negative', () => {
    expect(formatEta(-100)).toBe('-')
  })

  it('should format seconds correctly', () => {
    expect(formatEta(30)).toBe('约30秒')
  })

  it('should format minutes correctly', () => {
    expect(formatEta(120)).toBe('约2分')
  })

  it('should format hours correctly', () => {
    expect(formatEta(3600)).toBe('约1小时0分')
  })
})
