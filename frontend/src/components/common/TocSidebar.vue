<template>
  <div v-if="headings.length > 0" class="toc-sidebar">
    <div class="toc-title">目录</div>
    <nav class="toc-nav">
      <a
        v-for="item in headings"
        :key="item.id"
        :href="`#${item.id}`"
        :class="['toc-link', `toc-level-${item.level}`, { active: activeId === item.id }]"
        @click.prevent="scrollTo(item.id)"
      >
        {{ item.text }}
      </a>
    </nav>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch } from 'vue'
import type { HeadingItem } from '@/utils/markdown'

const props = defineProps<{
  headings: HeadingItem[]
}>()

const activeId = ref<string>('')

let observer: IntersectionObserver | null = null

function scrollTo(id: string) {
  activeId.value = id
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

function setupObserver() {
  if (observer) observer.disconnect()
  if (props.headings.length === 0) return

  observer = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          activeId.value = entry.target.id
        }
      }
    },
    { rootMargin: '-80px 0px -60% 0px', threshold: 0.1 }
  )

  for (const h of props.headings) {
    const el = document.getElementById(h.id)
    if (el) observer.observe(el)
  }
}

watch(() => props.headings, () => {
  setTimeout(setupObserver, 100)
})

onMounted(() => {
  setupObserver()
})

onUnmounted(() => {
  if (observer) observer.disconnect()
})
</script>

<style scoped>
.toc-sidebar {
  width: 220px;
  flex-shrink: 0;
  position: sticky;
  top: 24px;
  max-height: calc(100vh - 100px);
  overflow-y: auto;
  padding-left: 16px;
  border-left: 1px solid var(--el-border-color-lighter);
  align-self: flex-start;
}

.toc-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin-bottom: 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.toc-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.toc-link {
  font-size: 13px;
  color: var(--el-text-color-regular);
  text-decoration: none;
  padding: 4px 8px;
  border-radius: 4px;
  transition: all 0.15s ease;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  display: block;
}

.toc-link:hover {
  color: var(--el-color-primary);
  background: var(--el-fill-color-light);
}

.toc-link.active {
  color: var(--el-color-primary);
  font-weight: 500;
  background: var(--el-color-primary-light-9);
}

.toc-level-1 { padding-left: 8px; }
.toc-level-2 { padding-left: 20px; }
.toc-level-3 { padding-left: 32px; }
.toc-level-4 { padding-left: 44px; }
.toc-level-5 { padding-left: 56px; }
.toc-level-6 { padding-left: 68px; }

@media (max-width: 1024px) {
  .toc-sidebar {
    display: none;
  }
}
</style>
