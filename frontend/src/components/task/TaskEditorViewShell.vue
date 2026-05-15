<script setup lang="ts">
import AddTaskForm from './AddTaskForm.vue'
import type { CreateForm, PathBreadcrumb, PathBrowseItem } from './types'
import type { Task } from '../../types'
import { t } from '../../i18n'

const props = defineProps<{
  visible: boolean
  title: string
  commandMode: boolean
  commandText: string
  createForm: CreateForm
  remotes: string[]
  showSourcePathInput: boolean
  showTargetPathInput: boolean
  sourceBreadcrumbs: PathBreadcrumb[]
  sourceCurrentPath: string
  sourcePathOptions: PathBrowseItem[]
  targetBreadcrumbs: PathBreadcrumb[]
  targetCurrentPath: string
  targetPathOptions: PathBrowseItem[]
  showAdvancedOptions: boolean
  creatingState: string
  editingTask: Task | null
  setCommandMode: (value: boolean) => void
  setCommandText: (value: string) => void
  setShowSourcePathInput: (value: boolean) => void
  setShowTargetPathInput: (value: boolean) => void
  setShowAdvancedOptions: (value: boolean) => void
  onSourceRemoteChange: () => void
  onTargetRemoteChange: () => void
  onSourceBreadcrumbClick: (path: string) => void
  onTargetBreadcrumbClick: (path: string) => void
  onSourceArrow: (item: PathBrowseItem) => void
  onSourceClick: (item: PathBrowseItem) => void
  onTargetArrow: (item: PathBrowseItem) => void
  onTargetClick: (item: PathBrowseItem) => void
  createTask: () => void
  closeEditorModal: () => void
}>()

function handleClose() {
  if (props.creatingState === 'loading') return
  props.closeEditorModal()
}
</script>

<template>
  <div v-if="visible" class="modal-overlay task-editor-overlay" @click.self="handleClose">
    <div class="modal-content task-editor-modal">
      <div class="modal-header task-editor-header">
        <h3 class="task-editor-title">{{ title }}</h3>
        <button class="close-btn task-editor-close-btn" @click="handleClose" aria-label="关闭">×</button>
      </div>
      <div class="task-editor-body">
        <AddTaskForm
          :command-mode="commandMode"
          :command-text="commandText"
          :create-form="createForm"
          :remotes="remotes"
          :show-source-path-input="showSourcePathInput"
          :show-target-path-input="showTargetPathInput"
          :source-breadcrumbs="sourceBreadcrumbs"
          :source-current-path="sourceCurrentPath"
          :source-path-options="sourcePathOptions"
          :target-breadcrumbs="targetBreadcrumbs"
          :target-current-path="targetCurrentPath"
          :target-path-options="targetPathOptions"
          :show-advanced-options="showAdvancedOptions"
          :creating-state="creatingState"
          :editing-task="editingTask"
          @update:command-mode="setCommandMode"
          @update:command-text="setCommandText"
          @update:show-source-path-input="setShowSourcePathInput"
          @update:show-target-path-input="setShowTargetPathInput"
          @update:showAdvancedOptions="setShowAdvancedOptions"
          @source-remote-change="onSourceRemoteChange"
          @target-remote-change="onTargetRemoteChange"
          @source-breadcrumb-click="onSourceBreadcrumbClick"
          @target-breadcrumb-click="onTargetBreadcrumbClick"
          @source-arrow="onSourceArrow"
          @source-click="onSourceClick"
          @target-arrow="onTargetArrow"
          @target-click="onTargetClick"
          @submit="createTask"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.task-editor-overlay {
  z-index: 1200;
}

.task-editor-modal {
  width: min(1100px, calc(100vw - 32px));
  max-height: calc(100vh - 32px);
  overflow: hidden;
  padding: 0;
}

.task-editor-header {
  margin-bottom: 0;
  padding: 22px 28px 18px;
}

.task-editor-title {
  margin: 0;
  padding-right: 16px;
}

.task-editor-close-btn {
  margin-top: -2px;
  margin-right: 2px;
  padding: 6px;
  flex-shrink: 0;
}

.task-editor-body {
  max-height: calc(100vh - 120px);
  overflow: auto;
  padding: 6px 28px 28px;
}

.task-editor-body :deep(.form-content) {
  padding-top: 12px;
}

.task-editor-body :deep(.card) {
  border: none;
  box-shadow: none;
  background: transparent;
}

.task-editor-body :deep(.card-header) {
  display: none;
}
</style>
