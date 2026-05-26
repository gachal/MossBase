<template>
  <div class="install-layout">
    <div class="install-card">
      <h1 class="install-logo">MossBase</h1>
      <p class="install-subtitle">安装向导</p>
      <router-view />
    </div>
  </div>
</template>

<style scoped>
.install-layout { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #f5f7fa; padding: 20px; }
.install-card { width: 640px; max-width: 100%; padding: 40px; background: #fff; border-radius: 8px; box-shadow: 0 2px 12px rgba(0,0,0,0.1); text-align: center; }
.install-logo { font-size: 28px; color: var(--el-color-primary); margin: 0 0 8px; }
.install-subtitle { color: #909399; margin: 0 0 32px; }
</style>
