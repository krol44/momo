import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
	history: createWebHistory(import.meta.env.BASE_URL),
	routes: [
		{
			path: '/',
			name: 'home',
			component: () => import('../views/HomeView.vue')
		},
		{
			path: '/stats',
			name: 'stats',
			component: () => import('../views/StatsView.vue')
		},
		{
			path: '/alert',
			name: 'alert',
			component: () => import('../views/AlertView.vue')
		},
		{
			path: '/setting',
			name: 'setting',
			component: () => import('../views/SettingView.vue')
		}
	]
})

export default router
