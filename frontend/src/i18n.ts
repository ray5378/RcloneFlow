import { useI18n } from './composables/useI18n'
import type { Locale, I18nKey } from './composables/useI18n'

export type { Locale, I18nKey }

const instance = useI18n()

export const locale = instance.locale
export const setLocale = instance.setLocale
export const toggleLocale = instance.toggleLocale
export const t = instance.t