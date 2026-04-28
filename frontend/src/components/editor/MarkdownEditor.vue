<template>
  <div class="markdown-editor">
    <MdEditor
      v-model="internalContent"
      :editor-id="editorId"
      language="zh-CN"
      :theme="theme"
      :toolbars-exclude="excludedToolbars"
      @onHtmlChanged="handleHtmlChanged"
      @onUploadImg="handleUploadImage"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { MdEditor, type ToolbarNames } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { renderMarkdown } from '@/utils/markdown'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
  'update:html': [value: string]
}>()

const editorId = 'moss-md-editor'
const internalContent = ref(props.modelValue)

const theme = computed(() => {
  return document.documentElement.classList.contains('dark') ? 'dark' : 'light'
})

const excludedToolbars: ToolbarNames[] = ['github', 'mermaid', 'katex', 'htmlPreview']

watch(() => props.modelValue, (val) => {
  if (val !== internalContent.value) {
    internalContent.value = val
  }
})

watch(internalContent, (val) => {
  emit('update:modelValue', val)
})

function handleHtmlChanged() {
  emit('update:html', renderMarkdown(internalContent.value))
}

function handleUploadImage(_files: File[], _callback: (urls: string[]) => void) {
  // TODO: implement image upload to backend
}
</script>

<style scoped>
.markdown-editor {
  border: 1px solid var(--el-border-color);
  border-radius: 4px;
  overflow: hidden;
}

:deep(.md-editor) {
  border: none;
}

:deep(.md-editor-content) {
  min-height: 500px;
}
</style>
