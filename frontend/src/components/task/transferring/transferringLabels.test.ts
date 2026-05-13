import { describe, it, expect, vi } from 'vitest'
import { getTrackingLabels, getTransferStatusLabel } from './transferringLabels'

vi.mock('../../../i18n', () => ({
  t: (key: string) => ({
    'activeTransfer.currentProcessing': '当前处理',
    'activeTransfer.completedProcessing': '已处理',
    'activeTransfer.pendingProcessing': '未处理',
    'activeTransfer.currentTransfer': '当前传输',
    'activeTransfer.completedTransfer': '已传输',
    'activeTransfer.pendingTransfer': '未传输',
    'activeTransfer.statusCopied': '已复制',
    'activeTransfer.statusCasMatched': 'CAS 命中',
    'activeTransfer.statusInProgress': '进行中',
    'activeTransfer.statusPending': '等待中',
    'activeTransfer.statusSkipped': '已跳过',
    'activeTransfer.statusFailed': '失败',
    'activeTransfer.statusDeleted': '已删除',
  }[key] || key),
}))

describe('transferringLabels', () => {
  it('returns CAS tracking labels', () => {
    expect(getTrackingLabels('cas')).toEqual({
      current: '当前处理',
      completed: '已处理',
      pending: '未处理',
    })
  })

  it('returns normal tracking labels', () => {
    expect(getTrackingLabels('normal')).toEqual({
      current: '当前传输',
      completed: '已传输',
      pending: '未传输',
    })
  })

  it('maps known transfer statuses', () => {
    expect(getTransferStatusLabel('copied')).toBe('已复制')
    expect(getTransferStatusLabel('cas_matched')).toBe('CAS 命中')
    expect(getTransferStatusLabel('in_progress')).toBe('进行中')
    expect(getTransferStatusLabel('pending')).toBe('等待中')
    expect(getTransferStatusLabel('skipped')).toBe('已跳过')
    expect(getTransferStatusLabel('failed')).toBe('失败')
    expect(getTransferStatusLabel('deleted')).toBe('已删除')
  })

  it('falls back to raw status for unknown values', () => {
    expect(getTransferStatusLabel('mystery' as any)).toBe('mystery')
  })
})
