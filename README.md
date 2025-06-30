# LicenseCenter - OCR2Doc 离线授权管理系统

基于 Golang + Gin + GORM 构建的完整离线授权管理系统，支持授权码管理、设备激活、授权转移等功能。

## 📊 项目状态

![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)
![License](https://img.shields.io/badge/License-MIT-green.svg)
![Status](https://img.shields.io/badge/Status-Production_Ready-brightgreen.svg)
![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen.svg)
![Coverage](https://img.shields.io/badge/Coverage-45.5%25-green.svg)

**🎉 核心功能已完成，系统已可投入生产使用！**

## 🚀 功能特性

- **完全离线**: 客户端无需网络连接即可验证授权
- **安全可靠**: RSA-2048 数字签名，防止授权伪造
- **席位管理**: 灵活的席位分配和管理机制
- **授权转移**: 支持设备间安全的授权转移
- **管理控制台**: 完整的Web管理界面
- **多数据库支持**: SQLite、MySQL、PostgreSQL

## 📁 项目结构

```
LicenseCenter/
├── cmd/server/           # 应用程序入口
├── internal/
│   ├── config/          # 配置管理
│   ├── database/        # 数据库连接
│   ├── models/          # 数据模型
│   ├── services/        # 业务逻辑层
│   ├── handlers/        # HTTP处理器
│   ├── middleware/      # 中间件
│   └── router/          # 路由配置
├── pkg/
│   ├── crypto/          # 加密工具
│   ├── errors/          # 错误定义
│   ├── logger/          # 日志工具
│   └── utils/           # 通用工具
├── configs/             # 配置文件
├── test_client/         # 测试客户端
├── tests/               # 测试文件
└── docs/                # 文档
```

## 🛠️ 快速开始

### 1. 环境要求

- Go 1.21+
- SQLite3 (默认) 或 MySQL/PostgreSQL

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 初始化项目

```bash
# 初始化项目环境
make setup

# 构建所有程序
make build

# 初始化系统数据（创建默认管理员、RSA密钥等）
make init-system
```

### 4. 启动服务

```bash
# 启动服务器
make run

# 或使用开发模式
make dev
```

### 5. 测试客户端

```bash
# 显示机器信息
make run-client

# 生成绑定文件
make generate-bind
```

### 6. 默认管理员账号

系统初始化后的默认管理员账号：
- 用户名：`admin`
- 密码：`admin123`
- 登录地址：`http://localhost:8080/api/admin/login`

**请及时修改默认密码！**

## 🔧 开发工具

### 编译项目

```bash
# 编译所有程序
make build

# 或分别编译
make server     # 编译服务端
make client     # 编译测试客户端
make init-tool  # 编译初始化工具
```

### 运行测试

```bash
# 运行所有测试
make test

# 或直接使用 go test
go test ./tests/... -v

# 运行特定测试套件
go test ./tests/ -run TestLicenseServiceSuite -v
go test ./tests/ -run TestIntegrationTestSuite -v
```

### 数据库管理

```bash
# 系统会自动进行数据库迁移
# 如需重置数据库：
make reset-db

# 重新初始化系统数据：
make init-system
```

## 📋 API 接口

### 公开接口

- `GET /health` - 健康检查
- `GET /api/public-key` - 获取服务端公钥
- `POST /api/admin/login` - 管理员登录

### 客户端接口

- `POST /api/licenses/activate` - 批量激活设备
- `POST /api/licenses/transfer` - 授权转移
- `GET /api/licenses?auth_code=xxx` - 查询设备列表

### 管理员接口（需要JWT认证）

- `GET /api/admin/dashboard` - 管理员控制台
- `POST /api/admin/authorizations` - 创建授权码
- `GET /api/admin/authorizations` - 授权码列表
- `PUT /api/admin/authorizations/:id` - 更新授权码
- `DELETE /api/admin/authorizations/:id` - 删除授权码
- `DELETE /api/admin/licenses/:id/unbind` - 强制解绑设备
- `GET /api/admin/logs` - 查看操作日志

## 🔐 安全机制

1. **RSA数字签名**: 所有授权文件使用RSA-2048签名
2. **机器绑定**: 授权与硬件唯一标识绑定
3. **一次性密钥**: 解绑使用一次性密钥机制
4. **会话管理**: JWT令牌 + 超时控制
5. **操作日志**: 完整的管理员操作审计

## 📖 使用流程

### 1. 管理员创建授权码

1. 登录管理员控制台
2. 创建新的授权码，设置客户名称、席位数、有效期等
3. 将授权码提供给客户

### 2. 客户激活设备

1. 客户在需要激活的设备上生成 `.bind` 文件
2. 使用授权码登录客户控制台
3. 上传 `.bind` 文件进行批量激活
4. 下载生成的 `.license` 文件到对应设备

### 3. 授权转移

1. 在旧设备上生成 `.unbind` 解绑文件
2. 在新设备上生成 `.bind` 绑定文件
3. 在控制台上传两个文件完成转移
4. 下载新的 `.license` 文件到新设备

## 🗃️ 数据库设计

主要数据表：

- `authorizations` - 授权码管理
- `licenses` - 已激活设备
- `admin_users` - 管理员账户
- `admin_logs` - 操作日志
- `rsa_keys` - RSA密钥管理
- `system_config` - 系统配置

## 🧪 测试和质量保证

### 测试套件

项目包含完整的测试套件，确保代码质量和功能稳定性：

```bash
# 运行完整测试套件
make test

# 查看测试覆盖率  
go test ./tests/... -cover -coverpkg=./internal/... -v
```

### 测试类型

1. **单元测试**
   - License Service 测试：设备激活、席位管理、数据验证
   - RSA Service 测试：密钥生成、签名验证、加密解密  
   - Authorization Service 测试：授权码CRUD、业务逻辑

2. **集成测试**
   - 完整业务流程测试（管理员登录 → 创建授权码 → 设备激活 → 设备列表查询）
   - API接口端到端测试
   - 错误处理和边界条件测试

3. **性能测试**
   - 并发设备激活基准测试
   - RSA签名验证性能测试
   - 数据库操作性能测试

### 质量指标

- ✅ 所有核心功能单元测试通过
- ✅ 集成测试覆盖主要业务流程  
- ✅ 错误处理和异常情况测试
- ✅ 并发安全性验证
- ✅ 内存数据库测试隔离

## ✅ 开发状态

**核心功能已完成并可投入生产使用**

已完成的核心功能：
- [x] 基础框架搭建（Gin + GORM）
- [x] 数据模型设计（6个核心表）
- [x] RSA加密服务（密钥生成、数字签名、验证）
- [x] JWT认证中间件（令牌生成、验证、刷新）
- [x] 授权码管理服务（CRUD、席位管理、统计）
- [x] 设备激活服务（批量激活、授权转移、强制解绑）
- [x] 管理员认证服务（登录、会话管理、操作日志）
- [x] 完整REST API接口（客户端 + 管理员）
- [x] 测试客户端（机器ID生成、绑定文件生成）
- [x] 系统初始化工具（默认管理员、RSA密钥）
- [x] 完整单元测试（License Service、RSA Service等）
- [x] 完整集成测试（端到端业务流程测试）
- [x] 性能基准测试（并发激活测试）

**测试覆盖情况：**
- ✅ License Service 单元测试（设备激活、席位管理）
- ✅ RSA Service 单元测试（密钥生成、签名验证）
- ✅ Authorization Service 单元测试（授权码管理）
- ✅ 集成测试（完整的业务流程）
- ✅ 性能基准测试（并发场景）

**可选增强功能：**
- [ ] Web管理界面（React/Vue前端）
- [ ] 客户端SDK库（Java/C#/.NET）
- [ ] Docker部署配置
- [ ] Swagger API文档
- [ ] 监控和告警系统

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 📞 联系方式

如有问题或建议，请通过 GitHub Issues 联系我们。

## 📝 日志分层错误处理系统

### 概述

LicenseCenter 实现了完整的分层错误处理和日志记录系统，针对 Go 语言缺乏全局异常处理的特点，通过统一中间件和错误分类实现了类似其他语言的错误处理机制。

### 架构设计

#### 错误分类体系

```go
// 业务逻辑错误 (40xxx) - 记录到 app.log
ErrDuplicateMachine  = NewAppError(40016, "设备已被激活")
ErrInsufficientSeats = NewAppError(40015, "可用席位不足")
ErrAuthCodeDisabled  = NewAppError(40011, "授权码已被禁用")

// 系统错误 (50xxx) - 记录到 app.log + app_error.log
ErrCryptoError = NewAppError(50001, "加密操作失败")
```

#### 日志文件分离

- **`logs/app.log`**: 记录所有日志（INFO、WARN、ERROR）
- **`logs/app_error.log`**: 仅记录严重的系统错误（ERROR级别）

### 核心组件

#### 1. 统一错误处理中间件

**文件**: `internal/middleware/error_handler.go`

```go
// 统一错误处理入口点
func ErrorHandlerMiddleware() gin.HandlerFunc
func ErrorResponseHandler() gin.HandlerFunc

// 根据错误类型自动分类处理
func handleError(c *gin.Context, err error)
```

**功能特性**:
- 自动识别业务错误 vs 系统错误
- 根据错误代码自动分配日志级别
- 提取完整的请求上下文信息
- 处理 panic 恢复并记录堆栈

#### 2. 增强型日志系统

**文件**: `pkg/logger/logger.go`

```go
// 支持多文件输出的日志配置
type LogConfig struct {
    Level       string
    AppLogFile  string      // 应用日志文件
    ErrorLogFile string     // 错误日志文件
}

// 双日志器实例
var (
    Logger      *zap.Logger  // 应用日志器
    ErrorLogger *zap.Logger  // 错误日志器
)
```

**功能特性**:
- 自动生成错误日志文件路径（`app.log` → `app_error.log`）
- 控制台 + 文件双重输出
- JSON 格式化便于日志分析
- 结构化字段记录

#### 3. 前端错误处理优化

**文件**: `web/src/api/request.js`、`web/src/api/client.js`

```javascript
// 自动处理 blob 响应中的错误信息
if (data instanceof Blob && status >= 400) {
  // 将 blob 转换为 JSON 提取错误
}

// API 层错误包装
.catch(error => {
  if (error.response && error.response.status >= 400) {
    const businessError = new Error(data.error)
    businessError.code = data.code
    throw businessError
  }
})
```

### 日志级别映射

| 错误代码范围 | 错误类型 | 日志级别 | 记录位置 | 示例 |
|-------------|----------|----------|----------|------|
| 40000-40999 | 客户端错误 | WARN | app.log | 参数验证失败 |
| 41000-41999 | 业务逻辑错误 | WARN | app.log | 设备已被激活 |
| 43000-43999 | 资源不存在 | INFO | app.log | 授权码不存在 |
| 50000+ | 系统错误 | ERROR | app.log + app_error.log | 数据库连接失败 |

### 使用方法

#### 1. 处理器中的错误处理

**简化前**（繁琐的错误处理）:
```go
if err != nil {
    if appErr, ok := err.(*errors.AppError); ok {
        logger.GetLogger().Error("设备激活失败", /* 大量重复代码 */)
        c.JSON(appErr.HTTPStatus(), gin.H{"error": appErr.Message})
    } else {
        logger.GetLogger().Error("系统错误", /* 大量重复代码 */)
        c.JSON(500, gin.H{"error": "内部错误"})
    }
    return
}
```

**简化后**（统一处理）:
```go
if err != nil {
    c.Error(err)  // 让中间件统一处理
    return
}
```

#### 2. 业务层中记录关键错误

```go
// 在关键业务逻辑中添加详细日志
if err == nil {
    logger.GetLogger().Warn("设备重复激活被阻止",
        zap.String("auth_code", auth.AuthorizationCode),
        zap.String("machine_id", bindFile.MachineID),
        zap.String("hostname", bindFile.Hostname),
        zap.Uint("existing_license_id", existing.ID),
        zap.String("customer_name", auth.CustomerName),
    )
    return nil, errors.ErrDuplicateMachine
}
```

### 日志样例

#### 业务错误日志 (app.log)

```json
{
  "level": "WARN",
  "timestamp": "2025-06-30T18:30:00.000+0800",
  "caller": "middleware/error_handler.go:95",
  "msg": "业务警告",
  "error_type": "business",
  "error_code": 40016,
  "error_message": "设备已被激活",
  "path": "/api/actions/activate-licenses",
  "method": "POST",
  "client_ip": "::1",
  "username": "ABC-123-TEST"
}
```

#### 系统错误日志 (app_error.log)

```json
{
  "level": "ERROR",
  "timestamp": "2025-06-30T18:30:00.000+0800", 
  "caller": "middleware/error_handler.go:128",
  "msg": "系统错误",
  "error_type": "system",
  "error": "database connection failed",
  "path": "/api/actions/activate-licenses",
  "method": "POST",
  "client_ip": "::1",
  "stack": "goroutine 1 [running]:\n..."
}
```

### 前端错误显示优化

**优化前**: 显示通用错误
```
❌ 设备激活失败
```

**优化后**: 显示具体错误
```
❌ 设备已被激活
❌ 可用席位不足  
❌ 授权码已被禁用
```

### 配置选项

在 `configs/app.yaml` 中可以配置日志行为：

```yaml
logging:
  level: "debug"                # 日志级别
  file: "./logs/app.log"        # 应用日志文件
  enable_http_log: true         # 是否记录HTTP请求
```

错误日志文件自动生成为 `./logs/app_error.log`。

### 最佳实践

1. **处理器层**: 使用 `c.Error(err)` 统一交给中间件处理
2. **服务层**: 在关键业务节点记录详细的上下文日志
3. **前端**: 优先显示具体的错误消息，降级显示通用消息
4. **监控**: 重点关注 `app_error.log` 中的系统错误
5. **调试**: 查看 `app.log` 了解完整的业务流程

### 技术优势

相比传统的 Go 错误处理，该系统具有以下优势：

1. **统一处理**: 类似其他语言的全局异常处理器
2. **自动分类**: 根据错误类型自动选择日志级别和文件
3. **上下文丰富**: 自动记录请求信息、用户信息等
4. **前端友好**: 确保具体错误信息能正确传递到用户界面
5. **运维友好**: 严重错误单独记录，便于监控和告警
6. **开发友好**: 大大简化了处理器中的错误处理代码

## 安全特性

### 双因子认证 (TOTP)

系统支持基于时间的一次性密码 (TOTP) 双因子认证，兼容 Google Authenticator、Microsoft Authenticator 等主流认证应用。

#### 配置选项

在 `configs/app.yaml` 中可以配置：

```yaml
security:
  force_totp: true  # 强制启用双因子认证
```

#### 强制双因子认证

当 `force_totp` 设置为 `true` 时：

1. **新管理员账户**：创建时自动生成 TOTP 密钥
2. **登录验证**：必须提供双因子认证码才能登录
3. **禁用限制**：无法禁用已启用的双因子认证
4. **账户保护**：未设置双因子认证的账户无法登录

#### 使用流程

1. 管理员登录后，扫描二维码将账户添加到认证应用
2. 每次登录时输入6位认证码
3. 系统验证认证码后允许访问

#### API 接口

- `POST /admin/totp/enable` - 启用双因子认证
- `POST /admin/totp/disable` - 禁用双因子认证（强制模式下不可用）
- `GET /admin/totp/info/:id` - 获取双因子认证设置信息 