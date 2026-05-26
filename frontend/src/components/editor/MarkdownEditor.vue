<template>
  <div class="markdown-editor">
    <MdEditor
      v-model="internalContent"
      :editor-id="editorId"
      :mdHeadingId="generateHeadingId"
      language="zh-CN"
      :theme="theme"
      :toolbars-exclude="excludedToolbars"
      @onHtmlChanged="handleHtmlChanged"
      @onUploadImg="handleUploadImage"
      @onGetCatalog="handleGetCatalog"
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
  'update:catalog': [value: Array<{ text: string; level: number; id: string }>]
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

function generateHeadingId(text: string, _level: number, index: number): string {
  const base = text.toLowerCase().replace(/[^\w一-鿿]+/g, '-').replace(/^-|-$/g, '')
  return index === 0 ? base : `${base}-${index}`
}

function handleGetCatalog(list: Array<{ text: string; level: number }>) {
  const headings = list.map((item, index) => ({
    text: item.text,
    level: item.level,
    id: generateHeadingId(item.text, item.level, index + 1),
  }))
  emit('update:catalog', headings)
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
  --md-color: #487330;
  --md-color-hover: #3a5c26;
  --md-color-active: #2e4a1d;
}

:deep(.md-editor-content) {
  min-height: 500px;
}
</style>
