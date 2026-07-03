<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { useVirtualizer } from '@tanstack/vue-virtual'

import { useLogStore } from '@/stores/logs'

const store = useLogStore()
const parentRef = ref<HTMLElement | null>(null)

const rowVirtualizer = useVirtualizer(
  computed(() => ({
    count: store.filteredLogs.length,
    getScrollElement: () => parentRef.value,
    estimateSize: () => 32,
    overscan: 16
  }))
)

const virtualRows = computed(() => rowVirtualizer.value.getVirtualItems())
const totalSize = computed(() => rowVirtualizer.value.getTotalSize())

const scrollToBottom = useDebounceFn(() => {
  if (store.filteredLogs.length === 0) {
    return
  }
  rowVirtualizer.value.scrollToIndex(store.filteredLogs.length - 1, { align: 'end' })
}, 40)

watch(
  () => store.filteredLogs.length,
  async () => {
    if (store.paused) {
      return
    }
    await nextTick()
    scrollToBottom()
  }
)

function levelClass(level: string) {
  return `level level-${level || 'unknown'}`
}
</script>

<template>
  <section class="log-panel">
    <div class="log-header grid-row">
      <span>Time</span>
      <span>Level</span>
      <span>PID</span>
      <span>TID</span>
      <span>Package</span>
      <span>Tag</span>
      <span>Message</span>
    </div>

    <div ref="parentRef" class="log-scroll">
      <div v-if="store.filteredLogs.length === 0" class="empty-log">
        {{ store.tableEmptyMessage }}
        <button
          v-if="store.logs.length > 0 && store.filteredLogs.length === 0"
          class="link-button"
          type="button"
          @click="store.clearSearch"
        >
          Clear filters
        </button>
      </div>
      <div v-else class="virtual-spacer" :style="{ height: `${totalSize}px` }">
        <button
          v-for="virtualRow in virtualRows"
          :key="store.filteredLogs[virtualRow.index].id"
          class="log-row grid-row"
          :class="{ selected: store.selectedLog?.id === store.filteredLogs[virtualRow.index].id }"
          :style="{ transform: `translateY(${virtualRow.start}px)` }"
          @click="store.selectLog(store.filteredLogs[virtualRow.index])"
        >
          <span class="mono muted">{{ store.filteredLogs[virtualRow.index].timestamp || '-' }}</span>
          <span :class="levelClass(store.filteredLogs[virtualRow.index].level)">
            {{ store.filteredLogs[virtualRow.index].level || '?' }}
          </span>
          <span class="mono">{{ store.filteredLogs[virtualRow.index].pid || '-' }}</span>
          <span class="mono">{{ store.filteredLogs[virtualRow.index].tid || '-' }}</span>
          <span class="package-cell">{{ store.filteredLogs[virtualRow.index].packageName || '-' }}</span>
          <span class="tag-cell">{{ store.filteredLogs[virtualRow.index].tag || '-' }}</span>
          <span class="message-cell">{{ store.filteredLogs[virtualRow.index].message }}</span>
        </button>
      </div>
    </div>
  </section>
</template>
