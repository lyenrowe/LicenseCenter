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