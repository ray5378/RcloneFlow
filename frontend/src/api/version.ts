import { get } from './client'

export interface VersionInfo {
  commitHash: string
  rcloneVersion: string
}

export async function getVersion(): Promise<VersionInfo> {
  return get<VersionInfo>('/api/version')
}
