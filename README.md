# LicenseCenter - OCR2Doc 离线授权管理系统

基于 Golang + Gin + GORM 构建的完整离线授权管理系统，支持授权码管理、设备激活、授权转移等功能。

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
go test ./...

# 运行特定包的测试
go test ./internal/services/...
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
- `activated_licenses` - 已激活设备
- `admin_users` - 管理员账户
- `admin_logs` - 操作日志
- `rsa_keys` - RSA密钥管理
- `system_config` - 系统配置

## 🚧 开发状态

当前已完成：
- [x] 基础框架搭建（Gin + GORM）
- [x] 数据模型设计（6个核心表）
- [x] 加密工具包（RSA-2048签名）
- [x] 测试客户端（机器ID生成、绑定文件）
- [x] JWT认证服务（令牌生成、验证、刷新）
- [x] 授权码服务（CRUD、席位管理、统计）
- [x] 设备激活服务（批量激活、授权转移、强制解绑）
- [x] 管理员服务（TOTP双因素认证、操作日志）
- [x] 完整API接口（客户端 + 管理员）
- [x] 系统初始化工具
- [ ] Web管理界面
- [ ] 完整单元测试覆盖

## 📄 许可证

本项目采用 MIT 许可证。详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request 来改进项目。

## 📞 联系方式

如有问题或建议，请通过 GitHub Issues 联系我们。 