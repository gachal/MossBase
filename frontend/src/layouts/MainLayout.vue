<template>
  <el-container class="main-layout">
    <el-header class="main-header">
      <div class="header-left">
        <h1 class="logo" @click="$router.push('/')">MossBase</h1>
      </div>
      <div class="header-right">
        <el-dropdown v-if="authStore.isAuthenticated" @command="handleCommand">
          <span class="user-info">
            <el-avatar :size="32" :src="authStore.user?.avatar" />
            <span class="username">{{ authStore.user?.username }}</span>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="profile">个人资料</el-dropdown-item>
              <el-dropdown-item v-if="authStore.isAdmin" command="admin">管理后台</el-dropdown-item>
              <el-dropdown-item divided command="logout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()

function handleCommand(command: string) {
  switch (command) {
    case 'profile': router.push('/profile'); break
    case 'admin': router.push('/admin'); break
    case 'logout': authStore.clearAuth(); router.push('/login'); break
  }
}
</script>

<style scoped>
.main-layout { min-height: 100vh; }
.main-header { display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #e4e7ed; }
.logo { cursor: pointer; font-size: 20px; color: var(--el-color-primary); margin: 0; }
.user-info { display: flex; align-items: center; gap: 8px; cursor: pointer; }
.username { font-size: 14px; }
.main-layout :deep(.el-main) { overflow: visible; }
</style>
