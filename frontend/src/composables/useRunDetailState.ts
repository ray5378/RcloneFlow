import { ref } from 'vue'

export function useRunDetailState() {
  const showDetailModal = ref(false)
  const runDetail = ref<any>({})

  function openRunDetailModal(run: any) {
    runDetail.value = run
    showDetailModal.value = true
  }

  function closeRunDetailModal() {
    showDetailModal.value = false
  }

  return {
    showDetailModal,
    runDetail,
    openRunDetailModal,
    closeRunDetailModal,
  }
}
