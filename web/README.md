# OCR2Doc 离线授权管理系统 - Web界面

基于 Vue.js 3 + Element Plus 构建的现代化管理界面，为 OCR2Doc 离线授权系统提供完整的Web管理功能。

## 🚀 技术栈

- **前端框架**: Vue.js 3
- **UI组件库**: Element Plus
- **路由管理**: Vue Router 4
- **状态管理**: Pinia
- **HTTP客户端**: Axios
- **构建工具**: Vite
- **开发语言**: JavaScript

## 📁 项目结构

```
web/
├── public/                 # 静态资源
├── src/
│   ├── api/               # API接口封装
│   │   ├── request.js     # Axios请求配置
│   │   ├── auth.js        # 认证相关接口
│   │   ├── client.js      # 客户端接口
│   │   └── admin.js       # 管理员接口
│   ├── components/        # 公共组件
│   ├── router/            # 路由配置
│   │   └── index.js       # 路由定义和守卫
│   ├── stores/            # 状态管理
│   │   └── auth.js        # 认证状态store
│   ├── views/             # 页面组件
│   │   ├── client/        # 客户端页面
│   │   │   ├── Login.vue  # 客户端登录
│   │   │   └── Dashboard.vue # 客户端控制台
│   │   ├── admin/         # 管理员页面
│   │   │   ├── Login.vue  # 管理员登录
│   │   │   ├── Dashboard.vue # 管理员控制台
│   │   │   ├── AuthorizationManagement.vue # 授权码管理
│   │   │   ├── CustomerManagement.vue # 客户管理
│   │   │   └── SystemSettings.vue # 系统设置
│   │   └── NotFound.vue   # 404页面
│   ├── App.vue            # 根组件
│   └── main.js            # 应用入口
├── index.html             # HTML模板
├── package.json           # 项目依赖配置
├── vite.config.js         # Vite构建配置
└── README.md              # 项目说明
```

## 🛠️ 安装和运行

### 1. 安装依赖

```bash
cd web
npm install
```

### 2. 开发环境运行

```bash
npm run dev
```

访问 `http://localhost:3000` 查看应用

### 3. 生产环境构建

```bash
npm run build
```

构建产物将生成在 `dist/` 目录中

### 4. 预览生产构建

```bash
npm run preview
```

## 🎯 主要功能

### 客户端功能
- **授权码登录**: 支持人机验证的安全登录
- **设备管理**: 查看已激活设备列表
- **批量激活**: 上传多个 `.bind` 文件批量激活设备
- **授权转移**: 通过上传 `.unbind` 和 `.bind` 文件完成授权转移
- **文件下载**: 下载 `.license` 授权文件

### 管理员功能
- **系统控制台**: 查看系统统计和最近活动
- **授权码管理**: 创建、查询、修改、禁用授权码
- **客户管理**: 查看客户详情、强制解绑设备
- **系统设置**: 密钥管理、系统日志、数据备份

## 🔧 配置说明

### API代理配置

开发环境下，Vite会自动将 `/api` 请求代理到后端服务（默认 `http://localhost:8080`）。

如需修改后端地址，请编辑 `vite.config.js`:

```javascript
server: {
  proxy: {
    '/api': {
      target: 'http://your-backend-url:port',
      changeOrigin: true,
    }
  }
}
```

### 环境变量

项目支持以下环境变量：

- `VITE_API_BASE_URL`: API基础URL（可选，默认使用代理）
- `VITE_APP_TITLE`: 应用标题

## 🎨 界面预览

### 客户端界面
- **登录页面**: 简洁的授权码登录界面，集成人机验证
- **控制台**: 直观的席位使用情况和设备管理界面
- **文件上传**: 支持拖拽上传的现代化文件选择器

### 管理员界面
- **登录页面**: 支持双因子认证的安全登录
- **控制台**: 数据统计卡片和功能导航的仪表板设计
- **侧边栏导航**: 清晰的功能模块导航

## 🔒 安全特性

- **JWT令牌认证**: 安全的会话管理
- **路由权限控制**: 基于角色的页面访问控制
- **请求拦截**: 自动处理认证头和错误响应
- **登录状态持久化**: 本地存储用户认证信息

## 📱 响应式设计

- 支持桌面端和移动端访问
- 自适应布局设计
- 优化的触屏交互体验

## 🚀 部署建议

### 生产环境部署

1. 构建生产版本:
```bash
npm run build
```

2. 将 `dist/` 目录部署到Web服务器

3. 配置Nginx反向代理（推荐）:
```nginx
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        root /path/to/dist;
        try_files $uri $uri/ /index.html;
    }
    
    location /api {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Docker部署

创建 `Dockerfile`:
```dockerfile
FROM nginx:alpine
COPY dist/ /usr/share/nginx/html/
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## 🤝 开发规范

- 使用 Vue 3 Composition API
- 遵循 Element Plus 设计规范
- API调用统一使用 async/await
- 错误处理统一使用 Element Plus 消息提示
- 路由跳转前进行权限验证

## 📞 技术支持

如有问题或建议，请查看项目文档或提交Issue。 