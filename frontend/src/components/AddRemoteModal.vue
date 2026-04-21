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
  root_folder_id: '根目录 ID',
  project_number: '项目编号',
  project_id: '项目 ID',
  service_account_file: '服务账号文件',
  service_account_credentials: '服务账号凭据',
  region: '区域',
  location: '区域',
  endpoint: '接入点',
  endpoint_url: '接入点 URL',
  account: '账号',
  user: '用户名',
  username: '用户名',
  pass: '密码',
  password: '密码',
  key: '密钥',
  secret: '密钥',
  secret_access_key: 'Secret Key',
  access_key_id: 'Access Key ID',
  account_id: '账号 ID',
  account_name: '账号名称',
  sas_url: 'SAS 链接',
  bucket: '桶名称',
  container: '容器名称',
  host: '主机地址',
  port: '端口',
  url: '地址 URL',
  site: '站点地址',
  auth_url: '认证地址',
  auth_endpoint: '认证端点',
  scope: '授权范围',
  team_drive: '团队盘',
  shared_with_me: '共享给我',
  shared_files: '共享文件',
  root_folder_path: '根目录路径',
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
  description: '说明',
}

const optionHelpMap: Record<string, string> = {
  token: '授权后生成的访问令牌；需要网页授权的存储通常会用到。',
  client_id: '应用的客户端 ID；留空通常表示使用官方默认值。',
  client_secret: '应用的客户端密钥；留空通常表示使用官方默认值。',
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
  force_path_style: '是否强制使用 path style 访问。默认通常为 true；某些提供商会根据 provider 自动调整。如果桶名不符合 DNS 规范（如包含 . 或 _），通常需要启用。',
  v2_auth: '是否使用 v2 认证。默认使用 v4；仅在 v4 签名不可用时才建议开启。',
  use_dual_stack: '是否使用 AWS S3 双栈端点（支持 IPv6）。',
  use_arn_region: '是否启用 ARN 区域支持。',
  list_chunk: '每次 ListObject 请求返回的列表分块大小。多数服务即使设更大也会截断到 1000。',
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
    return `请输入${getOptionLabel(option)}`
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
  </Modal>
</template>

<style scoped>
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
</style>
