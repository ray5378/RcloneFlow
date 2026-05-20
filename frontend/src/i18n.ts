import { ref } from 'vue'
import zh from './i18n/zh'
import en from './i18n/en'

export type Locale = 'zh' | 'en'
const messages = { zh, en } as const

type MessageTree = typeof zh

type DotPath<T, Prefix extends string = ''> = {
  [K in keyof T & string]: T[K] extends string
    ? `${Prefix}${K}`
    : T[K] extends Record<string, any>
      ? DotPath<T[K], `${Prefix}${K}.`>
      : never
}[keyof T & string]

export type I18nKey = DotPath<MessageTree>

const stored = localStorage.getItem('ui-locale') as Locale | null
export const locale = ref<Locale>(stored === 'en' ? 'en' : 'zh')

export function setLocale(next: Locale) {
  locale.value = next
  localStorage.setItem('ui-locale', next)
}

export function toggleLocale() {
  setLocale(locale.value === 'zh' ? 'en' : 'zh')
}

function getByPath(obj: Record<string, any>, path: string): string | undefined {
  return path.split('.').reduce<any>((acc, key) => (acc == null ? undefined : acc[key]), obj)
}

export function t(key: I18nKey, params?: Record<string, any>, fallback?: string): string {
  let result = getByPath(messages[locale.value] as Record<string, any>, key) || fallback || key
  if (params) {
    result = result.replace(/\{(\w+)\}/g, (_, k) => (params[k] != null ? String(params[k]) : `{${k}}`))
  }
  return result
}
