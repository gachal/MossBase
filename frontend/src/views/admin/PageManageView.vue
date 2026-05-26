<template>
  <div class="page-manage">
    <h2>页面管理</h2>
    <el-table :data="pages" v-loading="loading" stripe style="margin-top: 16px;">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="title" label="标题" width="250" />
      <el-table-column prop="space_id" label="空间ID" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'published' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="version" label="版本" width="80" />
      <el-table-column prop="updated_at" label="更新时间" width="180">
        <template #default="{ row }">{{ new Date(row.updated_at).toLocaleString('zh-CN') }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" type="danger" text @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-if="total > pageSize" layout="prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="fetchPages" style="margin-top: 16px;" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminPage } from '@/api/admin'
import { listAdminPages, deleteAdminPage } from '@/api/admin'

const loading = ref(false)
const pages = ref<AdminPage[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

onMounted(() => fetchPages())

async function fetchPages() {
  loading.value = true
  try {
    const result = await listAdminPages(page.value, pageSize)
    pages.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

async function handleDelete(p: AdminPage) {
  try {
    await ElMessageBox.confirm(`确定删除页面「${p.title}」？`, '删除页面', { type: 'warning' })
    await deleteAdminPage(p.id)
    ElMessage.success('已删除')
    fetchPages()
  } catch { /* cancelled */ }
}
</script>
