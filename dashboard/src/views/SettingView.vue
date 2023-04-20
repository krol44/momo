<template>
  <div style="margin: 0 40px 0 40px; width: 100%">
    <el-input v-model="installUrl" @click="copyInstallUrl" size="large">
      <template #prepend><a href="#" @click="copyInstallUrl" style="text-decoration: none;">copy</a></template>
    </el-input>
  </div>
</template>

<script>
import { ElMessage } from 'element-plus'

export default {
	data () {
		return {
			installUrl: '',
		}
	},
	methods: {
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
		}
	},
	mounted () {
		window.ws.send('get-install-url')
		window.ws.addEventListener('message', (evt) => {
			let jp = JSON.parse(evt.data)

			if (jp.typeMess === 'install-url') {
				this.installUrl = jp.data
			}
		})
	}
}
</script>
