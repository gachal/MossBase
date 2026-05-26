<template>
  <div class="dashboard" v-loading="loading">
    <h2>系统概览</h2>
    <div class="stats-grid">
      <el-card shadow="hover">
        <div class="stat-value">{{ stats?.total_users ?? 0 }}</div>
        <div class="stat-label">用户总数</div>
      </el-card>
      <el-card shadow="hover">
        <div class="stat-value">{{ stats?.total_spaces ?? 0 }}</div>
        <div class="stat-label">空间总数</div>
      </el-card>
      <el-card shadow="hover">
        <div class="stat-value">{{ stats?.total_pages ?? 0 }}</div>
        <div class="stat-label">页面总数</div>
      </el-card>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import type { DashboardStats } from '@/api/admin'
import { getDashboardStats } from '@/api/admin'

const loading = ref(false)
const stats = ref<DashboardStats | null>(null)

onMounted(async () => {
  loading.value = true
  try {
    stats.value = await getDashboardStats()
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.stats-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 16px; margin-top: 16px; }
.stat-value { font-size: 32px; font-weight: 700; color: var(--el-color-primary); }
.stat-label { font-size: 14px; color: var(--el-text-color-secondary); margin-top: 4px; }
</style>
