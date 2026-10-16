<template>
  <div class="term-wrap"><div ref="el" class="term"></div></div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import { openTerminalWS } from '../api/api'

const props = defineProps({ agentId: String })
const emit = defineEmits(['closed'])
const el = ref(null)

let term, fit, ws, ro, cleanups = []

onMounted(() => {
  term = new Terminal({
    fontSize: 14,
    cursorBlink: true,
    fontFamily: 'Menlo, Consolas, "Courier New", monospace',
    theme: { background: '#16181d', foreground: '#d4d4d4' },
  })
  fit = new FitAddon()
  term.loadAddon(fit)
  term.open(el.value)
  fit.fit()

  ws = openTerminalWS(props.agentId)
  ws.binaryType = 'arraybuffer'

  ws.onopen = () => {
    ws.send(JSON.stringify({ cols: term.cols, rows: term.rows, cmd: '/bin/bash -l' }))
  }
  ws.onmessage = (ev) => {
    if (typeof ev.data === 'string') {
      // 文本 = 错误信息
      term.write(`\r\n\x1b[31m${ev.data}\x1b[0m`)
      emit('closed')
      return
    }
    // 帧: [1B kind][payload]。
    // 注意 ArrayBuffer 上没有 length（那是 TypedArray 的），只有 byteLength。
    // 之前写成 ev.data.length 恒为 undefined，两个分支都进不去，
    // 于是 PTY 的输出永远不写屏 —— 表现就是「敲命令回车只有光标闪，什么都不返回」。
    const u8 = new Uint8Array(ev.data)
    if (u8.byteLength === 0) return
    if (u8[0] === 0) {
      if (u8.byteLength > 1) term.write(u8.subarray(1))
    } else if (u8[0] === 2) {
      term.write('\r\n\x1b[33m[会话已结束]\x1b[0m')
      emit('closed')
    } else {
      // 兜底：万一服务端发来的是不带 kind 的裸输出，原样写屏，别把数据丢了
      term.write(u8)
    }
  }
  ws.onclose = () => {
    term.write('\r\n\x1b[33m[连接已断开]\x1b[0m')
  }

  cleanups.push(term.onData((d) => {
    if (ws.readyState !== WebSocket.OPEN) return
    const enc = new TextEncoder().encode(d)
    const frame = new Uint8Array(1 + enc.length)
    frame[0] = 0
    frame.set(enc, 1)
    ws.send(frame)
  }))

  ro = new ResizeObserver(() => {
    if (!term.element?.offsetParent) return
    fit.fit()
    if (ws.readyState !== WebSocket.OPEN) return
    const frame = new Uint8Array(5)
    frame[0] = 1
    new DataView(frame.buffer).setUint16(1, term.cols)
    new DataView(frame.buffer).setUint16(3, term.rows)
    ws.send(frame)
  })
  ro.observe(el.value.parentElement)
})

onBeforeUnmount(() => {
  cleanups.forEach((d) => d?.dispose?.())
  ro && ro.disconnect()
  try { ws && ws.close() } catch {}
  term && term.dispose()
})
</script>
