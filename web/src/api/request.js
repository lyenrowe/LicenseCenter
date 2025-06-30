import axios from 'axios'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'

// 创建axios实例
const request = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    const authStore = useAuthStore()
    if (authStore.token) {
      config.headers.Authorization = `Bearer ${authStore.token}`
    }
    return config
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    console.error('响应错误:', error)
    
    if (error.response) {
      const { status, data } = error.response
      
      // 如果响应是blob类型且为错误响应，尝试解析为JSON
      if (data instanceof Blob && status >= 400) {
        return new Promise((resolve, reject) => {
          const reader = new FileReader()
          reader.onload = () => {
            try {
              const errorData = JSON.parse(reader.result)
              error.response.data = errorData
              reject(error)
            } catch (e) {
              // 如果不是JSON，保持原始blob
              reject(error)
            }
          }
          reader.readAsText(data)
        })
      }
      
      // 检查是否有具体的业务错误信息
      const hasBusinessError = data && (data.error || data.message)
      
      switch (status) {
        case 401:
          // 未授权，清除token并跳转登录
          const authStore = useAuthStore()
          authStore.clearAuth()
          ElMessage.error('登录已过期，请重新登录')
          // 根据当前路径跳转到对应登录页
          const currentPath = window.location.pathname
          if (currentPath.startsWith('/admin')) {
            window.location.href = '/admin/login'
          } else {
            window.location.href = '/client/login'
          }
          break
        case 403:
          // 对于403错误，不显示通用消息，让调用方处理
          break
        case 404:
          // 只有在没有具体业务错误信息时才显示通用404消息
          if (!hasBusinessError) {
            ElMessage.error('请求的资源不存在')
          }
          break
        case 500:
          // 只有在没有具体业务错误信息时才显示通用500消息
          if (!hasBusinessError) {
            ElMessage.error('服务器内部错误')
          }
          break
        default:
          // 对于其他错误（如400业务错误），不在这里显示消息
          // 让调用方根据具体的错误信息决定如何处理
          break
      }
    } else if (error.code === 'ECONNABORTED') {
      ElMessage.error('请求超时')
    } else {
      ElMessage.error('网络错误')
    }
    
    return Promise.reject(error)
  }
)

export default request 