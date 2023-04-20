<script setup>
import { Check, Close } from '@element-plus/icons-vue'
</script>
<template>
  <el-container>
    <el-aside width="250px" :style="{height: heightWindow-85+'px', 'overflow-y': 'scroll',
         'padding-bottom': '10px', 'padding-left': '10px'}">
      <div v-if="Object.keys(containersMenu).length" v-for="val in containersMenu"
           v-bind:key="val.hostname" style="white-space: nowrap;">
        <el-divider content-position="left">
          <el-switch v-model="switchMain[val.hostname]" @click="subContainers(val.containers, val.hostname)"
                     :active-text="val.hostname" />
        </el-divider>
        <div style="margin-left: 10px;" v-for="cont in val.containers"
             v-bind:key="cont.md5Name" :id="cont.md5Name">
          <el-switch v-model="switchData[cont.md5Name]" size="small"
                     :active-text="cont.name" :disabled="!cont.running"
                     :active-color="cont.color.toHex()"
                     @click="subContainer(cont.md5Name)" />
          <sub style="position: relative; bottom: 7px;
               left: 2px; font-size: 10px;">{{ cont.running ? cont.status : 'not working' }}</sub>
        </div>
      </div>
      <el-skeleton v-else animated :rows="7" style="padding-top: 20px;" />
    </el-aside>
    <el-main>
      <div style="display: inline-flex; width: 650px; position: relative;">
        <el-switch v-model="filterSwitch"
                   style="position: absolute; top: 4px; left: 122px; z-index: 10; height: unset;"
                   size="large"
                   inline-prompt
                   :active-icon="Check"
                   :inactive-icon="Close"
                   active-color="var(--el-color-success)" />
        <el-input v-model="filter"
                  class="input-filter"
                  placeholder="RegExp">
          <template #prepend>

            <el-select v-model="filterSelect" :default-first-option="true" style="width: 115px">
              <el-option label="filter mask" value="filter" />
              <el-option label="highlight" value="highlight" />
            </el-select>
          </template>
        </el-input>
      </div>
      <el-switch v-model="beepFlag"
                 style="margin-left: 5px;"
                 inline-prompt
                 active-color="var(--el-color-success)"
                 active-text="Beep on"
                 inactive-text="Beep off" />

      <div :style="{height: heightWindow-140+'px', 'overflow-y': 'scroll'}"
           @wheel="scroll"
           @click="stopChat=true">
        <div ref="logScroll">
          <div v-if="Object.keys(logsData).length" class="row" v-for="(val, index) in logsData" :key="index"
               :style="{'background-color':containers[val.data.md5_name]
                ? containers[val.data.md5_name].color.alpha(0.2).toHex() : ''}">
            <div class="row-text">{{ val.timeCreate }} | <strong>{{ val.data.hostname }}</strong>
              <line v-html="val.data.name"></line>
              <line v-html="val.data.body"></line>
            </div>
          </div>
          <div v-else class="row" style="background-color: #f5f6fa;">
            <div class="row-text" style="text-align: center;">no data</div>
          </div>
        </div>
      </div>
      <div style="display: flex; justify-content: center; width: 100%;">
        <el-button class="btn-down" type="warning" v-if="stopChat" @click="scrollDown" plain>
          scrolling is stopped, new messages bellow ↓
        </el-button>
      </div>
    </el-main>
  </el-container>
</template>

<script>
import { colord } from 'colord'
import stc from 'string-to-color'
import md5 from 'crypto-js/md5'
import _ from 'lodash'

export default {
	data () {
		let filter = window.localStorage.getItem('filter')
		if (!filter) {
			filter = 'panic|fatal|error|warning'
		}
		return {
			containersMenu: [],
			containersAlive: new Map(),
			logsData: [],
			containers: {},
			switchData: {},
			switchMain: {},
			heightWindow: document.documentElement.clientHeight,
			stopChat: false,
			filter: filter,
			filterSwitch: false,
			filterSelect: 'filter',
			beep: new Audio('data:audio/wav;base64,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qJrx4KOQjahZxWbcZgztj2c49nKmkId44S71j0c8eV9yDK6uPRzx5X18eDvjvQ6yKo9ZSS6l//8elePK/Lf//IInrOF/FvDoADYAGBMGb7FtErm5MXMlmPAJQVgWta7Zx2go+8xJ0UiCb8LHHdftWyLJE0QIAIsI+UbXu67dZMjmgDGCGl1H+vpF4NSDckSIkk7Vd+sxEhBQMRU8j/12UIRhzSaUdQ+rQU5kGeFxm+hb1oh6pWWmv3uvmReDl0UnvtapVaIzo1jZbf/pD6ElLqSX+rUmOQNpJFa/r+sa4e/pBlAABoAAAAA3CUgShLdGIxsY7AUABPRrgCABdDuQ5GC7DqPQCgbbJUAoRSUj+NIEig0YfyWUho1VBBBA//uQZB4ABZx5zfMakeAAAAmwAAAAF5F3P0w9GtAAACfAAAAAwLhMDmAYWMgVEG1U0FIGCBgXBXAtfMH10000EEEEEECUBYln03TTTdNBDZopopYvrTTdNa325mImNg3TTPV9q3pmY0xoO6bv3r00y+IDGid/9aaaZTGMuj9mpu9Mpio1dXrr5HERTZSmqU36A3CumzN/9Robv/Xx4v9ijkSRSNLQhAWumap82WRSBUqXStV/YcS+XVLnSS+WLDroqArFkMEsAS+eWmrUzrO0oEmE40RlMZ5+ODIkAyKAGUwZ3mVKmcamcJnMW26MRPgUw6j+LkhyHGVGYjSUUKNpuJUQoOIAyDvEyG8S5yfK6dhZc0Tx1KI/gviKL6qvvFs1+bWtaz58uUNnryq6kt5RzOCkPWlVqVX2a/EEBUdU1KrXLf40GoiiFXK///qpoiDXrOgqDR38JB0bw7SoL+ZB9o1RCkQjQ2CBYZKd/+VJxZRRZlqSkKiws0WFxUyCwsKiMy7hUVFhIaCrNQsKkTIsLivwKKigsj8XYlwt/WKi2N4d//uQRCSAAjURNIHpMZBGYiaQPSYyAAABLAAAAAAAACWAAAAApUF/Mg+0aohSIRobBAsMlO//Kk4soosy1JSFRYWaLC4qZBYWFRGZdwqKiwkNBVmoWFSJkWFxX4FFRQWR+LsS4W/rFRb/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////VEFHAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAU291bmRib3kuZGUAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAMjAwNGh0dHA6Ly93d3cuc291bmRib3kuZGUAAAAAAAAAACU='),
			beepFlag: false,
			restoredSubs: false
		}
	},
	watch: {
		filter: function (val) {
			window.localStorage.setItem('filter', val)
		},
	},
	methods: {
		scrollDown: function () {
			this.stopChat = false
			this.$refs.logScroll.scrollIntoView({ behavior: 'smooth', block: 'end' })
		},
		scroll: function (e) {
			if (e.changedTouches) {
				this.stopChat = true
			}
			if (e.wheelDelta >= 0) {
				this.stopChat = true
			} else {
				let pos = this.$refs.logScroll.offsetHeight - this.heightWindow
				if (pos - 25 < this.$refs.logScroll.parentElement.scrollTop) {
					this.stopChat = false
				}
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
				this.$refs.logScroll?.scrollIntoView({ behavior: 'smooth', block: 'end' })
			})
		},
		subContainer: function (md5Name) {
			if (this.switchData[md5Name]) {
				this.sendSubContainer('sub', md5Name)
			} else {
				this.sendSubContainer('unsub', md5Name)
			}
		},
		subContainers: function (objs, mainName) {
			for (let val of objs) {
				if (!val.running) {
					continue
				}

				this.switchData[val.md5Name] = this.switchMain[mainName]

				if (this.switchMain[mainName]) {
					this.sendSubContainer('sub', val.md5Name)
				} else {
					this.sendSubContainer('unsub', val.md5Name)
				}
			}
		},
		sendSubContainer: function (act, id) {
			let subs = JSON.parse(window.localStorage.getItem('subs-log'))
			if (!subs) {
				subs = {}
			}
			if (act === 'sub') {
				window.ws.send('sub-log-' + id)
				subs[id] = true
			} else {
				window.ws.send('unsub-log-' + id)
				delete subs[id]
			}
			window.localStorage.setItem('subs-log', JSON.stringify(subs))
		},
		purifyMenu () {
			for (let cont of this.containersAlive) {
				if ((new Date).getTime() - this.containersAlive.get(cont[0]) > 10000) {
					document.getElementById(cont[0]).style.display = 'none'
					this.containersAlive.delete(cont[0])
				} else {
					let el = document.getElementById(cont[0])
					if (el) {
						el.style.display = 'block'
					}
				}
			}
		},
		restoreSubs () {
			if (!this.restoredSubs) {
				this.restoredSubs = true
				for (let val in JSON.parse(window.localStorage.getItem('subs-log'))) {
					this.switchData[val] = true
					this.sendSubContainer('sub', val)
				}
			}
		}
	},
	created () {
		window.addEventListener('resize', this.resizeWindow)
	},
	unmounted () {
		window.removeEventListener('resize', this.resizeWindow)
	},

	mounted () {
		this.$refs.logScroll?.parentElement.addEventListener('touchmove', this.scroll)

		this.$refs.login?.focus()

		window.addEventListener('focus', (event) => {
			if (this.stopChat === false) {
				this.$refs.logScroll?.scrollIntoView({ block: 'end' })
			}
		})

		window.ws.addEventListener('message', (evt) => {
			let jp = JSON.parse(evt.data)

			if (jp.typeMess === 'container') {
				let nameContainer = jp.data.Names[0].slice(1)

				let second = {
					id: jp.data.Id,
					name: nameContainer,
					md5Name: md5(jp.data.Hostname + jp.data.Names[0]).toString(),
					color: colord(stc(jp.data.Hostname + nameContainer)),
					running: jp.data.State === 'running',
					status: jp.data.Status
				}

				let mainIndex = this.containersMenu.findIndex(key => key.hostname === jp.data.Hostname)
				if (this.containersMenu[mainIndex]) {
					this.containersAlive.set(second.md5Name, (new Date).getTime())

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

				this.containers[jp.data.Md5Name] = second

				this.purifyMenu()

				this.restoreSubs()
			}
			if (jp.typeMess === 'log') {
				jp.data.body = _.escape(jp.data.body)

				if (this.filterSwitch && this.filter !== '') {
					let body = jp.data.hostname + jp.data.name + jp.data.body
					let match = body.toString().match(new RegExp(this.filter, 'gmi'))
					if (!match) {
						if (this.filterSelect !== 'highlight') {
							return false
						}
					}

					let um = _.uniq(match)
					for (let val of um) {
						if (val === '') {
							continue
						}
						jp.data.name = jp.data.name.replaceAll(val, `<span>${val}</span>`)
					}

					for (let val of um) {
						if (val === '') {
							continue
						}
						jp.data.body = jp.data.body.replaceAll(val, `<span>${val}</span>`)
					}

					if (this.beepFlag) {
						this.beep.play()
					}
				}

				this.logsData.push(
					{
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
		})

		setTimeout(() => {
			if (this.$store.state.isAuth) {
				window.ws.send('get-containers')
			}
		}, 100)
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

.row-text span {
	padding: 4px 6px;
	border: 1px solid rgb(64, 158, 255);
	background-color: ghostwhite;
	margin: 0 2px;
	border-radius: 4px;
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

.input-filter {
	margin-bottom: 10px;
}

.input-filter > .el-input__wrapper input {
	width: 500px;
	padding-left: 52px;
}
</style>
