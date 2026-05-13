import { describe, it, expect, vi } from 'vitest'

vi.mock('../../i18n', () => ({
  t: (key: string) => ({
    'runtime.sunday': '周日',
    'runtime.monday': '周一',
    'runtime.tuesday': '周二',
    'runtime.wednesday': '周三',
    'runtime.thursday': '周四',
    'runtime.friday': '周五',
    'runtime.saturday': '周六',
  }[key] || key),
}))

import {
  scheduleFieldOptions,
  getWeekLabels,
  weekLabels,
  createEmptyScheduleTempState,
  parseScheduleFormToTemp,
  buildScheduleFormFieldsFromTemp,
  toggleScheduleTempField,
  toggleAllScheduleTempField,
} from './scheduleOptions'

describe('scheduleOptions', () => {
  it('exposes field options and week labels', () => {
    expect(scheduleFieldOptions.month).toHaveLength(12)
    expect(scheduleFieldOptions.hour[0]).toBe('00')
    expect(scheduleFieldOptions.minute[59]).toBe('59')
    expect(getWeekLabels()).toEqual(['周日', '周一', '周二', '周三', '周四', '周五', '周六'])
    expect(weekLabels).toEqual(['周日', '周一', '周二', '周三', '周四', '周五', '周六'])
  })

  it('creates and parses empty temp state', () => {
    expect(createEmptyScheduleTempState()).toEqual({ minute: [], hour: [], day: [], month: [], week: [] })
    expect(parseScheduleFormToTemp()).toEqual({ minute: [], hour: [], day: [], month: [], week: [] })
    expect(parseScheduleFormToTemp({ scheduleMinute: '*', scheduleHour: '01, 03', scheduleDay: '1,2', scheduleMonth: '', scheduleWeek: '0' })).toEqual({
      minute: [],
      hour: ['01', '03'],
      day: ['1', '2'],
      month: [],
      week: ['0'],
    })
  })

  it('builds fields from temp state', () => {
    expect(buildScheduleFormFieldsFromTemp({ minute: [], hour: ['03', '01'], day: ['2'], month: ['12'], week: ['0', '6'] })).toEqual({
      scheduleMinute: '*',
      scheduleHour: '01,03',
      scheduleDay: '2',
      scheduleMonth: '12',
      scheduleWeek: '0,6',
    })
  })

  it('toggles single and all values', () => {
    const base = createEmptyScheduleTempState()
    const one = toggleScheduleTempField(base, 'week', '1')
    expect(one.week).toEqual(['1'])
    const removed = toggleScheduleTempField(one, 'week', '1')
    expect(removed.week).toEqual([])

    const all = toggleAllScheduleTempField(base, 'month')
    expect(all.month).toHaveLength(12)
    const none = toggleAllScheduleTempField(all, 'month')
    expect(none.month).toEqual([])
  })
})
