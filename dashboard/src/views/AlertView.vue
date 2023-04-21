<template>
  <div style="margin: 0 20px 20px 20px;">
    <el-alert type="info" :closable="false">
      Join to the <a target="_blank"
                     :href="'https://t.me/'+telegramBotName"
    >https://t.me/{{ telegramBotName }}</a> telegram bot and press to start, after update this page
      <el-divider style="margin: 5px 0;" />
      <el-input v-model="defaultKeys" placeholder="example: key1**key2**key3" size="small">
        <template #prepend>set default templates keys</template>
      </el-input>
    </el-alert>
    <el-table :data="containers" style="width: 900px; margin: 20px 0 0 20px;
              padding-bottom: 200px;" class="table-alert">
      <el-table-column type="index" width="50" />
      <el-table-column prop="hostname" label="Host" width="180" />
      <el-table-column prop="name" label="Name" width="180" />
      <el-table-column label="Key alerts" #default="scope">
      <span v-for="tag in alerts">
        <el-tag class="mx-1" style="margin: 0 5px 5px 0;"
                v-if="tag.container_md5 === scope.row.md5Name"
                :key="tag.id"
                :disable-transitions="true"
                @close="rmAlert(tag.id)" closable>
          [@{{ tag.telegram_name.String ? tag.telegram_name.String : tag.telegram_id }}] {{ tag.key_alert }}</el-tag>
      </span>
        <el-divider style="margin: 5px 0;" />
        <el-input v-model="inputs[scope.row.md5Name]" placeholder="key alert string" size="small">
          <template #prepend>
            <el-select v-model="selects[scope.row.md5Name]"
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
        <el-button size="small" @click="addCustonAlerts(scope.row.md5Name)">add default templates</el-button>
        <el-button type="primary" size="small" @click="addCustumAlert(scope.row.md5Name)">add custom from input</el-button>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { ElMessage } from 'element-plus'

export default {
	data () {
		return {
			containers: [],
			select: null,
			inputs: {},
			selects: {},
			telegrams: [],
			alerts: [],
			telegramBotName: '',
			defaultKeys: 'level=error**level=fatal**level=panic',
			interval: setInterval
		}
	},
	watch: {
		defaultKeys: function (val) {
			if (window.localStorage.getItem('default-keys') === val) {
				return
			}
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
			window.ws.send('alert-rm-' + id)
			window.ws.send('alerts')
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

			window.ws.send('alert-add-' + JSON.stringify({ telegram_id: i, key_alert: s, md5: md5 }))
			window.ws.send('alerts')
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
				window.ws.send('alert-add-' + JSON.stringify({ telegram_id: i, key_alert: k, md5: md5 }))
			}
			window.ws.send('alerts')
		}
	},
	unmounted () {
		clearInterval(this.interval)
	},
	mounted () {
		window.ws.addEventListener('message', (evt) => {
			let jp = JSON.parse(evt.data)

			if (jp.typeMess === 'alerts') {
				this.telegrams = jp.telegrams
				this.alerts = jp.alerts
				this.telegramBotName = jp.telegram_bot_name

				for (let machine of this.$store.state.containersMenu) {
					for (let cont of machine.containers) {
						cont['hostname'] = machine.hostname
						let mainIndex = this.containers.findIndex(key => key.md5Name === cont.md5Name)
						if (mainIndex === -1) {
							this.containers.push(cont)
						} else {
							this.containers[mainIndex] = cont
						}
					}
				}
			}
		})


		setTimeout(() => {
			window.ws.send('alerts')
		}, 100)
		this.interval = setInterval(() => {
			window.ws.send('alerts')
		}, 5000)
	}
}
</script>

<style>
.table-alert .el-table__body td {
	vertical-align: top;
}
</style>