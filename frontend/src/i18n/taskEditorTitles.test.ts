import { describe, it, expect } from 'vitest'
import zh from './zh'
import en from './en'

describe('task editor titles', () => {
  it('uses updated Chinese task editor titles', () => {
    expect(zh.taskEditor.createTitle).toBe('新建任务')
    expect(zh.taskEditor.editTitle).toBe('修改任务')
  })

  it('keeps English task editor titles stable', () => {
    expect(en.taskEditor.createTitle).toBe('Create Task')
    expect(en.taskEditor.editTitle).toBe('Edit Task')
  })
})
