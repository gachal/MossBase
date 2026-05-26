<template>
  <div class="settings-view" v-loading="loading">
    <h2>系统设置</h2>

    <el-alert v-if="showRestartHint" type="warning" :closable="true" show-icon style="margin-bottom: 16px">
      配置已保存，部分设置需要重启服务才能生效。
    </el-alert>

    <!-- MCP Settings -->
    <el-card shadow="hover" style="margin-bottom: 24px">
      <template #header>
        <div class="card-header">
          <span>MCP 服务配置</span>
          <el-switch v-model="mcpForm.enabled" active-text="启用" inactive-text="关闭" />
        </div>
      </template>

      <el-form v-if="mcpForm.enabled" label-position="top">
        <el-form-item label="传输方式">
          <el-select v-model="mcpForm.transport" style="width: 100%">
            <el-option label="stdio" value="stdio" />
            <el-option label="http" value="http" />
            <el-option label="both (stdio + http)" value="both" />
          </el-select>
        </el-form-item>

        <el-form-item v-if="mcpForm.transport !== 'stdio'" label="HTTP 端口">
          <el-input-number v-model="mcpForm.http_port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>

        <el-form-item label="API Keys（每行一个）">
          <div style="display: flex; align-items: center; gap: 8px; width: 100%">
            <el-input
              v-model="mcpApiKeysText"
              type="textarea"
              :rows="3"
              :placeholder="mcpKeysPlaceholder"
              style="flex: 1"
            />
            <el-button v-if="mcpKeysMasked" type="danger" plain @click="handleClearMcpKeys">清空全部</el-button>
          </div>
        </el-form-item>

        <el-form-item label="默认用户 ID">
          <el-input-number v-model="mcpForm.default_user_id" :min="0" style="width: 100%" />
        </el-form-item>
      </el-form>
    </el-card>

    <!-- RAG Settings -->
    <el-card shadow="hover" style="margin-bottom: 24px">
      <template #header>
        <div class="card-header">
          <span>RAG 服务配置</span>
          <el-switch v-model="ragForm.enabled" active-text="启用" inactive-text="关闭" />
        </div>
      </template>

      <el-form v-if="ragForm.enabled" label-position="top">
        <el-form-item label="服务地址">
          <el-input v-model="ragForm.base_url" placeholder="http://127.0.0.1:8090" />
        </el-form-item>

        <el-form-item label="API Key">
          <el-input
            v-model="ragForm.api_key"
            :placeholder="ragApiKeyMasked ? '已设置（未修改将保持原值）' : 'RAG 服务 API Key'"
            show-password
          />
        </el-form-item>

        <el-form-item label="超时时间（秒）">
          <el-input-number v-model="ragForm.timeout" :min="1" :max="300" style="width: 100%" />
        </el-form-item>

        <el-form-item>
          <el-button :loading="testingRag" @click="handleTestRAG">测试连接</el-button>
          <el-tag v-if="ragTestResult" :type="ragTestResult.connected ? 'success' : 'danger'" style="margin-left: 12px">
            {{ ragTestResult.message }}
          </el-tag>
        </el-form-item>
      </el-form>
    </el-card>

    <div class="actions">
      <el-button type="primary" :loading="saving" @click="handleSave">保存设置</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  getSettings,
  updateSettings,
  testRAGConnection,
  UNCHANGED,
  type SettingsResponse,
  type TestRAGResponse,
} from '@/api/settings'

const loading = ref(false)
const saving = ref(false)
const testingRag = ref(false)
const showRestartHint = ref(false)
const ragTestResult = ref<TestRAGResponse | null>(null)

const mcpForm = reactive({
  enabled: false,
  transport: 'stdio',
  http_port: 8095,
  default_user_id: 0,
})

const ragForm = reactive({
  enabled: false,
  base_url: 'http://127.0.0.1:8090',
  api_key: '',
  timeout: 30,
})

const mcpApiKeysText = ref('')
const mcpKeysMasked = ref(false)
const ragApiKeyMasked = ref(false)
const mcpKeysClearPending = ref(false)

const mcpKeysPlaceholder = computed(() => {
  if (mcpKeysMasked.value) {
    return '留空保持不变，输入新内容将替换全部'
  }
  return '留空则不设置'
})

onMounted(async () => {
  loading.value = true
  try {
    const settings: SettingsResponse = await getSettings()
    Object.assign(mcpForm, {
      enabled: settings.mcp.enabled,
      transport: settings.mcp.transport,
      http_port: settings.mcp.http_port,
      default_user_id: settings.mcp.default_user_id,
    })
    mcpKeysMasked.value = settings.mcp.api_keys_masked

    Object.assign(ragForm, {
      enabled: settings.rag.enabled,
      base_url: settings.rag.base_url,
      api_key: settings.rag.api_key_masked ? '' : settings.rag.api_key,
      timeout: settings.rag.timeout,
    })
    ragApiKeyMasked.value = settings.rag.api_key_masked
  } catch {
    ElMessage.error('加载设置失败')
  } finally {
    loading.value = false
  }
})

function resolveMcpKeysAction(): { action: 'keep' | 'replace' | 'clear'; keys: string[] } {
  if (mcpKeysClearPending.value) {
    return { action: 'clear', keys: [] }
  }
  const text = mcpApiKeysText.value.trim()
  if (text === '') {
    return { action: 'keep', keys: [] }
  }
  const keys = text.split('\n').map(k => k.trim()).filter(Boolean)
  return { action: 'replace', keys }
}

async function handleClearMcpKeys() {
  await ElMessageBox.confirm('确定要清空所有 MCP API Keys 吗？此操作在保存后生效。', '清空确认', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning',
  })
  mcpKeysClearPending.value = true
  mcpApiKeysText.value = ''
  mcpKeysMasked.value = false
  ElMessage.info('将在保存时清空所有 API Keys')
}

async function handleSave() {
  saving.value = true
  showRestartHint.value = false
  try {
    const req: Parameters<typeof updateSettings>[0] = {}

    const mcpKeys = resolveMcpKeysAction()
    req.mcp = {
      enabled: mcpForm.enabled,
      transport: mcpForm.transport,
      http_port: mcpForm.http_port,
      default_user_id: mcpForm.default_user_id,
      api_keys: mcpKeys.keys,
      api_keys_action: mcpKeys.action,
    }

    const ragApiKey = ragApiKeyMasked.value && ragForm.api_key === '' ? UNCHANGED : ragForm.api_key
    req.rag = {
      enabled: ragForm.enabled,
      base_url: ragForm.base_url,
      api_key: ragApiKey,
      timeout: ragForm.timeout,
    }

    await updateSettings(req)
    ElMessage.success('设置已保存')
    showRestartHint.value = true
    mcpKeysClearPending.value = false

    const settings = await getSettings()
    mcpKeysMasked.value = settings.mcp.api_keys_masked
    ragForm.api_key = settings.rag.api_key_masked ? '' : settings.rag.api_key
    ragApiKeyMasked.value = settings.rag.api_key_masked
    mcpApiKeysText.value = ''
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '保存失败'
    ElMessage.error(msg)
  } finally {
    saving.value = false
  }
}

async function handleTestRAG() {
  testingRag.value = true
  ragTestResult.value = null
  try {
    const useSavedKey = ragApiKeyMasked.value && ragForm.api_key === ''
    const apiKey = useSavedKey ? '' : ragForm.api_key
    ragTestResult.value = await testRAGConnection(ragForm.base_url, apiKey, useSavedKey)
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '测试失败'
    ragTestResult.value = { connected: false, message: msg }
  } finally {
    testingRag.value = false
  }
}
</script>

<style scoped>
.settings-view { max-width: 800px; }
.settings-view h2 { margin: 0 0 24px; font-size: 20px; color: #303133; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.actions { display: flex; justify-content: flex-end; }
</style>
