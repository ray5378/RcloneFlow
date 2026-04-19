<script setup lang="ts">
import RunItem from './RunItem.vue'
import type { Run } from '../../types'

const props = defineProps<{
  currentTotal: number
  runsPage: number
  runsPageSize: number
  currentTotalPages: number
  jumpPage: number
  historyFilterTaskId: number | null
  historyStatusFilter: string
  filteredRuns: Run[]
  getDbProgressStable: (run: Run) => any
  getFinalSummary: (run: Run) => any
}>()

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
</script>

<template>
  <div class="card">
    <div class="card-header history-header" @click="onHeaderClick">
      <div class="title clickable" @click.stop="emit('back')">任务历史记录 ←</div>
      <div class="history-filters" data-no-back @click.stop>
        <button :class="['filter-btn', historyStatusFilter === 'all' && 'active']" @click.stop="emit('set-status-filter', 'all')">全部</button>
        <button :class="['filter-btn', historyStatusFilter === 'finished' && 'active']" @click.stop="emit('set-status-filter', 'finished')">成功</button>
        <button :class="['filter-btn', historyStatusFilter === 'failed' && 'active']" @click.stop="emit('set-status-filter', 'failed')">失败</button>
        <button :class="['filter-btn', historyStatusFilter === 'skipped' && 'active']" @click.stop="emit('set-status-filter', 'skipped')">跳过</button>
        <button :class="['filter-btn', historyStatusFilter === 'hasTransfer' && 'active']" @click.stop="emit('set-status-filter', 'hasTransfer')">有传输</button>
      </div>
      <div class="pagination" v-if="currentTotal > runsPageSize" data-no-back @click.stop>
        <span class="page-current">第 {{ runsPage }} / {{ currentTotalPages }} 页</span>
        <button class="page-btn" :disabled="runsPage <= 1" @click.stop="emit('prev-page')">上一页</button>
        <button class="page-btn" :disabled="runsPage >= currentTotalPages" @click.stop="emit('next-page')">下一页</button>
        <input
          type="number"
          class="page-input"
          :value="jumpPage"
          :min="1"
          :max="currentTotalPages"
          @input="onJumpPageInput"
          @keyup.enter.stop="emit('jump-page')"
        />
        <button class="page-btn" @click.stop="emit('jump-page')">跳转</button>
      </div>
      <div class="header-actions" data-no-back @click.stop>
        <button
          v-if="historyFilterTaskId !== null && filteredRuns.length > 0"
          class="ghost small danger-text"
          @click.stop="emit('clear-all')"
        >
          删除所有
        </button>
        <button v-if="historyFilterTaskId !== null" class="ghost small" @click.stop="emit('back')">
          ← 返回
        </button>
      </div>
    </div>
    <div class="list">
      <RunItem
        v-for="run in filteredRuns"
        :key="run.id"
        :run="run"
        :progress="run.status === 'running' ? getDbProgressStable(run) : undefined"
        :summary="getFinalSummary(run)"
        @click="emit('view-detail', run)"
        @view-detail="emit('view-detail', run)"
        @view-log="emit('view-log', run)"
        @clear="emit('clear-run', run.id)"
      />
      <div v-if="!filteredRuns.length" class="empty">暂无历史记录</div>
    </div>
  </div>
</template>

<style scoped>
.history-header{
  cursor:pointer;
}
.history-header > *{
  cursor:auto;
}
.history-filters{
  display:flex;
  align-items:center;
  gap:8px;
  flex-wrap:wrap;
}
.pagination{
  display:flex;
  align-items:center;
  gap:8px;
  flex-wrap:wrap;
}
.page-btn{
  white-space:nowrap;
}
.page-input{
  width:64px;
  min-width:64px;
  padding:6px 8px;
  border:1px solid #333;
  border-radius:8px;
  background:#252525;
  color:#e0e0e0;
  box-sizing:border-box;
}
body.light .page-input{
  background:#fff;
  color:#111827;
  border-color:#ddd;
}
</style>
