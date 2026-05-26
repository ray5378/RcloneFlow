import { computed, ref } from 'vue'
import { useRunDetailState } from './useRunDetailState'
import { useRunDetailFiles, type RunDetailFilesReturn } from './useRunDetailFiles'
import { useRunDetailComputed, type RunDetailComputedReturn } from './useRunDetailComputed'

export interface RunApi {
  getRunDetail: (runId: number) => Promise<any>
  getRunFiles: (runId: number, page: number, pageSize: number) => Promise<any>
}

export interface UseRunDetailRuntimeOptions {
  runApi: RunApi
}

export interface RunDetailRuntimeReturn {
  showDetailModal: ReturnType<typeof useRunDetailState>['showDetailModal']
  runDetail: ReturnType<typeof useRunDetailState>['runDetail']
  openRunDetailModal: ReturnType<typeof useRunDetailState>['openRunDetailModal']
  closeRunDetailModal: ReturnType<typeof useRunDetailState>['closeRunDetailModal']
  runFilesTotal: ReturnType<typeof computed<number>>
  runFilesPage: RunDetailFilesReturn['runFilesPage']
  openRunDetailFiles: RunDetailFilesReturn['openRunDetailFiles']
  pagedRunFiles: RunDetailFilesReturn['pagedRunFiles']
  totalRunFilesPages: RunDetailFilesReturn['totalRunFilesPages']
  goPrevFilesPage: RunDetailFilesReturn['goPrevFilesPage']
  goNextFilesPage: RunDetailFilesReturn['goNextFilesPage']
  getFinalSummary: RunDetailComputedReturn['getFinalSummary']
  hasFinalSummaryFiles: RunDetailComputedReturn['hasFinalSummaryFiles']
  finalFiles: RunDetailComputedReturn['finalFiles']
  finalCountAll: RunDetailComputedReturn['finalCountAll']
  finalCountSuccess: RunDetailComputedReturn['finalCountSuccess']
  finalCountFailed: RunDetailComputedReturn['finalCountFailed']
  finalFilesPage: RunDetailComputedReturn['finalFilesPage']
  finalFilesTotal: RunDetailComputedReturn['finalFilesTotal']
  totalFinalFilesPages: RunDetailComputedReturn['totalFinalFilesPages']
  pagedFinalFiles: RunDetailComputedReturn['pagedFinalFiles']
  finalFilesJump: RunDetailComputedReturn['finalFilesJump']
  goPrevFinalFilesPage: RunDetailComputedReturn['goPrevFinalFilesPage']
  goNextFinalFilesPage: RunDetailComputedReturn['goNextFinalFilesPage']
  jumpFinalFilesPage: RunDetailComputedReturn['jumpFinalFilesPage']
}

export function useRunDetailRuntime(options: UseRunDetailRuntimeOptions): RunDetailRuntimeReturn {
  const {
    showDetailModal,
    runDetail,
    openRunDetailModal,
    closeRunDetailModal,
  } = useRunDetailState()

  const {
    runFiles,
    runFilesPage,
    openRunDetailFiles,
    visibleRunFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
  } = useRunDetailFiles({ runDetail, runApi: options.runApi })

  const {
    getFinalSummary,
    hasFinalSummaryFiles,
    finalFiles,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    finalFilesPage,
    finalFilesTotal,
    totalFinalFilesPages,
    pagedFinalFiles,
    finalFilesJump,
    goPrevFinalFilesPage,
    goNextFinalFilesPage,
    jumpFinalFilesPage,
  } = useRunDetailComputed({ runDetail, detailFiles: runFiles })

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
