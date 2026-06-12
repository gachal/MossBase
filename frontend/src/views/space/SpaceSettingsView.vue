<template>
  <div class="space-settings" v-loading="spaceStore.loading">
    <template v-if="spaceStore.currentSpace">
      <el-page-header @back="router.push(`/spaces/${spaceId}`)" :title="spaceStore.currentSpace.name" content="空间设置" />

      <el-card style="margin-top: 16px">
        <template #header>基本信息</template>
        <el-form :model="form" label-width="80px">
          <el-form-item label="名称">
            <el-input v-model="form.name" />
          </el-form-item>
          <el-form-item label="描述">
            <el-input v-model="form.description" type="textarea" :rows="3" />
          </el-form-item>
          <el-form-item label="可见性">
            <el-radio-group v-model="form.visibility">
              <el-radio value="private">私有</el-radio>
              <el-radio value="public">公开</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="封面图片">
            <ImageUpload v-model="form.cover" upload-type="space-cover" :preview-size="200" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="saving" @click="handleSave">保存</el-button>
          </el-form-item>
        </el-form>
      </el-card>

      <el-card style="margin-top: 16px">
        <template #header>
          <div style="display:flex;justify-content:space-between;align-items:center">
            <span>成员管理</span>
          </div>
        </template>
        <el-table :data="spaceStore.members" stripe>
          <el-table-column prop="username" label="用户名" />
          <el-table-column prop="email" label="邮箱" />
          <el-table-column prop="role" label="角色" width="120">
            <template #default="{ row }">
              <el-tag :type="row.role === 'admin' ? 'danger' : row.role === 'member' ? 'primary' : 'info'" size="small">
                {{ row.role === 'admin' ? '管理员' : row.role === 'member' ? '成员' : '访客' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button size="small" type="danger" text @click="handleRemove(row.user_id)" :disabled="row.role === 'admin'">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <el-card style="margin-top: 16px">
        <template #header><span style="color: #f56c6c">危险操作</span></template>
        <el-button type="danger" @click="handleDeleteSpace">删除空间</el-button>
      </el-card>
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSpaceStore } from '@/stores/space'
import { updateSpace, deleteSpace, removeMember } from '@/api/space'
import ImageUpload from '@/components/common/ImageUpload.vue'

const route = useRoute()
const router = useRouter()
const spaceStore = useSpaceStore()
const spaceId = computed(() => Number(route.params.id))
const saving = ref(false)
const form = reactive({ name: '', description: '', visibility: 'private', cover: '' })

onMounted(async () => {
  await spaceStore.fetchSpace(spaceId.value)
  await spaceStore.fetchMembers(spaceId.value)
  if (spaceStore.currentSpace) {
    form.name = spaceStore.currentSpace.name
    form.description = spaceStore.currentSpace.description
    form.visibility = spaceStore.currentSpace.visibility
    form.cover = spaceStore.currentSpace.cover || ''
  }
})

async function handleSave() {
  saving.value = true
  try {
    await updateSpace(spaceId.value, form)
    ElMessage.success('保存成功')
    spaceStore.fetchSpace(spaceId.value)
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '保存失败')
  } finally {
    saving.value = false
  }
}

async function handleRemove(userId: number) {
  try {
    await ElMessageBox.confirm('确定移除该成员？', '确认')
    await removeMember(spaceId.value, userId)
    ElMessage.success('已移除')
    spaceStore.fetchMembers(spaceId.value)
  } catch { /* cancelled */ }
}

async function handleDeleteSpace() {
  try {
    await ElMessageBox.confirm('删除空间将同时删除所有页面，此操作不可恢复！', '危险操作', { type: 'warning' })
    await deleteSpace(spaceId.value)
    ElMessage.success('空间已删除')
    router.push('/spaces')
  } catch { /* cancelled */ }
}
</script>

<style scoped>
.space-settings { padding: 24px; max-width: 800px; margin: 0 auto; }
</style>
