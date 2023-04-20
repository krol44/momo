<template>
  <div style="padding: 0 10px 10px 10px;">
    <el-table :data="containers" style="width: 100%;"
              :default-sort="{ prop: 'cpu', order: 'descending' }">
      <el-table-column prop="hostname" label="Host" sortable width="150">
        <template #default="scope">
          <span :style="'font-weight: bold; color: '+scope.row.color.alpha(1).toHex()">{{scope.row.hostname}}</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="Name" sortable width="200">
        <template #default="scope">
          <span :style="'color: '+scope.row.color.alpha(1).toHex()">{{scope.row.name}}</span>
        </template>
      </el-table-column>
      <el-table-column prop="cpu" label="CPU" sortable width="80">
        <template #default="scope">
          <span :style="scope.row.cpu >= 30? 'color: #eb656f':''">{{ scope.row.cpu }} %</span>
        </template>
      </el-table-column>
      <el-table-column prop="mem" label="Memory / Max" #default="scope" sortable width="250">
          <span v-if="stats[scope.row.md5Name]"
                :style="stats[scope.row.md5Name].mem > 200 ? 'color: #eb656f':''">
          {{ stats[scope.row.md5Name].mem }} MB</span><span v-if="stats[scope.row.md5Name]">
        / {{ stats[scope.row.md5Name].memMax }} MB</span>
      </el-table-column>
      <el-table-column prop="status" label="Status" sortable width="180" />
      <el-table-column label="Network I/O" #default="scope" width="250">
          <span v-if="stats[scope.row.md5Name]">
            {{ stats[scope.row.md5Name].netI }} MB / {{ stats[scope.row.md5Name].netO }} MB
          </span>
      </el-table-column>
      <el-table-column label="Disk R/W" #default="scope" width="250">
          <span v-if="stats[scope.row.md5Name]">
            {{ stats[scope.row.md5Name].diskRead }} MB / {{ stats[scope.row.md5Name].diskWrite }} MB
          </span>
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import md5 from 'crypto-js/md5'
import { colord } from 'colord'
import stc from 'string-to-color'
import _ from 'lodash'

export default {
	data () {
		return {
			containers: [],
			stats: {}
		}
	},
	methods: {},
	mounted () {
		window.ws.addEventListener('message', (evt) => {
			let jp = JSON.parse(evt.data)

			if (jp.typeMess === 'stats') {
				let o = JSON.parse(jp.data.body)

				let cpu = _.round(((o.cpu_stats.cpu_usage.total_usage - o.precpu_stats.cpu_usage.total_usage) /
					(o.cpu_stats.system_cpu_usage - o.precpu_stats.system_cpu_usage)) * o.cpu_stats.online_cpus * 100.0, 2)

				let mem = _.round((o.memory_stats.usage - o.memory_stats.stats.inactive_file) / 1024 / 1024, 2)
				let memMax = _.round(o.memory_stats.limit / 1024 / 1024, 2)

				let netI = 0
				let netO = 0
				if (o.networks['eth0']) {
					netI = _.round(o.networks['eth0'].rx_bytes / 1000 / 1000, 2)
					netO = _.round(o.networks['eth0'].tx_bytes / 1000 / 1000, 2)
				}

				let diskRead = 0
				let diskWrite = 0
				if (o.blkio_stats.io_service_bytes_recursive) {
					diskRead = _.round(o.blkio_stats.io_service_bytes_recursive[0].value / 1000 / 1000, 2)
					diskWrite = _.round(o.blkio_stats.io_service_bytes_recursive[1].value / 1000 / 1000, 2)
				}

				this.stats[jp.data.md5_name] = {
					cpu: cpu,
					mem: mem,
					memMax: memMax,
					netI: netI,
					netO: netO,
					diskRead: diskRead,
					diskWrite: diskWrite,
				}
			}

			if (jp.typeMess === 'container') {
				let nameContainer = jp.data.Names[0].slice(1)

				let second = {
					id: jp.data.Id,
					name: nameContainer,
					md5Name: md5(jp.data.Hostname + jp.data.Names[0]).toString(),
					color: colord(stc(jp.data.Hostname)).darken(),
					running: jp.data.State === 'running',
					status: jp.data.Status,
					hostname: jp.data.Hostname,
					AliveTime: (new Date()).getTime()
				}

				if (!second.running) {
					return
				}

				second.cpu = this.stats[second.md5Name] ? this.stats[second.md5Name].cpu : 0
				second.mem = this.stats[second.md5Name] ? this.stats[second.md5Name].mem : 0

				let i = this.containers.findIndex(key => key.md5Name === jp.data.Md5Name)
				if (this.containers[i]) {
					this.containers[i] = second
				} else {
					this.containers.push(second)
				}

				_.remove(this.containers, (v) => (new Date()).getTime() - 10000 > v.AliveTime)
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
