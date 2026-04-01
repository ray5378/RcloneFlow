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
const providers = ref<Provider[]>([])

const filteredProviders = computed(() => {
  if (!providers.value) return []
  return providers.value
    .filter(p => {
      if (!search.value) return true
      return p.Name.toLowerCase().startsWith(search.value.toLowerCase())
    })
    .sort((a, b) => a.Name.localeCompare(b.Name))
})

// Show all options except those tied to a specific provider (Provider field set means it's provider-specific)
const providerOptions = computed(() => {
  if (!selectedProvider.value) return []
  return (selectedProvider.value.Options || []).filter((opt: ProviderOption) => {
    if (opt.Provider && opt.Provider !== '') return false
    return true
  })
})

// Group options: required first, then optional, then advanced
const groupedOptions = computed(() => {
  const required: ProviderOption[] = []
  const optional: ProviderOption[] = []
  const advanced: ProviderOption[] = []
  
  for (const opt of providerOptions.value) {
    if (opt.Required) {
      required.push(opt)
    } else if (opt.Advanced) {
      advanced.push(opt)
    } else {
      optional.push(opt)
    }
  }
  
  return { required, optional, advanced }
})

watch(() => props.show, async (val) => {
  if (val) {
    step.value = 0
    search.value = ''
    selectedProviderName.value = ''
    selectedProvider.value = null
    remoteName.value = ''
    remoteOptions.value = {}
    creating.value = false
    success.value = false
    
    // Load providers
    const data = await api.listProviders()
    providers.value = data.providers || []
  }
})

function selectProvider(provider: Provider) {
  selectedProvider.value = provider
  selectedProviderName.value = provider.Name
  step.value = 1
  remoteOptions.value = {}
  
  // Set default values using DefaultStr (string representation)
  for (const opt of providerOptions.value) {
    if (opt.DefaultStr && opt.DefaultStr !== '') {
      remoteOptions.value[opt.Name] = opt.DefaultStr
    }
  }
  
  // SMB 存储特殊默认配置
  if (provider.Name === 'smb') {
    remoteOptions.value['idle_timeout'] = '0s'
    remoteOptions.value['encoding'] = 'None'
  }
}

function nextStep() {
  if (step.value === 0) {
    step.value = 1
  } else if (step.value === 1) {
    if (!props.editMode) {
      step.value = 2
    } else {
      step.value = 3
    }
  } else if (step.value === 2) {
    step.value = 3
  }
}

function prevStep() {
  if (step.value === 3) {
    step.value = props.editMode ? 1 : 2
  } else if (step.value === 2) {
    step.value = 1
  } else if (step.value === 1) {
    step.value = 0
  }
}

async function create() {
  creating.value = true
  try {
    const params: Record<string, unknown> = {}
    for (const key in remoteOptions.value) {
      if (remoteOptions.value[key] !== '') {
        params[key] = remoteOptions.value[key]
      }
    }
    
    if (props.editMode && props.editName) {
      await api.updateRemote(remoteName.value, selectedProvider.value!.Name, params)
    } else {
      await api.createRemote(remoteName.value, selectedProvider.value!.Name, params)
    }
    success.value = true
    setTimeout(() => {
      emit('success')
      emit('close')
    }, 1000)
  } catch (e) {
    alert((e as Error).message)
  } finally {
    creating.value = false
  }
}

defineExpose({ loadConfig: async (name: string) => {
  // Wait for providers to be loaded
  while (!providers.value.length) {
    await new Promise(resolve => setTimeout(resolve, 50))
  }
  
  const config = await api.getRemoteConfig(name)
  remoteName.value = name
  const type = config.type as string
  const provider = providers.value.find(p => p.Name === type)
  if (provider) {
    selectedProvider.value = provider
    selectedProviderName.value = type
    remoteOptions.value = {}
    for (const key in config) {
      if (key !== 'type' && key !== 'name') {
        remoteOptions.value[key] = String(config[key])
      }
    }
    // Go to configuration step (step 1), not save step
    step.value = 1
  }
}})
</script>

<template>
  <Modal :show="show" :title="editMode ? '修改存储配置' : '添加存储'" @close="emit('close')">
    <!-- Stepper -->
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

    <!-- Step 0: Select Provider -->
    <div v-if="step === 0">
      <input
        v-model="search"
        type="text"
        placeholder="搜索存储类型..."
        style="width: 100%; margin-bottom: 16px"
      />
      <div class="provider-grid">
        <div
          v-for="p in filteredProviders"
          :key="p.Name"
          class="provider-card"
          :class="{ selected: selectedProviderName === p.Name }"
          @click="selectProvider(p)"
        >
          <strong>{{ p.Name }}</strong>
        </div>
      </div>
    </div>

    <!-- Step 1: Configure Options -->
    <div v-if="step === 1">
      <!-- Required options -->
      <div v-if="groupedOptions.required.length" class="field-grid">
        <div v-for="opt in groupedOptions.required" :key="opt.Name" class="field-item">
          <label>{{ opt.Name }} <span style="color: #dc2626">*</span></label>
          <select v-if="opt.Examples && opt.Examples.length" v-model="remoteOptions[opt.Name]">
            <option value="">选择...</option>
            <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">
              {{ ex.Help }}
            </option>
          </select>
          <input
            v-else-if="opt.IsPassword"
            v-model="remoteOptions[opt.Name]"
            type="password"
            :placeholder="opt.Help"
          />
          <input
            v-else
            v-model="remoteOptions[opt.Name]"
            type="text"
            :placeholder="opt.Help"
          />
        </div>
      </div>

      <!-- Optional options -->
      <div v-if="groupedOptions.optional.length" class="field-grid" style="margin-top: 16px">
        <div v-for="opt in groupedOptions.optional" :key="opt.Name" class="field-item">
          <label>{{ opt.Name }}</label>
          <select v-if="opt.Examples && opt.Examples.length" v-model="remoteOptions[opt.Name]">
            <option value="">选择...</option>
            <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">
              {{ ex.Help }}
            </option>
          </select>
          <input
            v-else-if="opt.IsPassword"
            v-model="remoteOptions[opt.Name]"
            type="password"
            :placeholder="opt.Help"
          />
          <input
            v-else
            v-model="remoteOptions[opt.Name]"
            type="text"
            :placeholder="opt.Help"
          />
        </div>
      </div>

      <!-- Advanced options -->
      <details v-if="groupedOptions.advanced.length" style="margin-top: 16px">
        <summary style="cursor: pointer; color: #6b7280; font-size: 14px">
          高级选项 ({{ groupedOptions.advanced.length }})
        </summary>
        <div class="field-grid" style="margin-top: 12px">
          <div v-for="opt in groupedOptions.advanced" :key="opt.Name" class="field-item">
            <label>{{ opt.Name }}</label>
            <select v-if="opt.Examples && opt.Examples.length" v-model="remoteOptions[opt.Name]">
              <option value="">选择...</option>
              <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">
                {{ ex.Help }}
              </option>
            </select>
            <input
              v-else-if="opt.IsPassword"
              v-model="remoteOptions[opt.Name]"
              type="password"
              :placeholder="opt.Help"
            />
            <input
              v-else
              v-model="remoteOptions[opt.Name]"
              type="text"
              :placeholder="opt.Help"
            />
          </div>
        </div>
      </details>

      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button v-if="!editMode" class="ghost" @click="prevStep">上一步</button>
        <button @click="nextStep">{{ editMode ? '下一步' : '下一步' }}</button>
      </div>
    </div>

    <!-- Step 2: Storage Name (if not editMode) -->
    <div v-if="step === 2 && !editMode">
      <div class="field-item">
        <label>存储名称 <span style="color: #dc2626">*</span></label>
        <input
          v-model="remoteName"
          type="text"
          placeholder="输入存储名称，如: mydrive"
        />
      </div>
      <div class="actions" style="margin-top: 16px; justify-content: space-between">
        <button class="ghost" @click="prevStep">上一步</button>
        <button @click="nextStep" :disabled="!remoteName">下一步</button>
      </div>
    </div>

    <!-- Step 3: Save -->
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
