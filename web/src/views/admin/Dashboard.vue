<template>
  <div class="admin-dashboard">
    <!-- 侧边栏 -->
    <div class="sidebar">
      <div class="sidebar-header">
        <h2>管理控制台</h2>
      </div>
      <el-menu
        :default-active="activeMenu"
        class="sidebar-menu"
        @select="handleMenuSelect"
      >
        <el-menu-item index="dashboard">
          <el-icon><Monitor /></el-icon>
          <span>控制台</span>
        </el-menu-item>
        <el-menu-item index="authorizations">
          <el-icon><Key /></el-icon>
          <span>授权码管理</span>
        </el-menu-item>
        <el-menu-item index="customers">
          <el-icon><User /></el-icon>
          <span>客户管理</span>
        </el-menu-item>
        <el-menu-item index="system">
          <el-icon><Setting /></el-icon>
          <span>系统设置</span>
        </el-menu-item>
      </el-menu>
      
      <div class="sidebar-footer">
        <el-button type="danger" @click="handleLogout" style="width: 100%;">
          <el-icon><SwitchButton /></el-icon>
          注销登录
        </el-button>
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="main-content">
      <!-- 头部 -->
      <div class="content-header">
        <h1>欢迎, {{ userInfo.username || '管理员' }}</h1>
        <p>当前时间: {{ currentTime }}</p>
      </div>

      <!-- 快速统计 -->
      <div class="stats-grid">
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.total_authorizations || 0 }}</div>
            <div class="stat-label">总授权码</div>
          </div>
          <el-icon class="stat-icon" color="#409EFF"><Document /></el-icon>
        </el-card>
        
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.active_customers || 0 }}</div>
            <div class="stat-label">活跃客户</div>
          </div>
          <el-icon class="stat-icon" color="#67C23A"><UserFilled /></el-icon>
        </el-card>
        
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.total_seats || 0 }}</div>
            <div class="stat-label">总席位</div>
          </div>
          <el-icon class="stat-icon" color="#E6A23C"><Monitor /></el-icon>
        </el-card>
        
        <el-card class="stat-card">
          <div class="stat-content">
            <div class="stat-number">{{ dashboardData.used_seats || 0 }}</div>
            <div class="stat-label">已用席位</div>
          </div>
          <el-icon class="stat-icon" color="#F56C6C"><Checked /></el-icon>
        </el-card>
      </div>

      <!-- 功能导航 -->
      <div class="function-grid">
        <el-card class="function-card" @click="goToAuthorizations">
          <div class="function-content">
            <el-icon class="function-icon"><Key /></el-icon>
            <h3>授权码管理</h3>
            <ul>
              <li>创建授权码</li>
              <li>查询授权码</li>
              <li>修改授权码</li>
              <li>禁用授权码</li>
            </ul>
          </div>
        </el-card>
        
        <el-card class="function-card" @click="goToCustomers">
          <div class="function-content">
            <el-icon class="function-icon"><User /></el-icon>
            <h3>客户管理</h3>
            <ul>
              <li>查看客户列表</li>
              <li>客户详情</li>
              <li>强制解绑</li>
            </ul>
          </div>
        </el-card>
        
        <el-card class="function-card" @click="goToSystem">
          <div class="function-content">
            <el-icon class="function-icon"><Setting /></el-icon>
            <h3>系统设置</h3>
            <ul>
              <li>密钥管理</li>
              <li>系统日志</li>
              <li>数据备份</li>
            </ul>
          </div>
        </el-card>
      </div>

      <!-- 最近活动 -->
      <el-card class="activity-card">
        <template #header>
          <div class="card-header">
            <h3>最近活动</h3>
            <el-button type="primary" size="small" @click="refreshData">刷新</el-button>
          </div>
        </template>
        
        <el-table :data="dashboardData.recent_activities || []" style="width: 100%">
          <el-table-column prop="time" label="时间" width="180" />
          <el-table-column prop="action" label="操作" width="150" />
          <el-table-column prop="target" label="目标" />
          <el-table-column prop="operator" label="操作者" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag 
                :type="scope.row.status === 'success' ? 'success' : 'danger'"
                size="small"
              >
                {{ scope.row.status === 'success' ? '成功' : '失败' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getAdminDashboard } from '@/api/admin'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

const userInfo = computed(() => authStore.userInfo)
const dashboardData = ref({})
const currentTime = ref('')
const activeMenu = ref('dashboard')
let timeInterval = null

const updateCurrentTime = () => {
  currentTime.value = new Date().toLocaleString('zh-CN')
}

const loadDashboard = async () => {
  try {
    const response = await getAdminDashboard()
    dashboardData.value = response.data
  } catch (error) {
    ElMessage.error('加载数据失败')
  }
}

const refreshData = () => {
  loadDashboard()
}

const handleMenuSelect = (index) => {
  activeMenu.value = index
  if (index !== 'dashboard') {
    router.push(`/admin/${index}`)
  }
}

const goToAuthorizations = () => {
  router.push('/admin/authorizations')
}

const goToCustomers = () => {
  router.push('/admin/customers')
}

const goToSystem = () => {
  router.push('/admin/system')
}

const handleLogout = async () => {
  await authStore.logoutAction()
  router.push('/admin/login')
}

onMounted(() => {
  updateCurrentTime()
  timeInterval = setInterval(updateCurrentTime, 1000)
  loadDashboard()
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
})
</script>

<style scoped>
.admin-dashboard {
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
}

.sidebar-header {
  padding: 20px;
  text-align: center;
  border-bottom: 1px solid #34495e;
}

.sidebar-header h2 {
  margin: 0;
  color: white;
}

.sidebar-menu {
  flex: 1;
  border: none;
  background: transparent;
}

:deep(.el-menu-item) {
  color: #bdc3c7;
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

.main-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.content-header {
  margin-bottom: 20px;
}

.content-header h1 {
  margin: 0;
  color: #2c3e50;
  font-size: 24px;
}

.content-header p {
  margin: 4px 0 0 0;
  color: #7f8c8d;
  font-size: 14px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  cursor: default;
}

.stat-card .el-card__body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
}

.stat-content {
  flex: 1;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: #2c3e50;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: #7f8c8d;
}

.stat-icon {
  font-size: 32px;
}

.function-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

.function-card {
  cursor: pointer;
  transition: all 0.3s;
}

.function-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.function-content {
  text-align: center;
  padding: 10px;
}

.function-icon {
  font-size: 48px;
  color: #3498db;
  margin-bottom: 16px;
}

.function-content h3 {
  margin: 0 0 16px 0;
  color: #2c3e50;
}

.function-content ul {
  list-style: none;
  padding: 0;
  margin: 0;
}

.function-content li {
  color: #7f8c8d;
  font-size: 14px;
  margin-bottom: 4px;
}

.activity-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header h3 {
  margin: 0;
  color: #2c3e50;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }
  
  .function-grid {
    grid-template-columns: 1fr;
  }
}
</style> 