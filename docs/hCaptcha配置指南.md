# hCaptcha 人机验证配置指南

## 概述

本系统已集成 hCaptcha 人机验证功能，用于保护客户端登录入口，防止自动化脚本攻击。本指南将说明如何配置和使用hCaptcha。

## 获取 hCaptcha 密钥

1. 访问 [hCaptcha 官网](https://www.hcaptcha.com/)
2. 注册账户并登录
3. 在仪表板中点击"New Site"创建新站点
4. 填写站点信息：
   - **Site Key (公钥)**: 用于前端页面显示验证码
   - **Secret Key (私钥)**: 用于后端验证验证码响应

## 配置步骤

### 1. 后端配置

编辑配置文件 `configs/app.yaml` 或 `configs/app.local.yaml`：

```yaml
captcha:
  enabled: true                           # 启用验证码
  site_key: "your-hcaptcha-site-key"     # 替换为你的站点密钥
  secret_key: "your-hcaptcha-secret-key" # 替换为你的私钥
```

### 2. 前端配置（可选）

如果需要在前端直接配置站点密钥，可以创建 `web/.env` 文件：

```env
# hCaptcha 配置
VITE_HCAPTCHA_SITE_KEY=your-hcaptcha-site-key
```

> **注意**: 前端配置的密钥优先级低于后端API返回的配置。推荐通过后端API动态获取配置。

## 使用方式

### 1. 开发环境

在开发环境中，系统会自动提供降级验证码：

- 如果 hCaptcha 加载失败，会显示"开发模式"验证码
- 这些降级验证码仅在 `debug` 模式下有效

### 2. 生产环境

在生产环境中，必须使用真实的 hCaptcha 验证：

- 设置 `server.mode: "release"`
- 配置有效的 hCaptcha 密钥
- 降级验证码会被拒绝

## 验证流程

1. **前端获取配置**: 调用 `/api/captcha/config` 获取验证码配置
2. **用户完成验证**: 用户在登录页面完成 hCaptcha 验证
3. **获取验证令牌**: hCaptcha 返回一次性验证令牌
4. **提交登录**: 前端将授权码和验证令牌一起提交
5. **后端验证**: 服务器向 hCaptcha 验证令牌有效性
6. **验证授权码**: 验证通过后再验证授权码

## 错误处理

系统会在以下情况重新要求验证：

- 登录失败（无论是验证码错误还是授权码错误）
- 验证码过期
- 验证码验证失败

## 测试密钥

hCaptcha 提供测试密钥用于开发环境：

```yaml
captcha:
  enabled: true
  site_key: "10000000-ffff-ffff-ffff-000000000001"
  secret_key: "0x0000000000000000000000000000000000000000"
```

> **警告**: 测试密钥仅用于开发环境，生产环境必须使用真实密钥。

## 故障排除

### 1. 验证码加载失败

- 检查网络连接
- 确认站点密钥正确
- 查看浏览器控制台错误信息

### 2. 验证码验证失败

- 确认私钥配置正确
- 检查服务器网络连接
- 查看服务器日志中的详细错误信息

### 3. 降级验证码在生产环境被拒绝

这是正常的安全机制，请：
- 配置正确的 hCaptcha 密钥
- 确保 `server.mode` 设置为 `release`

## API 参考

### 获取验证码配置

```
GET /api/captcha/config
```

响应：
```json
{
  "enabled": true,
  "site_key": "your-hcaptcha-site-key"
}
```

### 客户端登录

```
POST /api/login
```

请求体：
```json
{
  "authorization_code": "your-auth-code",
  "captcha_token": "hcaptcha-response-token"
}
```

## 安全注意事项

1. **私钥安全**: 永远不要在前端代码或版本控制中暴露私钥
2. **环境分离**: 开发、测试、生产环境使用不同的密钥
3. **定期轮换**: 定期更新 hCaptcha 密钥
4. **监控**: 监控验证码的使用情况和异常

## 相关链接

- [hCaptcha 官方文档](https://docs.hcaptcha.com/)
- [hCaptcha 仪表板](https://dashboard.hcaptcha.com/)
- [hCaptcha API 参考](https://docs.hcaptcha.com/) 