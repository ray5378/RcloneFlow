<script setup lang="ts">
import { computed } from 'vue'
import RunItem from './RunItem.vue'
import type { Run } from '../../types'
import type { TaskProgressLike } from './progressText'
import { t } from '../../i18n'

const props = defineProps<{
  currentTotal: number
  runsPage: number
  runsPageSize: number
  currentTotalPages: number
  jumpPage: number
  historyFilterTaskId: number | null
  historyStatusFilter: string
  filteredRuns: Run[]
  getRunProgressFromSummary: (run: Run) => TaskProgressLike | null
  getRealtimeProgressByRun: (run: Run) => TaskProgressLike | null
  getFinalSummary: (run: Run) => any
}>()

const displayRuns = computed(() => props.filteredRuns.map(run => {
  const summary = (run as any)?.summary?.finalSummary && typeof (run as any).summary.finalSummary === 'object'
    ? (run as any).summary.finalSummary
    : (run as any)?.summary
  return {
    ...run,
    __title: run.taskName || `${t('runItem.taskFallback')} #${run.taskId}`,
    __triggerText: run.trigger === 'schedule' ? t('runItem.schedule') : run.trigger === 'manual' ? t('runItem.manual') : run.trigger === 'webhook' ? 'Webhook' : '',
    __statusText: run.status === 'finished' ? t('runItem.finished') : run.status === 'failed' ? t('runItem.failed') : run.status === 'skipped' ? t('runItem.skipped') : run.status === 'running' ? t('runItem.running') : run.status,
    __statusClass: run.status === 'finished' ? 'success' : run.status === 'failed' ? 'danger' : run.status === 'skipped' ? 'warning' : run.status === 'running' ? 'info' : '',
    __startedText: run.startedAt ? new Date(run.startedAt).toLocaleString('zh-CN', { month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' }) : '-',
    __summaryTotal: summary?.counts?.total ?? 0,
    __summarySuccess: ((summary?.counts?.copied || 0) + (summary?.counts?.deleted || 0)),
    __summaryFailed: summary?.counts?.failed || 0,
    __summaryTotalSize: summary?.totalBytes || 0,
    __summaryTransferred: summary?.transferredBytes || 0,
    __summaryMessage: summary?.message || '',
  }
}))

const emit = defineEmits<{
  (e: 'back'): void
  (e: 'set-status-filter', value: string): void
  (e: 'prev-page'): void
  (e: 'next-page'): void
  (e: 'update-jump-page', value: number): void
  (e: 'jump-page'): void
  (e: 'clear-all'): void
  (e: 'view-detail', run: Run): void
  (e: 'view-log', run: Run): void
  (e: 'clear-run', runId: number): void
}>()

function onJumpPageInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update-jump-page', Number(target.value || 1))
}

function onHeaderClick(event: MouseEvent) {
  const target = event.target as HTMLElement | null
  if (!target) return
  if (target.closest('[data-no-back]')) return
  emit('back')
}

function pageInfo(page: number, total: number) {
  return t('taskUI.pageInfo').replace('{page}', String(page)).replace('{total}', String(total))
}
</script>

<template>
  <div class="card">
    <div class="card-header history-header" @click="onHeaderClick">
      <div class="title clickable" @click.stop="emit('back')">{{ t('taskUI.historyTitle') }} ←</div>
      <div class="history-filters" data-no-back @click.stop>
        <button :class="['filter-btn', historyStatusFilter === 'all' && 'active']" @click.stop="emit('set-status-filter', 'all')">{{ t('taskUI.all') }}</button>
        <button :class="['filter-btn', historyStatusFilter === 'finished' && 'active']" @click.stop="emit('set-status-filter', 'finished')">{{ t('taskUI.success') }}</button>
        <button :class="['filter-btn', historyStatusFilter === 'failed' && 'active']" @click.stop="emit('set-status-filter', 'failed')">{{ t('taskUI.failed') }}</button>
        <button :class="['filter-btn', historyStatusFilter === 'skipped' && 'active']" @click.stop="emit('set-status-filter', 'skipped')">{{ t('taskUI.skipped') }}</button>
        <button :class="['filter-btn', historyStatusFilter === 'hasTransfer' && 'active']" @click.stop="emit('set-status-filter', 'hasTransfer')">{{ t('taskUI.hasTransfer') }}</button>
      </div>
      <div class="pagination" v-if="currentTotal > runsPageSize" data-no-back @click.stop>
        <span class="page-current">{{ pageInfo(runsPage, currentTotalPages) }}</span>
        <button class="page-btn" :disabled="runsPage <= 1" @click.stop="emit('prev-page')">{{ t('taskUI.prevPage') }}</button>
        <button class="page-btn" :disabled="runsPage >= currentTotalPages" @click.stop="emit('next-page')">{{ t('taskUI.nextPage') }}</button>
        <input type="number" class="page-input" :value="jumpPage" :min="1" :max="currentTotalPages" @input="onJumpPageInput" @keyup.enter.stop="emit('jump-page')" />
        <button class="page-btn" @click.stop="emit('jump-page')">{{ t('taskUI.jump') }}</button>
      </div>
      <div class="header-actions" data-no-back @click.stop>
        <button v-if="historyFilterTaskId !== null && props.filteredRuns.length > 0" class="ghost small danger-text" @click.stop="emit('clear-all')">{{ t('taskUI.deleteAll') }}</button>
        <button v-if="historyFilterTaskId !== null" class="ghost small" @click.stop="emit('back')">← {{ t('taskUI.back') }}</button>
      </div>
    </div>
    <div class="list history-list-cards">
      <RunItem
        v-for="run in displayRuns"
        :key="run.id"
        v-memo="[run.id, run.status, run.startedAt, run.finishedAt, run.bytesTransferred]"
        :run="run"
        :progress="run.status === 'running' ? getRealtimeProgressByRun(run) : undefined"
        :summary="(run as any).summary"
        @click="emit('view-detail', run)"
        @view-detail="emit('view-detail', run)"
        @view-log="emit('view-log', run)"
        @clear="emit('clear-run', run.id)"
      />
      <div v-if="!props.filteredRuns.length" class="empty">{{ t('taskUI.noHistory') }}</div>
    </div>
  </div>
</template>

<style scoped>
.history-list-cards{
  gap: 8px;
  margin-top: 16px;
  padding: 0 16px;
  box-sizing: border-box;
}
.history-list-cards :deep(.run-item){
  margin-bottom: 0;
}
.history-header{ cursor:pointer; }
.history-header > *{ cursor:auto; }
.history-filters{ display:flex; align-items:center; gap:8px; flex-wrap:wrap; }
.pagination{ display:flex; align-items:center; gap:8px; flex-wrap:wrap; }
.page-btn{ white-space:nowrap; }
.page-input{ width:64px; min-width:64px; padding:6px 8px; border:1px solid #333; border-radius:8px; background:#252525; color:#e0e0e0; box-sizing:border-box; }
body.light .page-input{ background:#fff; color:#111827; border-color:#ddd; }
</style>
