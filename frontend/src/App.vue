<script lang="ts" setup>
import { onMounted, ref } from 'vue'
import MainLayout from '@/components/MainLayout.vue'
import { useSettingStore } from '@/stores'
import { GetConfig } from '../wailsjs/go/main/App'

const ready = ref(false)
const setting = useSettingStore()

onMounted(async () => {
  try { const c = await GetConfig(); setting.setConfig(c as unknown as import('./types').AppConfig) } catch (e) { console.warn(e) }
  ready.value = true
})
</script>

<template>
  <div class="h-screen w-screen bg-white dark:bg-[#1e1e1e]">
    <MainLayout v-if="ready" />
    <div v-else class="flex items-center justify-center h-full text-gray-400 text-sm">Loading...</div>
  </div>
</template>
