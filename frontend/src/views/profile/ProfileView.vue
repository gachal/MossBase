<template>
  <div class="profile-page">
    <el-card>
      <template #header><span>个人资料</span></template>
      <el-form v-if="authStore.user" :model="form" label-width="80px">
        <el-form-item label="邮箱">
          <el-input :model-value="authStore.user.email" disabled />
        </el-form-item>
        <el-form-item label="用户名">
          <el-input v-model="form.username" />
        </el-form-item>
        <el-form-item label="角色">
          <el-tag>{{ authStore.user.role === 'admin' ? '管理员' : '用户' }}</el-tag>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { updateProfile } from '@/api/auth'

const authStore = useAuthStore()
const saving = ref(false)
const form = reactive({ username: '' })

onMounted(() => {
  if (authStore.user) form.username = authStore.user.username
})

async function handleSave() {
  saving.value = true
  try {
    const updated = await updateProfile({ username: form.username })
    authStore.user = updated
    ElMessage.success('保存成功')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.profile-page { max-width: 600px; margin: 0 auto; padding: 24px; }
</style>
