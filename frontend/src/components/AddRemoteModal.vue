<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import Modal from './Modal.vue'
import * as api from '../api'
import type { Provider, ProviderOption } from '../types'

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
      return `${p.Name} ${p.Description || ''}`.toLowerCase().includes(keyword)
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

async function create() {
  creating.value = true
  try {
    if (!selectedProvider.value?.Name) throw new Error('未选择存储类型')
    if (!validateName(remoteName.value)) {
      throw new Error('存储名称至少 2 位，且只能包含字母、数字、下划线和横线')
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
    alert((e as Error).message)
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
  <Modal :show="show" :title="editMode ? '修改存储配置' : '添加存储'" @close="emit('close')">
    <div v-if="!editMode" class="stepper">
      <div class="step" :class="{ active: step === 0, done: step > 0 }">1. 选择类型</div>
      <div class="step" :class="{ active: step === 1, done: step > 1 }">2. 配置选项</div>
      <div class="step" :class="{ active: step === 2, done: step > 2 }">3. 存储名称</div>
      <div class="step" :class="{ active: step === 3 }">4. 保存</div>
    </div>
    <div v-else class="stepper">
      <div class="step" :class="{ active: step === 1, done: step > 1 }">1. 配置选项</div>
      <div class="step" :class="{ active: step === 3 }">2. 确认保存</div>
    </div>

    <div v-if="step === 0">
      <input v-model="search" type="text" placeholder="搜索存储类型..." style="width: 100%; margin-bottom: 16px" />
      <div class="provider-grid">
        <div
          v-for="p in filteredProviders"
          :key="p.Name"
          class="provider-card"
          :class="{ selected: selectedProviderName === p.Name }"
          @click="selectProvider(p)"
        >
          <strong>{{ p.Name }}</strong>
          <div style="margin-top: 6px; color: #6b7280; font-size: 12px">{{ p.Description }}</div>
        </div>
      </div>
    </div>

    <div v-if="step === 1">
      <div class="field-grid" v-if="groupedOptions.required.length">
        <div v-for="opt in groupedOptions.required" :key="opt.Name" class="field-item">
          <label>{{ opt.Name }} <span style="color: #dc2626">*</span></label>
          <input
            v-if="!opt.Examples || !opt.Examples.length"
            v-model="remoteOptions[opt.Name]"
            :type="opt.IsPassword ? 'password' : 'text'"
            :placeholder="opt.Help"
          />
          <select v-else v-model="remoteOptions[opt.Name]">
            <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ ex.Help ? ` — ${ex.Help}` : '' }}</option>
          </select>
          <div class="field-help">{{ opt.Help }}</div>
        </div>
      </div>

      <div class="field-grid" v-if="groupedOptions.optional.length" style="margin-top: 16px">
        <div v-for="opt in groupedOptions.optional" :key="opt.Name" class="field-item">
          <label>{{ opt.Name }}</label>
          <input
            v-if="!opt.Examples || !opt.Examples.length"
            v-model="remoteOptions[opt.Name]"
            :type="opt.IsPassword ? 'password' : 'text'"
            :placeholder="opt.Help"
          />
          <select v-else v-model="remoteOptions[opt.Name]">
            <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ ex.Help ? ` — ${ex.Help}` : '' }}</option>
          </select>
          <div class="field-help">{{ opt.Help }}</div>
        </div>
      </div>

      <details v-if="allOptions.some(o => o.Advanced)" style="margin-top: 16px" :open="showAdvancedOptions">
        <summary style="cursor: pointer; color: #6b7280; font-size: 14px" @click.prevent="showAdvancedOptions = !showAdvancedOptions">
          {{ showAdvancedOptions ? '隐藏高级选项' : '显示高级选项' }}
        </summary>
        <div class="field-grid" v-if="groupedOptions.advanced.length" style="margin-top: 12px">
          <div v-for="opt in groupedOptions.advanced" :key="opt.Name" class="field-item">
            <label>{{ opt.Name }}</label>
            <input
              v-if="!opt.Examples || !opt.Examples.length"
              v-model="remoteOptions[opt.Name]"
              :type="opt.IsPassword ? 'password' : 'text'"
              :placeholder="opt.Help"
            />
            <select v-else v-model="remoteOptions[opt.Name]">
              <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ ex.Help ? ` — ${ex.Help}` : '' }}</option>
            </select>
            <div class="field-help">{{ opt.Help }}</div>
          </div>
        </div>
      </details>

      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button v-if="!editMode" class="ghost" @click="prevStep">上一步</button>
        <button @click="nextStep">下一步</button>
      </div>
    </div>

    <div v-if="step === 2 && !editMode">
      <div class="field-item">
        <label>存储名称 <span style="color: #dc2626">*</span></label>
        <input v-model="remoteName" type="text" placeholder="输入存储名称，如: mydrive" />
        <div class="field-help">至少 2 位，只能包含字母、数字、下划线和横线。</div>
      </div>
      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button class="ghost" @click="prevStep">上一步</button>
        <button @click="nextStep" :disabled="!validateName(remoteName)">下一步</button>
      </div>
    </div>

    <div v-if="step === 3">
      <div class="card" style="background: #f0f9ff; margin-bottom: 16px">
        <div><strong>存储类型:</strong> {{ selectedProvider?.Name }}</div>
        <div><strong>存储名称:</strong> {{ remoteName }}</div>
      </div>
      <div class="actions" style="justify-content: space-between">
        <button v-if="!editMode" class="ghost" @click="prevStep" :disabled="success">上一步</button>
        <button v-if="!success" color="primary" :disabled="creating" @click="create">
          {{ creating ? '保存中...' : (editMode ? '保存修改' : '保存') }}
        </button>
        <button v-else color="primary" @click="emit('close')">
          {{ editMode ? '修改成功，点击关闭' : '添加成功，点击关闭' }}
        </button>
      </div>
    </div>
  </Modal>
</template>

<style scoped>
.stepper { display: flex; gap: 8px; margin-bottom: 20px; }
/* Dark theme defaults */
.step { padding: 8px 12px; border-radius: 8px; background: #1f2937; color: #cbd5e1; font-size: 14px; border:1px solid #374151 }
.step.active { background: #1e3a5f; color: #60a5fa; border-color:#2563eb }
.step.done { background: #0a2f22; color: #34d399; border-color:#14532d }
/* Light theme overrides */
body.light .step { background: #f3f4f6; color: #6b7280; border-color:#e5e7eb }
body.light .step.active { background: #dbeafe; color: #1d4ed8; border-color:#93c5fd }
body.light .step.done { background: #dcfce7; color: #166534; border-color:#86efac }
.provider-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 12px; max-height: 360px; overflow: auto; }
/* Dark theme defaults */
.provider-card { padding: 14px; border: 1px solid #374151; border-radius: 12px; cursor: pointer; background: #1f2937; color: #e5e7eb; }
.provider-card:hover { border-color: #60a5fa; background: #111827; }
.provider-card.selected { border-color: #2563eb; background: #111827; }
/* Light theme overrides */
body.light .provider-card { border-color: #e5e7eb; background: #fff; color: #111827; }
body.light .provider-card:hover { border-color: #93c5fd; background: #eff6ff; }
body.light .provider-card.selected { border-color: #2563eb; background: #eff6ff; }
.field-grid { display: grid; gap: 16px; }
.field-item { display: grid; gap: 8px; }
.field-item label { font-weight: 600; }
.field-item input, .field-item select { width: 100%; padding: 10px 12px; border-radius: 10px; border: 1px solid #d1d5db; font: inherit; }
.field-help { font-size: 12px; color: #6b7280; line-height: 1.5; }
.actions { display: flex; gap: 8px; }
button { padding: 10px 14px; border-radius: 10px; border: none; background: #2563eb; color: white; cursor: pointer; }
button.ghost { background: #eef2ff; color: #1d4ed8; }
button:disabled { opacity: .6; cursor: not-allowed; }
</style>
