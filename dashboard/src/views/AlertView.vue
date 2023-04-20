<template>
  <div style="max-width: 900px; margin-left: 20px;">
    <el-alert v-if="telegramBotName" type="info" :closable="false">
      Join to the <a target="_blank"
                     :href="'https://t.me/'+telegramBotName"
    >https://t.me/{{ telegramBotName }}</a> telegram bot and press to start, after update this page
      <el-divider style="margin: 5px 0;" />
      <el-input v-model="defaultKeys" placeholder="example: key1**key2**key3" size="small">
        <template #prepend>set default templates keys</template>
      </el-input>
    </el-alert>
    <el-table :data="containers" style="max-width: 900px; margin-left: 20px;" class="table-alert">
      <el-table-column type="index" width="50" />
      <el-table-column prop="hostname" label="Host" width="180" />
      <el-table-column prop="name" label="Name" width="180" />
      <el-table-column label="Key alerts" #default="scope">
      <span v-for="tag in alerts">
        <el-tag class="mx-1" style="margin: 0 5px 5px 0;"
                v-if="tag.container_md5 === scope.row.md5"
                :key="tag.id"
                :disable-transitions="true"
                @close="rmAlert(tag.id)" closable>
          [@{{ tag.telegram_name.String ? tag.telegram_name.String : tag.telegram_id }}] {{ tag.key_alert }}</el-tag>
      </span>
        <el-divider style="margin: 5px 0;" />
        <el-input v-model="inputs[scope.row.md5]" placeholder="key alert string" size="small">
          <template #prepend>
            <el-select v-model="selects[scope.row.md5]"
                       :default-first-option="true" size="small" style="width: 110px;" placeholder="select chat">
              <el-option
                      v-for="item in telegrams"
                      :key="item.telegram_id"
                      :label="item.telegram_name ? item.telegram_name : item.telegram_id"
                      :value="item.telegram_id"
              />
            </el-select>
          </template>
        </el-input>
        <el-button size="small" @click="addCustonAlerts(scope.row.md5)">add default templates</el-button>
        <el-button type="primary" size="small" @click="addCustumAlert(scope.row.md5)">add custom from input</el-button>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { ElMessage } from 'element-plus'
import md5 from 'crypto-js/md5'
import { colord } from 'colord'
import stc from 'string-to-color'

export default {
	data () {
		return {
			containersMenu: [],
			containers: [],
			select: null,
			inputs: {},
			selects: {},
			telegrams: [],
			alerts: [],
			telegramBotName: '',
			defaultKeys: 'level=error**level=fatal**level=panic'
		}
	},
	watch: {
		defaultKeys: function (val) {
			window.localStorage.setItem('default-keys', val)
			ElMessage({
				showClose: false,
				grouping: true,
				message: 'saved',
				type: 'success',
				duration: 1000,
			})
		},
	},
  created () {
		let d = window.localStorage.getItem('default-keys')
	  if (!!d) {
	    this.defaultKeys = d
    }
  },
	methods: {
		rmAlert: function (id) {
			window.ws.send('rm-alert-' + id)
			window.ws.send('get-alerts')
		},
		addCustumAlert: function (md5) {
			let s = this.inputs[md5]
			let i = this.selects[md5]

			if (!s) {
				ElMessage({
					showClose: false,
					grouping: true,
					message: 'write key alert',
					type: 'warning',
					duration: 2000,
				})
				return
			}
			if (!i) {
				ElMessage({
					showClose: false,
					grouping: true,
					message: 'choose telegram chat',
					type: 'warning',
					duration: 2000,
				})
				return
			}

			window.ws.send('add-alert-' + JSON.stringify({ telegram_id: i, key_alert: s, md5: md5 }))
			window.ws.send('get-alerts')
		},
		addCustonAlerts: function (md5) {
			let i = this.selects[md5]
			if (!i) {
				ElMessage({
					showClose: false,
					grouping: true,
					message: 'choose telegram chat',
					type: 'warning',
					duration: 2000,
				})
				return
			}

			for (let k of this.defaultKeys.split('**')) {
				window.ws.send('add-alert-' + JSON.stringify({ telegram_id: i, key_alert: k, md5: md5 }))
      }
			window.ws.send('get-alerts')
		}
	},
	mounted () {
		window.ws.addEventListener('message', (evt) => {
			let jp = JSON.parse(evt.data)

			if (jp.typeMess === 'alerts') {
				this.telegrams = jp.telegrams
				this.alerts = jp.alerts
				this.telegramBotName = jp.telegram_bot_name
			}

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

				let containers = []
				for (let g of this.containersMenu) {
					for (let c of g.containers) {
						if (c.running) {
							containers.push({ hostname: g.hostname, name: c.name, md5: c.md5Name })
						}
					}
				}

				this.containers = containers
			}
		})

		setTimeout(() => {
			if (this.$store.state.isAuth) {
				window.ws.send('get-alerts')
				window.ws.send('get-containers')
			}
		}, 100)
	}
}
</script>

<style>
.table-alert .el-table__body td {
	vertical-align: top;
}
</style>