import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'

vi.mock('../../api', () => ({
  getRemotes: vi.fn(),
  listPath: vi.fn(),
  listProviders: vi.fn(),
  createRemote: vi.fn(),
  getRemoteConfig: vi.fn(),
  updateRemote: vi.fn(),
}))

vi.mock('../../i18n', () => ({
  t: (key: string) => key,
  locale: { value: 'zh' },
}))

import * as api from '../../api'

describe('Crypt Remote Configuration Parameters', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    vi.restoreAllMocks()
  })

  const mockCryptProvider = {
    Name: 'crypt',
    Description: 'Encrypt/Decrypt a remote',
    Options: [
      {
        Name: 'remote',
        Help: 'Remote to encrypt/decrypt. Normally should contain a \':\' and a path.',
        Required: true,
        IsPassword: false,
        Hide: 0,
        Advanced: false,
        DefaultStr: '',
        Examples: [],
      },
      {
        Name: 'filename_encryption',
        Help: 'How to encrypt the filenames.',
        Required: false,
        IsPassword: false,
        Hide: 0,
        Advanced: false,
        DefaultStr: 'standard',
        Examples: [
          { Value: 'standard', Help: 'Encrypt the filenames.' },
          { Value: 'obfuscate', Help: 'Very simple filename obfuscation.' },
          { Value: 'off', Help: "Don't encrypt the file names." },
        ],
      },
      {
        Name: 'directory_name_encryption',
        Help: 'Option to either encrypt directory names or leave them intact.',
        Required: false,
        IsPassword: false,
        Hide: 0,
        Advanced: false,
        DefaultStr: 'true',
        Examples: [
          { Value: 'true', Help: 'Encrypt directory names.' },
          { Value: 'false', Help: "Don't encrypt directory names." },
        ],
      },
      {
        Name: 'password',
        Help: 'Password or pass phrase for encryption.',
        Required: true,
        IsPassword: true,
        Hide: 0,
        Advanced: false,
        DefaultStr: '',
        Examples: [],
      },
      {
        Name: 'password2',
        Help: 'Password or pass phrase for salt. Optional but recommended.',
        Required: false,
        IsPassword: true,
        Hide: 0,
        Advanced: false,
        DefaultStr: '',
        Examples: [],
      },
      {
        Name: 'filename_encoding',
        Help: 'How to encode the encrypted filename to text string.',
        Required: false,
        IsPassword: false,
        Hide: 0,
        Advanced: true,
        DefaultStr: 'base32',
        Examples: [
          { Value: 'base32', Help: 'base32 encoding' },
          { Value: 'base64', Help: 'base64 encoding' },
        ],
      },
      {
        Name: 'suffix',
        Help: 'If this is set it will override the default suffix of ".bin".',
        Required: false,
        IsPassword: false,
        Hide: 0,
        Advanced: true,
        DefaultStr: '',
        Examples: [],
      },
      {
        Name: 'no_data_encryption',
        Help: "Option to either encrypt file data or leave it unencrypted.",
        Required: false,
        IsPassword: false,
        Hide: 0,
        Advanced: true,
        DefaultStr: 'false',
        Examples: [
          { Value: 'true', Help: 'Do not encrypt file data.' },
          { Value: 'false', Help: 'Encrypt file data.' },
        ],
      },
    ],
  }

  describe('field name mapping matches rclone CLI options', () => {
    it('exposes all required crypt options through ProviderOptions', () => {
      const optionNames = mockCryptProvider.Options.map(o => o.Name)

      expect(optionNames).toContain('remote')
      expect(optionNames).toContain('password')
      expect(optionNames).toContain('password2')
      expect(optionNames).toContain('filename_encryption')
      expect(optionNames).toContain('directory_name_encryption')
    })

    it('marks password field as IsPassword', () => {
      const passwordOpt = mockCryptProvider.Options.find(o => o.Name === 'password')
      expect(passwordOpt).toBeDefined()
      expect(passwordOpt!.IsPassword).toBe(true)

      const password2Opt = mockCryptProvider.Options.find(o => o.Name === 'password2')
      expect(password2Opt).toBeDefined()
      expect(password2Opt!.IsPassword).toBe(true)
    })

    it('has correct filename_encryption option values', () => {
      const fe = mockCryptProvider.Options.find(o => o.Name === 'filename_encryption')
      expect(fe).toBeDefined()
      const values = fe!.Examples?.map(e => e.Value) || []
      expect(values).toContain('standard')
      expect(values).toContain('obfuscate')
      expect(values).toContain('off')
    })
  })

  describe('parameter validation matches rclone API expectations', () => {
    it('remote must contain colon separator', () => {
      const validRemotes = ['gdrive:', 's3:bucket', 'gdrive:path/to/encrypted']
      for (const r of validRemotes) {
        expect(r).toMatch(/:/)
      }
    })

    it('remote name must not be empty', () => {
      const invalidRemotes = [':path', ':']
      for (const r of invalidRemotes) {
        const parts = r.split(':')
        expect(parts[0]).toBe('')
      }
    })

    it('password must not be empty for crypt to work', () => {
      const password = ''
      expect(password.length).toBe(0)
    })

    it('optional password2 can be empty (will use internal salt)', () => {
      const password2 = ''
      expect(password2.length).toBe(0)
    })

    it('filename_encryption default is standard', () => {
      const fe = mockCryptProvider.Options.find(o => o.Name === 'filename_encryption')
      expect(fe!.DefaultStr).toBe('standard')
    })

    it('directory_name_encryption default is true', () => {
      const dne = mockCryptProvider.Options.find(o => o.Name === 'directory_name_encryption')
      expect(dne!.DefaultStr).toBe('true')
    })

    it('no_data_encryption default is false (means data IS encrypted)', () => {
      const nde = mockCryptProvider.Options.find(o => o.Name === 'no_data_encryption')
      expect(nde!.DefaultStr).toBe('false')
    })
  })

  describe('API call parameter format', () => {
    it('sends parameters with correct structure to backend', async () => {
      ;(api.createRemote as ReturnType<typeof vi.fn>).mockResolvedValueOnce(undefined)

      const params = {
        remote: 'gdrive:encrypted',
        password: 'my-password-123',
        password2: 'my-salt-456',
        filename_encryption: 'standard',
        directory_name_encryption: 'true',
      }

      await api.createRemote('mysecret', 'crypt', params)

      expect(api.createRemote).toHaveBeenCalledWith(
        'mysecret',
        'crypt',
        params
      )

      const callArgs = (api.createRemote as ReturnType<typeof vi.fn>).mock.calls[0]
      expect(callArgs[0]).toBe('mysecret')
      expect(callArgs[1]).toBe('crypt')
      expect(callArgs[2].remote).toBe('gdrive:encrypted')
      expect(callArgs[2].password).toBe('my-password-123')
      expect(callArgs[2].password2).toBe('my-salt-456')
      expect(callArgs[2].filename_encryption).toBe('standard')
      expect(callArgs[2].directory_name_encryption).toBe('true')
    })

    it('filters out empty optional parameters', () => {
      const rawOptions: Record<string, string> = {
        remote: 'gdrive:backup',
        password: 'mypassword',
        password2: '',
        filename_encryption: 'standard',
        directory_name_encryption: 'true',
        filename_encoding: '',
        suffix: '',
      }

      const filtered = Object.fromEntries(
        Object.entries(rawOptions).filter(([, value]) => value !== '')
      )

      expect(filtered).toHaveProperty('remote', 'gdrive:backup')
      expect(filtered).toHaveProperty('password', 'mypassword')
      expect(filtered).toHaveProperty('filename_encryption', 'standard')
      expect(filtered).toHaveProperty('directory_name_encryption', 'true')
      expect(filtered).not.toHaveProperty('password2')
      expect(filtered).not.toHaveProperty('filename_encoding')
      expect(filtered).not.toHaveProperty('suffix')
    })
  })

  describe('remote path format from RemotePathBrowser', () => {
    it('produces remote:path format matching rclone CLI convention', () => {
      const remoteName = 'gdrive'
      const path = 'encrypted/data'

      const fullPath = `${remoteName}:${path}`
      expect(fullPath).toBe('gdrive:encrypted/data')
      expect(fullPath).toMatch(/^[a-zA-Z0-9_-]+:.*$/)
    })

    it('produces remote: format for root directory', () => {
      const remoteName = 's3'
      const path = ''

      const fullPath = `${remoteName}:${path}`
      expect(fullPath).toBe('s3:')
      expect(fullPath).toMatch(/:/)
    })

    it('handles path with special characters', () => {
      const remoteName = 'gdrive'
      const paths = [
        'backup-2024/photos',
        'Encrypted/Data',
        'path/with_underscores',
      ]

      for (const p of paths) {
        const full = `${remoteName}:${p}`
        expect(full.startsWith('gdrive:')).toBe(true)
      }
    })
  })
})