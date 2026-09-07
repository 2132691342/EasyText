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
  <div class="h-screen w-screen bg-[var(--et-bg)] text-[var(--et-fg)]">
    <MainLayout v-if="ready" />
    <div v-else class="flex items-center justify-center h-full text-[var(--et-fg-subtle)] text-sm">正在加载…</div>
  </div>
</template>
