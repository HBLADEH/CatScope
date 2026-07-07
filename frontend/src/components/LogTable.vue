<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { useVirtualizer } from '@tanstack/vue-virtual'
import { CopyOutline } from '@vicons/ionicons5'
import { NIcon, useMessage } from 'naive-ui'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'

import { t } from '@/i18n'
import { useLogStore } from '@/stores/logs'
import type { LogEntry } from '@/types/backend'

const store = useLogStore()
const message = useMessage()
const parentRef = ref<HTMLElement | null>(null)
const selectedLogIDs = ref<Set<number>>(new Set())
const lastSelectedIndex = ref<number | null>(null)

const rowVirtualizer = useVirtualizer(
  computed(() => ({
    count: store.filteredLogs.length,
    getScrollElement: () => parentRef.value,
    estimateSize: () => 26,
    overscan: 20
  }))
)

const virtualRows = computed(() => rowVirtualizer.value.getVirtualItems())
const totalSize = computed(() => rowVirtualizer.value.getTotalSize())
const selectedLogs = computed(() => store.filteredLogs.filter((entry) => selectedLogIDs.value.has(entry.id)))
const selectedCount = computed(() => selectedLogs.value.length)

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

function handleRowClick(entry: LogEntry, index: number, event: MouseEvent) {
  const nextSelectedIDs = new Set(selectedLogIDs.value)

  if (event.shiftKey && lastSelectedIndex.value !== null) {
    const start = Math.min(lastSelectedIndex.value, index)
    const end = Math.max(lastSelectedIndex.value, index)
    for (let current = start; current <= end; current += 1) {
      const rangeEntry = store.filteredLogs[current]
      if (rangeEntry) {
        nextSelectedIDs.add(rangeEntry.id)
      }
    }
  } else if (event.ctrlKey || event.metaKey) {
    if (nextSelectedIDs.has(entry.id)) {
      nextSelectedIDs.delete(entry.id)
    } else {
      nextSelectedIDs.add(entry.id)
    }
    lastSelectedIndex.value = index
  } else {
    nextSelectedIDs.clear()
    nextSelectedIDs.add(entry.id)
    lastSelectedIndex.value = index
  }

  selectedLogIDs.value = nextSelectedIDs
  store.selectLog(entry)
}

function selectAllVisible() {
  selectedLogIDs.value = new Set(store.filteredLogs.map((entry) => entry.id))
  lastSelectedIndex.value = store.filteredLogs.length > 0 ? store.filteredLogs.length - 1 : null
}

function clearSelection() {
  selectedLogIDs.value = new Set()
  lastSelectedIndex.value = null
}

function formatLogEntry(entry: LogEntry) {
  const primary = entry.raw || [
    entry.timestamp || '-',
    entry.level || '?',
    entry.pid || '-',
    entry.tid || '-',
    entry.packageName || '-',
    entry.tag || '-',
    entry.message
  ].join('\t')
  return [primary, ...(entry.multiline ?? [])].filter(Boolean).join('\n')
}

async function copySelectedLogs() {
  if (selectedLogs.value.length === 0) {
    return
  }
  try {
    await ClipboardSetText(selectedLogs.value.map(formatLogEntry).join('\n'))
    message.success(t('table.copySuccess', { count: selectedLogs.value.length }))
  } catch (err) {
    message.error(err instanceof Error ? err.message : String(err))
  }
}
</script>

<template>
  <section class="log-panel">
    <div class="log-actions">
      <span>{{ t('table.selected', { count: selectedCount }) }}</span>
      <div class="log-action-buttons">
        <n-button size="tiny" tertiary :disabled="selectedCount === 0" @click="copySelectedLogs">
          <template #icon>
            <n-icon :component="CopyOutline" />
          </template>
          {{ t('table.copy') }}
        </n-button>
        <n-button size="tiny" tertiary :disabled="store.filteredLogs.length === 0" @click="selectAllVisible">
          {{ t('table.selectVisible') }}
        </n-button>
        <n-button size="tiny" tertiary :disabled="selectedCount === 0" @click="clearSelection">
          {{ t('table.clearSelection') }}
        </n-button>
      </div>
    </div>

    <div class="log-header grid-row">
      <span></span>
      <span>{{ t('table.time') }}</span>
      <span>{{ t('table.level') }}</span>
      <span>{{ t('table.pid') }}</span>
      <span>{{ t('table.tid') }}</span>
      <span>{{ t('table.package') }}</span>
      <span>{{ t('table.tag') }}</span>
      <span>{{ t('table.message') }}</span>
    </div>

    <div
      ref="parentRef"
      class="log-scroll"
      tabindex="0"
      @keydown.ctrl.a.prevent="selectAllVisible"
      @keydown.meta.a.prevent="selectAllVisible"
      @keydown.ctrl.c.prevent="copySelectedLogs"
      @keydown.meta.c.prevent="copySelectedLogs"
    >
      <div v-if="store.filteredLogs.length === 0" class="empty-log">
        {{ store.tableEmptyMessage }}
        <button
          v-if="store.logs.length > 0 && store.filteredLogs.length === 0"
          class="link-button"
          type="button"
          @click="store.clearSearch"
        >
          {{ t('table.clearFilters') }}
        </button>
      </div>
      <div v-else class="virtual-spacer" :style="{ height: `${totalSize}px` }">
        <button
          v-for="virtualRow in virtualRows"
          :key="store.filteredLogs[virtualRow.index].id"
          class="log-row grid-row"
          :class="{
            selected: store.selectedLog?.id === store.filteredLogs[virtualRow.index].id,
            'multi-selected': selectedLogIDs.has(store.filteredLogs[virtualRow.index].id)
          }"
          :style="{ transform: `translateY(${virtualRow.start}px)` }"
          @click="handleRowClick(store.filteredLogs[virtualRow.index], virtualRow.index, $event)"
        >
          <span class="selection-cell" :aria-label="selectedLogIDs.has(store.filteredLogs[virtualRow.index].id) ? t('table.selectedAria') : t('table.notSelectedAria')">
            <span v-if="selectedLogIDs.has(store.filteredLogs[virtualRow.index].id)" class="selection-mark"></span>
          </span>
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
