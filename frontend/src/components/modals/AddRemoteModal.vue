<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import Modal from './Modal.vue'
import RemoteProviderSelect from '../remote/RemoteProviderSelect.vue'
import RemoteFormFields from '../remote/RemoteFormFields.vue'
import RemoteNameInput from '../remote/RemoteNameInput.vue'
import RemoteConfirmStep from '../remote/RemoteConfirmStep.vue'
import * as api from '../../api'
import type { Provider, ProviderOption } from '../../types'
import { t } from '../../i18n'
import { showErrorToast } from '../../api/errors'
import { getProviderDescription } from '../remote/remoteUtils'

const props = defineProps<{
  show: boolean
  editMode?: boolean
  editName?: string
}>()

const emit = defineEmits<{
  close: []
  success: []
}>()

const step = ref(0)
const search = ref('')
const selectedProviderName = ref('')
const selectedProvider = ref<Provider | null>(null)
const remoteName = ref('')
const remoteOptions = ref<Record<string, string>>({})
const creating = ref(false)
const success = ref(false)
const showAdvancedOptions = ref(false)
const providerNeedAuth = ref(false)
const providers = ref<Provider[]>([])

const filteredProviders = computed(() => {
  const keyword = search.value.trim().toLowerCase()
  return (providers.value || [])
    .filter(p => {
      if (!keyword) return true
      const haystack = `${p.Name} ${p.Description || ''} ${getProviderDescription(p)}`
      return haystack.toLowerCase().includes(keyword)
    })
    .sort((a, b) => a.Name.localeCompare(b.Name))
})

const allOptions = computed(() => selectedProvider.value?.Options || [])

function resetProviderOptions() {
  if (!selectedProvider.value) return
  remoteOptions.value = Object.fromEntries(
    selectedProvider.value.Options.map((option) => [option.Name, option.DefaultStr || ''])
  )
}

function shouldShowOption(option: ProviderOption) {
  if (option.Advanced && !showAdvancedOptions.value) return false
  if (
    selectedProvider.value?.Name === 's3' &&
    option.Provider &&
    option.Provider !== remoteOptions.value['provider']
  ) {
    return false
  }
  return true
}

const groupedOptions = computed(() => {
  const required: ProviderOption[] = []
  const optional: ProviderOption[] = []
  const advanced: ProviderOption[] = []
  for (const option of allOptions.value) {
    if (!shouldShowOption(option)) continue
    if (option.Advanced) advanced.push(option)
    else if (option.Required) required.push(option)
    else optional.push(option)
  }
  return { required, optional, advanced }
})

watch(() => props.show, async (val) => {
  if (!val) return
  step.value = props.editMode ? 1 : 0
  search.value = ''
  selectedProviderName.value = ''
  selectedProvider.value = null
  remoteName.value = props.editName || ''
  remoteOptions.value = {}
  creating.value = false
  success.value = false
  showAdvancedOptions.value = false
  providerNeedAuth.value = false
  const data = await api.listProviders()
  providers.value = data.providers || []
})

function selectProvider(provider: Provider) {
  selectedProvider.value = provider
  selectedProviderName.value = provider.Name
  providerNeedAuth.value = provider.Options.some((option) => option.Name === 'token')
  resetProviderOptions()
  step.value = 1
}

function nextStep() {
  if (step.value === 0) step.value = 1
  else if (step.value === 1) step.value = props.editMode ? 3 : 2
  else if (step.value === 2) step.value = 3
}

function prevStep() {
  if (step.value === 3) step.value = props.editMode ? 1 : 2
  else if (step.value === 2) step.value = 1
  else if (step.value === 1) step.value = 0
}

function validateName(name: string) {
  return !!name && name.length >= 2 && /^[a-zA-Z0-9_-]*$/.test(name)
}

function handleRemoteOptionsUpdate(options: Record<string, string>) {
  remoteOptions.value = options
}

async function create() {
  creating.value = true
  try {
    if (!selectedProvider.value?.Name) throw new Error(t('remote.notSelected'))
    if (!validateName(remoteName.value)) {
      throw new Error(t('remote.invalidName'))
    }
    const params: Record<string, unknown> = Object.fromEntries(
      Object.entries(remoteOptions.value).filter(([, value]) => value !== '')
    )
    if (providerNeedAuth.value && !params['token']) {
      params['token'] = ''
    }
    if (props.editMode && props.editName) {
      await api.updateRemote(remoteName.value, selectedProvider.value.Name, params)
    } else {
      await api.createRemote(remoteName.value, selectedProvider.value.Name, params)
    }
    success.value = true
    setTimeout(() => {
      emit('success')
      emit('close')
    }, 800)
  } catch (e) {
    showErrorToast((e as Error).message)
  } finally {
    creating.value = false
  }
}

defineExpose({
  loadConfig: async (name: string) => {
    while (!providers.value.length) {
      await new Promise(resolve => setTimeout(resolve, 50))
    }
    const config = await api.getRemoteConfig(name)
    remoteName.value = name
    const type = String(config.type || '')
    const provider = providers.value.find(p => p.Name === type)
    if (!provider) return
    selectedProvider.value = provider
    selectedProviderName.value = type
    providerNeedAuth.value = provider.Options.some((option) => option.Name === 'token')
    resetProviderOptions()
    for (const key in config) {
      if (key !== 'type' && key !== 'name') {
        remoteOptions.value[key] = String(config[key])
      }
    }
    step.value = 1
  },
})
</script>

<template>
  <Modal :show="show" :title="editMode ? $t('remote.editTitle') : $t('remote.addTitle')" @close="emit('close')">
    <div class="add-remote-body">
      <div v-if="!editMode" class="stepper">
      <div class="step" :class="{ active: step === 0, done: step > 0 }">1. {{ $t('remote.stepChooseType') }}</div>
      <div class="step" :class="{ active: step === 1, done: step > 1 }">2. {{ $t('remote.stepConfig') }}</div>
      <div class="step" :class="{ active: step === 2, done: step > 2 }">3. {{ $t('remote.stepName') }}</div>
      <div class="step" :class="{ active: step === 3 }">4. {{ $t('remote.stepSave') }}</div>
    </div>
    <div v-else class="stepper">
      <div class="step" :class="{ active: step === 1, done: step > 1 }">1. {{ $t('remote.stepConfig') }}</div>
      <div class="step" :class="{ active: step === 3 }">2. {{ $t('remote.stepConfirm') }}</div>
    </div>

    <div v-if="step === 0">
      <RemoteProviderSelect
        :providers="filteredProviders"
        :search="search"
        :selected-provider-name="selectedProviderName"
        @update:search="search = $event"
        @select="selectProvider"
      />
    </div>

    <div v-if="step === 1">
      <RemoteFormFields
        :required-options="groupedOptions.required"
        :optional-options="groupedOptions.optional"
        :advanced-options="groupedOptions.advanced"
        :all-options="allOptions"
        :remote-options="remoteOptions"
        :show-advanced-options="showAdvancedOptions"
        @update:remote-options="handleRemoteOptionsUpdate"
        @update:show-advanced-options="showAdvancedOptions = $event"
      />
      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button v-if="!editMode" class="ghost" @click="prevStep">{{ $t('remote.previous') }}</button>
        <button @click="nextStep">{{ $t('remote.next') }}</button>
      </div>
    </div>

    <div v-if="step === 2 && !editMode">
      <RemoteNameInput
        :remote-name="remoteName"
        @update:remote-name="remoteName = $event"
      />
      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button class="ghost" @click="prevStep">{{ $t('remote.previous') }}</button>
        <button @click="nextStep" :disabled="!validateName(remoteName)">{{ $t('remote.next') }}</button>
      </div>
    </div>

    <div v-if="step === 3">
      <RemoteConfirmStep
        :selected-provider="selectedProvider"
        :remote-name="remoteName"
        :creating="creating"
        :success="success"
        :edit-mode="editMode"
        @create="create"
        @close="prevStep"
      />
    </div>
    </div>
  </Modal>
</template>

<style scoped>
.add-remote-body {
  display: flex;
  flex-direction: column;
  min-height: 0;
  max-height: calc(85vh - 88px);
  overflow-y: auto;
}
.stepper { display: flex; gap: 8px; margin-bottom: 20px; }
.step { padding: 8px 12px; border-radius: 8px; background: #1f2937; color: #cbd5e1; font-size: 14px; border:1px solid #374151 }
.step.active { background: #1e3a5f; color: #60a5fa; border-color:#2563eb }
.step.done { background: #0a2f22; color: #34d399; border-color:#14532d }
body.light .step { background: #f3f4f6; color: #6b7280; border-color:#e5e7eb }
body.light .step.active { background: #dbeafe; color: #1d4ed8; border-color:#93c5fd }
body.light .step.done { background: #dcfce7; color: #166534; border-color:#86efac }
.actions { display: flex; gap: 8px; }
button { padding: 10px 14px; border-radius: 10px; border: none; background: #2563eb; color: white; cursor: pointer; }
button.ghost { background: #eef2ff; color: #1d4ed8; }
button:disabled { opacity: .6; cursor: not-allowed; }

@media (max-width: 768px) {
  .add-remote-body {
    max-height: calc(85vh - 72px);
    overflow-y: auto;
    -webkit-overflow-scrolling: touch;
    padding-right: 2px;
  }
  .stepper {
    flex-wrap: wrap;
    margin-bottom: 16px;
  }
  .actions {
    flex-wrap: wrap;
  }
  .actions button {
    flex: 1 1 auto;
  }
}
</style>