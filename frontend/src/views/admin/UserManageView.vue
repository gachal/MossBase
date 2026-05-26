<template>
  <div class="user-manage">
    <h2>用户管理</h2>
    <el-table :data="users" v-loading="loading" stripe style="margin-top: 16px;">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="username" label="用户名" width="150" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="role" label="角色" width="120">
        <template #default="{ row }">
          <el-tag :type="row.role === 'admin' ? 'danger' : 'info'" size="small">{{ row.role }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'warning'" size="small">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" text @click="handleRole(row)">{{ row.role === 'admin' ? '降为用户' : '升为管理员' }}</el-button>
          <el-button size="small" text :type="row.status === 'active' ? 'warning' : 'success'" @click="handleStatus(row)">{{ row.status === 'active' ? '禁用' : '启用' }}</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination v-if="total > pageSize" layout="prev, pager, next" :total="total" :page-size="pageSize" v-model:current-page="page" @current-change="fetchUsers" style="margin-top: 16px;" />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { AdminUser } from '@/api/admin'
import { listAdminUsers, updateUserRole, updateUserStatus } from '@/api/admin'

const loading = ref(false)
const users = ref<AdminUser[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = 20

onMounted(() => fetchUsers())

async function fetchUsers() {
  loading.value = true
  try {
    const result = await listAdminUsers(page.value, pageSize)
    users.value = result.items
    total.value = result.total
  } finally {
    loading.value = false
  }
}

async function handleRole(user: AdminUser) {
  const newRole = user.role === 'admin' ? 'user' : 'admin'
  try {
    await ElMessageBox.confirm(`确定将 ${user.username} 的角色改为 ${newRole}？`, '修改角色')
    await updateUserRole(user.id, newRole)
    ElMessage.success('角色已更新')
    fetchUsers()
  } catch { /* cancelled */ }
}

async function handleStatus(user: AdminUser) {
  const newStatus = user.status === 'active' ? 'disabled' : 'active'
  try {
    await ElMessageBox.confirm(`确定${newStatus === 'disabled' ? '禁用' : '启用'} ${user.username}？`, '修改状态')
    await updateUserStatus(user.id, newStatus)
    ElMessage.success('状态已更新')
    fetchUsers()
  } catch { /* cancelled */ }
}
</script>
