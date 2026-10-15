<template>
  <div class="app">
    <header class="topbar">
      <button class="menu-btn" @click="sidebarOpen = !sidebarOpen">☰</button>
      <span class="logo">⬡ hitest</span>
      <span class="hint" v-if="state.agents.length">
        {{ state.agents.filter(a => a.online).length }}/{{ state.agents.length }} 在线
      </span>
    </header>

    <div class="body">
      <Sidebar
        :class="{ open: sidebarOpen }"
        :agents="state.agents"
        :selected="agentId"
        @select="onSelect"
      />

      <main class="main" v-if="agentId">
        <div class="tabbar">
          <button :class="{ active: tab === 'monitor' }" @click="tab = 'monitor'">监控</button>
          <button :class="{ active: tab === 'term' }" @click="tab = 'term'">终端</button>
          <button :class="{ active: tab === 'files' }" @click="tab = 'files'">文件</button>
          <span class="spacer" />
          <template v-if="tab === 'term'">
            <button
              v-for="s in sessions"
              :key="s.key"
              class="session-tab"
              :class="{ active: s.key === activeSession }"
              @click="activeSession = s.key"
            >
              {{ s.agent }} #{{ s.n }}
              <span class="x" @click.stop="closeSession(s.key)">✕</span>
            </button>
            <button class="add" @click="newSession" title="新终端">＋</button>
          </template>
        </div>

        <MonitorView v-if="tab === 'monitor'" :agent-id="agentId" />
        <FilesView v-else-if="tab === 'files'" :agent-id="agentId" />
        <div v-else class="terms">
          <TerminalView
            v-for="s in sessions"
            :key="s.key"
            v-show="s.key === activeSession"
            :agent-id="s.agent"
            @closed="closeSession(s.key)"
          />
          <div v-if="!sessions.length" class="empty">点右上 ＋ 新开终端</div>
        </div>
      </main>

      <main class="main empty" v-else>
        <p>从左侧选择一台机器</p>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { state, startAgentPolling } from './api/api'
import Sidebar from './components/Sidebar.vue'
import TerminalView from './components/TerminalView.vue'
import FilesView from './components/FilesView.vue'
import MonitorView from './components/MonitorView.vue'

startAgentPolling()

const sidebarOpen = ref(false)
const agentId = ref('')
const tab = ref('monitor')
const sessions = ref([]) // [{key, agent, n}]
const activeSession = ref('')
let seq = 0

function onSelect(id) {
  agentId.value = id
  sidebarOpen.value = false
  tab.value = 'monitor'
}

function newSession() {
  if (!agentId.value) return
  const s = { key: `s${++seq}`, agent: agentId.value, n: seq }
  sessions.value.push(s)
  activeSession.value = s.key
}

function closeSession(key) {
  const idx = sessions.value.findIndex((s) => s.key === key)
  if (idx < 0) return
  sessions.value.splice(idx, 1)
  if (activeSession.value === key) {
    activeSession.value = sessions.value.length
      ? sessions.value[Math.max(0, idx - 1)].key
      : ''
  }
}
</script>
