import request from './request'

// 客户端登录
export const clientLogin = (authorizationCode, captchaToken) => {
  return request.post('/login', {
    authorization_code: authorizationCode,
    captcha_token: captchaToken
  })
}

// 管理员登录
export const adminLogin = (username, password, totpCode) => {
  return request.post('/admin/login', {
    username,
    password,
    totp_code: totpCode
  })
}

// 注销登录
export const logout = () => {
  return request.post('/logout')
}

// 获取公钥
export const getPublicKey = () => {
  return request.get('/public-key')
} 