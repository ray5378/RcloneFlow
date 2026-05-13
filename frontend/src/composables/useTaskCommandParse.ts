interface ParsedRcloneCommand {
  mode: string
  src: { remote: string; path: string }
  dst: { remote: string; path: string }
  options: Record<string, any>
}

import { t } from '../i18n'

export function parseRcloneCommand(cmd: string): ParsedRcloneCommand {
  if (!cmd) throw new Error(t('runtime.commandEmpty'))

  const tokens = cmd.match(/(?:"[^"]*"|'[^']*'|\S)+/g) || []
  if (tokens.length < 3) throw new Error(t('runtime.commandMissingSrcDst'))

  const sub = tokens[1]
  const mode = sub === 'sync' ? 'sync' : sub === 'move' ? 'move' : 'copy'
  const src = parseRemotePath(tokens[2])
  const dst = parseRemotePath(tokens[3])
  const options: Record<string, any> = {}

  for (let i = 4; i < tokens.length; i++) {
    const token = tokens[i]
    if (!token.startsWith('--')) continue

    const key = token.replace(/^--/, '')
    const next = tokens[i + 1]

    switch (key) {
      case 'bwlimit':
        options.bwLimit = stripQuotes(next)
        i++
        break
      case 'transfers':
        options.transfers = Number(next)
        i++
        break
      case 'use-server-modtime':
        options.useServerModtime = true
        break
      case 'size-only':
        options.sizeOnly = true
        break
      case 'verbose':
        break
      default:
        if (!next || next.startsWith('--')) {
          options[toCamel(key)] = true
        } else {
          options[toCamel(key)] = stripQuotes(next)
          i++
        }
    }
  }

  return { mode, src, dst, options }
}

function parseRemotePath(value: string) {
  const cleaned = stripQuotes(value) || value
  const parts = cleaned.split(':')
  if (parts.length < 2) throw new Error(t('runtime.commandInvalidPath').replace('{value}', value))
  return { remote: parts[0], path: parts.slice(1).join(':') || '' }
}

function stripQuotes(value?: string) {
  return value ? value.replace(/^['"]|['"]$/g, '') : value
}

function toCamel(value: string) {
  return value.replace(/-([a-z])/g, (_, char) => char.toUpperCase())
}
