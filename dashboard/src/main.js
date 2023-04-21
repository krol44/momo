import { createApp } from 'vue'
import app from './App.vue'
import router from './router'
import elementPlus from 'element-plus'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'element-plus/theme-chalk/el-message.css'
import 'element-plus/theme-chalk/el-badge.css'
import './assets/main.css'
import VueApexCharts from 'vue3-apexcharts'

import { createStore } from 'vuex'

const store = createStore({
	state () {
		return {
			containersMenu: [],
			containersColor: {}
		}
	}
})

createApp(app).use(router).use(store).use(elementPlus).use(VueApexCharts).mount('#app')