<template>
  <div class="login-container">
    <div class="login-card">
      <div class="login-header">
        <h1>授权管理系统</h1>
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
            <div id="hcaptcha" class="captcha-widget"></div>
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
      
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'
import { getCaptchaConfig } from '@/api/auth'

const router = useRouter()
const authStore = useAuthStore()

const loginFormRef = ref()
const loading = ref(false)
let hcaptchaWidgetId = null
const captchaConfig = ref({ enabled: true, site_key: '' })

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

// 加载hCaptcha脚本
const loadHCaptchaScript = () => {
  return new Promise((resolve, reject) => {
    // 检查是否已经完全加载并且render函数可用
    if (window.hcaptcha && typeof window.hcaptcha.render === 'function') {
      resolve()
      return
    }

    // 检查是否已经有脚本标签
    const existingScript = document.querySelector('script[src*="hcaptcha"]')
    if (existingScript) {
      // 等待脚本加载完成并且API可用
      const checkHCaptcha = () => {
        if (window.hcaptcha && typeof window.hcaptcha.render === 'function') {
          resolve()
        } else {
          setTimeout(checkHCaptcha, 100)
        }
      }
      checkHCaptcha()
      return
    }

    const script = document.createElement('script')
    script.src = 'https://js.hcaptcha.com/1/api.js'
    script.async = true
    script.defer = true
    
    script.onload = () => {
      // 脚本加载完成后，等待API完全可用
      const waitForAPI = () => {
        if (window.hcaptcha && typeof window.hcaptcha.render === 'function') {
          resolve()
        } else {
          setTimeout(waitForAPI, 50)
        }
      }
      waitForAPI()
    }
    
    script.onerror = (error) => {
      console.error('hCaptcha脚本加载失败:', error)
      reject(error)
    }
    
    document.head.appendChild(script)
  })
}

// 初始化验证码
const initCaptcha = async () => {
  try {
    console.log('开始初始化hCaptcha...')
    console.log('验证码配置:', captchaConfig.value)
    
    await loadHCaptchaScript()
    console.log('hCaptcha脚本加载成功')
    
    await nextTick()
    
    // 清理现有的验证码组件
    if (hcaptchaWidgetId !== null && window.hcaptcha) {
      try {
        window.hcaptcha.remove(hcaptchaWidgetId)
      } catch (e) {
        console.warn('清理hCaptcha组件时出错:', e)
      }
      hcaptchaWidgetId = null
    }

    // 清空token
    loginForm.captchaToken = ''

    const captchaContainer = document.getElementById('hcaptcha')
    if (!captchaContainer) {
      console.error('未找到hCaptcha容器')
      return
    }

    // 清空容器
    captchaContainer.innerHTML = ''

    // 检查API是否可用
    if (!window.hcaptcha || typeof window.hcaptcha.render !== 'function') {
      throw new Error('hCaptcha API未正确加载')
    }

    const siteKey = captchaConfig.value.site_key || import.meta.env.VITE_HCAPTCHA_SITE_KEY || '10000000-ffff-ffff-ffff-000000000001'
    console.log('使用的site_key:', siteKey)

    // 渲染hCaptcha
    hcaptchaWidgetId = window.hcaptcha.render('hcaptcha', {
      sitekey: siteKey,
      callback: (token) => {
        loginForm.captchaToken = token
        console.log('hCaptcha验证成功, token:', token.substring(0, 20) + '...')
      },
      'expired-callback': () => {
        loginForm.captchaToken = ''
        console.log('hCaptcha已过期')
        ElMessage.warning('验证码已过期，请重新验证')
      },
      'error-callback': (error) => {
        loginForm.captchaToken = ''
        console.error('hCaptcha验证出错:', error)
        ElMessage.error('验证码加载失败，请刷新页面重试')
      }
    })
    
    console.log('hCaptcha组件渲染成功, widgetId:', hcaptchaWidgetId)
  } catch (error) {
    console.error('初始化hCaptcha失败:', error)
    // 降级到本地验证码
    initFallbackCaptcha()
  }
}

// 降级验证码（开发环境或hCaptcha加载失败时使用）
const initFallbackCaptcha = () => {
  nextTick(() => {
    const captchaContainer = document.getElementById('hcaptcha')
    if (captchaContainer) {
      captchaContainer.innerHTML = `
        <div class="fallback-captcha" style="
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
        " onclick="fallbackCaptchaClick()">
          <span style="color: #666;">点击完成人机验证 (开发模式)</span>
        </div>
      `
      
      // 降级验证码回调
      window.fallbackCaptchaClick = () => {
        loginForm.captchaToken = 'fallback_captcha_token_' + Date.now()
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
            <span style="color: #67c23a;">✓ 验证成功 (开发模式)</span>
          </div>
        `
      }
    }
  })
}

// 重置验证码
const resetCaptcha = () => {
  if (window.hcaptcha && hcaptchaWidgetId !== null) {
    try {
      window.hcaptcha.reset(hcaptchaWidgetId)
      loginForm.captchaToken = ''
    } catch (error) {
      console.warn('重置hCaptcha失败:', error)
      // 重新初始化
      initCaptcha()
    }
  } else {
    // 降级验证码重置
    initFallbackCaptcha()
  }
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
    } else {
      // 登录失败，重置验证码
      resetCaptcha()
    }
  } catch (error) {
    // 登录失败，重置验证码
    resetCaptcha()
  } finally {
    loading.value = false
  }
}

// 清理资源
onUnmounted(() => {
  if (window.hcaptcha && hcaptchaWidgetId !== null) {
    try {
      window.hcaptcha.remove(hcaptchaWidgetId)
    } catch (e) {
      console.warn('清理hCaptcha组件时出错:', e)
    }
  }
  
  // 清理全局回调函数
  if (window.fallbackCaptchaClick) {
    delete window.fallbackCaptchaClick
  }
})

// 获取验证码配置
const fetchCaptchaConfig = async () => {
  try {
    const response = await getCaptchaConfig()
    captchaConfig.value = response.data
  } catch (error) {
    console.warn('获取验证码配置失败，使用默认配置:', error)
  }
}

onMounted(async () => {
  await fetchCaptchaConfig()
  if (captchaConfig.value.enabled) {
    initCaptcha()
  } else {
    // 如果验证码被禁用，设置一个默认token
    loginForm.captchaToken = 'disabled'
  }
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

.captcha-widget {
  margin-top: 8px;
}

.login-button {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
}

:deep(.el-divider__text) {
  background-color: white;
  color: #909399;
}

/* hCaptcha样式调整 */
:deep(.h-captcha) {
  margin-top: 8px;
}

.fallback-captcha {
  transition: all 0.3s ease;
}

.fallback-captcha:hover {
  background: #e8e8e8 !important;
  border-color: #999 !important;
}
</style> 