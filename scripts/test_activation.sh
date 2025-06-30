#!/bin/bash

# 测试激活接口的脚本
echo "🚀 启动授权文件激活测试"

# 设置变量
SERVER_URL="http://localhost:8080"
AUTH_CODE="TEST-AUTH-001"
BIND_FILE="test_data/TEST-PC-01.bind"

# 检查服务器是否运行
echo "📡 检查服务器连接..."
if ! curl -s "$SERVER_URL/api/public-key" > /dev/null; then
    echo "❌ 服务器未运行，请先启动服务器: go run cmd/server/main.go"
    exit 1
fi
echo "✅ 服务器连接正常"

# 生成测试文件
echo "📝 生成测试bind文件..."
go run cmd/test-file-generator/main.go generate-bind

# 检查文件是否存在
if [ ! -f "$BIND_FILE" ]; then
    echo "❌ 未找到bind文件: $BIND_FILE"
    exit 1
fi
echo "✅ 找到bind文件: $BIND_FILE"

# 获取客户端token (需要先有授权码)
echo "🔑 获取客户端认证token..."
TOKEN_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/auth/client/login" \
    -H "Content-Type: application/json" \
    -d "{\"authorization_code\": \"$AUTH_CODE\"}")

if [ $? -ne 0 ]; then
    echo "❌ 获取token失败，请确保授权码 $AUTH_CODE 存在"
    echo "可以通过管理后台创建授权码"
    exit 1
fi

TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
    echo "❌ 无法解析token，响应: $TOKEN_RESPONSE"
    exit 1
fi
echo "✅ 获取到token: ${TOKEN:0:20}..."

# 测试激活
echo "🔄 测试设备激活..."
ACTIVATION_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/actions/activate-licenses" \
    -H "Authorization: Bearer $TOKEN" \
    -F "bind_files=@$BIND_FILE" \
    --write-out "%{http_code}")

HTTP_CODE="${ACTIVATION_RESPONSE: -3}"
RESPONSE_BODY="${ACTIVATION_RESPONSE%???}"

echo "📊 激活结果:"
echo "   HTTP状态码: $HTTP_CODE"

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ 激活成功！"
    echo "📦 收到授权文件包 (licenses.zip)"
else
    echo "❌ 激活失败"
    echo "📄 响应内容: $RESPONSE_BODY"
fi

echo "🏁 测试完成" 