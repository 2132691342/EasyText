<template>
  <div class="script-manager flex h-full bg-[var(--theme-bg)] text-[var(--theme-fg)]">
    <!-- Left: script list -->
    <div class="w-2/5 border-r flex flex-col" style="border-color: var(--theme-gutter-bg)">
      <div class="p-2 border-b" style="border-color: var(--theme-gutter-bg)">
        <ElInput v-model="searchQuery" placeholder="搜索脚本..." size="small" :prefix-icon="SearchIcon" />
      </div>
      <ElTable :data="filteredScripts" size="small" highlight-current-row stripe
               style="width: 100%; height: 0; flex: 1; overflow-y: auto"
               @row-click="selectScript">
        <ElTableColumn prop="name" label="名称" min-width="100" />
        <ElTableColumn prop="language" label="语言" width="70">
          <template #default="{ row }">
            <ElTag type="info" size="small">Lua</ElTag>
          </template>
        </ElTableColumn>
        <ElTableColumn prop="enabled" label="启用" width="60">
          <template #default="{ row }">
            <ElSwitch v-model="row.enabled" size="small" @click.stop />
          </template>
        </ElTableColumn>
      </ElTable>
      <div class="p-2 border-t" style="border-color: var(--theme-gutter-bg)">
        <ElButton size="small" @click="createNew" type="primary">
          <Plus class="w-4 h-4 mr-1" /> 新建
        </ElButton>
      </div>
    </div>

    <!-- Right: editor -->
    <div class="flex-1 flex flex-col" v-if="editingScript">
      <div class="p-3 border-b space-y-2" style="border-color: var(--theme-gutter-bg)">
        <div class="flex gap-2">
          <ElInput v-model="editingScript.name" placeholder="脚本名称" size="small" class="flex-1" />
          <ElSelect v-model="editingScript.language" size="small" style="width: 120px">
            <ElOption value="lua" label="Lua" />
          </ElSelect>
          <ElInput v-model="editingScript.menuGroup" placeholder="菜单分组" size="small" style="width: 120px" />
        </div>
        <ElInput v-model="editingScript.description" placeholder="描述（可选）" size="small" />
      </div>
      <textarea v-model="editingScript.code"
                class="flex-1 p-3 font-mono text-sm resize-none outline-none bg-transparent"
                placeholder="-- 在此编写脚本代码..."
                style="color: var(--theme-fg)" />
      <div class="p-2 border-t flex gap-2" style="border-color: var(--theme-gutter-bg)">
        <ElButton size="small" type="primary" @click="saveScript">
          <Save class="w-4 h-4 mr-1" /> 保存
        </ElButton>
        <ElButton size="small" @click="runScript">
          <Play class="w-4 h-4 mr-1" /> 运行测试
        </ElButton>
        <ElButton size="small" type="danger" @click="deleteScript">
          <Trash2 class="w-4 h-4 mr-1" /> 删除
        </ElButton>
      </div>
    </div>

    <!-- Empty state -->
    <div v-else class="flex-1 flex items-center justify-center text-sm text-[var(--theme-comment)]">
      选择或新建一个脚本
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, h } from 'vue'
import { ElButton, ElInput, ElSelect, ElOption, ElTable, ElTableColumn, ElTag, ElSwitch, ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Save, Trash2, Play, Search } from 'lucide-vue-next'
import { ListScripts, GetScript, SaveScript as SaveScriptApi, DeleteScript as DeleteScriptApi, ExecuteScript } from '../../wailsjs/go/main/App'

interface ScriptInfo {
  id: string
  name: string
  description: string
  language: string
  code: string
  enabled: boolean
  menuGroup: string
  createdAt: string
}

const SearchIcon = h(Search)

const searchQuery = ref('')
const scripts = ref<ScriptInfo[]>([])
const editingScript = ref<ScriptInfo | null>(null)
const selectedId = ref('')

const filteredScripts = computed(() => {
  if (!searchQuery.value) return scripts.value
  const q = searchQuery.value.toLowerCase()
  return scripts.value.filter(s => s.name.toLowerCase().includes(q) || s.description.toLowerCase().includes(q))
})

async function loadScripts() {
  try {
    scripts.value = await ListScripts()
  } catch { scripts.value = [] }
}

function selectScript(row: ScriptInfo) {
  selectedId.value = row.id
  loadScriptDetail(row.id)
}

async function loadScriptDetail(id: string) {
  try {
    const detail = await GetScript(id)
    if (detail) editingScript.value = detail
  } catch (err: any) {
    ElMessage.error('加载脚本失败: ' + (err?.message || String(err)))
  }
}

function createNew() {
  const id = 'script-' + Date.now()
  editingScript.value = {
    id, name: '新脚本', description: '', language: 'lua',
    code: '', enabled: true, menuGroup: '', createdAt: new Date().toISOString(),
  }
  if (!scripts.value.find(s => s.id === id)) {
    scripts.value.push(editingScript.value)
  }
}

async function saveScript() {
  if (!editingScript.value) return
  try {
    await SaveScriptApi(editingScript.value)
    ElMessage.success('脚本已保存')
    await loadScripts()
  } catch (err: any) {
    ElMessage.error('保存失败: ' + (err?.message || String(err)))
  }
}

async function runScript() {
  if (!editingScript.value) return
  try {
    const result = await ExecuteScript(editingScript.value.id, {
      filePath: '', content: '', selection: '',
      cursorLine: 1, cursorCol: 1, language: '',
    })
    if (result?.success) {
      ElMessage.success('执行成功' + (result.output ? ': ' + result.output : ''))
    } else {
      ElMessage.error('执行失败: ' + (result?.error || '未知错误'))
    }
  } catch (err: any) {
    ElMessage.error('执行失败: ' + (err?.message || String(err)))
  }
}

async function deleteScript() {
  if (!editingScript.value) return
  try {
    await ElMessageBox.confirm('确定删除脚本 "' + editingScript.value.name + '"?', '确认删除')
    await DeleteScriptApi(editingScript.value.id)
    scripts.value = scripts.value.filter(s => s.id !== editingScript.value!.id)
    editingScript.value = null
    ElMessage.success('脚本已删除')
  } catch (e) { console.warn(e) }
}

onMounted(loadScripts)
</script>
