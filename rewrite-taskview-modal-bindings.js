const { execSync } = require('child_process')
const fs = require('fs')

let text = execSync('git show dfa47d9:frontend/src/views/TaskView.vue', { encoding: 'utf8' })

text = text.replace(
  "import { useTaskViewPagingBridge } from '../composables/useTaskViewPagingBridge'\n",
  "import { useTaskViewPagingBridge } from '../composables/useTaskViewPagingBridge'\nimport { useTaskViewModalBindings } from '../composables/useTaskViewModalBindings'\n"
)

const modalBlock = `const closeWebhookModal = () => { showWebhookModal.value = false }\nconst closeSingletonModal = () => { showSingletonModal.value = false }\nconst closeLogModal = () => { showLogModal.value = false }\nconst closeGlobalStatsModal = () => { showGlobalStatsModal.value = false }\nfunction ensureWebhookFormShape() {\n  if (!webhookForm.value.notify) {\n    webhookForm.value.notify = { manual: false, schedule: false, webhook: false }\n  }\n  if (!webhookForm.value.status) {\n    webhookForm.value.status = { success: true, failed: true }\n  }\n}\nconst setWebhookTriggerId = (value: string) => { webhookForm.value.triggerId = value }\nconst setWebhookPostUrl = (value: string) => { webhookForm.value.postUrl = value }\nconst setWebhookWecomUrl = (value: string) => { webhookForm.value.wecomUrl = value }\nconst setWebhookNotifyManual = (value: boolean) => { ensureWebhookFormShape(); webhookForm.value.notify.manual = value }\nconst setWebhookNotifySchedule = (value: boolean) => { ensureWebhookFormShape(); webhookForm.value.notify.schedule = value }\nconst setWebhookNotifyWebhook = (value: boolean) => { ensureWebhookFormShape(); webhookForm.value.notify.webhook = value }\nconst setWebhookStatusSuccess = (value: boolean) => { ensureWebhookFormShape(); webhookForm.value.status.success = value }\nconst setWebhookStatusFailed = (value: boolean) => { ensureWebhookFormShape(); webhookForm.value.status.failed = value }\nconst setSingletonEnabled = (value: boolean) => { singletonForm.value.singletonEnabled = value }\nconst setCommandMode = (value: boolean) => { commandMode.value = value }\nconst setCommandText = (value: string) => { commandText.value = value }\nconst setShowAdvancedOptions = (value: boolean) => { showAdvancedOptions.value = value }\n\n`
if (!text.includes(modalBlock)) throw new Error('modal block not found')
text = text.replace(modalBlock, '')

const anchor = `} = useTaskViewPagingBridge({\n  taskSearch,\n  tasksJumpPage,\n  historyStatusFilter,\n  jumpPage,\n  finalFilesJump,\n  tasksPage,\n  runsPage,\n  currentModule,\n  loadData,\n})\n\n`
if (!text.includes(anchor)) throw new Error('paging anchor not found')
text = text.replace(anchor, `${anchor}const {\n  closeWebhookModal,\n  closeSingletonModal,\n  closeLogModal,\n  closeGlobalStatsModal,\n  setWebhookTriggerId,\n  setWebhookPostUrl,\n  setWebhookWecomUrl,\n  setWebhookNotifyManual,\n  setWebhookNotifySchedule,\n  setWebhookNotifyWebhook,\n  setWebhookStatusSuccess,\n  setWebhookStatusFailed,\n  setSingletonEnabled,\n  setCommandMode,\n  setCommandText,\n  setShowAdvancedOptions,\n} = useTaskViewModalBindings({\n  showWebhookModal,\n  webhookForm,\n  showSingletonModal,\n  singletonForm,\n  showLogModal,\n  commandMode,\n  commandText,\n  showAdvancedOptions,\n  showGlobalStatsModal,\n})\n\n`)

fs.writeFileSync('frontend/src/views/TaskView.vue', text)
