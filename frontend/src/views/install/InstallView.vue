<template>
  <div class="install-wizard">
    <el-steps :active="currentStep" finish-status="success" align-center class="steps">
      <el-step title="数据库配置" />
      <el-step title="管理员账户" />
      <el-step title="可选服务" />
      <el-step title="确认安装" />
    </el-steps>

    <div class="step-content">
      <!-- Step 1: Database -->
      <el-form v-show="currentStep === 0" ref="dbFormRef" :model="dbForm" :rules="dbRules" label-position="top">
        <el-form-item label="主机地址" prop="host">
          <el-input v-model="dbForm.host" placeholder="127.0.0.1" />
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input-number v-model="dbForm.port" :min="1" :max="65535" style="width: 100%" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="dbForm.username" placeholder="root" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="dbForm.password" type="password" placeholder="数据库密码" show-password />
        </el-form-item>
        <el-form-item label="数据库名" prop="dbname">
          <el-input v-model="dbForm.dbname" placeholder="mossbase" />
        </el-form-item>
        <div class="step-actions">
          <el-button :loading="testing" @click="handleTestDb">测试连接</el-button>
          <el-button type="primary" :disabled="!dbTested" @click="nextStep">下一步</el-button>
        </div>
        <div v-if="dbTestResult" class="test-result">
          <el-tag :type="dbTestResult.connected ? 'success' : 'danger'">
            {{ dbTestResult.connected ? `连接成功 — MySQL ${dbTestResult.version}` : `连接失败 — ${dbTestResult.error}` }}
          </el-tag>
        </div>
      </el-form>

      <!-- Step 2: Admin -->
      <el-form v-show="currentStep === 1" ref="adminFormRef" :model="adminForm" :rules="adminRules" label-position="top">
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="adminForm.email" placeholder="admin@example.com" />
        </el-form-item>
        <el-form-item label="用户名" prop="username">
          <el-input v-model="adminForm.username" placeholder="admin" />
        </el-form-item>
        <el-form-item label="密码" prop="password">
          <el-input v-model="adminForm.password" type="password" placeholder="至少 8 位" show-password />
        </el-form-item>
        <el-form-item label="确认密码" prop="confirmPassword">
          <el-input v-model="adminForm.confirmPassword" type="password" placeholder="再次输入密码" show-password />
        </el-form-item>
        <div class="step-actions">
          <el-button @click="prevStep">上一步</el-button>
          <el-button type="primary" @click="nextStep">下一步</el-button>
        </div>
      </el-form>

      <!-- Step 3: Optional Services -->
      <div v-show="currentStep === 2">
        <div class="service-section">
          <h3>MCP 服务</h3>
          <el-switch v-model="mcpForm.enabled" active-text="启用" inactive-text="关闭" />
          <el-form v-if="mcpForm.enabled" label-position="top" style="margin-top: 16px">
            <el-form-item label="传输方式">
              <el-select v-model="mcpForm.transport" style="width: 100%">
                <el-option label="stdio" value="stdio" />
                <el-option label="http" value="http" />
                <el-option label="both" value="both" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="mcpForm.transport !== 'stdio'" label="HTTP 端口">
              <el-input-number v-model="mcpForm.http_port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
            <el-form-item label="API Keys（每行一个）">
              <el-input v-model="mcpApiKeysText" type="textarea" :rows="3" placeholder="留空则不设置" />
            </el-form-item>
          </el-form>
        </div>

        <div class="service-section">
          <h3>RAG 服务</h3>
          <el-switch v-model="ragForm.enabled" active-text="启用" inactive-text="关闭" />
          <el-form v-if="ragForm.enabled" label-position="top" style="margin-top: 16px">
            <el-form-item label="服务地址">
              <el-input v-model="ragForm.base_url" placeholder="http://127.0.0.1:8090" />
            </el-form-item>
            <el-form-item label="API Key">
              <el-input v-model="ragForm.api_key" placeholder="RAG 服务 API Key" />
            </el-form-item>
          </el-form>
        </div>

        <div class="step-actions">
          <el-button @click="prevStep">上一步</el-button>
          <el-button type="primary" @click="nextStep">下一步</el-button>
        </div>
      </div>

      <!-- Step 4: Confirm & Install -->
      <div v-show="currentStep === 3">
        <div class="summary">
          <h3>配置摘要</h3>
          <el-descriptions :column="1" border>
            <el-descriptions-item label="数据库主机">{{ dbForm.host }}:{{ dbForm.port }}</el-descriptions-item>
            <el-descriptions-item label="数据库用户">{{ dbForm.username }}</el-descriptions-item>
            <el-descriptions-item label="数据库名">{{ dbForm.dbname }}</el-descriptions-item>
            <el-descriptions-item label="管理员邮箱">{{ adminForm.email }}</el-descriptions-item>
            <el-descriptions-item label="管理员用户名">{{ adminForm.username }}</el-descriptions-item>
            <el-descriptions-item label="MCP">{{ mcpForm.enabled ? `启用 (${mcpForm.transport})` : '未启用' }}</el-descriptions-item>
            <el-descriptions-item label="RAG">{{ ragForm.enabled ? `启用 (${ragForm.base_url})` : '未启用' }}</el-descriptions-item>
          </el-descriptions>
        </div>

        <div v-if="installing" class="install-progress">
          <el-progress :percentage="installProgress" :status="installStatus" />
          <p class="progress-text">{{ installMessage }}</p>
        </div>

        <div v-if="installDone" class="install-done">
          <el-result icon="success" title="安装完成" sub-title="服务器正在重启，请稍候..." />
          <p class="restart-hint">正在等待服务器重启...</p>
        </div>

        <div v-if="installError" class="install-error">
          <el-result icon="error" title="安装失败" :sub-title="installError" />
          <el-button @click="installError = ''">重试</el-button>
        </div>

        <div v-if="!installing && !installDone && !installError" class="step-actions">
          <el-button @click="prevStep">上一步</el-button>
          <el-button type="primary" @click="handleInstall">开始安装</el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useInstallStore } from '@/stores/install'
import type { InstallRequest } from '@/api/install'

const router = useRouter()
const installStore = useInstallStore()

const currentStep = ref(0)

// Step 1: Database
const dbFormRef = ref<FormInstance>()
const dbForm = reactive({ host: '127.0.0.1', port: 3306, username: 'root', password: '', dbname: 'mossbase' })
const dbRules: FormRules = {
  host: [{ required: true, message: '请输入主机地址', trigger: 'blur' }],
  port: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  dbname: [{ required: true, message: '请输入数据库名', trigger: 'blur' }],
}
const testing = ref(false)
const dbTested = ref(false)
const dbTestResult = ref<{ connected: boolean; version?: string; error?: string } | null>(null)

async function handleTestDb() {
  const valid = await dbFormRef.value?.validate().catch(() => false)
  if (!valid) return
  testing.value = true
  try {
    dbTestResult.value = await installStore.testDb({ ...dbForm })
    if (dbTestResult.value?.connected) {
      dbTested.value = true
    }
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '测试失败'
    dbTestResult.value = { connected: false, error: msg }
  } finally {
    testing.value = false
  }
}

// Step 2: Admin
const adminFormRef = ref<FormInstance>()
const adminForm = reactive({ email: '', username: '', password: '', confirmPassword: '' })

const validateConfirm = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== adminForm.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const adminRules: FormRules = {
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' },
  ],
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 2, max: 50, message: '用户名长度 2-50 位', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, max: 128, message: '密码至少 8 位', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
}

// Step 3: Optional Services
const mcpForm = reactive({ enabled: false, transport: 'stdio', http_port: 8095, api_keys: [] as string[] })
const ragForm = reactive({ enabled: false, base_url: 'http://127.0.0.1:8090', api_key: '' })
const mcpApiKeysText = ref('')

// Step 4: Install
const installing = ref(false)
const installDone = ref(false)
const installError = ref('')
const installProgress = ref(0)
const installStatus = ref<'' | 'success' | 'exception' | 'warning'>('')
const installMessage = ref('')

function prevStep() {
  if (currentStep.value > 0) currentStep.value--
}

async function nextStep() {
  if (currentStep.value === 0) {
    const valid = await dbFormRef.value?.validate().catch(() => false)
    if (!valid) return
    if (!dbTested.value) {
      ElMessage.warning('请先测试数据库连接')
      return
    }
  } else if (currentStep.value === 1) {
    const valid = await adminFormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  currentStep.value++
}

async function handleInstall() {
  installing.value = true
  installError.value = ''
  installProgress.value = 10
  installMessage.value = '正在安装...'

  const req: InstallRequest = {
    database: { ...dbForm },
    admin: { email: adminForm.email, username: adminForm.username, password: adminForm.password },
  }

  if (mcpForm.enabled) {
    const keys = mcpApiKeysText.value.split('\n').map(k => k.trim()).filter(Boolean)
    req.mcp = { ...mcpForm, api_keys: keys }
  }
  if (ragForm.enabled) {
    req.rag = { ...ragForm }
  }

  try {
    installProgress.value = 50
    installMessage.value = '正在配置数据库和创建管理员账户...'
    await installStore.doInstall(req)
    installProgress.value = 80
    installMessage.value = '安装完成，等待服务器重启...'
    installStatus.value = 'success'
    installDone.value = true
    installing.value = false

    installProgress.value = 90
    await waitForRestart()
    installProgress.value = 100
    installMessage.value = '服务器已就绪'
    await new Promise(r => setTimeout(r, 1500))
    router.push('/login')
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '安装失败'
    installError.value = msg
    installStatus.value = 'exception'
    installing.value = false
  }
}

async function waitForRestart() {
  const maxAttempts = 30
  for (let i = 0; i < maxAttempts; i++) {
    await new Promise(r => setTimeout(r, 2000))
    try {
      await installStore.fetchStatus()
      if (installStore.isInstalled) return
    } catch {
      // Server not back yet, continue waiting
    }
  }
  ElMessage.warning('等待服务器重启超时，请手动刷新页面')
}
</script>

<style scoped>
.install-wizard { text-align: left; }
.steps { margin-bottom: 32px; }
.step-content { min-height: 300px; }
.step-actions { display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px; }
.test-result { text-align: center; margin-top: 12px; }
.service-section { margin-bottom: 24px; padding: 16px; border: 1px solid #ebeef5; border-radius: 4px; }
.service-section h3 { margin: 0 0 12px; font-size: 16px; color: #303133; }
.summary { margin-bottom: 24px; }
.summary h3 { margin: 0 0 16px; font-size: 16px; color: #303133; }
.install-progress { text-align: center; margin: 24px 0; }
.progress-text { color: #909399; margin-top: 12px; }
.install-done, .install-error { text-align: center; }
.restart-hint { color: #909399; animation: pulse 1.5s infinite; }
@keyframes pulse { 0%, 100% { opacity: 1; } 50% { opacity: 0.5; } }
</style>
