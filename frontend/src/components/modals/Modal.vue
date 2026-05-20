<script setup lang="ts">
defineProps<{
  title: string
  show: boolean
}>()

const emit = defineEmits<{
  close: []
}>()
</script>

<template>
  <Teleport to="body">
    <div v-if="show" class="modal-overlay" @click.self="emit('close')">
      <div class="modal">
        <div class="modal-header">
          <h2>{{ title }}</h2>
          <button class="modal-close" @click="emit('close')">&times;</button>
        </div>
        <slot />
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.modal{ background:#111827; color:#e5e7eb; border:1px solid #374151 }
.modal-header h2{ color:#e5e7eb }
body.light .modal{ background:#fff; color:#111827; border:1px solid #e5e7eb }
body.light .modal-header h2{ color:#111827 }
</style>
