import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { getInstallStatus, testDatabase, executeInstall } from '@/api/install'
import type { DatabaseInput, InstallRequest, TestDBResult } from '@/api/install'

export const useInstallStore = defineStore('install', () => {
  const installed = ref<boolean | null>(null)
  const loading = ref(false)

  const isInstalled = computed(() => installed.value === true)

  async function fetchStatus(): Promise<void> {
    try {
      const result = await getInstallStatus()
      installed.value = result.installed
    } catch {
      installed.value = true
    }
  }

  async function testDb(input: DatabaseInput): Promise<TestDBResult> {
    return testDatabase(input)
  }

  async function doInstall(req: InstallRequest): Promise<void> {
    loading.value = true
    try {
      await executeInstall(req)
      installed.value = true
    } finally {
      loading.value = false
    }
  }

  function reset() {
    installed.value = null
    loading.value = false
  }

  return { installed, loading, isInstalled, fetchStatus, testDb, doInstall, reset }
})
