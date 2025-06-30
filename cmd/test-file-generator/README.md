# 授权文件测试生成器

这个工具用于生成各种类型的测试文件，用于验证授权系统的功能。

## 🚀 功能特性

### 文件类型
- **`.bind`** - 设备绑定请求文件（加密版本，可直接用于API）
- **`.license`** - 授权许可文件（加密版本）
- **`.unbind`** - 设备解绑文件（加密版本）
- **`.bind.json/.license.json/.unbind.json`** - 对应的明文版本（用于调试）

### 加密机制
- 🔒 **真实加密**: 自动从运行中的服务器获取公钥，使用混合加密（RSA+AES）
- 🔄 **智能回退**: 如果服务器未运行，自动回退到模拟加密
- ✅ **API兼容**: 生成的加密文件可直接用于 `/api/actions/activate-licenses` 接口

## 📖 使用方法

### 基本命令

```bash
# 生成单个bind文件
go run cmd/test-file-generator/main.go generate-bind

# 生成单个license文件
go run cmd/test-file-generator/main.go generate-license

# 生成单个unbind文件
go run cmd/test-file-generator/main.go generate-unbind

# 生成完整的文件集合
go run cmd/test-file-generator/main.go generate-all

# 显示帮助信息
go run cmd/test-file-generator/main.go help
```

### 测试激活

```bash
# 运行完整的激活测试
./scripts/test_activation.sh
```

## 📁 文件结构

所有生成的文件保存在 `test_data/` 目录下：

```
test_data/
├── TEST-PC-01.bind          # 加密的绑定文件 (用于API)
├── TEST-PC-01.bind.json     # 明文的绑定文件 (用于调试)
├── TEST-PC-01.license       # 加密的授权文件
├── TEST-PC-01.license.json  # 明文的授权文件
├── TEST-PC-01.unbind        # 加密的解绑文件
└── TEST-PC-01.unbind.json   # 明文的解绑文件
```

## 🔧 技术细节

### 加密流程
1. 从 `http://localhost:8080/api/public-key` 获取服务器公钥
2. 使用混合加密（RSA+AES）加密JSON数据
3. 将加密结果进行Base64编码存储

### 数据格式
生成的测试数据包含真实的字段：
- **machine_id**: 32位MD5格式的机器标识
- **hostname**: 模拟的主机名
- **request_time**: 当前时间戳
- **授权期限**: 1年有效期

## 🧪 测试验证

### 前置条件
1. 服务器运行在 `localhost:8080`
2. 数据库中存在测试授权码 `TEST-AUTH-001`
3. 有足够的授权席位

### 验证步骤
```bash
# 1. 启动服务器
go run cmd/server/main.go

# 2. 生成测试文件
go run cmd/test-file-generator/main.go generate-bind

# 3. 运行激活测试
./scripts/test_activation.sh
```

### 预期结果
- ✅ 生成真实加密的bind文件
- ✅ 成功通过API激活验证
- ✅ 获得加密的license文件包

## 🐛 故障排除

### 无法获取公钥
**问题**: 显示 "无法获取服务器公钥，使用模拟加密"
**解决**: 确保服务器运行在 `localhost:8080`

### 激活失败
**问题**: 测试脚本激活失败
**解决**: 
1. 检查授权码 `TEST-AUTH-001` 是否存在
2. 确认席位数量充足
3. 验证服务器日志中的错误信息

### 文件格式错误
**问题**: API返回文件格式错误
**解决**: 确保使用真实加密生成的文件，而非模拟加密

## 📝 注意事项

- 加密文件只能由拥有对应私钥的服务器解密
- 明文文件仅用于调试，不应在生产环境使用
- 生成的测试数据不包含真实的签名验证
- 每次生成都会产生不同的机器ID和时间戳

## 技术实现

- 使用 `crypto/md5` 生成机器ID
- 使用 `crypto/rsa` 生成RSA密钥对
- 使用 `encoding/base64` 进行Base64编码
- 使用 `encoding/json` 进行JSON序列化
- 符合项目现有的文件格式规范 