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

  const {
    runFiles,
    runFilesPage,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
  } = useRunDetailFiles({ runDetail, runApi: options.runApi })

  const {
    getFinalSummary,
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
  } = useRunDetailComputed({ runDetail, detailFiles: runFiles })

  return {
    showDetailModal,
    runDetail,
    openRunDetailModal,
    closeRunDetailModal,
    runFilesPage,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
    getFinalSummary,
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
