<template>
  <div class="dashboard-content">
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
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getAdminDashboard } from '@/api/admin'
import { ElMessage, ElLoading } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

const userInfo = computed(() => authStore.userInfo)
const dashboardData = ref({
  total_authorizations: 0,
  active_customers: 0,
  total_seats: 0,
  used_seats: 0,
  today_new_authorizations: 0,
  today_new_devices: 0,
  expiring_licenses: 0,
  recent_activities: []
})
const currentTime = ref('')
const loading = ref(false)
let timeInterval = null

const updateCurrentTime = () => {
  currentTime.value = new Date().toLocaleString('zh-CN')
}

const loadDashboard = async () => {
  loading.value = true
  try {
    const response = await getAdminDashboard()
    if (response.data && response.data.data) {
      dashboardData.value = { ...dashboardData.value, ...response.data.data }
    }
    console.log('Dashboard data loaded:', dashboardData.value)
  } catch (error) {
    console.error('Dashboard load error:', error)
    ElMessage.error('加载数据失败：' + (error.response?.data?.error || error.message))
  } finally {
    loading.value = false
  }
}

const refreshData = () => {
  loadDashboard()
  ElMessage.success('数据已刷新')
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
.dashboard-content {
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
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