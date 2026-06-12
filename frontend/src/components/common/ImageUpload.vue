<template>
  <el-upload
    :show-file-list="false"
    :before-upload="beforeUpload"
    :http-request="customUpload"
    :disabled="uploading"
    accept="image/jpeg,image/png,image/gif,image/webp"
  >
    <div v-if="modelValue" class="preview-wrapper" :style="{ width: previewSize + 'px', height: previewSize + 'px' }">
      <img :src="modelValue" class="preview-image" />
      <div class="overlay">{{ uploading ? '上传中...' : '更换图片' }}</div>
    </div>
    <el-button v-else size="small" :loading="uploading">上传图片</el-button>
  </el-upload>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { uploadFile } from '@/api/upload'

const props = withDefaults(defineProps<{
  modelValue: string
  uploadType: 'avatar' | 'space-cover'
  previewSize?: number
}>(), {
  previewSize: 120,
})

const emit = defineEmits<{
  'update:modelValue': [url: string]
}>()

const uploading = ref(false)

function beforeUpload(file: File) {
  const validTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']
  if (!validTypes.includes(file.type)) {
    ElMessage.error('仅支持 JPEG/PNG/GIF/WebP 格式')
    return false
  }
  if (file.size > 5 * 1024 * 1024) {
    ElMessage.error('图片大小不能超过 5MB')
    return false
  }
  return true
}

async function customUpload({ file }: { file: File }) {
  uploading.value = true
  try {
    const url = await uploadFile(file, props.uploadType)
    emit('update:modelValue', url)
    ElMessage.success('上传成功')
  } catch (e: any) {
    ElMessage.error(e.response?.data?.message || '上传失败')
  } finally {
    uploading.value = false
  }
}
</script>

<style scoped>
.preview-wrapper {
  position: relative;
  overflow: hidden;
  border-radius: 8px;
  border: 1px solid #dcdfe6;
  cursor: pointer;
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  font-size: 14px;
  opacity: 0;
  transition: opacity 0.2s;
}

.preview-wrapper:hover .overlay {
  opacity: 1;
}
</style>
