import request from './request'

// 获取客户端控制台信息
export const getClientDashboard = () => {
  return request.get('/dashboard')
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