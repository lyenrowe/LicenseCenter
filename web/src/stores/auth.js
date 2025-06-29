import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { clientLogin, adminLogin, logout } from '@/api/auth'

export const useAuthStore = defineStore('auth', () => {
  // 状态
  const token = ref(localStorage.getItem('token') || '')
  const userRole = ref(localStorage.getItem('userRole') || '')
  const userInfo = ref(JSON.parse(localStorage.getItem('userInfo') || '{}'))

  // 计算属性
  const isAuthenticated = computed(() => !!token.value)
  const isAdmin = computed(() => userRole.value === 'admin')
  const isClient = computed(() => userRole.value === 'client')

  // 客户端登录
  const clientLoginAction = async (authCode, captchaToken) => {
    try {
      const response = await clientLogin(authCode, captchaToken)
      const { session_token, customer_info } = response.data
      
      setAuth(session_token, 'client', customer_info)
      ElMessage.success('登录成功')
      return true
    } catch (error) {
      ElMessage.error(error.response?.data?.error || error.response?.data?.message || '登录失败')
      return false
    }
  }

  // 管理员登录
  const adminLoginAction = async (username, password, totpCode) => {
    try {
      const response = await adminLogin(username, password, totpCode)
      // 后端返回的数据结构：{ token, expires_in, admin: { id, username } }
      const { token: adminToken, admin } = response.data
      
      setAuth(adminToken, 'admin', admin)
      ElMessage.success('登录成功')
      return true
    } catch (error) {
      ElMessage.error(error.response?.data?.error || error.response?.data?.message || '登录失败')
      return false
    }
  }

  // 注销登录
  const logoutAction = async () => {
    try {
      await logout()
    } catch (error) {
      console.error('注销请求失败:', error)
    } finally {
      clearAuth()
      ElMessage.success('已退出登录')
    }
  }

  // 设置认证信息
  const setAuth = (newToken, role, info) => {
    token.value = newToken
    userRole.value = role
    userInfo.value = info

    localStorage.setItem('token', newToken)
    localStorage.setItem('userRole', role)
    localStorage.setItem('userInfo', JSON.stringify(info))
  }

  // 清除认证信息
  const clearAuth = () => {
    token.value = ''
    userRole.value = ''
    userInfo.value = {}

    localStorage.removeItem('token')
    localStorage.removeItem('userRole')
    localStorage.removeItem('userInfo')
  }

  // 检查token是否有效
  const checkAuth = () => {
    if (!token.value) {
      clearAuth()
      return false
    }
    return true
  }

  return {
    // 状态
    token,
    userRole,
    userInfo,
    
    // 计算属性
    isAuthenticated,
    isAdmin,
    isClient,
    
    // 方法
    clientLoginAction,
    adminLoginAction,
    logoutAction,
    setAuth,
    clearAuth,
    checkAuth
  }
}) 