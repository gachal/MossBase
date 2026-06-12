<template>
  <div class="space-list-page">
    <div class="page-header">
      <h2>我的空间</h2>
      <el-button type="primary" @click="showCreateDialog = true">创建空间</el-button>
    </div>

    <el-row :gutter="16" v-loading="spaceStore.loading">
      <el-col :span="8" v-for="space in spaceStore.spaces" :key="space.id">
        <el-card shadow="hover" class="space-card" @click="router.push(`/spaces/${space.id}`)">
          <img v-if="space.cover" :src="space.cover" class="cover-image" />
          <template #header>
            <div class="card-header">
              <span>{{ space.name }}</span>
              <el-tag size="small">{{ space.visibility === 'public' ? '公开' : '私有' }}</el-tag>
            </div>
          </template>
          <p class="desc">{{ space.description || '暂无描述' }}</p>
          <div class="meta">创建于 {{ formatDate(space.created_at) }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-empty v-if="!spaceStore.loading && spaceStore.spaces.length === 0" description="还没有空间，点击上方按钮创建" />

    <el-dialog v-model="showCreateDialog" title="创建空间" width="480px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="空间名称" />
        </el-form-item>
        <el-form-item label="描述" prop="description">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="空间描述" />
        </el-form-item>
        <el-form-item label="可见性">
          <el-radio-group v-model="form.visibility">
            <el-radio value="private">私有</el-radio>
            <el-radio value="public">公开</el-radio>
          </el-radio-group>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" :loading="creating" @click="handleCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useSpaceStore } from '@/stores/space'
import { createSpace } from '@/api/space'

const router = useRouter()
const spaceStore = useSpaceStore()
const showCreateDialog = ref(false)
const creating = ref(false)
const formRef = ref<FormInstance>()
const form = reactive({ name: '', description: '', visibility: 'private' })

const rules: FormRules = {
  name: [{ required: true, message: '请输入空间名称', trigger: 'blur' }],
}

function formatDate(d: string) {
  return new Date(d).toLocaleDateString('zh-CN')
}

async function handleCreate() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return
  creating.value = true
  try {
    await createSpace(form)
    ElMessage.success('空间创建成功')
    showCreateDialog.value = false
    form.name = ''
    form.description = ''
    spaceStore.fetchSpaces()
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '创建失败')
  } finally {
    creating.value = false
  }
}

onMounted(() => spaceStore.fetchSpaces())
</script>

<style scoped>
.space-list-page { padding: 16px;padding-top:0px;margin-top:-18px; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 12px; }
.space-card { cursor: pointer; margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.desc { color: #606266; font-size: 14px; margin: 0; }
.meta { color: #909399; font-size: 12px; margin-top: 8px; }
.cover-image { width: 100%; height: 120px; object-fit: cover; border-radius: 4px; margin-bottom: 8px; }
</style>
