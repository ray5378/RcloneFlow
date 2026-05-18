export interface Task {
  id: number
  name: string
  mode: string
  sourceRemote: string
  sourcePath: string
  targetRemote: string
  targetPath: string
  options?: TaskOptions
  sortOrder?: number
  createdAt: string
}

// 高级任务选项
export interface TaskOptions {
  enableStreaming?: boolean
  // 过滤参数
  exclude?: string[]           // --exclude
  excludeFrom?: string[]       // --exclude-from
  excludeIfPresent?: string[]  // --exclude-if-present
  include?: string[]           // --include
  includeFrom?: string[]       // --include-from
  filter?: string[]            // --filter
  filterFrom?: string[]        // --filter-from
  filesFrom?: string[]         // --files-from
  filesFromRaw?: string[]      // --files-from-raw
  minSize?: string             // --min-size
  maxSize?: string             // --max-size
  minAge?: string              // --min-age
  maxAge?: string              // --max-age
  deleteExcluded?: boolean     // --delete-excluded
  ignoreCase?: boolean         // --ignore-case
  ignoreCaseSync?: boolean     // --ignore-case-sync
  ignoreExisting?: boolean     // --ignore-existing

  // 比较策略
  checksum?: boolean           // --checksum
  sizeOnly?: boolean           // --size-only
  ignoreSize?: boolean         // --ignore-size
  ignoreTimes?: boolean        // --ignore-times
  update?: boolean             // --update
  useServerModtime?: boolean   // --use-server-modtime
  modifyWindow?: string        // --modify-window
  refreshTimes?: boolean       // --refresh-times

  // 路径策略
  noTraverse?: boolean         // --no-traverse
  noCheckDest?: boolean        // --no-check-dest
  compareDest?: string         // --compare-dest
  copyDest?: string            // --copy-dest

  // 元数据
  ignoreChecksum?: boolean     // --ignore-checksum

  // 多线程
  multiThreadStreams?: number  // --multi-thread-streams
  multiThreadCutoff?: number   // --multi-thread-cutoff
  multiThreadWriteBufferSize?: number // --multi-thread-write-buffer-size

  // 其他复制标志
  inplace?: boolean            // --inplace
  immutable?: boolean          // --immutable
  orderBy?: string             // --order-by
  maxTransfer?: number        // --max-transfer
  maxDuration?: number        // --max-duration
  checkFirst?: boolean        // --check-first

  // 备份
  backupDir?: string          // --backup-dir
  suffix?: string             // --suffix

  // 传输控制
  cutoffMode?: string          // --cutoff-mode

  // 同步标志
  deleteBefore?: boolean       // --delete-before
  deleteDuring?: boolean       // --delete-during
  deleteAfter?: boolean        // --delete-after
  maxDelete?: number           // --max-delete
  maxDeleteSize?: number      // --max-delete-size
  trackRenames?: boolean       // --track-renames
  ignoreErrors?: boolean       // --ignore-errors

  // 网络参数
  bwLimit?: string             // --bwlimit
  bwLimitFile?: string         // --bwlimit-file
  bind?: string               // --bind
  connTimeout?: number         // --contimeout
  timeout?: number             // --timeout
  expectContinueTimeout?: number // --expect-continue-timeout
  header?: Record<string, string> // --header
  headerDownload?: Record<string, string> // --header-download
  headerUpload?: Record<string, string>   // --header-upload
  caCert?: string              // --ca-cert
  clientCert?: string          // --client-cert
  clientKey?: string           // --client-key
  noCheckCertificate?: boolean // --no-check-certificate
  tpsLimit?: number            // --tpslimit
  tpsLimitBurst?: number       // --tpslimit-burst
  userAgent?: string          // --user-agent
  useCookies?: boolean         // --use-cookies
  disableHttpKeepAlives?: boolean // --disable-http-keep-alives
  dscp?: string               // --dscp

  // 性能参数
  transfers?: number          // --transfers
  checkers?: number           // --checkers
  bufferSize?: number          // --buffer-size (MB)

  // 日志输出
  verbose?: boolean            // --verbose
  quiet?: boolean             // --quiet
  logFile?: string            // --log-file
  logFormat?: string          // --log-format
  humanReadable?: boolean     // --human-readable
  useJsonLog?: boolean        // --use-json-log

  // 配置参数
  configDir?: string           // --config
  cacheDir?: string            // --cache-dir
  tempDir?: string             // --temp-dir
  interactive?: boolean        // --interactive
  dryRun?: boolean             // --dry-run
  autoConfirm?: boolean        // --auto-confirm
  errorOnNoTransfer?: boolean  // --error-on-no-transfer
  retries?: number             // --retries
  lowLevelRetries?: number     // --low-level-retries
  askPassword?: boolean        // --ask-password
  passwordCommand?: string     // --password-command
  useMmap?: boolean            // --use-mmap
  noUnicodeNormalization?: boolean // --no-unicode-normalization
  color?: string               // --color

  // 其他
  serverSideAcrossConfigs?: boolean // --server-side-across-configs
  openlistCasCompatible?: boolean   // 目标端启用 OpenList-CAS .cas 等效原文件兼容
}

export interface Schedule {
  id: number
  taskId: number
  spec: string
  enabled: boolean
  createdAt: string
}

export type { ActiveRun }

import type { ActiveRun, FinalSummary, RunSummaryProgress } from '../api/run'

export interface RunSummaryPayload {
  /**
   * 历史记录里的运行中快照。
   * 仅用于历史视图回看 run 当时的进度帧，不得替代 active runs 主链。
   */
  progress?: RunSummaryProgress
  /**
   * 历史详情 / 最终总结主字段。
   * 不得拿来驱动任务卡片运行中 UI。
   */
  finalSummary?: FinalSummary
  [key: string]: unknown
}

export interface Run {
  id: number
  taskId: number
  status: string
  trigger: string
  startedAt: string
  finishedAt?: string
  taskName?: string
  taskMode?: string
  sourceRemote?: string
  sourcePath?: string
  targetRemote?: string
  targetPath?: string
  bytesTransferred?: number
  speed?: string
  summary?: string | RunSummaryPayload
  error?: string
}

export interface FileItem {
  Path: string
  Name: string
  Size: string
  IsDir: boolean
  ModTime: string
  MimeType?: string
}

export interface Provider {
  Name: string
  Description?: string
  ArchInPath?: boolean
  HashTypes?: string[]
  Flags?: ProviderFlag[]
  Options: ProviderOption[]
}

export interface ProviderFlag {
  Name: string
  Short: string
  Default: unknown
  Provider: unknown
  Required: boolean
  Password: boolean
  Hide: number
  Advanced: boolean
}

export interface ProviderOption {
  Name: string
  Help: string
  Provider?: string
  Default?: unknown
  Required: boolean
  IsPassword: boolean
  Hide: 0 | 2 | 3
  Advanced: boolean
  DefaultStr: string
  ValueStr?: string
  Examples?: ProviderExample[]
}

export interface ProviderExample {
  Value: string
  Help: string
  Provider?: string
}

export type RemoteTestState = 'idle' | 'testing' | 'success' | 'failed'

export interface RemoteDescription {
  [key: string]: string
}
