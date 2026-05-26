<template>
  <div class="auth-layout">
    <div class="auth-card">
      <h1 class="auth-logo">MossBase</h1>
      <p class="auth-subtitle">苔藓知识库，小而有机</p>
      <router-view />
    </div>
  </div>
</template>

<style scoped>
.auth-layout { min-height: 100vh; display: flex; align-items: center; justify-content: center; background: #f5f7fa; }
.auth-card { width: 400px; padding: 40px; background: #fff; border-radius: 8px; box-shadow: 0 2px 12px rgba(0,0,0,0.1); text-align: center; }
.auth-logo { font-size: 28px; color: var(--el-color-primary); margin: 0 0 8px; }
.auth-subtitle { color: #909399; margin: 0 0 32px; }
</style>
