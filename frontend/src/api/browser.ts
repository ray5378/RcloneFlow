/**
 * Browser API 模块
 * 对应后端 BrowserService - 文件浏览器操作
 */
import { get, post } from './client'

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
  return get<{ fs: string; items: FileItem[] }>(
    `/api/browser/list?remote=${encodeURIComponent(remote)}&path=${encodeURIComponent(path)}`
  )
}

/** 复制文件 */
export async function copyFile(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return post('/api/fs/copy', {
    srcFs: srcRemote + ':',
    srcRemote: srcPath,
    dstFs: dstRemote + ':',
    dstRemote: dstPath,
  })
}

/** 移动文件 */
export async function moveFile(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return post('/api/fs/move', {
    srcFs: srcRemote + ':',
    srcRemote: srcPath,
    dstFs: dstRemote + ':',
    dstRemote: dstPath,
  })
}

/** 复制目录 */
export async function copyDir(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return post('/api/fs/copyDir', {
    srcFs: srcRemote + ':' + srcPath,
    dstFs: dstRemote + ':' + dstPath,
    createEmptySrcDirs: true,
  })
}

/** 移动目录 */
export async function moveDir(
  srcRemote: string,
  srcPath: string,
  dstRemote: string,
  dstPath: string
): Promise<void> {
  return post('/api/fs/moveDir', {
    srcFs: srcRemote + ':' + srcPath,
    dstFs: dstRemote + ':' + dstPath,
    createEmptySrcDirs: true,
    deleteEmptySrcDirs: true,
  })
}

/** 删除文件 */
export async function deleteFile(remote: string, path: string): Promise<void> {
  return post('/api/fs/delete', { fs: remote + ':', remote: path })
}

/** 删除目录及内容 */
export async function purgeDir(remote: string, path: string): Promise<void> {
  return post('/api/fs/purge', { fs: remote + ':', remote: path })
}

/** 创建目录 */
export async function mkdir(remote: string, path: string): Promise<void> {
  return post('/api/fs/mkdir', { fs: remote + ':', remote: path })
}
