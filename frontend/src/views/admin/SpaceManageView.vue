<template>
  <div class="space-manage">
    <h2>空间管理</h2>
    <el-table :data="spaces" v-loading="loading" stripe style="margin-top: 16px;">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="名称" width="200" />
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="visibility" label="可见性" width="100">
        <template #default="{ row }">
          <el-tag :type="row.visibility === 'public' ? 'success' : 'info'" size="small">{{ row.visibility }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="owner_id" label="创建者ID" width="100" />
      <el-table-column label="操作" width="100">
        <template #default="{ row }">
          <el-button size="small" type="danger" text @click="handleDelete(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-if="total > pageSize" layout="prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="fetchSpaces" style="margin-top: 16px;" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminSpace } from '@/api/admin'
import { listAdminSpaces, deleteAdminSpace } from '@/api/admin'

const loading = ref(false)
const spaces = ref<AdminSpace[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

onMounted(() => fetchSpaces())

async function fetchSpaces() {
  loading.value = true
  try {
    const result = await listAdminSpaces(page.value, pageSize)
    spaces.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

async function handleDelete(space: AdminSpace) {
  try {
    await ElMessageBox.confirm(`确定删除空间「${space.name}」？所有页面将被删除。`, '危险操作', { type: 'warning' })
    await deleteAdminSpace(space.id)
    ElMessage.success('已删除')
    fetchSpaces()
  } catch { /* cancelled */ }
}
</script>
