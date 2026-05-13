import { describe, it, expect, vi } from 'vitest'

vi.mock('../i18n', () => ({
  t: (key: string) => ({
    'runtime.commandEmpty': '命令为空',
    'runtime.commandMissingSrcDst': '缺少源和目标',
    'runtime.commandInvalidPath': '路径非法: {value}',
  }[key] || key),
}))

import { parseRcloneCommand } from './useTaskCommandParse'

describe('parseRcloneCommand', () => {
  it('throws on empty or incomplete command', () => {
    expect(() => parseRcloneCommand('')).toThrow('命令为空')
    expect(() => parseRcloneCommand('rclone copy')).toThrow('缺少源和目标')
  })

  it('parses sync command and typed options', () => {
    const result = parseRcloneCommand('rclone sync src:/a dst:/b --bwlimit "10M" --transfers 4 --use-server-modtime --size-only --fast-list')
    expect(result.mode).toBe('sync')
    expect(result.src).toEqual({ remote: 'src', path: '/a' })
    expect(result.dst).toEqual({ remote: 'dst', path: '/b' })
    expect(result.options).toEqual({
      bwLimit: '10M',
      transfers: 4,
      useServerModtime: true,
      sizeOnly: true,
      fastList: true,
    })
  })

  it('parses move/copy mode and generic options', () => {
    const move = parseRcloneCommand("rclone move 'src:/a:b' dst:/b --checkers 8 --metadata")
    expect(move.mode).toBe('move')
    expect(move.src).toEqual({ remote: 'src', path: '/a:b' })
    expect(move.options).toEqual({ checkers: '8', metadata: true })

    const copy = parseRcloneCommand('rclone copy src:/a dst:/b --verbose')
    expect(copy.mode).toBe('copy')
    expect(copy.options).toEqual({})
  })

  it('rejects invalid remote paths', () => {
    expect(() => parseRcloneCommand('rclone copy invalid dst:/b')).toThrow('路径非法: invalid')
  })
})
