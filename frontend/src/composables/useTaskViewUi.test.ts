import { describe, it, expect } from 'vitest'
import { useTaskViewUi } from './useTaskViewUi'

describe('useTaskViewUi', () => {
  it('should toggle menu', () => {
    const { openMenuId, toggleMenu, closeMenus } = useTaskViewUi()
    toggleMenu(1)
    expect(openMenuId.value).toBe(1)
    toggleMenu(1)
    expect(openMenuId.value).toBeNull()
    toggleMenu(2)
    expect(openMenuId.value).toBe(2)
    closeMenus()
    expect(openMenuId.value).toBeNull()
  })

  it('should show and close confirm modal', () => {
    const { confirmModal, showConfirm, closeConfirm } = useTaskViewUi()
    expect(confirmModal.value.show).toBe(false)
    showConfirm('Title', 'Message', () => {})
    expect(confirmModal.value.show).toBe(true)
    expect(confirmModal.value.title).toBe('Title')
    closeConfirm()
    expect(confirmModal.value.show).toBe(false)
  })

  it('should execute callback on confirm', () => {
    const { showConfirm, confirmAndClose } = useTaskViewUi()
    let executed = false
    showConfirm('Title', 'Message', () => { executed = true })
    confirmAndClose()
    expect(executed).toBe(true)
  })
})
