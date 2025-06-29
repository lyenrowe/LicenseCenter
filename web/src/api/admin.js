import request from './request'

// 获取管理员控制台
export const getAdminDashboard = () => {
  return request.get('/admin/dashboard')
}

// 授权码管理
export const getAuthorizations = (params) => {
  return request.get('/admin/authorizations', { params })
}

export const createAuthorization = (data) => {
  return request.post('/admin/authorizations', data)
}

export const updateAuthorization = (id, data) => {
  return request.put(`/admin/authorizations/${id}`, data)
}

export const deleteAuthorization = (id) => {
  return request.delete(`/admin/authorizations/${id}`)
}

export const getAuthorizationDetails = (id) => {
  return request.get(`/admin/authorizations/${id}/details`)
}

// 设备管理
export const forceUnbindLicense = (licenseId, reason = '') => {
  return request.post(`/admin/licenses/${licenseId}/force-unbind`, {
    reason
  })
}

// 系统管理
export const getSystemLogs = (params) => {
  return request.get('/admin/logs', { params })
}

export const createSystemBackup = () => {
  return request.post('/admin/system/backup', {}, {
    timeout: 60000 // 备份可能需要更长时间
  })
}

export const getSystemConfig = () => {
  return request.get('/admin/system/config')
}

export const updateSystemConfig = (data) => {
  return request.post('/admin/system/config', data)
}

// 获取管理员列表
export const getAdminList = (params) => {
  return request.get('/admin/admins', { params })
}

// 创建管理员
export const createAdmin = (data) => {
  return request.post('/admin/admins', data)
}

// 更新管理员
export const updateAdmin = (id, data) => {
  return request.put(`/admin/admins/${id}`, data)
}

// 删除管理员
export const deleteAdmin = (id) => {
  return request.delete(`/admin/admins/${id}`)
}

// TOTP相关API
export const getTOTPInfo = (adminId) => {
  return request.get(`/admin/totp/info/${adminId}`)
}

export const enableTOTP = (adminId) => {
  return request.post(`/admin/totp/enable/${adminId}`)
}

export const disableTOTP = (adminId) => {
  return request.post(`/admin/totp/disable/${adminId}`)
}

export const verifyTOTPSetup = (adminId, totpCode) => {
  return request.post(`/admin/totp/verify/${adminId}`, {
    totp_code: totpCode
  })
}

// 获取控制台数据
export const getDashboard = () => {
  return request.get('/admin/dashboard')
}

// 获取操作日志
export const getLogs = (params) => {
  return request.get('/admin/logs', { params })
} 