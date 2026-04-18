interface UseRunDetailEntryOptions {
  openRunningHint: (run: any) => void
  openRunDetailModal: (run: any) => void
  openRunDetailFiles: (run: any) => void
  closeRunDetailModal: () => void
}

export function useRunDetailEntry(options: UseRunDetailEntryOptions) {
  function showRunDetail(run: any) {
    if (run.status === 'running') {
      // 运行中的记录不进入历史详情弹窗，而是走轻量提示小窗
      options.openRunningHint(run)
      return
    }

    // 历史详情入口：页面层只负责装配，详情状态与文件链已分别下沉
    options.openRunDetailModal(run)
    options.openRunDetailFiles(run)
  }

  function closeRunDetail() {
    // 历史详情出口：具体状态由 useRunDetailState 管理
    options.closeRunDetailModal()
  }

  return {
    showRunDetail,
    closeRunDetail,
  }
}
