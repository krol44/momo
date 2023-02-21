<!--<script setup>-->
<!--  import { RouterLink, RouterView } from 'vue-router'-->
<!--</script>-->
<template>
  <div v-if="!isAuth" class="common-layout"
       style="display: flex; justify-content: center; width: 100%; padding-top: 120px;">
    <el-input v-model="pass" type="password" placeholder="password..."
              size="large" style="width: 200px;" class="w-50 m-2" ref="login" @keydown.enter="login" />
    <el-button type="primary" size="large" plain style="margin-left: 10px;" @click.prevent="login">login</el-button>
  </div>

  <div v-else class="common-layout">
    <el-container>
      <el-header>
        <el-row justify="space-between">
          <el-col :span="1">
            <img title="Mo mo" src="/logo.png" style="width: 60px;" alt="Mo mo">
          </el-col>
          <el-col :span="6">
            <div style="display: flex; height: 60px; flex-direction: column; justify-content: center;">
              <el-input v-if="installUrl" v-model="installUrl">
                <template #append><a href="#" @click="copyInstallUrl" style="text-decoration: none;">copy</a></template>
              </el-input>
              <a v-else href="#" @click="getInstallUrl">install grabber on the new machine</a>
            </div>
          </el-col>
          <el-col :span="10">
            <!--            <el-row>-->
            <!--              <el-col :span="6">-->
            <!--                <el-statistic title="Daily active users" :value="268500" />-->
            <!--              </el-col>-->
            <!--              <el-col :span="6">-->
            <!--                <el-statistic :value="138">-->
            <!--                  <template #title>-->
            <!--                    <div style="display: inline-flex; align-items: center">-->
            <!--                      Ratio of men to women-->
            <!--                      <el-icon style="margin-left: 4px" :size="12">-->
            <!--                        <Male />-->
            <!--                      </el-icon>-->
            <!--                    </div>-->
            <!--                  </template>-->
            <!--                  <template #suffix>/100</template>-->
            <!--                </el-statistic>-->
            <!--              </el-col>-->
            <!--              <el-col :span="6">-->
            <!--                <el-statistic title="Total Transactions" :value="172000" />-->
            <!--              </el-col>-->
            <!--              <el-col :span="6">-->
            <!--                <el-statistic title="Feedback number" :value="562">-->
            <!--                  <template #suffix>-->
            <!--                    <el-icon style="vertical-align: -0.125em">-->
            <!--                      <ChatLineRound />-->
            <!--                    </el-icon>-->
            <!--                  </template>-->
            <!--                </el-statistic>-->
            <!--              </el-col>-->
            <!--            </el-row>-->
          </el-col>
        </el-row>
      </el-header>
      <el-container>
        <el-aside width="250px" :style="{height: heightWindow-100+'px', 'overflow-y': 'scroll',
         'padding-bottom': '10px', 'padding-left': '10px'}">
          <div v-if="Object.keys(containersMenu).length" v-for="(val, name) in containersMenu"
               v-bind:key="name" style="white-space: nowrap;">
            <el-divider content-position="left">
              <el-switch v-model="this.switchMain[name]" @click="subContainers(val, name)" :active-text="name" />
            </el-divider>
            <div style="margin-left: 10px;" v-for="cont in val" v-bind:key="cont.id">
              <el-switch v-model="this.switch[cont.id]" size="small"
                         :active-text="cont.name" :disabled="!cont.running"
                         :active-color="cont.color.toHex()"
                         @click="subContainer(cont.id)" />
              <sub style="position: relative; bottom: 7px;
               left: 2px; font-size: 10px;">{{ cont.running ? cont.status : 'not working' }}</sub>
            </div>
          </div>
          <el-skeleton v-else animated :rows="7" style="padding-top: 20px;" />
        </el-aside>
        <el-main>

          <div :style="{height: heightWindow-100+'px', 'overflow-y': 'scroll'}" @scroll="scroll">
            <div ref="logScroll">
              <div v-if="Object.keys(logsData).length" class="row" v-for="(val, index) in logsData" :key="index"
                   :style="{'background-color':this.containers[val.data.container_id].color.alpha(0.2).toHex()}">
                <div class="row-text">{{ val.timeCreate }} - {{ val.containerName }} - {{ val.data.log_line }}</div>
              </div>
              <div v-else class="row" style="background-color: #f5f6fa;">
                <div class="row-text" style="text-align: center;">no data</div>
              </div>
            </div>
          </div>
          <div style="display: flex; justify-content: center; width: 100%;">
            <el-button class="btn-down" type="warning" v-if="this.stopChat" @click="scrollDown" plain>
              scrolling is stopped, new messages bellow ↓
            </el-button>
          </div>
        </el-main>
      </el-container>
    </el-container>
  </div>
  <!--  <RouterView />-->
</template>

<script>
import { colord } from 'colord'
import stc from 'string-to-color'
import { ElMessage } from 'element-plus'

export default {
	data () {
		return {
			isAuth: false,
			pass: '',
			ws: new WebSocket('ws' + (window.location.protocol !== 'http:' ? 's' : '') + '://'
				+ window.location.hostname + ':8844/ws'),
			containersMenu: {},
			logsData: [],
			containers: {},
			switch: {},
			switchMain: [],
			heightWindow: document.documentElement.clientHeight,
			stopChat: false,
			installUrl: ''
		}
	},

	methods: {
		login: function () {
			this.ws.send('pass-' + this.pass)
		},
		getInstallUrl: function () {
			this.ws.send('get-install-url')
		},
		copyInstallUrl: function () {
			navigator.clipboard.writeText(this.installUrl).then(_ => {
				ElMessage({
					showClose: false,
					grouping: true,
					message: 'copied',
					type: 'success',
					duration: 1000,
				})
			})
		},
		scrollDown: function () {
			this.$refs.logScroll.scrollIntoView({ behavior: 'smooth', block: 'end' })
		},
		scroll: function () {
			let pos = this.$refs.logScroll.offsetHeight - this.heightWindow
			if (pos - 300 > this.$refs.logScroll.parentElement.scrollTop) {
				this.stopChat = true
			}
			if (pos - 100 < this.$refs.logScroll.parentElement.scrollTop) {
				this.stopChat = false
			}
		},
		resizeWindow: function () {
			this.heightWindow = document.documentElement.clientHeight
		},
		logScrollDown: function () {
			if (this.stopChat === true) {
				return true
			}
			this.$nextTick(() => {
				this.$refs.logScroll.scrollIntoView({ behavior: 'smooth', block: 'end' })
			})
		},
		subContainer: function (contId) {
			if (this.switch[contId]) {
				console.log('sub-log-' + contId)
				this.ws.send('sub-log-' + contId)
			} else {
				console.log('unsub-log-' + contId)
				this.ws.send('unsub-log-' + contId)
			}
		},
		subContainers: function (objs, mainName) {
			for (let val of objs) {
				if (!val.running) {
					continue
				}

				this.switch[val.id] = this.switchMain[mainName]

				if (this.switchMain[mainName]) {
					console.log('sub-log-' + val.id)
					this.ws.send('sub-log-' + val.id)
				} else {
					console.log('unsub-log-' + val.id)
					this.ws.send('unsub-log-' + val.id)
				}
			}
		}
	},

	created () {
		this.isAuth = window.localStorage.getItem('isAuth') === 'yes'
		window.addEventListener('resize', this.resizeWindow)
	},
	unmounted () {
		window.removeEventListener('resize', this.resizeWindow)
	},

	mounted () {
		this.$refs.login?.focus()

		this.ws.onclose = () => {
			ElMessage({
				showClose: false,
				grouping: true,
				dangerouslyUseHTMLString: true,
				message: 'websocket connection is closed, <a href="#" onclick="window.location.reload()">reload page</a>',
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
						message: 'oops, password is wrong',
						type: 'error',
					})

					this.isAuth = false
					window.localStorage.setItem('isAuth', 'no')
				} else {
					// if used domain and without port so can set HttpOnly=true;
					document.cookie = 'token=' + jp.data + '; path=/; SameSite=Strict; Secure=true;' +
						'expires=' + new Date(new Date().getTime() + 365 * 24 * 3600 * 1000).toUTCString()

					window.localStorage.setItem('isAuth', 'yes')
					this.isAuth = true

					window.location.reload()
				}
			}

			if (jp.typeMess === 'install-url') {
				this.installUrl = jp.data
			}

			if (jp.typeMess === 'container') {
				let nameContainer = jp.data.Names[0].slice(1)

				let second = {
					id: jp.data.Id,
					name: nameContainer,
					color: colord(stc(jp.data.Hostname + nameContainer)),
					running: jp.data.State === 'running',
					status: jp.data.Status
				}

				if (this.containersMenu[jp.data.Hostname]) {
					let index = this.containersMenu[jp.data.Hostname].findIndex(key => key.id === jp.data.Id)
					if (this.containersMenu[jp.data.Hostname][index]) {
						this.containersMenu[jp.data.Hostname][index].running = second.running
					} else {
						this.containersMenu[jp.data.Hostname].push(second)
					}
					this.containersMenu[jp.data.Hostname].sort((a, b) => {
						return a.name > b.name
					})
				} else {
					this.containersMenu[jp.data.Hostname] = [second]
				}

				this.containers[jp.data.Id] = second
			}

			if (jp.type_mess === 'log') {
				this.logsData.push(
					{
						containerName: this.containers[jp.data.container_id].name,
						timeCreate: new Date().toString().split(' ')[4],
						data: jp.data,
					}
				)

				let cutNub = 1000
				if (this.logsData.length >= cutNub) {
					this.logsData = this.logsData.slice(cutNub / 4, cutNub)
				}

				this.logScrollDown()
			}
		}
		this.ws.onerror = (evt) => {
			console.log(evt.data)
		}

		setTimeout(() => {
			if (this.isAuth) {
				this.ws.send('get-containers')
			}
		}, 500)
	}
}
</script>

<style>
.row:first-child {
	border-top: 1px solid #ebeef5;
}

.row {
	font-size: 14px;
	color: #606266;
	min-height: 40px;
	border-bottom: 1px solid #ebeef5;
	background-color: rgba(244, 221, 159, 0.2);
	transition: background-color .25s ease;
}

.row:hover {
	background-color: #f5f7fa !important;
}

.row-text {
	text-overflow: ellipsis;
	white-space: normal;
	word-break: break-word;
	line-height: 23px;
	padding: 8px 12px;
}

.btn-down {
	position: absolute;
	width: 400px;
	bottom: 30px;
}

.el-switch__label {
	font-weight: 400;
}

.el-main {
	padding-top: 0;
	padding-bottom: 0;
}
</style>
