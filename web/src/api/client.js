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