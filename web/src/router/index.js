import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes = [
  {
    path: '/',
    redirect: '/client/login'
  },
  // 客户端路由
  {
    path: '/client',
    children: [
      {
        path: 'login',
        name: 'ClientLogin',
        component: () => import('@/views/client/Login.vue'),
        meta: { requiresAuth: false }
      },
      {
        path: 'dashboard',
        name: 'ClientDashboard',
        component: () => import('@/views/client/Dashboard.vue'),
        meta: { requiresAuth: true, role: 'client' }
      }
    ]
  },
  // 管理员路由
  {
    path: '/admin',
    children: [
      {
        path: 'login',
        name: 'AdminLogin',
        component: () => import('@/views/admin/Login.vue'),
        meta: { requiresAuth: false }
      },
      {
        path: '',
        component: () => import('@/views/admin/AdminLayout.vue'),
        meta: { requiresAuth: true, role: 'admin' },
        redirect: '/admin/dashboard',
        children: [
          {
            path: 'dashboard',
            name: 'AdminDashboard',
            component: () => import('@/views/admin/Dashboard.vue')
          },
          {
            path: 'authorizations',
            name: 'AuthorizationManagement',
            component: () => import('@/views/admin/AuthorizationManagement.vue')
          },
          {
            path: 'customers',
            name: 'CustomerManagement',
            component: () => import('@/views/admin/CustomerManagement.vue')
          },
          {
            path: 'system',
            name: 'SystemSettings',
            component: () => import('@/views/admin/SystemSettings.vue')
          }
        ]
      }
    ]
  },
  // 404页面 - 必须放在最后
  {
    path: '/:pathMatch(.*)*',
    name: 'NotFound',
    component: () => import('@/views/NotFound.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  console.log('路由守卫:', { to: to.path, from: from.path })
  
  try {
    const authStore = useAuthStore()
    
    if (to.meta.requiresAuth) {
      if (!authStore.isAuthenticated) {
        console.log('未登录，重定向到登录页')
        // 未登录，跳转到对应的登录页面
        if (to.meta.role === 'admin') {
          next('/admin/login')
        } else {
          next('/client/login')
        }
        return
      }
      
      // 检查角色权限
      if (to.meta.role && authStore.userRole !== to.meta.role) {
        console.log('权限不足，重定向到首页')
        next('/') // 权限不足，跳转到首页
        return
      }
    }
    
    console.log('路由守卫通过')
    next()
  } catch (error) {
    console.error('路由守卫错误:', error)
    next('/client/login')
  }
})

// 路由错误处理
router.onError((error) => {
  console.error('路由错误:', error)
})

export default router 