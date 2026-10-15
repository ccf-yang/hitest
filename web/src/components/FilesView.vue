<template>
  <div class="files">
    <div class="toolbar">
      <button @click="goUp" :disabled="path === '/'">⬆ 上级</button>
      <span class="path" @click="promptPath">{{ path }}</span>
      <span class="spacer" />
      <label class="btn">
        上传
        <input
          type="file"
          multiple
          hidden
          @change="onUpload($event.target.files); $event.target.value = ''"
        />
      </label>
      <button @click="load">刷新</button>
      <button @click="mkdir">新建目录</button>
    </div>

    <table>
      <thead>
        <tr><th>名称</th><th>大小</th><th>修改时间</th><th>操作</th></tr>
      </thead>
      <tbody>
        <tr v-for="e in entries" :key="e.name">
          <td>
            <a @click="e.is_dir ? enter(e) : download(e)">
              {{ e.is_dir ? '📁' : '📄' }} {{ e.name }}
            </a>
          </td>
          <td>{{ e.is_dir ? '—' : fmtSize(e.size) }}</td>
          <td>{{ fmtTime(e.mtime) }}</td>
          <td class="ops">
            <button v-if="!e.is_dir" @click="download(e)">下载</button>
            <button @click="rename(e)">改名</button>
            <button class="danger" @click="remove(e)">删除</button>
          </td>
        </tr>
      </tbody>
    </table>
    <p v-if="!entries.length" class="empty">空目录</p>
    <p v-if="err" class="error">{{ err }}</p>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { fsList, fsDelete, fsRename, fsMkdir, fsUpload, fsDownload } from '../api/api'

const props = defineProps({ agentId: String })

const path = ref('/')
const entries = ref([])
const err = ref('')

async function load() {
  err.value = ''
  try {
    const res = await fsList(props.agentId, path.value)
    entries.value = res.entries || []
  } catch (e) {
    err.value = e.message
    entries.value = []
  }
}

function enter(e) {
  path.value = e.name === '..'
    ? parentOf(path.value)
    : join(path.value, e.name)
  load()
}

function goUp() {
  path.value = parentOf(path.value)
  load()
}

function promptPath() {
  const p = prompt('跳转到目录:', path.value)
  if (p) {
    path.value = p
    load()
  }
}

const join = (a, b) => (a === '/' ? '/' + b : a + '/' + b)
const parentOf = (p) => {
  const i = p.lastIndexOf('/')
  return i <= 0 ? '/' : p.slice(0, i)
}

// ---------- 上传: 流式 PUT, 支持大文件 ----------
async function onUpload(fileList) {
  for (const file of fileList) {
    try {
      await fsUpload(props.agentId, join(path.value, file.name), file)
    } catch (e) {
      err.value = `上传 ${file.name}: ${e.message}`
    }
  }
  load()
}

function download(e) {
  fsDownload(props.agentId, join(path.value, e.name), e.name)
}

async function remove(e) {
  if (!confirm(`确认删除 ${e.is_dir ? '目录' : '文件'} ${e.name}?`)) return
  try {
    await fsDelete(props.agentId, join(path.value, e.name))
    load()
  } catch (ex) {
    err.value = ex.message
  }
}

async function rename(e) {
  const newName = prompt('新名称:', e.name)
  if (!newName || newName === e.name) return
  try {
    await fsRename(props.agentId, join(path.value, e.name), join(path.value, newName))
    load()
  } catch (ex) {
    err.value = ex.message
  }
}

async function mkdir() {
  const name = prompt('目录名:')
  if (!name) return
  try {
    await fsMkdir(props.agentId, join(path.value, name))
    load()
  } catch (ex) {
    err.value = ex.message
  }
}

// ---------- 展示 ----------
const fmtSize = (n) => {
  if (n < 1024) return n + ' B'
  for (const u of ['KB', 'MB', 'GB', 'TB']) {
    n /= 1024
    if (n < 1024) return n.toFixed(1) + ' ' + u
  }
  return n.toFixed(1) + ' PB'
}
const fmtTime = (t) => new Date(t * 1000).toLocaleString()

watch(() => props.agentId, () => {
  path.value = '/'
  load()
})
load()
</script>
