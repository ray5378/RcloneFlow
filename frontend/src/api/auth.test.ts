import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import {
  login,
  register,
  logout,
  setTokens,
  getToken,
  getRefreshToken,
  isLoggedIn,
  getUser,
  changePassword
} from './auth'

const mockFetch = vi.fn()
globalThis.fetch = mockFetch as any

describe('auth.ts', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  describe('setTokens', () => {
    it('should store tokens in localStorage', () => {
      setTokens('access123', 'refresh456')
      expect(localStorage.getItem('authToken')).toBe('access123')
      expect(localStorage.getItem('refreshToken')).toBe('refresh456')
    })
  })

  describe('getToken', () => {
    it('should return null when no token', () => {
      expect(getToken()).toBeNull()
    })

    it('should return token when set', () => {
      localStorage.setItem('authToken', 'test-token')
      expect(getToken()).toBe('test-token')
    })
  })

  describe('getRefreshToken', () => {
    it('should return null when no refresh token', () => {
      expect(getRefreshToken()).toBeNull()
    })

    it('should return refresh token when set', () => {
      localStorage.setItem('refreshToken', 'refresh-token')
      expect(getRefreshToken()).toBe('refresh-token')
    })
  })

  describe('isLoggedIn', () => {
    it('should return false when not logged in', () => {
      expect(isLoggedIn()).toBe(false)
    })

    it('should return true when logged in', () => {
      localStorage.setItem('authToken', 'some-token')
      expect(isLoggedIn()).toBe(true)
    })
  })

  describe('getUser', () => {
    it('should return null when no user', () => {
      expect(getUser()).toBeNull()
    })

    it('should return parsed user when set', () => {
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin' }))
      expect(getUser()).toEqual({ id: 1, username: 'admin' })
    })
  })

  describe('logout', () => {
    it('should remove all auth items', () => {
      localStorage.setItem('authToken', 'token')
      localStorage.setItem('refreshToken', 'refresh')
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin' }))

      logout()

      expect(localStorage.getItem('authToken')).toBeNull()
      expect(localStorage.getItem('refreshToken')).toBeNull()
      expect(localStorage.getItem('user')).toBeNull()
    })
  })

  describe('login', () => {
    it('should login successfully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          accessToken: 'new-access',
          refreshToken: 'new-refresh',
          user: { id: 1, username: 'admin' }
        })
      })

      const result = await login('admin', 'admin')

      expect(mockFetch).toHaveBeenCalledWith('/api/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'admin', password: 'admin' })
      })
      expect(result.accessToken).toBe('new-access')
      expect(getToken()).toBe('new-access')
      expect(getUser()).toEqual({ id: 1, username: 'admin' })
    })

    it('should throw on login failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: () => Promise.resolve({ error: '用户名或密码错误' })
      })

      await expect(login('admin', 'wrong')).rejects.toThrow('用户名或密码错误')
    })

    it('should throw default error on parse failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: () => Promise.reject(new Error('parse error'))
      })

      await expect(login('admin', 'pass')).rejects.toThrow('登录失败')
    })
  })

  describe('register', () => {
    it('should register successfully', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({
          accessToken: 'reg-access',
          refreshToken: 'reg-refresh',
          user: { id: 2, username: 'newuser' }
        })
      })

      const result = await register('newuser', 'password123')

      expect(mockFetch).toHaveBeenCalledWith('/api/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: 'newuser', password: 'password123' })
      })
      expect(result.accessToken).toBe('reg-access')
    })

    it('should throw on register failure', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: () => Promise.resolve({ error: '用户名已存在' })
      })

      await expect(register('exists', 'pass')).rejects.toThrow('用户名已存在')
    })
  })

  describe('changePassword', () => {
    it('should change password successfully', async () => {
      localStorage.setItem('authToken', 'valid-token')
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({ message: '修改成功' })
      })

      await changePassword('old', 'new')

      expect(mockFetch).toHaveBeenCalledWith('/api/auth/change-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer valid-token'
        },
        body: JSON.stringify({ oldPassword: 'old', newPassword: 'new', username: undefined })
      })
    })

    it('should include username when provided', async () => {
      localStorage.setItem('authToken', 'valid-token')
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: () => Promise.resolve({})
      })

      await changePassword('old', 'new', 'newname')

      expect(mockFetch).toHaveBeenCalledWith('/api/auth/change-password', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer valid-token'
        },
        body: JSON.stringify({ oldPassword: 'old', newPassword: 'new', username: 'newname' })
      })
    })

    it('should throw on failure', async () => {
      localStorage.setItem('authToken', 'valid-token')
      mockFetch.mockResolvedValueOnce({
        ok: false,
        json: () => Promise.resolve({ error: '旧密码错误' })
      })

      await expect(changePassword('wrong', 'new')).rejects.toThrow('旧密码错误')
    })
  })
})
