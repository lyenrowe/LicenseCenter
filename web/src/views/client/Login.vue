<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>OCR2Doc 授权管理系统</h1>
        <p>客户端登录</p>
      </div>
      
      <el-form 
        ref="loginFormRef" 
        :model="loginForm" 
        :rules="loginRules" 
        class="login-form"
        @submit.prevent="handleLogin"
      >
        <el-form-item prop="authorizationCode">
          <el-input
            v-model="loginForm.authorizationCode"
            placeholder="请输入授权码"
            size="large"
            prefix-icon="Key"
            :disabled="loading"
          />
        </el-form-item>
        
        <el-form-item prop="captchaToken">
          <div class="captcha-container">
            <el-input
              v-model="loginForm.captchaToken"
              placeholder="请完成人机验证"
              size="large"
              prefix-icon="Shield"
              :disabled="loading"
              readonly
            />
            <div id="captcha" class="captcha-widget"></div>
          </div>
        </el-form-item>
        
        <el-form-item>
          <el-button 
            type="primary" 
            size="large" 
            :loading="loading" 
            :disabled="!loginForm.captchaToken"
            @click="handleLogin"
            class="login-button"
          >
            登录
          </el-button>
        </el-form-item>
      </el-form>
      
      <div class="login-footer">
        <el-divider>或</el-divider>
        <el-button 
          text 
          type="primary" 
          @click="goToAdminLogin"
        >
          管理员登录
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

const loginFormRef = ref()
const loading = ref(false)

const loginForm = reactive({
  authorizationCode: '',
  captchaToken: ''
})

const loginRules = {
  authorizationCode: [
    { required: true, message: '请输入授权码', trigger: 'blur' },
    { min: 3, message: '授权码长度不能少于3位', trigger: 'blur' }
  ]
}

// 初始化验证码
const initCaptcha = () => {
  // 这里模拟hCaptcha初始化
  // 在实际项目中，需要引入hCaptcha或reCAPTCHA的SDK
  nextTick(() => {
    const captchaContainer = document.getElementById('captcha')
    if (captchaContainer) {
      captchaContainer.innerHTML = `
        <div class="mock-captcha" style="
          width: 100%;
          height: 78px;
          border: 1px solid #ddd;
          display: flex;
          align-items: center;
          justify-content: center;
          background: #f5f5f5;
          margin-top: 8px;
          cursor: pointer;
          border-radius: 4px;
        " onclick="mockCaptchaClick()">
          <span style="color: #666;">点击完成人机验证 (模拟)</span>
        </div>
      `
      
      // 模拟验证码回调
      window.mockCaptchaClick = () => {
        loginForm.captchaToken = 'mock_captcha_token_' + Date.now()
        captchaContainer.innerHTML = `
          <div style="
            width: 100%;
            height: 78px;
            border: 1px solid #67c23a;
            display: flex;
            align-items: center;
            justify-content: center;
            background: #f0f9ff;
            margin-top: 8px;
            border-radius: 4px;
          ">
            <el-icon style="color: #67c23a; margin-right: 8px;"><SuccessFilled /></el-icon>
            <span style="color: #67c23a;">验证成功</span>
          </div>
        `
      }
    }
  })
}

const handleLogin = async () => {
  if (!loginFormRef.value) return
  
  const valid = await loginFormRef.value.validate()
  if (!valid) return
  
  if (!loginForm.captchaToken) {
    ElMessage.warning('请完成人机验证')
    return
  }
  
  loading.value = true
  
  try {
    const success = await authStore.clientLoginAction(
      loginForm.authorizationCode,
      loginForm.captchaToken
    )
    
    if (success) {
      router.push('/client/dashboard')
    }
  } finally {
    loading.value = false
  }
}

const goToAdminLogin = () => {
  router.push('/admin/login')
}

onMounted(() => {
  initCaptcha()
})
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  padding: 32px;
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-header h1 {
  font-size: 24px;
  font-weight: 600;
  color: #2c3e50;
  margin-bottom: 8px;
}

.login-header p {
  font-size: 14px;
  color: #7f8c8d;
  margin: 0;
}

.login-form {
  margin-bottom: 20px;
}

.captcha-container {
  width: 100%;
}

.login-button {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
}

.login-footer {
  text-align: center;
}

:deep(.el-divider__text) {
  background-color: white;
  color: #909399;
}
</style> 