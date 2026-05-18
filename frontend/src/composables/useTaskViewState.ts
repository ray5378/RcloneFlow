import { ref } from 'vue'
import type { Task, Schedule, Run } from '../types'

export function useTaskViewState() {
  const tasks = ref<Task[]>([])
  const schedules = ref<Schedule[]>([])
  const runs = ref<Run[]>([])
  const runsTotal = ref(0)
  const taskRuns = ref<Run[]>([])
  const runsPage = ref(1)
  const runsPageSize = 50
  const jumpPage = ref(1)

  const remotes = ref<string[]>([])
  const currentModule = ref<'history' | 'add' | 'tasks'>('tasks')
  const historyFilterTaskId = ref<number | null>(null)
  const historyStatusFilter = ref<string>('all')

  return {
    tasks,
    schedules,
    runs,
    runsTotal,
    taskRuns,
    runsPage,
    runsPageSize,
    jumpPage,
    remotes,
    currentModule,
    historyFilterTaskId,
    historyStatusFilter,
  }
}
