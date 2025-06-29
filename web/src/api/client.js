import request from './request'

// 客户端登录
export function clientLogin(data) {
  return request({
    url: '/api/login',  // 客户端登录接口
    method: 'post',
    data
  })
}

// 客户端登出
export function clientLogout() {
  return request({
    url: '/api/logout',  // 客户端登出接口
    method: 'post'
  })
}

// 获取客户端控制台数据
export function getClientDashboard() {
  return request({
    url: '/api/client/dashboard',  // 客户端控制台接口
    method: 'post'
  })
}

// 批量激活设备
export const uploadFiles = (files, authorizationCode) => {
  const formData = new FormData()
  files.forEach((file, index) => {
    formData.append(`files`, file.raw)
  })
  formData.append('authorization_code', authorizationCode)
  
  return request.post('/actions/activate-licenses', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    timeout: 60000 // 文件处理可能需要更长时间
  })
}

// 转移授权
export const transferLicense = (fromAuthCode, toAuthCode, deviceIds) => {
  const formData = new FormData()
  formData.append('from_authorization_code', fromAuthCode)
  formData.append('to_authorization_code', toAuthCode)
  deviceIds.forEach(id => formData.append('device_ids', id))
  
  return request.post('/actions/transfer-license', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 下载license文件
export const downloadLicense = (licenseId) => {
  return request.get(`/licenses/${licenseId}/download`, {
    responseType: 'blob'
  })
} 