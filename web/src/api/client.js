import request from './request'

// 客户端登录
export function clientLogin(data) {
  return request({
    url: '/login',  // 客户端登录接口
    method: 'post',
    data
  })
}

// 客户端登出
export function clientLogout() {
  return request({
    url: '/logout',  // 客户端登出接口
    method: 'post'
  })
}

// 获取客户端控制台数据
export function getDashboard() {
  return request({
    url: '/client/dashboard',  // 客户端控制台接口
    method: 'get'
  })
}

// 批量激活设备
export function activateLicenses(bindFiles) {
  const formData = new FormData()
  bindFiles.forEach((file) => {
    formData.append('bind_files', file)
  })
  
  return request({
    url: '/actions/activate-licenses',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    responseType: 'blob', // 返回ZIP文件
    timeout: 60000
  }).catch(error => {
    // 如果是业务错误（400-499），尝试提取具体错误信息
    if (error.response && error.response.status >= 400 && error.response.status < 500) {
      const data = error.response.data
      if (data && data.error) {
        // 创建一个新的错误对象，包含具体的错误信息
        const businessError = new Error(data.error)
        businessError.code = data.code
        businessError.response = error.response
        throw businessError
      }
    }
    // 其他错误直接抛出
    throw error
  })
}

// 转移授权
export function transferLicense(unbindFile, bindFile) {
  const formData = new FormData()
  formData.append('unbind_file', unbindFile)
  formData.append('bind_file', bindFile)
  
  return request({
    url: '/actions/transfer-license',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    },
    responseType: 'blob', // 返回license文件
    timeout: 30000
  }).catch(error => {
    // 如果是业务错误（400-499），尝试提取具体错误信息
    if (error.response && error.response.status >= 400 && error.response.status < 500) {
      const data = error.response.data
      if (data && data.error) {
        // 创建一个新的错误对象，包含具体的错误信息
        const businessError = new Error(data.error)
        businessError.code = data.code
        businessError.response = error.response
        throw businessError
      }
    }
    // 其他错误直接抛出
    throw error
  })
}

// 下载license文件
export function downloadLicense(licenseId) {
  return request({
    url: `/licenses/${licenseId}/download`,
    method: 'get',
    responseType: 'blob'
  })
} 