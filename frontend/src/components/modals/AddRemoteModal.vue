<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import Modal from './Modal.vue'
import * as api from '../api'
import type { Provider, ProviderOption } from '../types'
import { t, locale } from '../i18n'

const props = defineProps<{
  show: boolean
  editMode?: boolean
  editName?: string
}>()

const emit = defineEmits<{
  close: []
  success: []
}>()

const step = ref(0)
const search = ref('')
const selectedProviderName = ref('')
const selectedProvider = ref<Provider | null>(null)
const remoteName = ref('')
const remoteOptions = ref<Record<string, string>>({})
const creating = ref(false)
const success = ref(false)
const showAdvancedOptions = ref(false)
const providerNeedAuth = ref(false)
const providers = ref<Provider[]>([])

const providerDescriptionMap: Record<string, string> = {
  alias: '别名 / 路径映射',
  azureblob: 'Azure Blob 对象存储',
  b2: 'Backblaze B2 对象存储',
  cache: '缓存层',
  chunker: '分片包装存储',
  compress: '压缩包装存储',
  crypt: '加密包装存储',
  drive: 'Google Drive 网盘',
  dropbox: 'Dropbox 网盘',
  ftp: 'FTP 文件服务器',
  hasher: '哈希包装存储',
  hdfs: 'HDFS 分布式存储',
  http: 'HTTP 只读文件源',
  internetarchive: 'Internet Archive 存储',
  jottacloud: 'Jottacloud 网盘',
  koofr: 'Koofr 网盘',
  local: '本地磁盘目录',
  mailru: 'Mail.ru 云盘',
  mega: 'MEGA 网盘',
  memory: '内存存储',
  onedrive: 'OneDrive 网盘',
  opendrive: 'OpenDrive 网盘',
  oracleobjectstorage: 'Oracle 对象存储',
  pcloud: 'pCloud 网盘',
  pikpak: 'PikPak 网盘',
  premiumizeme: 'Premiumize 云存储',
  protondrive: 'Proton Drive 网盘',
  putio: 'Put.io 云存储',
  qingstor: '青云 QingStor 对象存储',
  quatrix: 'Quatrix 文件传输',
  s3: 'S3 兼容对象存储',
  seafile: 'Seafile 文件库',
  sftp: 'SFTP / SSH 文件服务器',
  sharefile: 'Citrix ShareFile',
  smb: 'SMB / Windows 共享',
  storj: 'Storj 去中心化对象存储',
  swift: 'OpenStack Swift 对象存储',
  union: '联合视图存储',
  uptobox: 'Uptobox 网盘',
  webdav: 'WebDAV 文件服务',
  yandex: 'Yandex Disk 网盘',
  zoho: 'Zoho WorkDrive',
}

const optionLabelMap: Record<string, string> = {
  type: '存储类型',
  provider: '服务商',
  token: '授权令牌',
  client_id: '客户端 ID',
  client_secret: '客户端密钥',
  client_credentials: '客户端凭据',
  tenant: '租户 ID',
  tenant_id: '租户 ID',
  drive_id: '云盘 ID',
  drive_type: '云盘类型',
  root_folder_id: '根目录 ID',
  root_folder_path: '根目录路径',
  project_number: '项目编号',
  project_id: '项目 ID',
  service_account_file: '服务账号文件',
  service_account_credentials: '服务账号凭据',
  region: '区域',
  location: '区域',
  endpoint: '接入点',
  endpoint_url: '接入点 URL',
  account: '账号',
  account_id: '账号 ID',
  account_name: '账号名称',
  user: '用户名',
  username: '用户名',
  pass: '密码',
  password: '密码',
  password2: '密码 2 / 盐值口令',
  key: '密钥',
  secret: '密钥',
  secret_access_key: 'Secret Key',
  access_key_id: 'Access Key ID',
  sas_url: 'SAS 链接',
  bucket: '桶名称',
  container: '容器名称',
  host: '主机地址',
  port: '端口',
  url: '地址 URL',
  site: '站点地址',
  auth_url: '认证地址',
  auth_endpoint: '认证端点',
  token_url: 'Token 地址',
  access_scopes: '访问范围',
  disable_site_permission: '禁用站点权限',
  expose_onenote_files: '显示 OneNote 文件',
  no_versions: '不保留历史版本',
  link_scope: '链接范围',
  link_type: '链接类型',
  link_password: '链接密码',
  hash_type: '哈希类型',
  bearer_token: 'Bearer Token',
  bearer_token_command: 'Bearer Token 命令',
  headers: '请求头',
  nextcloud_chunk_size: 'Nextcloud 分片大小',
  owncloud_exclude_shares: '排除 ownCloud 共享',
  owncloud_exclude_mounts: '排除 ownCloud 挂载',
  unix_socket: 'Unix Socket',
  auth_redirect: '认证重定向',
  nounc: '禁用 UNC 转换',
  copy_links: '复制链接目标',
  links: '链接转普通文件',
  skip_links: '跳过链接且不警告',
  skip_specials: '跳过特殊文件且不警告',
  zero_size_links: '链接视为零大小',
  unicode_normalization: 'Unicode 归一化',
  no_check_updated: '不检查上传时更新',
  one_file_system: '单一文件系统',
  case_sensitive: '区分大小写',
  case_insensitive: '不区分大小写',
  no_clone: '禁用克隆',
  remote: '远程存储路径',
  filename_encryption: '文件名加密',
  directory_name_encryption: '目录名加密',
  show_mapping: '显示映射关系',
  no_data_encryption: '不加密文件内容',
  pass_bad_blocks: '传递坏块',
  strict_names: '严格名称检查',
  filename_encoding: '文件名编码',
  suffix: '后缀',
  scope: '授权范围',
  team_drive: '团队盘',
  shared_with_me: '共享给我',
  shared_files: '共享文件',
  directory: '目录',
  directory_id: '目录 ID',
  encoding: '编码设置',
  chunk_size: '分片大小',
  upload_cutoff: '直传阈值',
  copy_cutoff: '复制阈值',
  pacer_min_sleep: '限速器最小等待',
  pacer_burst: '限速器突发值',
  hard_delete: '彻底删除',
  no_check_bucket: '不检查桶',
  use_mmap: '使用内存映射',
  disable_http2: '禁用 HTTP/2',
  server_side_across_configs: '跨配置服务端复制',
  env_auth: '使用环境变量认证',
  acl: 'ACL 权限',
  storage_class: '存储类型',
  max_upload_parts: '最大上传分片数',
  memory_pool_flush_time: '内存池刷新时间',
  directory_markers: '目录标记',
  no_head: '不使用 HEAD',
  no_traverse: '不遍历',
  domain: '域 / 工作组',
  spn: '服务主体名称',
  use_kerberos: '启用 Kerberos',
  idle_timeout: '空闲超时',
  hide_special_share: '隐藏特殊共享',
  kerberos_ccache: 'Kerberos 凭据缓存',
  description: '说明',
}

const optionHelpMap: Record<string, string> = {
  token: '授权后生成的访问令牌；需要网页授权的存储通常会用到。',
  client_id: '应用的客户端 ID；留空通常表示使用官方默认值。',
  client_secret: '应用的客户端密钥；留空通常表示使用官方默认值。',
  client_credentials: '使用 OAuth2 客户端凭据模式。并非所有后端都支持。',
  scope: '授权范围；多数情况下保持默认即可。',
  provider: '选择具体服务商或接口兼容实现。',
  region: '对象存储所在区域，例如 ap-east-1、us-east-1。',
  endpoint: '自定义对象存储接入点，常用于 S3 兼容服务。',
  endpoint_url: '自定义服务地址 URL。',
  access_key_id: '对象存储访问密钥 ID。',
  secret_access_key: '对象存储访问密钥 Secret。',
  bucket: '对象存储的桶名称。',
  container: '对象存储的容器名称。',
  account: '账号或账户名。',
  account_id: '服务端分配的账号 ID。',
  account_name: '服务端分配的账号名称。',
  user: '登录用户名。',
  username: '登录用户名。',
  pass: '登录密码。',
  password: '登录密码。',
  key: '访问密钥或 API Key。',
  secret: '访问密钥对应的 Secret。',
  sas_url: 'Azure 等服务提供的 SAS 授权链接。',
  url: '服务访问地址，例如 https://example.com/dav。',
  host: '服务器主机名或 IP 地址。',
  port: '服务监听端口。',
  root_folder_id: '根目录 ID；不知道时通常留空。',
  root_folder_path: '根目录路径；留空表示使用默认根目录。',
  drive_id: '要使用的云盘 ID。',
  drive_type: '云盘类型，例如 personal、business 或 documentLibrary。',
  access_scopes: '设置 rclone 请求的 OAuth scopes。可选择预设值，或手动输入以空格分隔的自定义 scopes 列表。',
  disable_site_permission: '禁用 Sites.Read.All 权限请求。开启后将无法在配置 drive ID 时搜索 SharePoint 站点。',
  expose_onenote_files: '在目录列表中显示 OneNote 文件。默认会隐藏，因为很多打开/更新操作对它们无效。',
  no_versions: '修改文件时移除旧版本，只保留最后一个版本，可减少 OneDrive for Business 的版本占用空间。',
  link_scope: '设置 link 命令创建的分享链接范围，例如 anonymous。',
  link_type: '设置 link 命令创建的链接类型，例如 view 只读链接。',
  link_password: '设置 link 命令创建的链接密码。目前通常仅 OneDrive 个人付费版支持。',
  hash_type: '指定 OneDrive 后端使用的哈希类型。设为 auto 时会自动选择最佳哈希；较新的 OneDrive 通常默认使用 QuickXorHash，也可设为 none 表示不使用哈希。',
  av_override: '允许下载被 OneDrive / SharePoint 服务器判定为含病毒的文件。仅在你 100% 确认文件安全、确实需要继续下载时再开启。',
  delta: '启用 delta listing 以加速递归列表、size、vfs/refresh 等操作。该能力只在云盘根目录附近最划算；若数据多数不在 rclone 根目录下，反而可能更慢。',
  metadata_permissions: '控制是否在元数据中读取或写入权限信息。读取通常很快，但并不总是希望按元数据恢复权限。',
  token_url: 'Token 服务器地址。留空时使用 provider 默认值。',
  service_account_file: '服务账号 JSON 文件路径。',
  service_account_credentials: '服务账号 JSON 内容。',
  project_number: 'Google Cloud 项目编号。',
  project_id: 'Google Cloud 项目 ID。',
  team_drive: '团队盘/共享盘 ID。',
  shared_with_me: '启用后显示“与我共享”的内容。',
  shared_files: '启用后显示共享文件。',
  directory: '目录路径。',
  directory_id: '目录 ID。',
  encoding: '文件名编码规则；通常保持默认。',
  chunk_size: '大文件上传时每个分片的大小。',
  upload_cutoff: '小于该大小时直接上传，大于时使用分片上传。',
  copy_cutoff: '服务端复制的阈值。',
  pacer_min_sleep: '请求限速器的最小等待时间。',
  pacer_burst: '限速器允许的突发请求数量。',
  hard_delete: '删除时不进回收站，直接彻底删除。',
  no_check_bucket: '跳过桶存在性检查。',
  use_mmap: '使用内存映射文件提高部分场景性能。',
  disable_http2: '禁用 HTTP/2；只有兼容性问题时再开启。',
  server_side_across_configs: '允许跨配置执行服务端复制。',
  env_auth: '优先从环境变量中读取认证信息。',
  acl: '上传对象的默认访问权限。',
  storage_class: '对象存储默认存储类型。',
  max_upload_parts: '分片上传时允许的最大分片数。',
  memory_pool_flush_time: '内存池定时刷新时间。',
  directory_markers: '为目录写入占位标记对象。',
  no_head: '跳过 HEAD 请求；仅在兼容性有问题时使用。',
  no_traverse: '跳过远端遍历，可减少部分场景请求量。',
  disable_checksum: '禁用将 MD5 校验和写入对象元数据。通常 rclone 会在上传前计算 MD5 以做完整性校验，但对大文件可能会明显拖慢开始上传的时间。',
  shared_credentials_file: '共享凭据文件路径。若 env_auth=true，可从该文件读取凭据；留空时会尝试环境变量 AWS_SHARED_CREDENTIALS_FILE，否则回退到用户主目录下的默认凭据文件。',
  profile: '共享凭据文件中使用的 profile 名称。留空时会尝试环境变量 AWS_PROFILE，否则默认使用 default。',
  session_token: 'AWS 会话令牌。',
  role_arn: '要承担的 IAM 角色 ARN。不使用角色扮演时可留空。',
  role_session_name: '承担角色时使用的会话名称；留空则自动生成。',
  role_session_duration: '承担角色时使用的会话时长；留空则使用默认时长。',
  role_external_id: '承担角色时使用的 External ID；不用时留空。',
  upload_concurrency: '分片上传或复制同一文件时的并发块数。若高速链路下上传少量大文件仍跑不满带宽，可适当调高。',
  force_path_style: '是否强制使用 path style 访问。默认通常为 true；某些提供商会根据 provider 自动调整。如果桶名不符合 DNS 规范，通常需要启用。',
  v2_auth: '是否使用 v2 认证。默认使用 v4；仅在 v4 签名不可用时才建议开启。',
  use_dual_stack: '是否使用 AWS S3 双栈端点（支持 IPv6）。',
  use_arn_region: '是否启用 ARN 区域支持。',
  list_chunk: '列表请求每次返回的分块大小。不同 provider 的上限可能不同。',
  list_version: 'ListObjects 接口版本：1、2，或 0 表示自动。通常优先建议使用性能更好的 V2。',
  list_url_encode: '是否对列表结果做 URL 编码：true / false / unset。某些服务在文件名包含控制字符时会更可靠。',
  no_head_object: '获取对象时跳过 GET 前的 HEAD 请求。',
  memory_pool_use_mmap: '是否在内部内存池中使用 mmap 缓冲区。（当前已基本不再使用）',
  download_url: '自定义下载端点。常用于配置 CloudFront 等 CDN 下载地址。',
  use_multipart_etag: '是否在分片上传校验时使用 ETag。可设为 true / false / unset。',
  use_unsigned_payload: '是否在 PutObject 时使用未签名负载。某些提供商可用它绕开 AWS SDK 对请求体 seek 的要求。',
  use_presigned_request: '单分片上传时是否使用预签名请求。通常不需要，除非特殊兼容场景或测试。',
  use_data_integrity_protections: '是否启用 AWS S3 数据完整性保护。',
  versions: '在目录列表中包含旧版本对象。',
  version_at: '查看指定时间点的文件版本。可填日期、日期时间或相对时长；启用后通常不能执行上传/删除等写操作。',
  version_deleted: '启用版本功能时显示已删除标记。它们会以 0 字节文件显示，仅可执行删除操作。',
  decompress: '下载时自动解压 gzip 编码对象。启用后文件内容会被解压，但大小和哈希校验可能不可用。',
  might_gzip: '当后端可能对对象自动 gzip 时启用。可减少某些 provider 因透明压缩导致的大小不一致问题。',
  use_accept_encoding_gzip: '是否发送 Accept-Encoding: gzip 请求头。某些服务会因此改写响应头并导致签名不匹配，必要时可关闭。',
  no_system_metadata: '禁止设置和读取系统元数据。',
  use_already_exists: '控制在创建桶时如何处理 BucketAlreadyExists / AlreadyOwnedByYou 等语义差异。通常无需手动修改。',
  use_multipart_uploads: '是否使用分片上传。一般不需要改，除非明确想禁用。',
  use_x_id: '是否附加 x-id URL 参数。一般不需要改，除非兼容性排障。',
  sign_accept_encoding: '是否将 Accept-Encoding 纳入签名计算。一般不需要改。',
  sdk_log_mode: 'AWS SDK 日志模式。可用于调试 Signing / Retries / Request / Response 等行为。',
  domain: 'NTLM 认证使用的域或工作组名称。',
  spn: '服务主体名称（SPN）。某些集群或特殊环境会要求填写，例如 cifs/remotehost:1020；不确定时留空。',
  use_kerberos: '启用 Kerberos 认证。开启后会改用 Kerberos 而不是 NTLM，并要求系统中存在可用的 Kerberos 配置与凭据缓存。',
  idle_timeout: '空闲连接关闭前的最长等待时间。设为 0 表示长期保持连接。',
  hide_special_share: '隐藏特殊共享（例如 print$）等通常不应直接访问的共享。',
  kerberos_ccache: 'Kerberos 凭据缓存（krb5cc）路径。可覆盖默认的 KRB5CCNAME，用于多凭据或多用户场景。支持 FILE:/path、DIR:/path 或直接填写文件路径。',
  bearer_token: 'Bearer token 认证，替代用户名/密码（例如 Macaroon）。',
  bearer_token_command: '用于获取 bearer token 的命令。',
  headers: '为所有请求设置额外的 HTTP 请求头。格式为逗号分隔的 key,value 列表，必要时可使用 CSV 编码；例如设置 Cookie 或 Authorization。',
  nextcloud_chunk_size: 'Nextcloud 上传分片大小。通常建议把服务端最大分片调大以提升大文件上传性能；设为 0 可禁用分片上传。',
  owncloud_exclude_shares: '排除 ownCloud 的共享内容。',
  owncloud_exclude_mounts: '排除 ownCloud 的挂载存储。',
  unix_socket: '改为通过 Unix Domain Socket 连接，而不是直接建立 TCP 连接。',
  auth_redirect: '在重定向时保留认证信息。仅在某些 WebDAV 服务读取文件时出现 401 且确有需要时再开启。',
  nounc: '在 Windows 上禁用 UNC（长路径）转换。',
  copy_links: '跟随符号链接，并复制其指向的目标。',
  links: '在 local 后端中，将符号链接与带 .rclonelink 扩展名的普通文件相互转换。',
  skip_links: '跳过符号链接或 junction 点时不再输出警告，表示你明确接受跳过这些链接。',
  skip_specials: '跳过管道、套接字和设备对象时不再输出警告。',
  zero_size_links: '将链接的 Stat 大小视为 0 并读取其内容（已废弃，一般无需开启）。',
  unicode_normalization: '对路径和文件名应用 Unicode NFC 归一化。通常仅在 macOS 等提供 NFD 文件名的场景下按需使用。',
  no_check_updated: '上传时不检查文件是否仍在变化。仅在某些文件系统修改时间检查失效时使用，否则可能导致传输内容不一致。',
  one_file_system: '不要跨越文件系统边界（仅 Unix/macOS）。',
  case_sensitive: '强制或声明按区分大小写处理路径。local 中用于覆盖本地文件系统默认判断；某些远端中表示服务端行为。',
  case_insensitive: '强制或声明按不区分大小写处理路径。local 中用于覆盖本地文件系统默认判断；某些远端中表示服务端行为。',
  no_clone: '禁用 clone / copy_file_range 等服务端拷贝优化。',
  remote: '要进行加密/解密的远程存储路径。通常应包含冒号和路径，例如 myremote:path/to/dir、myremote:bucket，或不太推荐的 myremote:。',
  filename_encryption: '如何加密文件名。standard 表示对文件名进行加密。',
  directory_name_encryption: '是否加密目录名。若 filename_encryption 为 off，则该选项不会生效。',
  password2: '用于盐值的第二个密码/口令。可选但推荐，并且应与前一个密码不同。',
  show_mapping: '列出文件时显示加密前后的名称映射，便于排障或核对加密文件名。',
  no_data_encryption: '是否不加密文件内容；关闭时会加密文件数据。',
  pass_bad_blocks: '遇到坏块时将其按全 0 继续传递。仅在抢救损坏的加密文件时使用，正常情况下不要开启。',
  strict_names: '遇到无法解密的文件名时直接报错，而不是仅记录日志继续处理。',
  filename_encoding: '加密后的文件名使用何种文本编码。可影响文件名长度，也与远端是否区分大小写有关。',
  suffix: '覆盖默认的 .bin 后缀；设为 none 表示不使用后缀。路径长度敏感时可能有用。',
  description: '该存储的备注说明。',
}

function formatOptionKey(name: string) {
  return name.replace(/_/g, ' ').replace(/\b\w/g, (m) => m.toUpperCase())
}

function getProviderDescription(provider: Provider) {
  if (locale.value === 'zh') {
    return providerDescriptionMap[provider.Name] || '该类型支持通过 rclone 接入。'
  }
  return provider.Description || 'This storage type is supported by rclone.'
}

function getOptionLabel(option: ProviderOption) {
  if (locale.value === 'zh') {
    return optionLabelMap[option.Name] || '配置项'
  }
  return option.Name ? formatOptionKey(option.Name) : 'Option'
}

function getOptionHelp(option: ProviderOption) {
  if (locale.value === 'zh') {
    return optionHelpMap[option.Name] || '请根据你的存储服务提供的信息填写。'
  }
  return option.Help || 'Fill this based on your storage provider documentation.'
}

function getOptionPlaceholder(option: ProviderOption) {
  if (locale.value === 'zh') {
    return `${t('remote.enterPrefix')}${getOptionLabel(option)}`
  }
  return `Enter ${getOptionLabel(option)}`
}

function getExampleHelp(help?: string) {
  if (!help) return ''
  const normalized = help.trim()
  if (locale.value !== 'zh') return normalized
  const map: Record<string, string> = {
    'string': '文本',
    'true': '是',
    'false': '否',
    'auto': '自动',
    'standard': '标准',
    'private': '私有',
    'public-read': '公开可读',
  }
  return map[normalized] || normalized
}

const filteredProviders = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  return (providers.value || [])
    .filter(p => {
      if (!keyword) return true
      const haystack = locale.value === 'zh'
        ? `${p.Name} ${p.Description || ''} ${getProviderDescription(p)}`
        : `${p.Name} ${p.Description || ''}`
      return haystack.toLowerCase().includes(keyword)
    })
    .sort((a, b) => a.Name.localeCompare(b.Name))
})

const allOptions = computed(() => selectedProvider.value?.Options || [])

function resetProviderOptions() {
  if (!selectedProvider.value) return
  remoteOptions.value = Object.fromEntries(
    selectedProvider.value.Options.map((option) => [option.Name, option.DefaultStr || ''])
  )
}

function shouldShowOption(option: ProviderOption) {
  if (option.Advanced && !showAdvancedOptions.value) return false
  if (
    selectedProvider.value?.Name === 's3' &&
    option.Provider &&
    option.Provider !== remoteOptions.value['provider']
  ) {
    return false
  }
  return true
}

const groupedOptions = computed(() => {
  const required: ProviderOption[] = []
  const optional: ProviderOption[] = []
  const advanced: ProviderOption[] = []
  for (const option of allOptions.value) {
    if (!shouldShowOption(option)) continue
    if (option.Advanced) advanced.push(option)
    else if (option.Required) required.push(option)
    else optional.push(option)
  }
  return { required, optional, advanced }
})

watch(() => props.show, async (val) => {
  if (!val) return
  step.value = props.editMode ? 1 : 0
  search.value = ''
  selectedProviderName.value = ''
  selectedProvider.value = null
  remoteName.value = props.editName || ''
  remoteOptions.value = {}
  creating.value = false
  success.value = false
  showAdvancedOptions.value = false
  providerNeedAuth.value = false
  const data = await api.listProviders()
  providers.value = data.providers || []
})

function selectProvider(provider: Provider) {
  selectedProvider.value = provider
  selectedProviderName.value = provider.Name
  providerNeedAuth.value = provider.Options.some((option) => option.Name === 'token')
  resetProviderOptions()
  step.value = 1
}

function nextStep() {
  if (step.value === 0) step.value = 1
  else if (step.value === 1) step.value = props.editMode ? 3 : 2
  else if (step.value === 2) step.value = 3
}

function prevStep() {
  if (step.value === 3) step.value = props.editMode ? 1 : 2
  else if (step.value === 2) step.value = 1
  else if (step.value === 1) step.value = 0
}

function validateName(name: string) {
  return !!name && name.length >= 2 && /^[a-zA-Z0-9_-]*$/.test(name)
}

async function create() {
  creating.value = true
  try {
    if (!selectedProvider.value?.Name) throw new Error(t('remote.notSelected'))
    if (!validateName(remoteName.value)) {
      throw new Error(t('remote.invalidName'))
    }
    const params: Record<string, unknown> = Object.fromEntries(
      Object.entries(remoteOptions.value).filter(([, value]) => value !== '')
    )
    if (providerNeedAuth.value && !params['token']) {
      params['token'] = ''
    }
    if (props.editMode && props.editName) {
      await api.updateRemote(remoteName.value, selectedProvider.value.Name, params)
    } else {
      await api.createRemote(remoteName.value, selectedProvider.value.Name, params)
    }
    success.value = true
    setTimeout(() => {
      emit('success')
      emit('close')
    }, 800)
  } catch (e) {
    alert((e as Error).message)
  } finally {
    creating.value = false
  }
}

defineExpose({
  loadConfig: async (name: string) => {
    while (!providers.value.length) {
      await new Promise(resolve => setTimeout(resolve, 50))
    }
    const config = await api.getRemoteConfig(name)
    remoteName.value = name
    const type = String(config.type || '')
    const provider = providers.value.find(p => p.Name === type)
    if (!provider) return
    selectedProvider.value = provider
    selectedProviderName.value = type
    providerNeedAuth.value = provider.Options.some((option) => option.Name === 'token')
    resetProviderOptions()
    for (const key in config) {
      if (key !== 'type' && key !== 'name') {
        remoteOptions.value[key] = String(config[key])
      }
    }
    step.value = 1
  },
})
</script>

<template>
  <Modal :show="show" :title="editMode ? t('remote.editTitle') : t('remote.addTitle')" @close="emit('close')">
    <div class="add-remote-body">
      <div v-if="!editMode" class="stepper">
      <div class="step" :class="{ active: step === 0, done: step > 0 }">1. {{ t('remote.stepChooseType') }}</div>
      <div class="step" :class="{ active: step === 1, done: step > 1 }">2. {{ t('remote.stepConfig') }}</div>
      <div class="step" :class="{ active: step === 2, done: step > 2 }">3. {{ t('remote.stepName') }}</div>
      <div class="step" :class="{ active: step === 3 }">4. {{ t('remote.stepSave') }}</div>
    </div>
    <div v-else class="stepper">
      <div class="step" :class="{ active: step === 1, done: step > 1 }">1. {{ t('remote.stepConfig') }}</div>
      <div class="step" :class="{ active: step === 3 }">2. {{ t('remote.stepConfirm') }}</div>
    </div>

    <div v-if="step === 0">
      <input v-model="search" type="text" :placeholder="t('remote.searchPlaceholder')" style="width: 100%; margin-bottom: 16px" />
      <div class="provider-grid">
        <div
          v-for="p in filteredProviders"
          :key="p.Name"
          class="provider-card"
          :class="{ selected: selectedProviderName === p.Name }"
          @click="selectProvider(p)"
        >
          <strong>{{ p.Name }}</strong>
          <div style="margin-top: 6px; color: #6b7280; font-size: 12px">{{ getProviderDescription(p) }}</div>
        </div>
      </div>
    </div>

    <div v-if="step === 1">
      <div class="field-grid" v-if="groupedOptions.required.length">
        <div v-for="opt in groupedOptions.required" :key="opt.Name" class="field-item">
          <label>{{ getOptionLabel(opt) }} <small v-if="getOptionLabel(opt) !== opt.Name" class="subkey">{{ opt.Name }}</small> <span style="color: #dc2626">*</span></label>
          <input
            v-if="!opt.Examples || !opt.Examples.length"
            v-model="remoteOptions[opt.Name]"
            :type="opt.IsPassword ? 'password' : 'text'"
            :placeholder="getOptionPlaceholder(opt)"
          />
          <select v-else v-model="remoteOptions[opt.Name]">
            <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ getExampleHelp(ex.Help) ? ` — ${getExampleHelp(ex.Help)}` : '' }}</option>
          </select>
          <div class="field-help">{{ getOptionHelp(opt) }}</div>
        </div>
      </div>

      <div class="field-grid" v-if="groupedOptions.optional.length" style="margin-top: 16px">
        <div v-for="opt in groupedOptions.optional" :key="opt.Name" class="field-item">
          <label>{{ getOptionLabel(opt) }} <small v-if="getOptionLabel(opt) !== opt.Name" class="subkey">{{ opt.Name }}</small></label>
          <input
            v-if="!opt.Examples || !opt.Examples.length"
            v-model="remoteOptions[opt.Name]"
            :type="opt.IsPassword ? 'password' : 'text'"
            :placeholder="getOptionPlaceholder(opt)"
          />
          <select v-else v-model="remoteOptions[opt.Name]">
            <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ getExampleHelp(ex.Help) ? ` — ${getExampleHelp(ex.Help)}` : '' }}</option>
          </select>
          <div class="field-help">{{ getOptionHelp(opt) }}</div>
        </div>
      </div>

      <details v-if="allOptions.some(o => o.Advanced)" style="margin-top: 16px" :open="showAdvancedOptions">
        <summary style="cursor: pointer; color: #6b7280; font-size: 14px" @click.prevent="showAdvancedOptions = !showAdvancedOptions">
          {{ showAdvancedOptions ? t('remote.hideAdvanced') : t('remote.showAdvanced') }}
        </summary>
        <div class="field-grid" v-if="groupedOptions.advanced.length" style="margin-top: 12px">
          <div v-for="opt in groupedOptions.advanced" :key="opt.Name" class="field-item">
            <label>{{ getOptionLabel(opt) }} <small v-if="getOptionLabel(opt) !== opt.Name" class="subkey">{{ opt.Name }}</small></label>
            <input
              v-if="!opt.Examples || !opt.Examples.length"
              v-model="remoteOptions[opt.Name]"
              :type="opt.IsPassword ? 'password' : 'text'"
              :placeholder="getOptionPlaceholder(opt)"
            />
            <select v-else v-model="remoteOptions[opt.Name]">
              <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ getExampleHelp(ex.Help) ? ` — ${getExampleHelp(ex.Help)}` : '' }}</option>
            </select>
            <div class="field-help">{{ getOptionHelp(opt) }}</div>
          </div>
        </div>
      </details>

      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button v-if="!editMode" class="ghost" @click="prevStep">{{ t('remote.previous') }}</button>
        <button @click="nextStep">{{ t('remote.next') }}</button>
      </div>
    </div>

    <div v-if="step === 2 && !editMode">
      <div class="field-item">
        <label>{{ t('remote.nameLabel') }} <span style="color: #dc2626">*</span></label>
        <input v-model="remoteName" type="text" :placeholder="t('remote.namePlaceholder')" />
        <div class="field-help">{{ t('remote.nameHelp') }}</div>
      </div>
      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button class="ghost" @click="prevStep">{{ t('remote.previous') }}</button>
        <button @click="nextStep" :disabled="!validateName(remoteName)">{{ t('remote.next') }}</button>
      </div>
    </div>

    <div v-if="step === 3">
      <div class="card" style="background: #f0f9ff; margin-bottom: 16px">
        <div><strong>{{ t('remote.typeLabel') }}:</strong> {{ selectedProvider?.Name }}</div>
        <div><strong>{{ t('remote.nameLabel') }}:</strong> {{ remoteName }}</div>
      </div>
      <div class="actions" style="justify-content: space-between">
        <button v-if="!editMode" class="ghost" @click="prevStep" :disabled="success">{{ t('remote.previous') }}</button>
        <button v-if="!success" color="primary" :disabled="creating" @click="create">
          {{ creating ? t('remote.saving') : (editMode ? t('remote.saveEdit') : t('remote.save')) }}
        </button>
        <button v-else color="primary" @click="emit('close')">
          {{ editMode ? t('remote.editSuccessClose') : t('remote.saveSuccessClose') }}
        </button>
      </div>
    </div>
    </div>
  </Modal>
</template>

<style scoped>
.add-remote-body {
  display: flex;
  flex-direction: column;
  min-height: 0;
  max-height: calc(85vh - 88px);
  overflow-y: auto;
}
.stepper { display: flex; gap: 8px; margin-bottom: 20px; }
.step { padding: 8px 12px; border-radius: 8px; background: #1f2937; color: #cbd5e1; font-size: 14px; border:1px solid #374151 }
.step.active { background: #1e3a5f; color: #60a5fa; border-color:#2563eb }
.step.done { background: #0a2f22; color: #34d399; border-color:#14532d }
body.light .step { background: #f3f4f6; color: #6b7280; border-color:#e5e7eb }
body.light .step.active { background: #dbeafe; color: #1d4ed8; border-color:#93c5fd }
body.light .step.done { background: #dcfce7; color: #166534; border-color:#86efac }
.provider-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 12px; max-height: 360px; overflow: auto; }
.provider-card { padding: 14px; border: 1px solid #374151; border-radius: 12px; cursor: pointer; background: #1f2937; color: #e5e7eb; }
.provider-card:hover { border-color: #60a5fa; background: #111827; }
.provider-card.selected { border-color: #2563eb; background: #111827; }
body.light .provider-card { border-color: #e5e7eb; background: #fff; color: #111827; }
body.light .provider-card:hover { border-color: #93c5fd; background: #eff6ff; }
body.light .provider-card.selected { border-color: #2563eb; background: #eff6ff; }
.field-grid { display: grid; gap: 16px; }
.field-item { display: grid; gap: 8px; }
.field-item label { font-weight: 600; }
.field-item input, .field-item select { width: 100%; padding: 10px 12px; border-radius: 10px; border: 1px solid #d1d5db; font: inherit; }
.field-help { font-size: 12px; color: #6b7280; line-height: 1.5; }
.subkey { opacity: .65; margin-left: 6px; font-weight: 400; }
.actions { display: flex; gap: 8px; }
button { padding: 10px 14px; border-radius: 10px; border: none; background: #2563eb; color: white; cursor: pointer; }
button.ghost { background: #eef2ff; color: #1d4ed8; }
button:disabled { opacity: .6; cursor: not-allowed; }

@media (max-width: 768px) {
  .add-remote-body {
    max-height: calc(85vh - 72px);
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    padding-right: 2px;
  }

  .stepper {
    flex-wrap: wrap;
    margin-bottom: 16px;
  }

  .provider-grid {
    grid-template-columns: 1fr;
    max-height: none;
  }

  .actions {
    flex-wrap: wrap;
  }

  .actions button {
    flex: 1 1 auto;
  }
}
</style>
