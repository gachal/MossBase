<template>
  <el-container class="admin-layout">
    <el-aside width="220px" class="admin-aside">
      <div class="admin-logo" @click="$router.push('/')">MossBase</div>
      <el-menu :default-active="route.path" router>
        <el-menu-item index="/spaces" data-test="back-to-wiki">
          <el-icon><Back /></el-icon>
          <span>返回 Wiki</span>
        </el-menu-item>
        <el-menu-item index="/admin/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/admin/users">
          <el-icon><User /></el-icon>
          <span>用户管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/spaces">
          <el-icon><Folder /></el-icon>
          <span>空间管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/pages">
          <el-icon><Document /></el-icon>
          <span>页面管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/settings">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { DataAnalysis, User, Folder, Document, Back, Setting } from '@element-plus/icons-vue'

const route = useRoute()
</script>

<style scoped>
.admin-layout { min-height: 100vh; }
.admin-aside { border-right: 1px solid #e4e7ed; }
.admin-logo { padding: 16px 20px; font-size: 20px; color: var(--el-color-primary); cursor: pointer; border-bottom: 1px solid #e4e7ed; }
</style>
