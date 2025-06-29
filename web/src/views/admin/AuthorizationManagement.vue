<template>
  <div class="authorization-management">
    <div class="page-header">
      <h1>授权码管理</h1>
      <el-button type="primary" @click="showCreateDialog = true">
        <el-icon><Plus /></el-icon>
        创建授权码
      </el-button>
    </div>

    <!-- 搜索和筛选 -->
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
        <el-form-item label="状态">
          <el-select v-model="searchForm.status" placeholder="全部状态" clearable style="width: 120px">
            <el-option label="正常" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadAuthorizations">搜索</el-button>
          <el-button @click="resetSearch">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <!-- 授权码列表 -->
    <el-card>
      <el-table :data="authorizations" v-loading="loading" stripe>
        <el-table-column prop="authorization_code" label="授权码" width="240" />
        <el-table-column prop="customer_name" label="客户名称" />
        <el-table-column label="席位使用" width="120">
          <template #default="scope">
            {{ scope.row.used_seats }} / {{ scope.row.max_seats }}
          </template>
        </el-table-column>
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
            <el-button size="small" type="primary" @click="editAuthorization(scope.row)">编辑</el-button>
            <el-button 
              size="small" 
              :type="scope.row.status === 1 ? 'warning' : 'success'"
              @click="toggleStatus(scope.row)"
            >
              {{ scope.row.status === 1 ? '禁用' : '启用' }}
            </el-button>
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
          @size-change="loadAuthorizations"
          @current-change="loadAuthorizations"
        />
      </div>
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog 
      :title="editingAuth ? '编辑授权码' : '创建授权码'"
      v-model="showCreateDialog"
      width="500px"
    >
      <el-form :model="authForm" :rules="authRules" ref="authFormRef" label-width="100px">
        <el-form-item label="客户名称" prop="customer_name">
          <el-input v-model="authForm.customer_name" placeholder="请输入客户名称" />
        </el-form-item>
        <el-form-item label="最大席位" prop="max_seats">
          <el-input-number v-model="authForm.max_seats" :min="1" :max="1000" />
        </el-form-item>
        <el-form-item label="授权年限" prop="duration_years">
          <el-input-number v-model="authForm.duration_years" :min="1" :max="99" />
          <div class="form-tip">99表示永久授权</div>
        </el-form-item>
        <el-form-item label="最晚到期时间">
          <el-date-picker
            v-model="authForm.latest_expiry_date"
            type="datetime"
            placeholder="选择最晚到期时间"
            format="YYYY-MM-DD HH:mm:ss"
          />
          <div class="form-tip">可选，优先级高于授权年限</div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="showCreateDialog = false">取消</el-button>
          <el-button type="primary" @click="saveAuthorization" :loading="saving">
            {{ editingAuth ? '更新' : '创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { getAuthorizations, createAuthorization, updateAuthorization, deleteAuthorization } from '@/api/admin'

const loading = ref(false)
const saving = ref(false)
const showCreateDialog = ref(false)
const editingAuth = ref(null)
const authFormRef = ref()

const authorizations = ref([])
const searchForm = reactive({
  search: '',
  status: null
})

const pagination = reactive({
  page: 1,
  limit: 20,
  total: 0
})

const authForm = reactive({
  customer_name: '',
  max_seats: 1,
  duration_years: 1,
  latest_expiry_date: null
})

const authRules = {
  customer_name: [
    { required: true, message: '请输入客户名称', trigger: 'blur' }
  ],
  max_seats: [
    { required: true, message: '请设置最大席位数', trigger: 'blur' }
  ],
  duration_years: [
    { required: true, message: '请设置授权年限', trigger: 'blur' }
  ]
}

const loadAuthorizations = async () => {
  loading.value = true
  try {
    const params = {
      page: pagination.page,
      limit: pagination.limit,
      search: searchForm.search || undefined,
      status: searchForm.status !== null ? searchForm.status : undefined
    }
    
    const response = await getAuthorizations(params)
    authorizations.value = response.data.data || []
    pagination.total = response.data.pagination?.total || 0
  } catch (error) {
    ElMessage.error('加载授权码列表失败')
  } finally {
    loading.value = false
  }
}

const resetSearch = () => {
  searchForm.search = ''
  searchForm.status = null
  pagination.page = 1
  loadAuthorizations()
}

const viewDetails = (auth) => {
  // TODO: 实现详情查看
  ElMessage.info('详情功能待实现')
}

const editAuthorization = (auth) => {
  editingAuth.value = auth
  authForm.customer_name = auth.customer_name
  authForm.max_seats = auth.max_seats
  authForm.duration_years = auth.duration_years || 1
  authForm.latest_expiry_date = auth.latest_expiry_date ? new Date(auth.latest_expiry_date) : null
  showCreateDialog.value = true
}

const saveAuthorization = async () => {
  if (!authFormRef.value) return
  
  const valid = await authFormRef.value.validate()
  if (!valid) return
  
  saving.value = true
  try {
    const data = {
      customer_name: authForm.customer_name,
      max_seats: authForm.max_seats,
      duration_years: authForm.duration_years,
      latest_expiry_date: authForm.latest_expiry_date?.toISOString()
    }
    
    if (editingAuth.value) {
      await updateAuthorization(editingAuth.value.id, data)
      ElMessage.success('授权码更新成功')
    } else {
      await createAuthorization(data)
      ElMessage.success('授权码创建成功')
    }
    
    showCreateDialog.value = false
    resetForm()
    loadAuthorizations()
  } catch (error) {
    ElMessage.error(editingAuth.value ? '更新失败' : '创建失败')
  } finally {
    saving.value = false
  }
}

const toggleStatus = async (auth) => {
  const newStatus = auth.status === 1 ? 0 : 1
  const action = newStatus === 1 ? '启用' : '禁用'
  
  try {
    await ElMessageBox.confirm(`确定要${action}这个授权码吗？`, '确认操作')
    
    await updateAuthorization(auth.id, { status: newStatus })
    ElMessage.success(`${action}成功`)
    loadAuthorizations()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  }
}

const resetForm = () => {
  editingAuth.value = null
  authForm.customer_name = ''
  authForm.max_seats = 1
  authForm.duration_years = 1
  authForm.latest_expiry_date = null
}

onMounted(() => {
  loadAuthorizations()
})
</script>

<style scoped>
.authorization-management {
  padding: 20px;
  background: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
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

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style> 