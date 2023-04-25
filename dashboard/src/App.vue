<script setup>
import { ref } from 'vue'
import router from './router'

const activeIndex = ref(window.location.pathname)
const handleSelect = (key) => {
	if (key === '/statmq') {
		let a = document.createElement('a')
		a.target = '_blank'
		a.href = 'https://' + window.location.hostname + ':15671'
		a.click()
		return false
	}
	router.push(key)
}

</script>
<template>
  <div v-if="!isAuth" class="common-layout"
       style="display: flex; justify-content: center; width: 100%; padding-top: 150px;">
    <el-input v-model="pass" type="password" placeholder="password..."
              size="large" style="width: 200px;" class="w-50 m-2" ref="login" @keydown.enter="login" />
    <el-button type="primary" size="large" plain style="margin-left: 10px;" @click.prevent="login">login</el-button>
  </div>

  <div v-else class="common-layout">
    <el-menu mode="horizontal"
             :ellipsis="false"
             :default-active="activeIndex"
             @select="handleSelect">
      <el-menu-item index="logo" class="revert" disabled>
        <a title="MoMo" href="/"><img title="Mo Mo" src="/logo.png" style="width: 50px; margin-top: 4px;" alt="Mo Mo"></a>
      </el-menu-item>
      <el-menu-item index="/">Realtime logs</el-menu-item>
      <el-menu-item index="/stats">Stats</el-menu-item>
      <el-menu-item index="/alert">Alert list</el-menu-item>
      <el-menu-item index="/setting">Setting</el-menu-item>
      <div class="flex-grow" />
      <el-menu-item index="/statmq" class="hover-disable">
        <apexchart type="line" height="56" width="250"
                   :options="chartOptions"
                   :series="series" />
      </el-menu-item>
      <el-menu-item index="/statmq" class="hover-disable">
        <div class="statmq">
          <div>Delivery/s RabbitMQ</div>
          <div style="color: #409eff;">{{ deliverySec }}/s</div>
        </div>
      </el-menu-item>
      <el-menu-item index="/statmq" class="hover-disable">
        <div class="statmq">
          <div>Transactions RabbitMQ</div>
          <div style="color: #409eff;">{{ transactionsRabbitMQ }}</div>
        </div>
      </el-menu-item>
    </el-menu>

    <el-container style="margin-top: 20px;">
      <router-view />
    </el-container>
  </div>
</template>

<script>
import { ElMessage } from 'element-plus'
import _ from 'lodash'
import md5 from 'crypto-js/md5'
import { colord } from 'colord'
import stc from 'string-to-color'

export default {
	data () {
		return {
			isAuth: window.localStorage.getItem('is-auth') === 'yes',
			pass: '',
			ws: new WebSocket('ws' + (window.location.protocol !== 'http:' ? 's' : '') + '://'
				+ window.location.hostname + ':8844/ws'),
			deliverySec: 0,
			transactionsRabbitMQ: 0,
			containersMenu: [],
			series: [{ data: [{ x: new Date().getTime(), y: 0 }] }],
			chartOptions: {
				chart: {
					id: 'realtime',
					height: 56,
					type: 'line',
					animations: {
						enabled: true,
						easing: 'linear',
						dynamicAnimation: {
							speed: 500
						}
					},
					dropShadow: {
						enabled: true,
						top: 9,
						left: 2,
						blur: 5,
						opacity: 0.10
					},
					toolbar: {
						show: false
					},
					zoom: {
						enabled: false
					},
					sparkline: {
						enabled: false
					},
					parentHeightOffset: 0,
				},
				fill: {
					type: 'gradient',
					gradient: {
						type: 'vertical',
						shadeIntensity: 1,
						opacityFrom: 0.7,
						opacityTo: 0.9,
						colorStops: [
							{
								offset: 0,
								color: '#eb656f',
								opacity: 1
							},
							{
								offset: 50,
								color: '#fad375',
								opacity: 1
							},
							{
								offset: 100,
								color: '#95da74',
								opacity: 1
							}
						]
					}
				},
				dataLabels: {
					enabled: false
				},
				stroke: {
					width: 3,
					curve: 'smooth'
				},
				markers: {
					size: 0
				},
				grid: {
					show: false,
					padding: {
						top: -28,
						bottom: -14,
						left: 0,
						right: 0,
					},
					borderColor: 'whitesmoke',
					xaxis: {
						lines: {
							show: false
						}
					},
					yaxis: {
						lines: {
							show: false
						}
					},
				},
				yaxis: {
					opposite: true,
					labels: {
						show: false,
						formatter: (val) => { return _.round(val, 2) },
					},
					axisBorder: {
						show: false
					},
					axisTicks: {
						show: false
					}
				},
				xaxis: {
					type: 'datetime',
					range: 20000,
					show: false,
					labels: {
						show: false
					},
					axisBorder: {
						show: false
					},
					axisTicks: {
						show: false
					},
					tooltip: {
						enabled: false,
					},
					crosshairs: {
						show: false
					}
				},
				tooltip: {
					fixed: {
						enabled: false
					},
					x: {
						show: false
					},
					y: {
						show: false
					},
					marker: {
						show: false
					},
					custom: function ({ series, seriesIndex, dataPointIndex }) {
						return '<b style="padding: 5px;">' + series[seriesIndex][dataPointIndex] + ' ops</b>'
					}
				},
			},
		}
	},
	methods: {
		login: function () {
			this.ws.send('pass-' + this.pass)
		},
		purifyMenu () {
			for (let i in this.containersMenu) {
				_.remove(this.containersMenu[i].containers, (v) => {
					if ((new Date).getTime() - v.lastUpdate > 10000) {
						return true
					}
				})
			}
		}
	},
	mounted () {
		window.ws = this.ws
		this.ws.onclose = () => {
			ElMessage({
				showClose: false,
				grouping: true,
				dangerouslyUseHTMLString: true,
				message: 'websocket connection is closed, <a href="#" onclick="window.location.href=\'/\'">reload page</a>',
				type: 'error',
				duration: 0
			})
		}
		this.ws.onmessage = (evt) => {
			let jp = JSON.parse(evt.data)

			if (jp.typeMess === 'auth') {
				if (jp.data === 'fail') {
					ElMessage({
						showClose: true,
						grouping: true,
						message: this.isAuth ? 'session is outdated ' : 'oops, password is wrong',
						type: 'error',
					})

					this.$store.state.isAuth = false
					this.isAuth = false
					window.localStorage.setItem('is-auth', 'no')
				} else {
					// if used domain and without port so can set HttpOnly=true;
					document.cookie = 'token=' + jp.data + '; path=/; SameSite=Strict; Secure=true;' +
						'expires=' + new Date(new Date().getTime() + 365 * 24 * 3600 * 1000).toUTCString()

					window.localStorage.setItem('is-auth', 'yes')
					this.$store.state.isAuth = true

					window.location.reload()
				}
			}

			if (jp.typeMess === 'statistic') {
				this.transactionsRabbitMQ = jp.data.message_stats.deliver_get.toString().replace(/\B(?=(\d{3})+(?!\d))/g, '.')
				this.deliverySec = jp.data.message_stats.deliver_get_details.rate

				this.series[0].data.push({ x: new Date().getTime(), y: this.deliverySec })
				if (this.series[0].data.length > 1000) {
					this.series[0].data = this.series[0].data.slice(500)
				}
			}

			if (jp.typeMess === 'container') {
				let nameContainer = jp.data.Names[0].slice(1)

				let second = {
					id: jp.data.Id,
					name: nameContainer,
					md5Name: md5(jp.data.Hostname + jp.data.Names[0]).toString(),
					color: colord(stc(jp.data.Hostname + nameContainer)),
					running: jp.data.State === 'running',
					status: jp.data.Status,
					lastUpdate: (new Date).getTime()
				}

				this.$store.state.containersColor[second.md5Name] = second.color

				let mainIndex = this.containersMenu.findIndex(key => key.hostname === jp.data.Hostname)
				if (this.containersMenu[mainIndex]) {
					let i = this.containersMenu[mainIndex]['containers'].findIndex(key => key.md5Name === jp.data.Md5Name)
					if (this.containersMenu[mainIndex]['containers'][i]) {
						this.containersMenu[mainIndex]['containers'][i] = second
					} else {
						this.containersMenu[mainIndex]['containers'].push(second)
					}
					this.containersMenu[mainIndex]['containers'].sort((a, b) => a.name > b.name)

				} else {
					this.containersMenu.push({ hostname: jp.data.Hostname, containers: [second] })
				}
				this.containersMenu.sort((a, b) => a.hostname > b.hostname)

				this.$store.state.containersMenu = this.containersMenu
			}
		}
		this.ws.onerror = (evt) => {
			console.log(evt)
		}

		if (this.isAuth) {
			setTimeout(() => {
				window.ws.send('containers')
			}, 100)
			setInterval(() => {
				window.ws.send('containers')
				this.purifyMenu()
			}, 2000)
		}
	}
}
</script>

<style>
.flex-grow {
	flex-grow: 1;
}

.el-menu-item.is-disabled.revert {
	opacity: 1;
	cursor: default;
}

.statmq {
	line-height: 20px;
	text-align: center;
	margin-top: 10px;
}

.hover-disable:hover {
	background-color: white !important;
}

.is-active.hover-disable {
	border-bottom: none !important;
	background: none !important;
}
</style>