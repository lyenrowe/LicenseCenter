<template>
  <div class="admin-layout">
    <!-- 侧边栏 -->
    <div class="sidebar">
      <div class="sidebar-header">
        <h2>管理控制台</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        @select="handleMenuSelect"
        router
      >
        <el-menu-item index="/admin/dashboard">
          <el-icon><Monitor /></el-icon>
          <span>控制台</span>
        </el-menu-item>
        <el-menu-item index="/admin/authorizations">
          <el-icon><Key /></el-icon>
          <span>授权码管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/customers">
          <el-icon><User /></el-icon>
          <span>客户管理</span>
        </el-menu-item>
        <el-menu-item index="/admin/system">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
      
      <div class="sidebar-footer">
        <div class="user-info">
          <div class="user-name">{{ userInfo.username || '管理员' }}</div>
          <div class="user-role">系统管理员</div>
        </div>
        <el-button type="danger" @click="handleLogout" style="width: 100%;">
          <el-icon><SwitchButton /></el-icon>
          注销登录
        </el-button>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <router-view />
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { Monitor, Key, User, Setting, SwitchButton } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const userInfo = computed(() => authStore.userInfo)
const activeMenu = ref(route.path)

// 监听路由变化，更新激活的菜单项
watch(() => route.path, (newPath) => {
  activeMenu.value = newPath
}, { immediate: true })

const handleMenuSelect = (index) => {
  if (index !== route.path) {
    router.push(index)
  }
}

const handleLogout = async () => {
  await authStore.logoutAction()
  router.push('/admin/login')
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  background: #f5f5f5;
}

.sidebar {
  width: 250px;
  background: #2c3e50;
  color: white;
  display: flex;
  flex-direction: column;
  position: fixed;
  height: 100vh;
  z-index: 1000;
}

.sidebar-header {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #34495e;
}

.sidebar-header h2 {
  margin: 0;
  color: white;
  font-size: 18px;
}

.sidebar-menu {
  flex: 1;
  border: none;
  background: transparent;
}

:deep(.el-menu-item) {
  color: #bdc3c7;
  border-radius: 0;
}

:deep(.el-menu-item.is-active) {
  background-color: #3498db;
  color: white;
}

:deep(.el-menu-item:hover) {
  background-color: #34495e;
  color: white;
}

.sidebar-footer {
  padding: 20px;
  border-top: 1px solid #34495e;
}

.user-info {
  margin-bottom: 15px;
  text-align: center;
}

.user-name {
  font-size: 14px;
  font-weight: bold;
  color: white;
  margin-bottom: 4px;
}

.user-role {
  font-size: 12px;
  color: #bdc3c7;
}

.main-content {
  flex: 1;
  margin-left: 250px;
  overflow-y: auto;
  min-height: 100vh;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .sidebar {
    width: 200px;
  }
  
  .main-content {
    margin-left: 200px;
  }
}
</style> 