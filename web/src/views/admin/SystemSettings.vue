<template>
  <div class="system-settings">
    <div class="page-header">
      <h1>系统设置</h1>
    </div>
    
    <!-- TOTP双因子认证设置 -->
    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <span>双因子认证 (TOTP)</span>
          <el-tag v-if="totpInfo.has_totp_secret" type="success">已启用</el-tag>
          <el-tag v-else type="warning">未启用</el-tag>
        </div>
      </template>
      
      <div class="totp-section">
        <div class="totp-description">
          <p>双因子认证为您的账户提供额外的安全保护。启用后，每次登录都需要输入手机认证应用生成的6位验证码。</p>
          <p>支持的认证应用：Google Authenticator、Microsoft Authenticator、Authy 等。</p>
        </div>

        <!-- 强制TOTP提示 -->
        <el-alert 
          v-if="totpInfo.force_totp" 
          title="系统已启用强制双因子认证" 
          type="info" 
          :closable="false"
          class="force-totp-alert"
        >
          <p>管理员已启用强制双因子认证，所有账户必须设置双因子认证才能登录。</p>
        </el-alert>

        <!-- 未启用状态 -->
        <div v-if="!totpInfo.has_totp_secret" class="totp-setup">
          <el-button 
            type="primary" 
            @click="enableTOTP"
            :loading="loading"
            size="large"
          >
            <el-icon><Lock /></el-icon>
            启用双因子认证
          </el-button>
        </div>

        <!-- 已启用状态 -->
        <div v-else class="totp-enabled">
          <div class="totp-status">
            <el-icon class="success-icon" size="24"><CircleCheck /></el-icon>
            <span>双因子认证已启用</span>
          </div>
          
          <div class="totp-actions">
            <el-button @click="showQRCode" :loading="loading">
              <el-icon><View /></el-icon>
              查看二维码
            </el-button>
            <el-button 
              v-if="totpInfo.can_disable" 
              type="danger" 
              @click="disableTOTP"
              :loading="loading"
            >
              <el-icon><Close /></el-icon>
              禁用双因子认证
            </el-button>
            <el-tag v-else type="info">强制模式下无法禁用</el-tag>
          </div>
        </div>
      </div>
    </el-card>

    <!-- 其他系统设置可以在这里添加 -->
    <el-card class="setting-card">
      <template #header>
        <div class="card-header">
          <span>其他设置</span>
        </div>
      </template>
      
      <div class="other-settings">
        <p class="placeholder">更多系统设置功能开发中...</p>
      </div>
    </el-card>

    <!-- TOTP设置对话框 -->
    <el-dialog
      v-model="setupDialogVisible"
      title="设置双因子认证"
      width="500px"
      :close-on-click-modal="false"
    >
      <div class="totp-setup-dialog">
        <div class="setup-steps">
          <el-steps :active="currentStep" align-center>
            <el-step title="生成密钥" />
            <el-step title="扫描二维码" />
            <el-step title="验证设置" />
          </el-steps>
        </div>

        <!-- 步骤1: 生成密钥 -->
        <div v-if="currentStep === 0" class="step-content">
          <div class="step-description">
            <p>我们将为您生成一个专用的认证密钥。</p>
          </div>
          <div class="step-actions">
            <el-button type="primary" @click="generateTOTPKey" :loading="loading">
              生成认证密钥
            </el-button>
          </div>
        </div>

        <!-- 步骤2: 扫描二维码 -->
        <div v-if="currentStep === 1" class="step-content">
          <div class="step-description">
            <p>请使用您的认证应用扫描下方二维码：</p>
          </div>
          <div class="qr-code-container" v-if="qrCodeUrl">
            <div id="qr-code" ref="qrCodeRef"></div>
            <p class="qr-code-tip">
              如果无法扫描，请手动输入密钥：
              <el-input 
                v-model="totpSecret" 
                readonly 
                class="secret-input"
                size="small"
              >
                <template #append>
                  <el-button @click="copySecret" size="small">复制</el-button>
                </template>
              </el-input>
            </p>
          </div>
          <div class="step-actions">
            <el-button @click="currentStep = 0">上一步</el-button>
            <el-button type="primary" @click="currentStep = 2">下一步</el-button>
          </div>
        </div>

        <!-- 步骤3: 验证设置 -->
        <div v-if="currentStep === 2" class="step-content">
          <div class="step-description">
            <p>请输入认证应用显示的6位验证码以完成设置：</p>
          </div>
          <div class="verification-form">
            <el-form :model="verificationForm" :rules="verificationRules" ref="verificationFormRef">
              <el-form-item prop="code">
                <el-input
                  v-model="verificationForm.code"
                  placeholder="请输入6位验证码"
                  maxlength="6"
                  size="large"
                  class="verification-input"
                  @keyup.enter="completeTOTPSetup"
                />
              </el-form-item>
            </el-form>
          </div>
          <div class="step-actions">
            <el-button @click="currentStep = 1">上一步</el-button>
            <el-button type="primary" @click="completeTOTPSetup" :loading="loading">
              完成设置
            </el-button>
          </div>
        </div>
      </div>
    </el-dialog>

    <!-- 二维码查看对话框 -->
    <el-dialog
      v-model="qrDialogVisible"
      title="双因子认证二维码"
      width="400px"
    >
      <div class="qr-view-dialog">
        <div class="qr-code-container" v-if="qrCodeUrl">
          <div id="qr-code-view" ref="qrCodeViewRef"></div>
          <p class="qr-code-tip">使用认证应用扫描此二维码</p>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Lock, CircleCheck, Close, View } from '@element-plus/icons-vue'
import QRCode from 'qrcode'
import { useAuthStore } from '@/stores/auth'
import { 
  getTOTPInfo, 
  enableTOTP as enableTOTPAPI, 
  disableTOTP as disableTOTPAPI, 
  verifyTOTPSetup as verifyTOTPSetupAPI 
} from '@/api/admin'

const authStore = useAuthStore()

// 响应式数据
const loading = ref(false)
const setupDialogVisible = ref(false)
const qrDialogVisible = ref(false)
const currentStep = ref(0)
const qrCodeUrl = ref('')
const totpSecret = ref('')
const qrCodeRef = ref()
const qrCodeViewRef = ref()

// TOTP信息
const totpInfo = reactive({
  has_totp_secret: false,
  force_totp: false,
  can_disable: true,
  qr_code_url: ''
})

// 验证表单
const verificationForm = reactive({
  code: ''
})

const verificationFormRef = ref()

const verificationRules = {
  code: [
    { required: true, message: '请输入验证码', trigger: 'blur' },
    { len: 6, message: '验证码必须为6位', trigger: 'blur' },
    { pattern: /^\d{6}$/, message: '验证码只能包含数字', trigger: 'blur' }
  ]
}

// 获取TOTP信息
const fetchTOTPInfo = async () => {
  try {
    const response = await getTOTPInfo(authStore.userInfo.id)
    Object.assign(totpInfo, response.data.data)
  } catch (error) {
    console.error('获取TOTP信息失败:', error)
  }
}

// 启用TOTP
const enableTOTP = () => {
  setupDialogVisible.value = true
  currentStep.value = 0
}

// 生成TOTP密钥
const generateTOTPKey = async () => {
  loading.value = true
  try {
    const response = await enableTOTPAPI(authStore.userInfo.id)
    qrCodeUrl.value = response.data.qr_code_url
    
    // 从URL中提取密钥
    const url = new URL(qrCodeUrl.value)
    totpSecret.value = url.searchParams.get('secret')
    
    currentStep.value = 1
    
    // 下一帧渲染二维码
    await nextTick()
    generateQRCode()
    
    ElMessage.success('认证密钥生成成功')
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '生成密钥失败')
  } finally {
    loading.value = false
  }
}

// 生成二维码
const generateQRCode = async () => {
  if (qrCodeRef.value && qrCodeUrl.value) {
    try {
      qrCodeRef.value.innerHTML = ''
      const canvas = await QRCode.toCanvas(qrCodeUrl.value, {
        width: 200,
        margin: 2
      })
      qrCodeRef.value.appendChild(canvas)
    } catch (error) {
      console.error('生成二维码失败:', error)
    }
  }
}

// 生成查看二维码
const generateViewQRCode = async () => {
  if (qrCodeViewRef.value && qrCodeUrl.value) {
    try {
      qrCodeViewRef.value.innerHTML = ''
      const canvas = await QRCode.toCanvas(qrCodeUrl.value, {
        width: 200,
        margin: 2
      })
      qrCodeViewRef.value.appendChild(canvas)
    } catch (error) {
      console.error('生成二维码失败:', error)
    }
  }
}

// 复制密钥
const copySecret = async () => {
  try {
    await navigator.clipboard.writeText(totpSecret.value)
    ElMessage.success('密钥已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败，请手动复制')
  }
}

// 完成TOTP设置
const completeTOTPSetup = async () => {
  if (!verificationFormRef.value) return
  
  const valid = await verificationFormRef.value.validate()
  if (!valid) return
  
  loading.value = true
  try {
    // 调用验证TOTP码的接口
    await verifyTOTPSetupAPI(authStore.userInfo.id, verificationForm.code)
    
    ElMessage.success('双因子认证设置成功！')
    setupDialogVisible.value = false
    currentStep.value = 0
    verificationForm.code = ''
    
    // 刷新TOTP信息
    await fetchTOTPInfo()
  } catch (error) {
    ElMessage.error(error.response?.data?.error || '验证失败')
  } finally {
    loading.value = false
  }
}

// 查看二维码
const showQRCode = async () => {
  qrCodeUrl.value = totpInfo.qr_code_url
  qrDialogVisible.value = true
  
  await nextTick()
  generateViewQRCode()
}

// 禁用TOTP
const disableTOTP = async () => {
  try {
    await ElMessageBox.confirm(
      '禁用双因子认证将降低账户安全性，确定要继续吗？',
      '确认禁用',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    loading.value = true
    await disableTOTPAPI(authStore.userInfo.id)
    ElMessage.success('双因子认证已禁用')
    
    // 刷新TOTP信息
    await fetchTOTPInfo()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(error.response?.data?.error || '禁用失败')
    }
  } finally {
    loading.value = false
  }
}

// 组件挂载时获取TOTP信息
onMounted(() => {
  fetchTOTPInfo()
})
</script>

<style scoped>
.system-settings {
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  color: #2c3e50;
}

.setting-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.totp-section {
  padding: 20px 0;
}

.totp-description {
  margin-bottom: 20px;
  color: #606266;
  line-height: 1.6;
}

.force-totp-alert {
  margin-bottom: 20px;
}

.totp-setup {
  text-align: center;
  padding: 40px 0;
}

.totp-enabled {
  padding: 20px 0;
}

.totp-status {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  font-size: 16px;
  color: #67c23a;
}

.success-icon {
  margin-right: 8px;
}

.totp-actions {
  display: flex;
  gap: 12px;
  align-items: center;
}

.other-settings {
  padding: 40px 0;
  text-align: center;
  color: #909399;
}

.placeholder {
  font-size: 16px;
  margin: 0;
}

/* 对话框样式 */
.totp-setup-dialog {
  padding: 20px 0;
}

.setup-steps {
  margin-bottom: 30px;
}

.step-content {
  min-height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}

.step-description {
  text-align: center;
  margin-bottom: 20px;
  color: #606266;
}

.qr-code-container {
  text-align: center;
  margin: 20px 0;
}

.qr-code-tip {
  margin-top: 15px;
  color: #909399;
  font-size: 14px;
}

.secret-input {
  margin-top: 10px;
  max-width: 300px;
}

.verification-form {
  text-align: center;
  margin: 20px 0;
}

.verification-input {
  max-width: 200px;
  text-align: center;
}

.step-actions {
  text-align: center;
  margin-top: 20px;
}

.qr-view-dialog {
  text-align: center;
}

/* 响应式 */
@media (max-width: 768px) {
  .system-settings {
    padding: 10px;
  }
  
  .totp-actions {
    flex-direction: column;
    align-items: stretch;
  }
  
  .totp-actions .el-button {
    margin-bottom: 8px;
  }
}
</style> 