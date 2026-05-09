import { computed, ref } from 'vue'
import { useRunDetailState } from './useRunDetailState'
import { useRunDetailFiles } from './useRunDetailFiles'
import { useRunDetailComputed } from './useRunDetailComputed'

export function useRunDetailRuntime(options: {
  runApi: any
}) {
  const {
    showDetailModal,
    runDetail,
    openRunDetailModal,
    closeRunDetailModal,
  } = useRunDetailState()

  const currentFinalFilter = ref<'all' | 'success' | 'failed' | 'other'>('all')

  const {
    runFiles,
    runFilesPage,
    openRunDetailFiles,
    visibleRunFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
  } = useRunDetailFiles({ runDetail, currentFinalFilter, runApi: options.runApi })

  const {
    getFinalSummary,
    hasFinalSummaryFiles,
    finalFiles,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    finalCountOther,
    setFinalFilter,
    finalFilesPage,
    finalFilesTotal,
    totalFinalFilesPages,
    pagedFinalFiles,
    finalFilesJump,
    goPrevFinalFilesPage,
    goNextFinalFilesPage,
    jumpFinalFilesPage,
  } = useRunDetailComputed({ runDetail, detailFiles: runFiles, currentFinalFilter })

  const runFilesTotal = computed(() => visibleRunFiles.value.length)

  return {
    showDetailModal,
    runDetail,
    openRunDetailModal,
    closeRunDetailModal,
    runFilesTotal,
    runFilesPage,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
    getFinalSummary,
    hasFinalSummaryFiles,
    finalFiles,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    finalCountOther,
    setFinalFilter,
    finalFilesPage,
    finalFilesTotal,
    totalFinalFilesPages,
    pagedFinalFiles,
    finalFilesJump,
    goPrevFinalFilesPage,
    goNextFinalFilesPage,
    jumpFinalFilesPage,
  }
}
