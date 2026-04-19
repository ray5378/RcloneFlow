<script setup lang="ts">
defineProps<{
  commandMode: boolean
  commandText: string
  createForm: any
  remotes: string[]
  showSourcePathInput: boolean
  showTargetPathInput: boolean
  sourceBreadcrumbs: any[]
  sourceCurrentPath: string
  sourcePathOptions: any[]
  targetBreadcrumbs: any[]
  targetCurrentPath: string
  targetPathOptions: any[]
  showAdvancedOptions: boolean
  creatingState: string
  editingTask: any
  setCommandMode: (value: boolean) => void
  setCommandText: (value: string) => void
  setShowSourcePathInput: (value: boolean) => void
  setShowTargetPathInput: (value: boolean) => void
  setShowAdvancedOptions: (value: boolean) => void
  onSourceRemoteChange: (...args: any[]) => void
  onTargetRemoteChange: (...args: any[]) => void
  onSourceBreadcrumbClick: (...args: any[]) => void
  onTargetBreadcrumbClick: (...args: any[]) => void
  onSourceArrow: (...args: any[]) => void
  onSourceClick: (...args: any[]) => void
  onTargetArrow: (...args: any[]) => void
  onTargetClick: (...args: any[]) => void
  createTask: () => void
}>()
</script>

<template>
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
    @update:show-advanced-options="setShowAdvancedOptions"
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
</template>
