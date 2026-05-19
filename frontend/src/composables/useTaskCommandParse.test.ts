import { describe, it, expect, vi } from 'vitest'
import { parseRcloneCommand } from './useTaskCommandParse'

vi.mock('../i18n', () => ({
  t: (key: string) => key,
}))

describe('useTaskCommandParse', () => {
  it('should throw on empty command', () => {
    expect(() => parseRcloneCommand('')).toThrow('runtime.commandEmpty')
  })

  it('should throw on too few tokens', () => {
    expect(() => parseRcloneCommand('rclone copy')).toThrow('runtime.commandMissingSrcDst')
  })

  it('should parse copy command', () => {
    const result = parseRcloneCommand('rclone copy remote1:/src remote2:/dst')
    expect(result.mode).toBe('copy')
    expect(result.src).toEqual({ remote: 'remote1', path: '/src' })
    expect(result.dst).toEqual({ remote: 'remote2', path: '/dst' })
  })

  it('should parse sync command', () => {
    const result = parseRcloneCommand('rclone sync remote1:/src remote2:/dst')
    expect(result.mode).toBe('sync')
  })

  it('should parse move command', () => {
    const result = parseRcloneCommand('rclone move remote1:/src remote2:/dst')
    expect(result.mode).toBe('move')
  })

  it('should parse options', () => {
    const result = parseRcloneCommand('rclone copy remote1:/src remote2:/dst --bwlimit 10M --transfers 4')
    expect(result.options.bwLimit).toBe('10M')
    expect(result.options.transfers).toBe(4)
  })

  it('should parse boolean flags', () => {
    const result = parseRcloneCommand('rclone copy remote1:/src remote2:/dst --use-server-modtime --size-only')
    expect(result.options.useServerModtime).toBe(true)
    expect(result.options.sizeOnly).toBe(true)
  })

  it('should handle quoted values', () => {
    const result = parseRcloneCommand('rclone copy remote1:/src remote2:/dst --bwlimit "10M"')
    expect(result.options.bwLimit).toBe('10M')
  })

  it('should handle unknown options', () => {
    const result = parseRcloneCommand('rclone copy remote1:/src remote2:/dst --dry-run --log-level INFO')
    expect(result.options.dryRun).toBe(true)
    expect(result.options.logLevel).toBe('INFO')
  })

  it('should handle paths with colons', () => {
    const result = parseRcloneCommand('rclone copy remote1:/src:with:colons remote2:/dst')
    expect(result.src.path).toBe('/src:with:colons')
  })
})
