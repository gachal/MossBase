<template>
  <component :is="layout">
    <router-view />
  </component>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import MainLayout from '@/layouts/MainLayout.vue'
import AuthLayout from '@/layouts/AuthLayout.vue'
import AdminLayout from '@/layouts/AdminLayout.vue'
import InstallLayout from '@/layouts/InstallLayout.vue'

const route = useRoute()
const layout = computed(() => {
  const l = route.meta.layout
  if (l === 'auth') return AuthLayout
  if (l === 'admin') return AdminLayout
  if (l === 'install') return InstallLayout
  return MainLayout
})
</script>

<style>
:root {
  --el-color-primary: #487330;
  --el-color-primary-light-3: #6d9457;
  --el-color-primary-light-5: #96b388;
  --el-color-primary-light-7: #b8ceae;
  --el-color-primary-light-8: #c9d9c2;
  --el-color-primary-light-9: #dbe5d7;
  --el-color-primary-dark-2: #3a5c26;
}
</style>
