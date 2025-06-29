<template>
  <div class="dashboard">
    <div class="header">
      <h1>OCR2Doc 授权管理控制台</h1>
      <el-button type="danger" @click="handleLogout">注销登录</el-button>
    </div>
    
    <el-card class="info-card">
      <h3>授权信息</h3>
      <p>授权码: {{ userInfo.authorization_code }}</p>
      <p>席位使用: {{ dashboardData.seats_info?.used }} / {{ dashboardData.seats_info?.total }}</p>
    </el-card>

    <el-row :gutter="20">
      <el-col :span="12">
        <el-card>
          <h3>新设备授权</h3>
          <el-upload
            :auto-upload="false"
            :on-change="handleBindFiles"
            accept=".bind"
            multiple
          >
            <el-button>选择 .bind 文件</el-button>
          </el-upload>
          <el-button 
            type="primary" 
            @click="activateDevices"
            :disabled="bindFiles.length === 0"
            style="margin-top: 10px;"
          >
            激活设备
          </el-button>
        </el-card>
      </el-col>
      
      <el-col :span="12">
        <el-card>
          <h3>转移授权</h3>
          <div>
            <el-upload
              :auto-upload="false"
              :on-change="handleUnbindFile"
              accept=".unbind"
              :limit="1"
            >
              <el-button size="small">选择 .unbind 文件</el-button>
            </el-upload>
            
            <el-upload
              :auto-upload="false"
              :on-change="handleTransferBindFile"
              accept=".bind"
              :limit="1"
              style="margin-top: 8px;"
            >
              <el-button size="small">选择 .bind 文件</el-button>
            </el-upload>
          </div>
          
          <el-button 
            type="warning" 
            @click="transferLicense"
            :disabled="!unbindFile || !transferBindFile"
            style="margin-top: 10px;"
          >
            转移授权
          </el-button>
        </el-card>
      </el-col>
    </el-row>

    <el-card style="margin-top: 20px;">
      <h3>已激活设备列表</h3>
      <el-table :data="dashboardData.activated_devices || []">
        <el-table-column prop="hostname" label="主机名" />
        <el-table-column prop="machine_id" label="机器ID" width="200">
          <template #default="scope">
            {{ scope.row.machine_id.substring(0, 12) }}...
          </template>
        </el-table-column>
        <el-table-column prop="activated_at" label="激活日期" />
        <el-table-column label="操作">
          <template #default="scope">
            <el-button size="small" @click="downloadLicense(scope.row.id)">
              下载授权文件
            </el-button>
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
const unbindFile = ref(null)
const transferBindFile = ref(null)

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
}

const handleUnbindFile = (file) => {
  unbindFile.value = file.raw
}

const handleTransferBindFile = (file) => {
  transferBindFile.value = file.raw
}

const activateDevices = async () => {
  try {
    const response = await activateLicenses(bindFiles.value)
    const blob = new Blob([response.data], { type: 'application/zip' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'licenses.zip'
    a.click()
    URL.revokeObjectURL(url)
    
    ElMessage.success('设备激活成功')
    bindFiles.value = []
    loadDashboard()
  } catch (error) {
    ElMessage.error('设备激活失败')
  }
}

const transferLicense = async () => {
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
    ElMessage.error('授权转移失败')
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
    ElMessage.error('下载失败')
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
</style> 