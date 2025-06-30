import request from './request'

// 获取验证码配置
export const getCaptchaConfig = () => {
  return request.get('/captcha/config')
}

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

// 客户端注销登录
export const clientLogout = () => {
  return request.post('/logout')
}

// 管理员注销登录
export const adminLogout = () => {
  return request.post('/admin/logout')
}

// 通用注销登录（兼容旧版本）
export const logout = () => {
  return request.post('/logout')
}

// 获取公钥
export const getPublicKey = () => {
  return request.get('/public-key')
} 