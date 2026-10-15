// HTTP API 封装: 一切功能 = server 的 HTTP API（设计支柱①的前端对应物）
import { reactive } from 'vue'

export const state = reactive({
  agents: [], // [{id, hostname, os, online}]
})

export function machineBase(id) {
  return id === 'local' ? '/api/v1/local' : `/api/v1/m/${encodeURIComponent(id)}`
}

export async function api(method, path, body, raw) {
  const opts = { method }
  if (body !== undefined && body !== null) {
    if (raw) {
      opts.body = body // 流式二进制（上传）
    } else {
      opts.headers = { 'Content-Type': 'application/json' }
      opts.body = JSON.stringify(body)
    }
  }
  const resp = await fetch(path, opts)
  if (!resp.ok) {
    let msg = `HTTP ${resp.status}`
    try {
      const e = await resp.json()
      if (e.error) msg = e.error
    } catch {}
    throw new Error(msg)
  }
  return resp
}

async function jsonCall(method, path, body) {
  const resp = await api(method, path, body)
  return resp.json()
}

// ---------- 机器列表（轮询） ----------
let pollTimer = null

export function startAgentPolling(intervalMs = 5000) {
  refreshAgents()
  pollTimer = setInterval(refreshAgents, intervalMs)
}

async function refreshAgents() {
  try {
    const data = await jsonCall('GET', '/api/v1/agents')
    state.agents = data.agents || []
  } catch {
    state.agents = []
  }
}

// ---------- 监控 ----------
export const getBaseOS = (id) => jsonCall('GET', `${machineBase(id)}/monitor/base/os`)
export const getCurrent = (id, io = 'all', netOpt = 'all') =>
  jsonCall('GET', `${machineBase(id)}/monitor/current/${io}/${netOpt}`)
export const getTop = (id, kind) => jsonCall('GET', `${machineBase(id)}/monitor/top/${kind}`)
export const searchMonitor = (id, body) => jsonCall('POST', `${machineBase(id)}/monitor/search`, body)
export const getMonitorSetting = (id) => jsonCall('GET', `${machineBase(id)}/monitor/setting`)

// ---------- 文件（纯 HTTP, 无自研协议） ----------
export const fsList = (id, path) =>
  jsonCall('GET', `${machineBase(id)}/files/list?path=${encodeURIComponent(path)}`)
export const fsDelete = (id, path) => jsonCall('POST', `${machineBase(id)}/files/delete`, { path })
export const fsRename = (id, path, newPath) =>
  jsonCall('POST', `${machineBase(id)}/files/rename`, { path, new_path: newPath })
export const fsMkdir = (id, path) => jsonCall('POST', `${machineBase(id)}/files/mkdir`, { path })

export async function fsUpload(id, path, file) {
  const resp = await api('PUT', `${machineBase(id)}/files/file?path=${encodeURIComponent(path)}`, file, true)
  await resp.body?.cancel?.()
}

export async function fsDownload(id, path, name) {
  const resp = await api('GET', `${machineBase(id)}/files/file?path=${encodeURIComponent(path)}`)
  const blob = await resp.blob()
  const a = document.createElement('a')
  a.href = URL.createObjectURL(blob)
  a.download = name
  a.click()
  URL.revokeObjectURL(a.href)
}

// ---------- exec ----------
export const execCmd = (id, cmd, timeoutSec = 30) =>
  jsonCall('POST', `${machineBase(id)}/exec/run`, { cmd, timeout_sec: timeoutSec })

// ---------- 终端 WS ----------
// 协议: 首条文本消息 = {cols,rows,cmd}; 之后二进制 = [1B kind][payload]
//   kind 0=数据 1=resize(cols,rows 各2B) 2=关闭
export function openTerminalWS(id) {
  const scheme = location.protocol === 'https:' ? 'wss' : 'ws'
  return new WebSocket(`${scheme}://${location.host}${machineBase(id)}/terminal/ws`)
}
