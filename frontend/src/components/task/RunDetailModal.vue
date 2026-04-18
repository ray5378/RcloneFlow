<script setup lang="ts">
import { FileItem } from '../files'

defineProps<{
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
  finalFilesTotal: number
  finalFilesPage: number
  totalFinalFilesPages: number
  finalFilesJump: number | null
  pagedFinalFiles: any[]
  finalFiles: any[]
  pagedRunFiles: any[]
  runFilesPage: number
  totalRunFilesPages: number
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'set-final-filter', value: 'all' | 'success' | 'failed' | 'other'): void
  (e: 'prev-final-files-page'): void
  (e: 'next-final-files-page'): void
  (e: 'update-final-files-jump', value: number | null): void
  (e: 'jump-final-files-page'): void
  (e: 'prev-files-page'): void
  (e: 'next-files-page'): void
}>()

function onFinalFilesJumpInput(event: Event) {
  const target = event.target as HTMLInputElement
  const value = target.value === '' ? null : Number(target.value)
  emit('update-final-files-jump', value)
}
</script>

<template>
  <div v-if="visible" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content detail-modal">
      <div class="modal-header">
        <h3>运行详情</h3>
        <button class="close-btn" @click="emit('close')">×</button>
      </div>
      <div class="modal-body">
        <div class="detail-item">
          <label>任务名称：</label>
          <span>{{ runDetail.taskName || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>执行模式：</label>
          <span>{{ runDetail.taskMode || '-' }}</span>
        </div>
        <div class="detail-item">
          <label>状态：</label>
          <span :class="['status', getStatusClass(runDetail.status)]">{{ getStatusText(runDetail.status) }}</span>
        </div>
        <div class="detail-item">
          <label>触发方式：</label>
          <span>{{ runDetail.trigger === 'schedule' ? '定时任务' : (runDetail.trigger === 'webhook' ? 'Webhook 触发' : '手动执行') }}</span>
        </div>
        <div class="detail-item full-width">
          <label>源路径：</label>
          <span>{{ runDetail.sourceRemote }}:{{ runDetail.sourcePath || '/' }}</span>
        </div>
        <div class="detail-item full-width">
          <label>目标路径：</label>
          <span>{{ runDetail.targetRemote }}:{{ runDetail.targetPath || '/' }}</span>
        </div>
        <div class="detail-item full-width">
          <label>运行总结：</label>
          <div class="summary-box" v-if="getFinalSummary(runDetail)">
            <div class="summary-title">统计概览（可筛选）</div>
            <div class="summary-grid">
              <div class="summary-cell clickable" @click="emit('set-final-filter', 'all')">
                <div class="summary-key">总计</div>
                <div class="summary-val">{{ finalCountAll }}</div>
              </div>
              <div class="summary-cell clickable" @click="emit('set-final-filter', 'success')">
                <div class="summary-key">{{ runDetail.taskMode === 'move' ? '移动' : '成功' }}</div>
                <div class="summary-val">{{ finalCountSuccess }}</div>
              </div>
              <div class="summary-cell clickable" @click="emit('set-final-filter', 'failed')">
                <div class="summary-key">失败</div>
                <div class="summary-val error-text">{{ finalCountFailed }}</div>
              </div>
              <div class="summary-cell clickable" @click="emit('set-final-filter', 'other')">
                <div class="summary-key">其他</div>
                <div class="summary-val">{{ finalCountOther }}</div>
              </div>
              <div class="summary-cell" v-if="getPreflight(runDetail)">
                <div class="summary-key">总体积</div>
                <div class="summary-val est">{{ formatBytes(getPreflight(runDetail).totalBytes || 0) }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">已传输体积</div>
                <div class="summary-val act">{{ formatBytes(getFinalSummary(runDetail)?.transferredBytes || 0) }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">开始时间</div>
                <div class="summary-val">{{ formatTime(runDetail.startedAt) }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">结束时间</div>
                <div class="summary-val">{{ formatTime(runDetail.finishedAt) }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">耗时</div>
                <div class="summary-val">{{ getFinalSummary(runDetail)?.durationText || '-' }}</div>
              </div>
              <div class="summary-cell">
                <div class="summary-key">平均速度</div>
                <div class="summary-val">{{ formatBps(getFinalSummary(runDetail)?.avgSpeedBps || 0) }}</div>
              </div>
            </div>
          </div>
          <pre v-else class="summary-pre">{{ JSON.stringify({ note: '无总结，可查看传输日志' }, null, 2) }}</pre>
        </div>

        <div class="detail-item full-width">
          <label>传输明细：</label>
          <div>
            <div class="files-toolbar">
              <span>共 {{ finalFilesTotal }} 条</span>
              <div class="pager-inline">
                <button class="ghost small" :disabled="finalFilesPage <= 1" @click="emit('prev-final-files-page')">上一页</button>
                <span>{{ finalFilesPage }}/{{ totalFinalFilesPages }}</span>
                <button class="ghost small" :disabled="finalFilesPage >= totalFinalFilesPages" @click="emit('next-final-files-page')">下一页</button>
                <input class="page-input jump-input" :value="finalFilesJump ?? ''" type="number" min="1" :max="totalFinalFilesPages" @input="onFinalFilesJumpInput" />
                <button class="ghost small" @click="emit('jump-final-files-page')">跳转</button>
              </div>
            </div>
            <div class="files-table large">
              <div class="files-header">
                <span class="name">文件</span>
                <span class="status">结果</span>
                <span class="time">时间</span>
                <span class="size">大小</span>
              </div>
              <div class="files-body">
                <template v-if="finalFiles && finalFiles.length">
                  <FileItem v-for="it in pagedFinalFiles" :key="(it.path || it.name) + (it.at || '') + (it.status || '')" :item="it" />
                </template>
                <template v-else>
                  <FileItem v-for="it in pagedRunFiles" :key="it.name + it.at + it.status" :item="it" />
                  <div v-if="!pagedRunFiles.length" class="path-empty">无明细（可能日志为空或历史记录较旧）</div>
                </template>
              </div>
            </div>
            <div class="files-pager" v-if="!finalFiles || !finalFiles.length">
              <button class="ghost small" :disabled="runFilesPage <= 1" @click="emit('prev-files-page')">上一页</button>
              <span>{{ runFilesPage }}/{{ totalRunFilesPages }}</span>
              <button class="ghost small" :disabled="runFilesPage >= totalRunFilesPages" @click="emit('next-files-page')">下一页</button>
            </div>
          </div>
        </div>
        <div v-if="runDetail.error" class="detail-item full-width">
          <label>错误信息：</label>
          <span class="error-text">{{ runDetail.error }}</span>
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
body.light .summary-cell.clickable:hover{background:#f0f4f8;border-color:#cbd5e1;box-shadow:0 0 0 1px #cbd5e1 inset}
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
