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
        path: 'dashboard',
        name: 'AdminDashboard',
        component: () => import('@/views/admin/Dashboard.vue'),
        meta: { requiresAuth: true, role: 'admin' }
      },
      {
        path: 'authorizations',
        name: 'AuthorizationManagement',
        component: () => import('@/views/admin/AuthorizationManagement.vue'),
        meta: { requiresAuth: true, role: 'admin' }
      },
      {
        path: 'customers',
        name: 'CustomerManagement',
        component: () => import('@/views/admin/CustomerManagement.vue'),
        meta: { requiresAuth: true, role: 'admin' }
      },
      {
        path: 'system',
        name: 'SystemSettings',
        component: () => import('@/views/admin/SystemSettings.vue'),
        meta: { requiresAuth: true, role: 'admin' }
      }
    ]
  },
  // 404页面
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
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth) {
    if (!authStore.isAuthenticated) {
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
      next('/') // 权限不足，跳转到首页
      return
    }
  }
  
  next()
})

export default router 