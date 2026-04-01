/**
 * Browser API 模块
 * 对应后端 BrowserService - 文件浏览器操作
 */

const BASE = '/api'

async function api<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(BASE + path, {
    headers: { 'Content-Type': 'application/json' },
    ...options,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new Error((data as { error?: string }).error || res.statusText)
  return data as T
}

export interface FileItem {
  Path: string
  Name: string
  Size: string
  IsDir: boolean
  ModTime: string
  MimeType?: string
}

/** 列出目录内容 */
export async function listPath(remote: string, path: string): Promise<{ fs: string; items: FileItem[] }> {
  return api<{ fs: string; items: FileItem[] }>(
    `/browser/list?remote=${encodeURIComponent(remote)}&path=${encodeURIComponent(path)}`
  )
}

/** 复制文件 */
export async function copyFile(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return api('/fs/copy', {
    method: 'POST',
    body: JSON.stringify({
      srcFs: srcRemote + ':',
      srcRemote: srcPath,
      dstFs: dstRemote + ':',
      dstRemote: dstPath,
    }),
  })
}

/** 移动文件 */
export async function moveFile(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return api('/fs/move', {
    method: 'POST',
    body: JSON.stringify({
      srcFs: srcRemote + ':',
      srcRemote: srcPath,
      dstFs: dstRemote + ':',
      dstRemote: dstPath,
    }),
  })
}

/** 复制目录 */
export async function copyDir(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return api('/fs/copyDir', {
    method: 'POST',
    body: JSON.stringify({
      srcFs: srcRemote + ':' + srcPath,
      dstFs: dstRemote + ':' + dstPath,
      createEmptySrcDirs: true,
    }),
  })
}

/** 移动目录 */
export async function moveDir(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return api('/fs/moveDir', {
    method: 'POST',
    body: JSON.stringify({
      srcFs: srcRemote + ':' + srcPath,
      dstFs: dstRemote + ':' + dstPath,
      createEmptySrcDirs: true,
      deleteEmptySrcDirs: true,
    }),
  })
}

/** 删除文件 */
export async function deleteFile(remote: string, path: string): Promise<void> {
  return api('/fs/delete', {
    method: 'POST',
    body: JSON.stringify({ fs: remote + ':', remote: path }),
  })
}

/** 删除目录及内容 */
export async function purgeDir(remote: string, path: string): Promise<void> {
  return api('/fs/purge', {
    method: 'POST',
    body: JSON.stringify({ fs: remote + ':', remote: path }),
  })
}

/** 创建目录 */
export async function mkdir(remote: string, path: string): Promise<void> {
  return api('/fs/mkdir', {
    method: 'POST',
    body: JSON.stringify({ fs: remote + ':', remote: path }),
  })
}
