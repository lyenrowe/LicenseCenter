<template>
  <div class="customer-management">
    <div class="page-header">
      <h1>客户管理</h1>
    </div>

    <!-- 搜索 -->
    <el-card class="search-card">
      <el-form :model="searchForm" inline>
        <el-form-item label="搜索">
          <el-input
            v-model="searchForm.search"
            placeholder="客户名称或授权码"
            clearable
            style="width: 200px"
          />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadCustomers">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 客户列表 -->
    <el-card>
      <el-table :data="customers" v-loading="loading" stripe>
        <el-table-column prop="customer_name" label="客户名称" />
        <el-table-column prop="authorization_code" label="授权码" width="240" />
        <el-table-column label="席位使用" width="120">
          <template #default="scope">
            <el-progress 
              :percentage="(scope.row.used_seats / scope.row.max_seats) * 100"
              :format="() => `${scope.row.used_seats}/${scope.row.max_seats}`"
            />
          </template>
        </el-table-column>
        <el-table-column prop="active_devices" label="活跃设备" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="300" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-tag :type="scope.row.status === 1 ? 'success' : 'danger'">
              {{ scope.row.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <el-button size="small" @click="viewDetails(scope.row)">详情</el-button>
            <el-button size="small" type="primary" @click="manageDevices(scope.row)">设备管理</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          v-model:current-page="pagination.page"
          v-model:page-size="pagination.limit"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadCustomers"
          @current-change="loadCustomers"
        />
      </div>
    </el-card>

    <!-- 客户详情对话框 -->
    <el-dialog title="客户详情" v-model="showDetailsDialog" width="800px">
      <div v-if="selectedCustomer">
        <el-descriptions :column="2" border>
          <el-descriptions-item label="客户名称">{{ selectedCustomer.customer_name }}</el-descriptions-item>
          <el-descriptions-item label="授权码">{{ selectedCustomer.authorization_code }}</el-descriptions-item>
          <el-descriptions-item label="最大席位">{{ selectedCustomer.max_seats }}</el-descriptions-item>
          <el-descriptions-item label="已用席位">{{ selectedCustomer.used_seats }}</el-descriptions-item>
          <el-descriptions-item label="创建时间">{{ formatDateTime(selectedCustomer.created_at) }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="selectedCustomer.status === 1 ? 'success' : 'danger'">
              {{ selectedCustomer.status === 1 ? '正常' : '禁用' }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedCustomer.duration_years" label="授权年限">
            {{ selectedCustomer.duration_years }} 年
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedCustomer.latest_expiry_date" label="最晚到期时间">
            {{ formatDateTime(selectedCustomer.latest_expiry_date) }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 设备列表 -->
        <div style="margin: 20px 0 10px 0; display: flex; justify-content: space-between; align-items: center;">
          <h3 style="margin: 0;">已激活设备 ({{ customerDevices.length }})</h3>
          <el-button size="small" @click="refreshDevices">刷新</el-button>
        </div>
        
        <el-table :data="customerDevices" stripe v-loading="devicesLoading">
          <el-table-column prop="hostname" label="主机名" width="150" />
          <el-table-column prop="machine_id" label="机器ID" show-overflow-tooltip width="200" />
          <el-table-column prop="activated_at" label="激活时间" width="160">
            <template #default="scope">
              {{ formatDateTime(scope.row.activated_at) }}
            </template>
          </el-table-column>
          <el-table-column prop="expires_at" label="到期时间" width="160">
            <template #default="scope">
              {{ formatDateTime(scope.row.expires_at) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '正常' : '已解绑' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button 
                v-if="scope.row.status === 'active'"
                size="small" 
                type="danger" 
                @click="forceUnbind(scope.row)"
              >
                强制解绑
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="customerDevices.length === 0" description="暂无激活设备" />
      </div>
    </el-dialog>

    <!-- 设备管理对话框 -->
    <el-dialog title="设备管理" v-model="showDeviceDialog" width="900px">
      <div v-if="selectedCustomer">
        <div class="device-header">
          <div class="customer-info">
            <h4>{{ selectedCustomer.customer_name }}</h4>
            <p>授权码: {{ selectedCustomer.authorization_code }}</p>
            <p>席位使用: {{ selectedCustomer.used_seats }} / {{ selectedCustomer.max_seats }}</p>
          </div>
          <el-button @click="refreshDevices">刷新设备列表</el-button>
        </div>

        <el-table :data="customerDevices" stripe v-loading="devicesLoading">
          <el-table-column prop="hostname" label="主机名" width="150" />
          <el-table-column prop="machine_id" label="机器ID" show-overflow-tooltip width="250" />
          <el-table-column prop="activated_at" label="激活时间" width="160">
            <template #default="scope">
              {{ formatDateTime(scope.row.activated_at) }}
            </template>
          </el-table-column>
          <el-table-column prop="expires_at" label="到期时间" width="160">
            <template #default="scope">
              {{ formatDateTime(scope.row.expires_at) }}
            </template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="scope">
              <el-tag :type="scope.row.status === 'active' ? 'success' : 'danger'">
                {{ scope.row.status === 'active' ? '正常' : '已解绑' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="scope">
              <el-button 
                v-if="scope.row.status === 'active'"
                size="small" 
                type="danger" 
                @click="forceUnbind(scope.row)"
              >
                强制解绑
              </el-button>
            </template>
          </el-table-column>
        </el-table>

        <el-empty v-if="customerDevices.length === 0" description="暂无激活设备" />
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAuthorizations, getAuthorizationDetails, forceUnbindLicense } from '@/api/admin'

const loading = ref(false)
const showDetailsDialog = ref(false)
const showDeviceDialog = ref(false)
const devicesLoading = ref(false)
const selectedCustomer = ref(null)
const customerDevices = ref([])

const customers = ref([])
const searchForm = reactive({
  search: ''
})

const pagination = reactive({
  page: 1,
  limit: 20,
  total: 0
})

const loadCustomers = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      limit: pagination.limit,
      search: searchForm.search || undefined
    }
    
    const response = await getAuthorizations(params)
    customers.value = response.data.data || []
    pagination.total = response.data.pagination?.total || 0
  } catch (error) {
    ElMessage.error('加载客户列表失败')
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.search = ''
  pagination.page = 1
  loadCustomers()
}

const viewDetails = async (customer) => {
  selectedCustomer.value = customer
  showDetailsDialog.value = true
  await loadCustomerDevices()
}

const manageDevices = async (customer) => {
  selectedCustomer.value = customer
  showDeviceDialog.value = true
  await loadCustomerDevices()
}

const loadCustomerDevices = async () => {
  if (!selectedCustomer.value) return
  
  devicesLoading.value = true
  try {
    const response = await getAuthorizationDetails(selectedCustomer.value.id)
    customerDevices.value = response.data.data.devices || []
    // 更新客户信息（可能有最新的席位使用情况）
    selectedCustomer.value = response.data.data
  } catch (error) {
    ElMessage.error('加载客户设备信息失败')
    customerDevices.value = []
  } finally {
    devicesLoading.value = false
  }
}

const refreshDevices = () => {
  loadCustomerDevices()
}

const forceUnbind = async (device) => {
  try {
    await ElMessageBox.confirm(
      `确定要强制解绑设备 "${device.hostname}" 吗？此操作不可恢复。`, 
      '确认强制解绑',
      { type: 'warning' }
    )
    
    const reason = await ElMessageBox.prompt('请输入解绑原因：', '解绑原因', {
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })
    
    await forceUnbindLicense(device.id, reason.value)
    ElMessage.success('设备已强制解绑')
    
    // 重新加载设备列表和客户列表
    await loadCustomerDevices()
    await loadCustomers()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('强制解绑失败')
    }
  }
}

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  const date = new Date(dateTime)
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

onMounted(() => {
  loadCustomers()
})
</script>

<style scoped>
.customer-management {
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  margin-bottom: 20px;
}

.page-header h1 {
  margin: 0;
  color: #2c3e50;
}

.search-card {
  margin-bottom: 20px;
}

.pagination-wrapper {
  margin-top: 20px;
  text-align: right;
}

.device-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 4px;
}

.customer-info h4 {
  margin: 0 0 8px 0;
  color: #2c3e50;
  font-size: 16px;
}

.customer-info p {
  margin: 4px 0;
  color: #606266;
  font-size: 14px;
}
</style> 