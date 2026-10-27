<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'
import { useDebounceFn } from '@vueuse/core'
import { useVirtualizer } from '@tanstack/vue-virtual'
import { CopyOutline } from '@vicons/ionicons5'
import { NIcon, useMessage, type DropdownOption } from 'naive-ui'
import { ClipboardSetText } from '../../wailsjs/runtime/runtime'

import { t } from '@/i18n'
import { useLogStore } from '@/stores/logs'
import type { LogEntry } from '@/types/backend'

const props = withDefaults(defineProps<{ focusEntryIds?: number[] }>(), { focusEntryIds: () => [] })
const emit = defineEmits<{ select: []; 'return-filter': [] }>()
const store = useLogStore()
const displayLogs = computed(() => props.focusEntryIds.length ? store.logs.filter(entry => props.focusEntryIds.includes(entry.id)) : store.filteredLogs)
const message = useMessage()
const parentRef = ref<HTMLElement | null>(null)
const headerRef = ref<HTMLElement | null>(null)
function syncHeaderScroll() {
  if (headerRef.value && parentRef.value) headerRef.value.scrollLeft = parentRef.value.scrollLeft
}
const selectedLogIDs = ref<Set<number>>(new Set())
const lastSelectedIndex = ref<number | null>(null)
const tagMenuOpen = ref(false)
const tagMenuX = ref(0)
const tagMenuY = ref(0)
const tagMenuTag = ref('')

const rowVirtualizer = useVirtualizer(
  computed(() => {
    const logs = displayLogs.value
    return {
      count: logs.length,
      getScrollElement: () => parentRef.value,
      getItemKey: (index: number) => logs[index]?.id ?? index,
      estimateSize: () => 32,
      overscan: 20
    }
  })
)

async function scrollToLog(id: number) {
  await nextTick()
  const index = displayLogs.value.findIndex(entry => entry.id === id)
  if (index >= 0) rowVirtualizer.value.scrollToIndex(index, { align: 'auto' })
}
defineExpose({ scrollToLog })

const virtualRows = computed(() => rowVirtualizer.value.getVirtualItems())
const totalSize = computed(() => rowVirtualizer.value.getTotalSize())
const selectedLogs = computed(() => displayLogs.value.filter((entry) => selectedLogIDs.value.has(entry.id)))
const selectedCount = computed(() => selectedLogs.value.length)
const tagMenuOptions = computed<DropdownOption[]>(() => [
  { label: t('table.filterTagOnly', { tag: tagMenuTag.value }), key: 'include' },
  { label: t('table.filterTagExclude', { tag: tagMenuTag.value }), key: 'exclude' }
])

const scrollToBottom = useDebounceFn(() => {
  if (displayLogs.value.length === 0) {
    return
  }
  rowVirtualizer.value.scrollToIndex(displayLogs.value.length - 1, { align: 'end' })
}, 40)

watch(
  [
    () => displayLogs.value.length,
    () => displayLogs.value[displayLogs.value.length - 1]?.id
  ],
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

function measureRow(element: unknown) {
  if (element instanceof HTMLElement) {
    rowVirtualizer.value.measureElement(element)
  }
}

function handleRowClick(entry: LogEntry, index: number, event: MouseEvent) {
  const nextSelectedIDs = new Set(selectedLogIDs.value)

  if (event.shiftKey && lastSelectedIndex.value !== null) {
    const start = Math.min(lastSelectedIndex.value, index)
    const end = Math.max(lastSelectedIndex.value, index)
    for (let current = start; current <= end; current += 1) {
      const rangeEntry = displayLogs.value[current]
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
  emit('select')
}

function selectAllVisible() {
  selectedLogIDs.value = new Set(displayLogs.value.map((entry) => entry.id))
  lastSelectedIndex.value = displayLogs.value.length > 0 ? displayLogs.value.length - 1 : null
}

function clearSelection() {
  selectedLogIDs.value = new Set()
  lastSelectedIndex.value = null
}

function handleTagContextMenu(event: MouseEvent, entry: LogEntry) {
  const tag = entry.tag?.trim()
  if (!tag) {
    return
  }
  tagMenuTag.value = tag
  tagMenuX.value = event.clientX
  tagMenuY.value = event.clientY
  tagMenuOpen.value = true
}

function handleTagMenuSelect(key: string | number) {
  const tag = tagMenuTag.value
  if (!tag) {
    return
  }
  const negative = key === 'exclude'
  store.setTagSearchFilter(tag, negative)
  tagMenuOpen.value = false
  message.success(t(negative ? 'table.filterTagExcludeApplied' : 'table.filterTagOnlyApplied', { tag }))
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
    <n-dropdown
      trigger="manual"
      placement="bottom-start"
      :show="tagMenuOpen"
      :x="tagMenuX"
      :y="tagMenuY"
      :options="tagMenuOptions"
      @select="handleTagMenuSelect"
      @clickoutside="tagMenuOpen = false"
    />

    <div class="log-actions">
      <div class="log-scope-summary"><span v-if="props.focusEntryIds.length">{{ t('workbench.relatedCount', { count: displayLogs.length }) }}</span><span>{{ t('table.selected', { count: selectedCount }) }}</span><n-button v-if="props.focusEntryIds.length" size="tiny" text type="primary" @click="emit('return-filter')">{{ t('workbench.returnFilters') }}</n-button></div>
      <div class="log-action-buttons">
        <n-button size="tiny" tertiary :disabled="selectedCount === 0" @click="copySelectedLogs">
          <template #icon>
            <n-icon :component="CopyOutline" />
          </template>
          {{ t('table.copy') }}
        </n-button>
        <n-button size="tiny" tertiary :disabled="displayLogs.length === 0" @click="selectAllVisible">
          {{ t('table.selectVisible') }}
        </n-button>
        <n-button size="tiny" tertiary :disabled="selectedCount === 0" @click="clearSelection">
          {{ t('table.clearSelection') }}
        </n-button>
      </div>
    </div>

    <div ref="headerRef" class="log-header-wrap"><div class="log-header grid-row">
      <span></span>
      <span>{{ t('table.time') }}</span>
      <span>{{ t('table.level') }}</span>
      <span>{{ t('table.pid') }}</span>
      <span>{{ t('table.tid') }}</span>
      <span>{{ t('table.package') }}</span>
      <span>{{ t('table.tag') }}</span>
      <span>{{ t('table.message') }}</span>
    </div></div>

    <div
      ref="parentRef"
      class="log-scroll"
      @scroll="syncHeaderScroll"
      tabindex="0"
      @keydown.ctrl.a.prevent="selectAllVisible"
      @keydown.meta.a.prevent="selectAllVisible"
      @keydown.ctrl.c.prevent="copySelectedLogs"
      @keydown.meta.c.prevent="copySelectedLogs"
    >
      <div v-if="displayLogs.length === 0" class="empty-log">
        <template v-if="store.logs.length === 0 && !store.running">
          <strong>{{ t('workbench.emptyTitle') }}</strong>
          <p>{{ store.tableEmptyMessage }}</p>
          <div class="empty-actions"><n-button type="primary" :loading="store.offlineLoading" @click="store.openLogFile('')">{{ t('workbench.openFile') }}</n-button><n-button :loading="store.sessionLoading" @click="store.openSession('')">{{ t('source.openSession') }}</n-button></div>
        </template>
        <p v-else>{{ store.tableEmptyMessage }}</p>
        <button
          v-if="store.logs.length > 0 && displayLogs.length === 0"
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
          :key="String(virtualRow.key)"
          :ref="measureRow"
          class="log-row grid-row"
          :data-index="virtualRow.index"
          :class="{
            selected: store.selectedLog?.id === displayLogs[virtualRow.index].id,
            'multi-selected': selectedLogIDs.has(displayLogs[virtualRow.index].id)
          }"
          :style="{ transform: `translateY(${virtualRow.start}px)` }"
          @click="handleRowClick(displayLogs[virtualRow.index], virtualRow.index, $event)"
        >
          <span class="selection-cell" :aria-label="selectedLogIDs.has(displayLogs[virtualRow.index].id) ? t('table.selectedAria') : t('table.notSelectedAria')">
            <span v-if="selectedLogIDs.has(displayLogs[virtualRow.index].id)" class="selection-mark"></span>
          </span>
          <span class="mono muted">{{ displayLogs[virtualRow.index].timestamp || '-' }}</span>
          <span :class="levelClass(displayLogs[virtualRow.index].level)">
            {{ displayLogs[virtualRow.index].level || '?' }}
          </span>
          <span class="mono">{{ displayLogs[virtualRow.index].pid || '-' }}</span>
          <span class="mono">{{ displayLogs[virtualRow.index].tid || '-' }}</span>
          <span class="package-cell" :title="displayLogs[virtualRow.index].packageName">{{ displayLogs[virtualRow.index].packageName || '-' }}</span>
          <span
            class="tag-cell tag-cell-action"
            :title="displayLogs[virtualRow.index].tag"
            @contextmenu.stop.prevent="handleTagContextMenu($event, displayLogs[virtualRow.index])"
          >
            {{ displayLogs[virtualRow.index].tag || '-' }}
          </span>
          <span class="message-cell">{{ displayLogs[virtualRow.index].message }}</span>
        </button>
      </div>
    </div>
  </section>
</template>
