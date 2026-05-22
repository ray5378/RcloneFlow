<script setup lang="ts">
import type { ProviderOption } from '../../types'
import { getOptionLabel, getOptionHelp, getOptionPlaceholder, getExampleHelp } from './remoteUtils'

const props = defineProps<{
  requiredOptions: ProviderOption[]
  optionalOptions: ProviderOption[]
  advancedOptions: ProviderOption[]
  allOptions: ProviderOption[]
  remoteOptions: Record<string, string>
  showAdvancedOptions: boolean
}>()

const emit = defineEmits<{
  'update:remoteOptions': [options: Record<string, string>]
  'update:showAdvancedOptions': [value: boolean]
}>()

function updateOption(name: string, value: string) {
  const updated = { ...props.remoteOptions, [name]: value }
  emit('update:remoteOptions', updated)
}
</script>

<template>
  <div>
    <div class="field-grid" v-if="requiredOptions.length">
      <div v-for="opt in requiredOptions" :key="opt.Name" class="field-item">
        <label>{{ getOptionLabel(opt) }} <small v-if="getOptionLabel(opt) !== opt.Name" class="subkey">{{ opt.Name }}</small> <span style="color: #dc2626">*</span></label>
        <input
          v-if="!opt.Examples || !opt.Examples.length"
          :value="remoteOptions[opt.Name] || ''"
          :type="opt.IsPassword ? 'password' : 'text'"
          :placeholder="getOptionPlaceholder(opt)"
          @input="updateOption(opt.Name, ($event.target as HTMLInputElement).value)"
        />
        <select v-else :value="remoteOptions[opt.Name] || ''" @change="updateOption(opt.Name, ($event.target as HTMLSelectElement).value)">
          <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ getExampleHelp(ex.Help) ? ` — ${getExampleHelp(ex.Help)}` : '' }}</option>
        </select>
        <div class="field-help">{{ getOptionHelp(opt) }}</div>
      </div>
    </div>

    <div class="field-grid" v-if="optionalOptions.length" style="margin-top: 16px">
      <div v-for="opt in optionalOptions" :key="opt.Name" class="field-item">
        <label>{{ getOptionLabel(opt) }} <small v-if="getOptionLabel(opt) !== opt.Name" class="subkey">{{ opt.Name }}</small></label>
        <input
          v-if="!opt.Examples || !opt.Examples.length"
          :value="remoteOptions[opt.Name] || ''"
          :type="opt.IsPassword ? 'password' : 'text'"
          :placeholder="getOptionPlaceholder(opt)"
          @input="updateOption(opt.Name, ($event.target as HTMLInputElement).value)"
        />
        <select v-else :value="remoteOptions[opt.Name] || ''" @change="updateOption(opt.Name, ($event.target as HTMLSelectElement).value)">
          <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ getExampleHelp(ex.Help) ? ` — ${getExampleHelp(ex.Help)}` : '' }}</option>
        </select>
        <div class="field-help">{{ getOptionHelp(opt) }}</div>
      </div>
    </div>

    <details v-if="allOptions.some(o => o.Advanced)" style="margin-top: 16px" :open="showAdvancedOptions">
      <summary style="cursor: pointer; color: #6b7280; font-size: 14px" @click.prevent="emit('update:showAdvancedOptions', !showAdvancedOptions)">
        {{ showAdvancedOptions ? $t('remote.hideAdvanced') : $t('remote.showAdvanced') }}
      </summary>
      <div class="field-grid" v-if="advancedOptions.length" style="margin-top: 12px">
        <div v-for="opt in advancedOptions" :key="opt.Name" class="field-item">
          <label>{{ getOptionLabel(opt) }} <small v-if="getOptionLabel(opt) !== opt.Name" class="subkey">{{ opt.Name }}</small></label>
          <input
            v-if="!opt.Examples || !opt.Examples.length"
            :value="remoteOptions[opt.Name] || ''"
            :type="opt.IsPassword ? 'password' : 'text'"
            :placeholder="getOptionPlaceholder(opt)"
            @input="updateOption(opt.Name, ($event.target as HTMLInputElement).value)"
          />
          <select v-else :value="remoteOptions[opt.Name] || ''" @change="updateOption(opt.Name, ($event.target as HTMLSelectElement).value)">
            <option v-for="ex in opt.Examples" :key="ex.Value" :value="ex.Value">{{ ex.Value }}{{ getExampleHelp(ex.Help) ? ` — ${getExampleHelp(ex.Help)}` : '' }}</option>
          </select>
          <div class="field-help">{{ getOptionHelp(opt) }}</div>
        </div>
      </div>
    </details>
  </div>
</template>

<style scoped>
.field-grid { display: grid; gap: 16px; }
.field-item { display: grid; gap: 8px; }
.field-item label { font-weight: 600; }
.field-item input, .field-item select { width: 100%; padding: 10px 12px; border-radius: 10px; border: 1px solid #d1d5db; font: inherit; }
.field-help { font-size: 12px; color: #6b7280; line-height: 1.5; }
.subkey { opacity: .65; margin-left: 6px; font-weight: 400; }
</style>