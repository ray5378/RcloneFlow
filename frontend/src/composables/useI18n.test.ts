import { describe, it, expect, vi, beforeEach } from 'vitest'
import { useI18n } from './useI18n'

vi.mock('../i18n/zh', () => ({
  default: {
    browserView: {
      clipboardEmpty: '剪贴板为空',
      selectStorageFirst: '请先选择存储',
      pathTraversalRisk: '路径遍历风险 {name}',
    },
    taskCard: {
      source: '源',
      target: '目标',
    },
  },
}))

vi.mock('../i18n/en', () => ({
  default: {
    browserView: {
      clipboardEmpty: 'Clipboard empty',
      selectStorageFirst: 'Select storage first',
      pathTraversalRisk: 'Path traversal risk {name}',
    },
    taskCard: {
      source: 'Source',
      target: 'Target',
    },
  },
}))

describe('useI18n.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  describe('initialization', () => {
    it('should return expected interface', () => {
      const i18n = useI18n()
      expect(i18n).toHaveProperty('locale')
      expect(i18n).toHaveProperty('setLocale')
      expect(i18n).toHaveProperty('toggleLocale')
      expect(i18n).toHaveProperty('t')
      expect(typeof i18n.setLocale).toBe('function')
      expect(typeof i18n.toggleLocale).toBe('function')
      expect(typeof i18n.t).toBe('function')
    })

    it('should default to zh locale when no stored preference', () => {
      const i18n = useI18n()
      expect(i18n.locale.value).toBe('zh')
    })

    it('should load stored en locale from localStorage', () => {
      localStorage.setItem('ui-locale', 'en')
      const i18n = useI18n()
      expect(i18n.locale.value).toBe('en')
    })

    it('should default to zh when stored value is invalid', () => {
      localStorage.setItem('ui-locale', 'fr')
      const i18n = useI18n()
      expect(i18n.locale.value).toBe('zh')
    })
  })

  describe('setLocale', () => {
    it('should set locale to zh', () => {
      const i18n = useI18n()
      i18n.setLocale('zh')
      expect(i18n.locale.value).toBe('zh')
    })

    it('should set locale to en', () => {
      const i18n = useI18n()
      i18n.setLocale('en')
      expect(i18n.locale.value).toBe('en')
    })

    it('should persist locale to localStorage', () => {
      const i18n = useI18n()
      i18n.setLocale('en')
      expect(localStorage.getItem('ui-locale')).toBe('en')
    })
  })

  describe('toggleLocale', () => {
    it('should toggle from zh to en', () => {
      const i18n = useI18n()
      i18n.locale.value = 'zh'
      i18n.toggleLocale()
      expect(i18n.locale.value).toBe('en')
    })

    it('should toggle from en to zh', () => {
      const i18n = useI18n()
      i18n.locale.value = 'en'
      i18n.toggleLocale()
      expect(i18n.locale.value).toBe('zh')
    })

    it('should persist toggle result to localStorage', () => {
      const i18n = useI18n()
      i18n.locale.value = 'zh'
      i18n.toggleLocale()
      expect(localStorage.getItem('ui-locale')).toBe('en')
    })
  })

  describe('translation function t', () => {
    it('should return translated string for zh locale', () => {
      const i18n = useI18n()
      i18n.setLocale('zh')
      const result = i18n.t('browserView.clipboardEmpty')
      expect(result).toBe('剪贴板为空')
    })

    it('should return translated string for en locale', () => {
      const i18n = useI18n()
      i18n.setLocale('en')
      const result = i18n.t('browserView.clipboardEmpty')
      expect(result).toBe('Clipboard empty')
    })

    it('should replace parameters in translation', () => {
      const i18n = useI18n()
      i18n.setLocale('zh')
      const result = i18n.t('browserView.pathTraversalRisk', { name: 'test' })
      expect(result).toBe('路径遍历风险 test')
    })

    it('should use fallback when key not found', () => {
      const i18n = useI18n()
      const result = i18n.t('non.existent.key', {}, 'fallback value')
      expect(result).toBe('fallback value')
    })

    it('should return key itself when no translation and no fallback', () => {
      const i18n = useI18n()
      const result = i18n.t('non.existent.key')
      expect(result).toBe('non.existent.key')
    })

    it('should handle multiple parameters', () => {
      const i18n = useI18n()
      i18n.setLocale('zh')
      const result = i18n.t('taskCard.source')
      expect(result).toBe('源')
    })

    it('should preserve missing parameters in string', () => {
      const i18n = useI18n()
      i18n.setLocale('zh')
      const result = i18n.t('browserView.pathTraversalRisk', {})
      expect(result).toBe('路径遍历风险 {name}')
    })

    it('should handle null parameter values', () => {
      const i18n = useI18n()
      i18n.setLocale('zh')
      const result = i18n.t('browserView.pathTraversalRisk', { name: null })
      expect(result).toBe('路径遍历风险 {name}')
    })

    it('should handle undefined parameter values', () => {
      const i18n = useI18n()
      i18n.setLocale('zh')
      const result = i18n.t('browserView.pathTraversalRisk', { name: undefined })
      expect(result).toBe('路径遍历风险 {name}')
    })
  })
})
