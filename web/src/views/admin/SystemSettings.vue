<template>
  <div class="system-settings">
    <div class="page-header">
      <h1>系统设置</h1>
    </div>
    
    <el-card>
      <div class="development-notice">
        <el-icon class="notice-icon" size="48"><Tools /></el-icon>
        <h3>功能开发中</h3>
        <p>系统设置功能正在开发中，敬请期待...</p>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { Tools } from '@element-plus/icons-vue'
</script>

<style scoped>
.system-settings {
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

.development-notice {
  text-align: center;
  padding: 60px 20px;
  color: #909399;
}

.notice-icon {
  margin-bottom: 20px;
}

.development-notice h3 {
  margin: 0 0 12px 0;
  font-size: 20px;
}

.development-notice p {
  margin: 0;
  font-size: 16px;
}
</style> 