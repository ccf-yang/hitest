<template>
  <div class="monitor">
    <div v-if="err" class="error">{{ err }}</div>

    <div class="cards" v-if="cur">
      <div class="card">
        <h4>CPU</h4>
        <div class="big">{{ cur.cpuUsedPercent?.toFixed(1) }}%</div>
        <div class="sub">{{ cur.cpuTotal }} 核 · 负载 {{ cur.load1?.toFixed(2) }}</div>
        <canvas ref="cpuChart" class="spark"></canvas>
      </div>
      <div class="card">
        <h4>内存</h4>
        <div class="big">{{ cur.memoryUsedPercent?.toFixed(1) }}%</div>
        <div class="sub">{{ fmtBytes(cur.memoryUsed) }} / {{ fmtBytes(cur.memoryTotal) }}</div>
        <canvas ref="memChart" class="spark"></canvas>
      </div>
      <div class="card">
        <h4>运行时间</h4>
        <div class="big">{{ cur.runningTime?.days }}<small>天</small></div>
        <div class="sub">{{ cur.runningTime?.hours }}小时 {{ cur.runningTime?.minutes }}分</div>
        <div class="sub">进程 {{ cur.procs }}</div>
      </div>
      <div class="card">
        <h4>Swap</h4>
        <div class="big">{{ cur.swapMemoryTotal ? cur.swapMemoryUsedPercent?.toFixed(0) + '%' : '—' }}</div>
        <div class="sub" v-if="cur.swapMemoryTotal">
          {{ fmtBytes(cur.swapMemoryUsed) }} / {{ fmtBytes(cur.swapMemoryTotal) }}
        </div>
      </div>
    </div>

    <div class="row2" v-if="cur">
      <div class="card wide">
        <h4>网络 ({{ netName }})</h4>
        <div class="sub big-sub">
          ↑ {{ fmtBytes(cur.netBytesSent) }} &nbsp;&nbsp; ↓ {{ fmtBytes(cur.netBytesRecv) }}
        </div>
      </div>
      <div class="card wide">
        <h4>磁盘</h4>
        <table class="mini">
          <tr><th>挂载点</th><th>已用</th><th>总量</th><th></th></tr>
          <tr v-for="d in cur.diskData" :key="d.path">
            <td>{{ d.path }}</td>
            <td>{{ fmtBytes(d.used) }}</td>
            <td>{{ fmtBytes(d.total) }}</td>
            <td class="bar"><span :style="{ width: d.usedPercent + '%' }" :class="{ hot: d.usedPercent > 90 }"></span></td>
          </tr>
        </table>
      </div>
    </div>

    <div class="card wide" v-if="top.length">
      <h4>Top 进程 (CPU)</h4>
      <table class="mini">
        <tr><th>PID</th><th>名称</th><th>用户</th><th>CPU%</th><th>内存</th></tr>
        <tr v-for="p in top" :key="p.pid">
          <td>{{ p.pid }}</td><td>{{ p.name }}</td><td>{{ p.user }}</td>
          <td>{{ p.percent?.toFixed(1) }}</td>
          <td>{{ fmtBytes(p.memory) }}</td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onBeforeUnmount, computed } from 'vue'
import { getCurrent, getTop, getBaseOS } from '../api/api'

const props = defineProps({ agentId: String })

const cur = ref(null)
const top = ref([])
const err = ref('')
const cpuChart = ref(null)
const memChart = ref(null)
const cpuHist = ref([])
const memHist = ref([])
let timer = null

const netName = computed(() => props.agentId)

function drawSpark(canvas, hist, color) {
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  const w = canvas.width = canvas.offsetWidth * 2
  const h = canvas.height = 80
  ctx.clearRect(0, 0, w, h)
  if (hist.length < 2) return
  ctx.beginPath()
  hist.forEach((v, i) => {
    const x = (i / (hist.length - 1)) * w
    const y = h - (v / 100) * h
    i === 0 ? ctx.moveTo(x, y) : ctx.lineTo(x, y)
  })
  ctx.strokeStyle = color
  ctx.lineWidth = 3
  ctx.stroke()
  ctx.lineTo(w, h)
  ctx.lineTo(0, h)
  ctx.closePath()
  ctx.globalAlpha = 0.15
  ctx.fillStyle = color
  ctx.fill()
  ctx.globalAlpha = 1
}

async function tick() {
  try {
    const c = await getCurrent(props.agentId)
    cur.value = c
    err.value = ''
    cpuHist.value.push(c.cpuUsedPercent || 0)
    memHist.value.push(c.memoryUsedPercent || 0)
    if (cpuHist.value.length > 60) { cpuHist.value.shift(); memHist.value.shift() }
    setTimeout(() => {
      drawSpark(cpuChart.value, cpuHist.value, '#4f9cf9')
      drawSpark(memChart.value, memHist.value, '#3fb950')
    }, 0)
  } catch (e) {
    err.value = e.message
  }
}

async function tickTop() {
  try {
    top.value = (await getTop(props.agentId, 'cpu')).slice(0, 5)
  } catch {}
}

let topTimer = null

watch(() => props.agentId, () => {
  cpuHist.value = []
  memHist.value = []
  cur.value = null
  top.value = []
  tick()
  tickTop()
})
tick()
tickTop()

timer = setInterval(tick, 3000)
topTimer = setInterval(tickTop, 10000)
onBeforeUnmount(() => { clearInterval(timer); clearInterval(topTimer) })

function fmtBytes(n) {
  if (n === undefined || n === null) return '—'
  if (n < 1024) return n + ' B'
  let u = ['KB', 'MB', 'GB', 'TB']
  for (const unit of u) {
    n /= 1024
    if (n < 1024) return n.toFixed(1) + ' ' + unit
  }
  return n.toFixed(1) + ' PB'
}
</script>
