<script lang="ts" setup>
/**
 * 列块编辑
 *  - 用 ModalOverlay 替代手撸 Teleport
 *  - 按钮 token 化
 */
import { ref, watch, nextTick } from 'vue'
import ModalOverlay from './ModalOverlay.vue'

const props = defineProps<{ visible: boolean }>()
const emit = defineEmits<{ (e: 'close'): void }>()

const insertTextEnabled = ref(true)
const insertText = ref('')
const insertDateEnabled = ref(false)
const dateFormat = ref<'short' | 'long' | 'iso' | 'unix'>('short')
const insertNumEnabled = ref(false)
const initNum = ref('1')
const incNum = ref(1)
const repeNum = ref(1)
const addPrefix = ref(false)
const prefix = ref('')
const radix = ref<10 | 16 | 8 | 2>(10)
const capital = ref(true)
const caseConversion = ref<'' | 'upper' | 'lower' | 'pascal' | 'camel' | 'snake' | 'kebab'>('')

const textInputRef = ref<HTMLInputElement | null>(null)
watch(() => props.visible, async (v) => {
  if (v) {
    insertTextEnabled.value = true
    insertNumEnabled.value = false
    insertDateEnabled.value = false
    insertText.value = ''
    initNum.value = '1'
    incNum.value = 1
    repeNum.value = 1
    addPrefix.value = false
    prefix.value = ''
    radix.value = 10
    capital.value = true
    dateFormat.value = 'short'
    caseConversion.value = ''
    await nextTick()
    textInputRef.value?.focus()
  }
})

function getDateString(): string {
  const now = new Date()
  switch (dateFormat.value) {
    case 'short':
      return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
    case 'long':
      return `${now.getFullYear()}年${now.getMonth() + 1}月${now.getDate()}日 ${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}:${String(now.getSeconds()).padStart(2, '0')}`
    case 'iso':
      return now.toISOString()
    case 'unix':
      return String(Math.floor(now.getTime() / 1000))
    default:
      return ''
  }
}

function ok() {
  if (insertTextEnabled.value && !insertNumEnabled.value && !insertDateEnabled.value) {
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: 'column-insert-text', args: [insertText.value] }
    }))
  } else if (insertDateEnabled.value && !insertTextEnabled.value && !insertNumEnabled.value) {
    const dateStr = getDateString()
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: 'column-insert-text', args: [dateStr] }
    }))
  } else if (insertNumEnabled.value && !insertTextEnabled.value && !insertDateEnabled.value) {
    const init = parseInt(initNum.value) || 0
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: {
        cmd: 'column-insert-num',
        args: [{
          init,
          inc: incNum.value,
          repeat: repeNum.value,
          prefix: addPrefix.value ? prefix.value : '',
          radix: radix.value,
          capital: capital.value,
        }]
      }
    }))
  }
  if (caseConversion.value) {
    document.dispatchEvent(new CustomEvent('editor-command', {
      detail: { cmd: `case-${caseConversion.value}`, args: [] }
    }))
  }
  emit('close')
}
</script>

<template>
  <ModalOverlay :visible="visible" title="列块编辑" size="md" @close="emit('close')">
    <div class="ce-grid">
      <!-- 左：4 个分组 -->
      <div class="ce-fields">
        <fieldset class="ce-field">
          <legend class="ce-legend-checkable">
            <input type="checkbox" v-model="insertTextEnabled"
              @change="insertTextEnabled && (insertNumEnabled = false, insertDateEnabled = false)" />
            插入文本
          </legend>
          <input
            ref="textInputRef"
            v-model="insertText"
            :disabled="!insertTextEnabled"
            maxlength="1024"
            class="et-input"
            @keydown.enter="ok"
          />
        </fieldset>

        <fieldset class="ce-field" :class="{ 'is-dim': !insertDateEnabled }">
          <legend class="ce-legend-checkable">
            <input type="checkbox" v-model="insertDateEnabled"
              @change="insertDateEnabled && (insertTextEnabled = false, insertNumEnabled = false)" />
            插入日期/时间
          </legend>
          <select v-model="dateFormat" class="et-select" :disabled="!insertDateEnabled">
            <option value="short">短日期 (2026-07-17)</option>
            <option value="long">长日期 (2026年7月17日 HH:mm:ss)</option>
            <option value="iso">ISO 8601</option>
            <option value="unix">Unix 时间戳</option>
          </select>
        </fieldset>

        <fieldset class="ce-field">
          <legend>大小写转换</legend>
          <select v-model="caseConversion" class="et-select">
            <option value="">不转换</option>
            <option value="upper">大写 (UPPERCASE)</option>
            <option value="lower">小写 (lowercase)</option>
            <option value="pascal">PascalCase</option>
            <option value="camel">camelCase</option>
            <option value="snake">snake_case</option>
            <option value="kebab">kebab-case</option>
          </select>
        </fieldset>

        <fieldset class="ce-field" :class="{ 'is-dim': !insertNumEnabled }">
          <legend class="ce-legend-checkable">
            <input type="checkbox" v-model="insertNumEnabled"
              @change="insertNumEnabled && (insertTextEnabled = false, insertDateEnabled = false)" />
            插入数字
          </legend>
          <div class="ce-num" :class="{ 'is-off': !insertNumEnabled }">
            <label class="ce-num-row">
              <span>初始值</span>
              <input v-model="initNum" maxlength="11" class="et-input" />
            </label>
            <label class="ce-num-row">
              <span>步进</span>
              <input type="number" v-model.number="incNum" min="-100" class="et-input" />
            </label>
            <label class="ce-num-row">
              <span>重复次数</span>
              <input type="number" v-model.number="repeNum" min="1" class="et-input" />
            </label>
            <div class="ce-num-row">
              <label class="ce-inline-check">
                <input type="checkbox" v-model="addPrefix" /> 前缀字符串
              </label>
              <input v-model="prefix" :disabled="!addPrefix" class="et-input" />
            </div>
            <fieldset class="ce-format">
              <legend>格式</legend>
              <div class="ce-format-grid">
                <label><input type="radio" v-model="radix" :value="10" /> 十进制</label>
                <label><input type="radio" v-model="radix" :value="16" /> 十六进制</label>
                <label><input type="radio" v-model="radix" :value="8" /> 八进制</label>
                <label><input type="radio" v-model="radix" :value="2" /> 二进制</label>
              </div>
              <label v-if="radix === 16" class="ce-inline-check">
                <input type="checkbox" v-model="capital" /> 大写
              </label>
            </fieldset>
          </div>
        </fieldset>
      </div>
    </div>

    <template #footer>
      <button class="et-btn" @click="emit('close')">取消</button>
      <button class="et-btn et-btn-primary" @click="ok">确定</button>
    </template>
  </ModalOverlay>
</template>

<style scoped>
.ce-grid {
  display: flex;
  gap: var(--et-space-3);
}
.ce-fields {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-2);
  flex: 1;
}
.ce-field {
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  padding: var(--et-space-2) var(--et-space-3);
  margin: 0;
}
.ce-field.is-dim { opacity: .6; }
.ce-field legend {
  padding: 0 var(--et-space-1);
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
  font-weight: var(--et-fw-medium);
}
.ce-legend-checkable {
  display: inline-flex;
  align-items: center;
  gap: var(--et-space-1);
}
.ce-legend-checkable input { accent-color: var(--et-accent); cursor: pointer; }

.ce-num {
  display: flex;
  flex-direction: column;
  gap: var(--et-space-2);
  margin-top: var(--et-space-1);
}
.ce-num.is-off { pointer-events: none; }
.ce-num-row {
  display: flex;
  align-items: center;
  gap: var(--et-space-2);
  font-size: var(--et-text-sm);
  color: var(--et-fg);
}
.ce-num-row > span { width: 60px; flex-shrink: 0; color: var(--et-fg-muted); font-size: var(--et-text-xs); }
.ce-inline-check {
  display: inline-flex;
  align-items: center;
  gap: var(--et-space-1);
  font-size: var(--et-text-sm);
  color: var(--et-fg);
  cursor: pointer;
  width: 90px;
  flex-shrink: 0;
}
.ce-inline-check input { accent-color: var(--et-accent); cursor: pointer; }

.ce-format {
  border: 1px solid var(--et-border);
  border-radius: var(--et-radius-sm);
  padding: var(--et-space-2);
  margin: 0;
}
.ce-format legend {
  padding: 0 var(--et-space-1);
  font-size: var(--et-text-xs);
  color: var(--et-fg-muted);
}
.ce-format-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--et-space-1) var(--et-space-2);
  font-size: var(--et-text-sm);
  color: var(--et-fg);
}
.ce-format-grid label { display: inline-flex; align-items: center; gap: var(--et-space-1); cursor: pointer; }
.ce-format-grid input { accent-color: var(--et-accent); cursor: pointer; }
</style>