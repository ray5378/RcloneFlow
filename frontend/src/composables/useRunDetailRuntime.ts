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
    runFilesPage,
    openRunDetailFiles,
    pagedRunFiles,
    totalRunFilesPages,
    goPrevFilesPage,
    goNextFilesPage,
  } = useRunDetailFiles({ runDetail, runApi: options.runApi })

  const {
    getFinalSummary,
    getPreflight,
    finalFiles,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    finalCountOther,
    setFinalFilter,
    finalFilesTotal,
    totalFinalFilesPages,
    pagedFinalFiles,
    finalFilesJump,
    goPrevFinalFilesPage,
    goNextFinalFilesPage,
    jumpFinalFilesPage,
  } = useRunDetailComputed({ runDetail })

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
    getPreflight,
    finalFiles,
    finalCountAll,
    finalCountSuccess,
    finalCountFailed,
    finalCountOther,
    setFinalFilter,
    finalFilesTotal,
    totalFinalFilesPages,
    pagedFinalFiles,
    finalFilesJump,
    goPrevFinalFilesPage,
    goNextFinalFilesPage,
    jumpFinalFilesPage,
  }
}
