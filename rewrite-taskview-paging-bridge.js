const { execSync } = require('child_process')
const fs = require('fs')

let text = execSync('git show 38b27ef:frontend/src/views/TaskView.vue', { encoding: 'utf8' })

const importNeedle = "import { useTaskListView } from '../composables/useTaskListView'\n"
if (!text.includes(importNeedle)) throw new Error('import needle not found')
text = text.replace(
  importNeedle,
  importNeedle + "import { useTaskViewPagingBridge } from '../composables/useTaskViewPagingBridge'\n"
)

const topGlue = `const closeWebhookModal = () => { showWebhookModal.value = false }\nconst closeSingletonModal = () => { showSingletonModal.value = false }\nconst closeLogModal = () => { showLogModal.value = false }\nconst closeGlobalStatsModal = () => { showGlobalStatsModal.value = false }\nconst setTaskSearch = (value: string) => { taskSearch.value = value }\nconst setTasksJumpPageValue = (value: number | null) => { tasksJumpPage.value = value }\nconst setHistoryStatusFilter = (value: string) => { historyStatusFilter.value = value }\nconst setJumpPageValue = (value: number) => { jumpPage.value = value }\nconst setFinalFilesJumpValue = (value: number | null) => { finalFilesJump.value = value }\nconst prevTasksPage = () => { tasksPage.value-- }\nconst nextTasksPage = () => { tasksPage.value++ }\nconst backToTasks = () => { currentModule.value = tasks }\n`
if (!text.includes(topGlue)) throw new Error('top glue block not found')
text = text.replace(topGlue, `const closeWebhookModal = () => { showWebhookModal.value = false }\nconst closeSingletonModal = () => { showSingletonModal.value = false }\nconst closeLogModal = () => { showLogModal.value = false }\nconst closeGlobalStatsModal = () => { showGlobalStatsModal.value = false }\n`)

const runsGlue = `const prevRunsPage = async () => { runsPage.value--; await loadData() }\nconst nextRunsPage = async () => { runsPage.value++; await loadData() }\n\n`
if (!text.includes(runsGlue)) throw new Error('runs glue block not found')
text = text.replace(runsGlue, `const {\n  setTaskSearch,\n  setTasksJumpPageValue,\n  setHistoryStatusFilter,\n  setJumpPageValue,\n  setFinalFilesJumpValue,\n  prevTasksPage,\n  nextTasksPage,\n  backToTasks,\n  prevRunsPage,\n  nextRunsPage,\n} = useTaskViewPagingBridge({\n  taskSearch,\n  tasksJumpPage,\n  historyStatusFilter,\n  jumpPage,\n  finalFilesJump,\n  tasksPage,\n  runsPage,\n  currentModule,\n  loadData,\n})\n\n`)

text = text.replace(':back-to-tasks="() => { currentModule = \'tasks\' }"', ':back-to-tasks="backToTasks"')

fs.writeFileSync('frontend/src/views/TaskView.vue', text)
