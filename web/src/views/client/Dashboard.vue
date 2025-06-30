<template>
  <div class="dashboard">
    <div class="header">
      <h1>OCR2Doc 授权管理控制台</h1>
      <el-button type="danger" @click="handleLogout">注销登录</el-button>
    </div>
    
    <el-card class="info-card">
      <h3>欢迎，{{ dashboardData.authorization?.customer_name || userInfo.customer_name }}</h3>
      <p>您的授权码: {{ dashboardData.authorization?.authorization_code || userInfo.authorization_code }}</p>
      <p>授权席位状态: {{ dashboardData.authorization?.used_seats || 0 }} / {{ dashboardData.authorization?.max_seats || 0 }} (已用/总量)</p>
      <p>可用席位: {{ dashboardData.authorization?.available_seats || 0 }}</p>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>新设备授权</span>
            </div>
          </template>
          <div class="upload-section">
            <el-upload
              :auto-upload="false"
              :on-change="handleBindFiles"
              :file-list="bindFileList"
              accept=".bind"
              multiple
            >
              <el-button type="primary" plain>选择 .bind 文件</el-button>
              <template #tip>
                <div class="el-upload__tip">
                  支持批量上传多个.bind文件进行设备激活
                </div>
              </template>
            </el-upload>
            <el-button 
              type="primary" 
              @click="activateDevices"
              :disabled="bindFiles.length === 0"
              :loading="activating"
              style="margin-top: 15px; width: 100%;"
            >
              激活设备 ({{ bindFiles.length }}个)
            </el-button>
          </div>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>转移授权</span>
            </div>
          </template>
          <div class="transfer-section">
            <div class="upload-item">
              <label>1. 选择解绑文件：</label>
              <el-upload
                :auto-upload="false"
                :on-change="handleUnbindFile"
                :limit="1"
                accept=".unbind"
              >
                <el-button type="warning" plain size="small">选择 .unbind 文件</el-button>
              </el-upload>
            </div>
            <div class="upload-item">
              <label>2. 选择新设备绑定文件：</label>
              <el-upload
                :auto-upload="false"
                :on-change="handleTransferBindFile"
                :limit="1"
                accept=".bind"
              >
                <el-button type="primary" plain size="small">选择新设备 .bind 文件</el-button>
              </el-upload>
            </div>
            <el-button 
              type="warning" 
              @click="transferLicense"
              :disabled="!unbindFile || !transferBindFile"
              :loading="transferring"
              style="margin-top: 15px; width: 100%;"
            >
              转移授权
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 已激活设备列表 -->
    <el-card style="margin-top: 20px;">
      <template #header>
        <div class="card-header">
          <span>已激活设备列表 ({{ (dashboardData.devices?.active || []).length }})</span>
        </div>
      </template>
      <el-table :data="dashboardData.devices?.active || []" stripe>
        <el-table-column prop="hostname" label="主机名" min-width="150" />
        <el-table-column prop="machine_id" label="机器ID (部分)" width="150">
          <template #default="scope">
            <code>{{ scope.row.machine_id }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="issued_at" label="激活日期" width="180">
          <template #default="scope">
            {{ new Date(scope.row.issued_at).toLocaleDateString() }}
          </template>
        </el-table-column>
        <el-table-column prop="expires_at" label="到期日期" width="180">
          <template #default="scope">
            {{ new Date(scope.row.expires_at).toLocaleDateString() }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button 
              size="small" 
              type="primary" 
              link
              @click="downloadLicense(scope.row.id)"
            >
              下载.license
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-empty v-if="!(dashboardData.devices?.active || []).length" description="暂无已激活设备" />
    </el-card>

    <!-- 已解绑/历史设备列表 -->
    <el-card style="margin-top: 20px;" v-if="(dashboardData.devices?.historical || []).length">
      <template #header>
        <div class="card-header">
          <span>已解绑/历史设备列表 ({{ (dashboardData.devices?.historical || []).length }})</span>
        </div>
      </template>
      <el-table :data="dashboardData.devices?.historical || []" stripe>
        <el-table-column prop="hostname" label="主机名" min-width="150" />
        <el-table-column prop="machine_id" label="机器ID (部分)" width="150">
          <template #default="scope">
            <code>{{ scope.row.machine_id }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="unbound_at" label="解绑日期" width="180">
          <template #default="scope">
            {{ scope.row.unbound_at ? new Date(scope.row.unbound_at).toLocaleDateString() : '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="expires_at" label="原始到期日" width="180">
          <template #default="scope">
            {{ new Date(scope.row.expires_at).toLocaleDateString() }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="scope">
            <el-tag v-if="scope.row.status === 'unbound'" type="warning">已解绑</el-tag>
            <el-tag v-else-if="scope.row.status === 'force_unbound'" type="danger">强制解绑</el-tag>
            <el-tag v-else type="info">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getDashboard, activateLicenses, transferLicense as transferLicenseApi, downloadLicense as downloadLicenseApi } from '@/api/client'
import { ElMessage } from 'element-plus'

const router = useRouter()
const authStore = useAuthStore()

const userInfo = computed(() => authStore.userInfo)
const dashboardData = ref({})
const bindFiles = ref([])
const bindFileList = ref([])
const unbindFile = ref(null)
const transferBindFile = ref(null)
const activating = ref(false)
const transferring = ref(false)

const loadDashboard = async () => {
  try {
    const response = await getDashboard()
    dashboardData.value = response.data
  } catch (error) {
    ElMessage.error('加载数据失败')
  }
}

const handleBindFiles = (file, fileList) => {
  bindFiles.value = fileList.map(item => item.raw)
  bindFileList.value = fileList
}

const handleUnbindFile = (file) => {
  unbindFile.value = file.raw
}

const handleTransferBindFile = (file) => {
  transferBindFile.value = file.raw
}

const activateDevices = async () => {
  if (bindFiles.value.length === 0) {
    ElMessage.warning('请先选择.bind文件')
    return
  }
  
  activating.value = true
  try {
    const response = await activateLicenses(bindFiles.value)
    const blob = new Blob([response.data], { type: 'application/zip' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'licenses.zip'
    a.click()
    URL.revokeObjectURL(url)
    
    ElMessage.success(`成功激活 ${bindFiles.value.length} 个设备`)
    bindFiles.value = []
    bindFileList.value = []
    loadDashboard()
  } catch (error) {
    console.error('设备激活错误:', error)
    
    // 优先显示具体的错误信息
    let errorMessage = '设备激活失败'
    
    if (error.message) {
      // 如果是我们在API中处理过的业务错误
      errorMessage = error.message
    } else if (error.response?.data?.error) {
      // 如果响应中有具体错误信息
      errorMessage = error.response.data.error
    } else if (error.response?.data?.message) {
      // 备用：检查message字段
      errorMessage = error.response.data.message
    }
    
    ElMessage.error(errorMessage)
  } finally {
    activating.value = false
  }
}

const transferLicense = async () => {
  if (!unbindFile.value || !transferBindFile.value) {
    ElMessage.warning('请先选择解绑文件和新设备绑定文件')
    return
  }
  
  transferring.value = true
  try {
    const response = await transferLicenseApi(unbindFile.value, transferBindFile.value)
    const blob = new Blob([response.data])
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'transferred_license.license'
    a.click()
    URL.revokeObjectURL(url)
    
    ElMessage.success('授权转移成功')
    unbindFile.value = null
    transferBindFile.value = null
    loadDashboard()
  } catch (error) {
    console.error('授权转移错误:', error)
    
    // 优先显示具体的错误信息
    let errorMessage = '授权转移失败'
    
    if (error.message) {
      errorMessage = error.message
    } else if (error.response?.data?.error) {
      errorMessage = error.response.data.error
    } else if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    }
    
    ElMessage.error(errorMessage)
  } finally {
    transferring.value = false
  }
}

const downloadLicense = async (licenseId) => {
  try {
    const response = await downloadLicenseApi(licenseId)
    const blob = new Blob([response.data])
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `license_${licenseId}.license`
    a.click()
    URL.revokeObjectURL(url)
    
    ElMessage.success('授权文件下载成功')
  } catch (error) {
    console.error('下载授权文件错误:', error)
    
    // 优先显示具体的错误信息
    let errorMessage = '下载失败'
    
    if (error.message) {
      errorMessage = error.message
    } else if (error.response?.data?.error) {
      errorMessage = error.response.data.error
    } else if (error.response?.data?.message) {
      errorMessage = error.response.data.message
    }
    
    ElMessage.error(errorMessage)
  }
}

const handleLogout = async () => {
  await authStore.logoutAction()
  router.push('/client/login')
}

onMounted(() => {
  loadDashboard()
})
</script>

<style scoped>
.dashboard {
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.info-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}

.upload-section {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.transfer-section {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.upload-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.upload-item label {
  font-size: 14px;
  font-weight: 500;
  color: #606266;
}

:deep(.el-upload__tip) {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
}

:deep(.el-table) {
  font-size: 14px;
}

:deep(.el-table code) {
  background: #f5f5f5;
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 12px;
}
</style> 