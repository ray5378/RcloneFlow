import { t } from '../../../i18n'
import type { TrackingMode, ActiveTransferFileStatus } from '../../../api/activeTransfer'

export function getTrackingLabels(mode: TrackingMode) {
  if (mode === 'cas') {
    return {
      current: t('activeTransfer.currentProcessing'),
      completed: t('activeTransfer.completedProcessing'),
      pending: t('activeTransfer.pendingProcessing'),
    }
  }
  return {
    current: t('activeTransfer.currentTransfer'),
    completed: t('activeTransfer.completedTransfer'),
    pending: t('activeTransfer.pendingTransfer'),
  }
}

export function getTransferStatusLabel(status: ActiveTransferFileStatus) {
  switch (status) {
    case 'copied': return t('activeTransfer.statusCopied')
    case 'cas_matched': return t('activeTransfer.statusCasMatched')
    case 'in_progress': return t('activeTransfer.statusInProgress')
    case 'pending': return t('activeTransfer.statusPending')
    case 'skipped': return t('activeTransfer.statusSkipped')
    case 'failed': return t('activeTransfer.statusFailed')
    case 'deleted': return t('activeTransfer.statusDeleted')
    default: return status
  }
}
