<template>
  <div style="margin: 0 20px 0 20px;">
    <el-table :data="containers" style="width: 100%; padding-bottom: 200px;"
              :default-sort="{ prop: 'cpu', order: 'descending' }">
      <el-table-column prop="hostname" label="Host" sortable width="150">
        <template #default="scope">
          <span :style="'font-weight: bold; color: '+
          scope.row.hostname_color.alpha(1).toHex()">{{ scope.row.hostname }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="Name" sortable width="200">
        <template #default="scope">
          <span :style="'color: '+scope.row.hostname_color.alpha(1).toHex()">{{ scope.row.name }}</span>
        </template>
      </el-table-column>
      <el-table-column prop="cpu" label="CPU" sortable width="80">
        <template #default="scope">
          <span :style="scope.row.cpu >= 30?'color: #eb656f':''">{{ scope.row.cpu }} %</span>
        </template>
      </el-table-column>
      <el-table-column prop="mem" label="Memory / Max" #default="scope" sortable width="250">
          <span :style="scope.row.mem > 200 ? 'color: #eb656f':''">{{ scope.row.mem }} MB</span
          > {{ scope.row.mem_max }} MB
      </el-table-column>
      <el-table-column prop="status" label="Status" sortable width="180" />
      <el-table-column label="Network I/O" #default="scope" width="250">
        {{ scope.row.net_i }} MB / {{ scope.row.net_o }} MB
      </el-table-column>
      <el-table-column label="Disk R/W" #default="scope" width="250">
        {{ scope.row.d_r }} MB / {{ scope.row.d_w }} MB
      </el-table-column>
    </el-table>
  </div>
</template>

<script>
import { colord } from 'colord'
import stc from 'string-to-color'

export default {
	data () {
		return {
			containers: [],
			interval: setInterval
		}
	},
	unmounted () {
		clearInterval(this.interval)
	},
	mounted () {
		window.ws.addEventListener('message', (evt) => {
			let jp = JSON.parse(evt.data)

			if (jp.typeMess === 'stats') {
				for (let machine of this.$store.state.containersMenu) {
					for (let cont of machine.containers) {

						let stats = jp.data[cont.md5Name]
						cont['cpu'] = stats?.cpu ? stats.cpu : 0
						cont['d_r'] = stats?.d_r ? stats.d_r : 0
						cont['d_w'] = stats?.d_w ? stats.d_w : 0
						cont['mem'] = stats?.mem ? stats.mem : 0
						cont['mem_max'] = stats?.mem_max ? stats.mem_max : 0
						cont['net_i'] = stats?.net_i ? stats.net_i : 0
						cont['net_o'] = stats?.net_o ? stats.net_o : 0

						cont['hostname_color'] = colord(stc(machine.hostname)).darken()
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
			window.ws.send('stats')
		}, 100)
		this.interval = setInterval(() => {
			window.ws.send('stats')
		}, 2000)
	}
}
</script>
