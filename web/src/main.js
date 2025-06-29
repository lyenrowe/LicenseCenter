import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'

import App from './App.vue'
import router from './router'

console.log('开始初始化Vue应用...')

try {
  const app = createApp(App)

  // 注册Element Plus图标
  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }

  console.log('注册插件...')
  app.use(createPinia())
  app.use(router)
  app.use(ElementPlus, {
    locale: zhCn,
  })

  // 全局错误处理
  app.config.errorHandler = (err, instance, info) => {
    console.error('Vue应用错误:', err)
    console.error('错误信息:', info)
  }

  console.log('挂载应用到DOM...')
  app.mount('#app')
  
  console.log('Vue应用初始化完成！')
} catch (error) {
  console.error('Vue应用初始化失败:', error)
} 