<template>
  <aside class="sidebar">
    <h3>机器</h3>
    <ul>
      <li
        v-for="a in agents"
        :key="a.id"
        :class="{ active: a.id === selected, off: !a.online }"
        @click="a.online && $emit('select', a.id)"
      >
        <span class="dot" :class="a.online ? 'on' : 'off'">●</span>
        <span class="name">{{ a.id }}</span>
        <span class="os">{{ a.os }}</span>
      </li>
    </ul>
    <p v-if="!agents.length" class="empty">暂无机器接入</p>

    <details class="add-help">
      <summary>＋ 添加机器</summary>
      <ol>
        <li>把 <code>hitest</code> 二进制拷到目标机器</li>
        <li>
          执行 <pre>sudo sh install.sh \
  -s {{ serverURL }} \
  -t &lt;config.yaml 的 token&gt; \
  -n nas1</pre>
        </li>
        <li>上线后自动出现在列表</li>
      </ol>
    </details>
  </aside>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  agents: { type: Array, default: () => [] },
  selected: { type: String, default: '' },
})
defineEmits(['select'])

const serverURL = computed(() => {
  const scheme = location.protocol === 'https:' ? 'wss' : 'ws'
  return `${scheme}://${location.host}`
})
</script>
