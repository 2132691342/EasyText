import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
// EP 官方暗色 CSS 变量（html.dark 选择器，与本项目暗色机制一致）。
// 此前只手写覆盖了少数 --el-* 变量，ElMessage/表格/日期选择等组件在
// 暗色下仍是白底；官方变量表补全其余组件，style.css 中的手写覆盖
// 在级联顺序上仍优先生效。
import 'element-plus/theme-chalk/dark/css-vars.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import './style.css'

const app = createApp(App)
const pinia = createPinia()

// Register all Element Plus icons
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

app.use(pinia)
app.use(ElementPlus, { locale: zhCn })
app.mount('#app')