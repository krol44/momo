import { createApp } from 'vue'
import app from './App.vue'
import router from './router'
import elementPlus from 'element-plus'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'element-plus/theme-chalk/el-message.css'
import 'element-plus/theme-chalk/el-badge.css'
import './assets/main.css'

createApp(app).use(router).use(elementPlus).mount('#app')