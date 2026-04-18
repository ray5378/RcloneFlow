export type ScheduleField = 'month' | 'week' | 'day' | 'hour' | 'minute'

export interface ScheduleFormLike {
  enableSchedule: boolean
  scheduleMinute: string
  scheduleHour: string
  scheduleDay: string
  scheduleMonth: string
  scheduleWeek: string
}

export interface ScheduleTempState {
  minute: string[]
  hour: string[]
  day: string[]
  month: string[]
  week: string[]
}

export const scheduleFieldOptions: Record<ScheduleField, string[]> = {
  month: ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12'],
  week: ['0', '1', '2', '3', '4', '5', '6'],
  day: ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12', '13', '14', '15', '16', '17', '18', '19', '20', '21', '22', '23', '24', '25', '26', '27', '28', '29', '30', '31'],
  hour: Array.from({ length: 24 }, (_, i) => String(i).padStart(2, '0')),
  minute: Array.from({ length: 60 }, (_, i) => String(i).padStart(2, '0')),
}

export const weekLabels = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']

export function createEmptyScheduleTempState(): ScheduleTempState {
  return {
    minute: [],
    hour: [],
    day: [],
    month: [],
    week: [],
  }
}

export function parseScheduleFormToTemp(form?: Partial<ScheduleFormLike> | null): ScheduleTempState {
  return {
    minute: parseScheduleField(form?.scheduleMinute),
    hour: parseScheduleField(form?.scheduleHour),
    day: parseScheduleField(form?.scheduleDay),
    month: parseScheduleField(form?.scheduleMonth),
    week: parseScheduleField(form?.scheduleWeek),
  }
}

export function buildScheduleFormFieldsFromTemp(temp: ScheduleTempState): Omit<ScheduleFormLike, 'enableSchedule'> {
  return {
    scheduleMinute: formatScheduleField('minute', temp.minute),
    scheduleHour: formatScheduleField('hour', temp.hour),
    scheduleDay: formatScheduleField('day', temp.day),
    scheduleMonth: formatScheduleField('month', temp.month),
    scheduleWeek: formatScheduleField('week', temp.week),
  }
}

export function toggleScheduleTempField(temp: ScheduleTempState, field: ScheduleField, value: string): ScheduleTempState {
  const current = temp[field]
  const nextField = current.includes(value)
    ? current.filter(item => item !== value)
    : [...current, value]

  return {
    ...temp,
    [field]: nextField,
  }
}

export function toggleAllScheduleTempField(temp: ScheduleTempState, field: ScheduleField): ScheduleTempState {
  const all = scheduleFieldOptions[field]
  const current = temp[field]

  return {
    ...temp,
    [field]: current.length === all.length ? [] : [...all],
  }
}

export function isScheduleFieldAllSelected(field: ScheduleField, value: string[]): boolean {
  return value.join(',') === scheduleFieldOptions[field].join(',')
}

function parseScheduleField(value?: string | null): string[] {
  if (!value || value === '*') {
    return []
  }
  return value.split(',')
}

function formatScheduleField(field: ScheduleField, value: string[]): string {
  if (!value.length || isScheduleFieldAllSelected(field, value)) {
    return '*'
  }
  return value.join(',')
}
