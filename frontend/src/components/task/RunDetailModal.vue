<script setup lang="ts">
import { computed } from 'vue'
import { FileItem } from '../files'
import { t } from '../../i18n'

const props = defineProps<{
  visible: boolean
  runDetail: any
  getStatusClass: (status: string) => string
  getStatusText: (status: string) => string
  getFinalSummary: (run: any) => any
  getPreflight: (run: any) => any
  formatBytes: (bytes: number) => string
  formatTime: (value: any) => string
  formatBps: (bps: number) => string
  finalCountAll: number
  finalCountSuccess: number
  finalCountFailed: number
  finalCountOther: number
  pagedRunFiles: any[]
  runFilesTotal: number
  runFilesPage: number
  totalRunFilesPages: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'prev-files-page'): void
  (e: 'next-files-page'): void
}>()

function getSuccessLabel(mode?: string) {
  if (mode === 'move') return t('modal.moved')
  if (mode === 'sync') return t('modal.synced')
  return t('modal.copied')
}

function countLine(template: string, count: number) {
  return template.replace('{count}', String(count))
}

const activeFilesPage = computed(() => Math.max(1, Number(props.runFilesPage) || 1))
const activeTotalFilesPages = computed(() => Math.max(1, Number(props.totalRunFilesPages) || 1))
const activeFilesTotal = computed(() => Math.max(0, Number(props.runFilesTotal) || 0))
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content detail-modal">
      <div class="modal-header">
        <h3>{{ t('modal.runDetail') }}</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item">
          <label>{{ t('modal.taskName') }}</label>
          <span>{{ props.runDetail.taskName || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>{{ t('modal.runMode') }}</label>
          <span>{{ props.runDetail.taskMode || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>{{ t('modal.status') }}</label>
          <span :class="['status', props.getStatusClass(props.runDetail.status)]">{{ props.getStatusText(props.runDetail.status) }}</span>
        </div>
        <div class="detail-item">
          <label>{{ t('modal.trigger') }}</label>
          <span>{{ props.runDetail.trigger === 'schedule' ? t('modal.triggerSchedule') : (props.runDetail.trigger === 'webhook' ? t('modal.triggerWebhook') : t('modal.triggerManual')) }}</span>
        </div>
        <div class="detail-item full-width">
          <label>{{ t('modal.sourcePath') }}</label>
          <span>{{ props.runDetail.sourceRemote }}:{{ props.runDetail.sourcePath || '/' }}</span>
        </div>
        <div class="detail-item full-width">
          <label>{{ t('modal.targetPath') }}</label>
          <span>{{ props.runDetail.targetRemote }}:{{ props.runDetail.targetPath || '/' }}</span>
        </div>
        <div class="detail-item full-width">
          <label>{{ t('modal.summary') }}</label>
          <div class="summary-box" v-if="props.getFinalSummary(props.runDetail)">
            <div class="summary-title">{{ t('modal.summaryStats') }}</div>
            <div class="summary-grid">
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.total') }}</div>
                <div class="summary-val">{{ finalCountAll }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ getSuccessLabel(props.runDetail.taskMode) }}</div>
                <div class="summary-val">{{ finalCountSuccess }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.failed') }}</div>
                <div class="summary-val error-text">{{ finalCountFailed }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.other') }}</div>
                <div class="summary-val">{{ finalCountOther }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.transferredBytes') }}</div>
                <div class="summary-val act">{{ props.formatBytes(props.getFinalSummary(props.runDetail)?.transferredBytes || 0) }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.startedAt') }}</div>
                <div class="summary-val">{{ props.formatTime(props.runDetail.startedAt) }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.finishedAt') }}</div>
                <div class="summary-val">{{ props.formatTime(props.runDetail.finishedAt) }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.duration') }}</div>
                <div class="summary-val">{{ props.getFinalSummary(props.runDetail)?.durationText || '-' }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">{{ t('modal.avgSpeed') }}</div>
                <div class="summary-val">{{ props.formatBps(props.getFinalSummary(props.runDetail)?.avgSpeedBps || 0) }}</div>
              </div>
            </div>
          </div>
          <pre v-else class="summary-pre">{{ JSON.stringify({ note: t('modal.noSummary') }, null, 2) }}</pre>
        </div>

        <div class="detail-item full-width">
          <label>{{ t('modal.details') }}</label>
          <div>
            <div class="files-toolbar">
              <span>{{ countLine(t('modal.countLine'), activeFilesTotal) }}</span>
              <div class="pager-inline">
                <button class="ghost small" :disabled="activeFilesPage <= 1" @click="emit('prev-files-page')">{{ t('modal.prevPage') }}</button>
                <span>{{ activeFilesPage }}/{{ activeTotalFilesPages }}</span>
                <button class="ghost small" :disabled="activeFilesPage >= activeTotalFilesPages" @click="emit('next-files-page')">{{ t('modal.nextPage') }}</button>
              </div>
            </div>
            <div class="files-table large">
              <div class="files-header">
                <span class="name">{{ t('modal.file') }}</span>
                <span class="status">{{ t('modal.result') }}</span>
                <span class="time">{{ t('modal.time') }}</span>
                <span class="size">{{ t('modal.size') }}</span>
              </div>
              <div class="files-body">
                <FileItem v-for="it in pagedRunFiles" :key="it.name + it.at + it.status" :item="it" />
                <div v-if="!pagedRunFiles.length" class="path-empty">{{ t('modal.noDetail') }}</div>
              </div>
            </div>
          </div>
        </div>
        <div v-if="props.runDetail.error" class="detail-item full-width">
          <label>{{ t('modal.error') }}</label>
          <span class="error-text">{{ props.runDetail.error }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.detail-modal{ width: 135% !important; max-width: 1200px !important; }
.summary-box{background:#111827;border:1px solid #333;border-radius:10px;padding:12px 14px;margin-top:6px;max-width:1200px}
.summary-title{font-size:13px;color:#94a3b8;margin-bottom:10px}
.summary-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:12px}
.summary-cell{background:#0f172a;border:1px solid #334155;border-radius:10px;padding:10px 12px}
.summary-key{font-size:12px;color:#94a3b8;margin-bottom:6px}
.summary-val{font-size:18px;font-weight:700;color:#e5e7eb}
.summary-val.est{color:#93c5fd}
.summary-val.act{color:#86efac}
body.light .summary-box{background:#ffffff;border-color:#e5e7eb}
body.light .summary-title{color:#64748b}
body.light .summary-cell{background:#f8fafc;border-color:#e5e7eb}
body.light .summary-key{color:#64748b}
body.light .summary-val{color:#111827}
.summary-cell.clickable{cursor:pointer;transition:background .15s ease,border-color .15s ease,box-shadow .15s ease}
.summary-cell.clickable:hover{background:#1f2937;border-color:#475569;box-shadow:0 0 0 1px #334155 inset}
.files-toolbar{display:flex;align-items:center;justify-content:space-between;gap:12px;margin-top:8px;max-width:1200px}
.files-table{margin-top:14px;border:1px solid #333;border-radius:10px;overflow:hidden}
.files-table.large{max-width:1200px}
.files-header,.files-row{display:grid;grid-template-columns:1fr 140px 200px 140px;gap:18px;align-items:center}
.files-header{padding:12px 16px;background:#252525;color:#cbd5e1;font-size:13px}
.files-body{max-height:540px;overflow:auto}
body.light .files-table{border-color:#e5e7eb}
body.light .files-header{background:#f5f5f5;color:#4b5563}
.files-table .status{width:auto;padding:0;background:transparent;border-radius:0;font-weight:600}
.files-table .status.success{color:#34d399}
.files-table .status.failed{color:#f87171}
.files-table .status.skipped{color:#fbbf24}
.pager-inline{display:flex;align-items:center;gap:8px;flex-wrap:wrap}
.files-pager{display:flex;align-items:center;gap:8px;margin-top:10px;flex-wrap:wrap}
.pager-inline button,
.files-pager button,
.pager-inline span,
.files-pager span{white-space:nowrap}
.jump-input{width:72px;min-width:72px;padding:4px 8px}
@media (max-width: 1100px){
  .summary-grid{grid-template-columns:repeat(2,1fr)}
  .files-toolbar{flex-direction:column;align-items:flex-start}
}
</style>
