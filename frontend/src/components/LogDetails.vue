<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { NButton, NTag, useMessage } from 'naive-ui'

import {
  analysisSuggestions,
  analysisSummary,
  analysisTitle,
  analysisTypeLabel,
  severityLabel,
  t
} from '@/i18n'
import { useLogStore } from '@/stores/logs'

const props = withDefaults(defineProps<{
  defaultTab?: 'details' | 'analysis'
}>(), {
  defaultTab: 'details'
})

const store = useLogStore()
const message = useMessage()
const activeTab = ref<'details' | 'analysis'>(props.defaultTab)

watch(
  () => props.defaultTab,
  (tab) => {
    activeTab.value = tab
  }
)

const fullText = computed(() => {
  const entry = store.selectedLog
  if (!entry) {
    return ''
  }
  return [entry.raw, ...(entry.multiline ?? [])].filter(Boolean).join('\n')
})

async function copyAIContext(resultID?: string) {
  try {
    await store.copyAIContext(resultID)
    message.success(t('analysis.copied'))
  } catch (err) {
    message.error(err instanceof Error ? err.message : String(err))
  }
}

async function exportAIContext(resultID?: string) {
  try {
    const path = await store.exportAIContext(resultID)
    message.success(t('analysis.exported', { path }))
  } catch (err) {
    message.error(err instanceof Error ? err.message : String(err))
  }
}

function severityTagType(severity: string) {
  if (severity === 'fatal' || severity === 'error') {
    return 'error'
  }
  if (severity === 'warning') {
    return 'warning'
  }
  return 'info'
}
</script>

<template>
  <aside class="details-panel">
    <n-tabs v-model:value="activeTab" type="line" animated>
      <n-tab-pane name="details" :tab="t('details.title')">
        <template v-if="store.selectedLog">
          <dl>
            <dt>{{ t('details.id') }}</dt>
            <dd>{{ store.selectedLog.id }}</dd>
            <dt>{{ t('details.time') }}</dt>
            <dd>{{ store.selectedLog.timestamp || '-' }}</dd>
            <dt>{{ t('details.level') }}</dt>
            <dd>{{ store.selectedLog.level || '-' }}</dd>
            <dt>PID / TID</dt>
            <dd>{{ store.selectedLog.pid || '-' }} / {{ store.selectedLog.tid || '-' }}</dd>
            <dt>{{ t('details.package') }}</dt>
            <dd>{{ store.selectedLog.packageName || '-' }}</dd>
            <dt>{{ t('details.tag') }}</dt>
            <dd>{{ store.selectedLog.tag || '-' }}</dd>
            <dt>{{ t('details.message') }}</dt>
            <dd>{{ store.selectedLog.message || '-' }}</dd>
            <dt>{{ t('details.raw') }}</dt>
            <dd>{{ store.selectedLog.raw || '-' }}</dd>
            <dt>{{ t('details.lines') }}</dt>
            <dd>{{ 1 + (store.selectedLog.multiline?.length ?? 0) }}</dd>
          </dl>
          <section class="multiline-block">
            <h3>{{ t('details.multiline') }}</h3>
            <pre>{{ store.selectedLog.multiline?.join('\n') || '-' }}</pre>
          </section>
          <h3>{{ t('details.fullRaw') }}</h3>
          <pre>{{ fullText }}</pre>
        </template>
        <p v-else class="empty-copy">{{ t('details.selectLog') }}</p>
      </n-tab-pane>

      <n-tab-pane name="analysis" :tab="t('details.analysisTitle')">
        <div class="analysis-toolbar">
          <span>{{ t('common.issueCount', { count: store.analysisResults.length }) }}</span>
          <div class="analysis-actions">
            <n-button size="small" tertiary :disabled="store.filteredLogs.length === 0" @click="store.analyzeCurrentLogs">
              {{ t('analysis.analyzeCurrent') }}
            </n-button>
            <n-button size="small" tertiary :disabled="!store.selectedAnalysis" @click="copyAIContext()">
              {{ t('analysis.generateSelected') }}
            </n-button>
          </div>
        </div>

        <div v-if="store.analysisResults.length === 0" class="empty-copy">
          {{ t('analysis.empty') }}
        </div>

        <div v-else class="analysis-list">
          <div
            v-for="result in store.analysisResults"
            :key="result.id"
            class="analysis-item"
            :class="{ selected: store.selectedAnalysis?.id === result.id }"
            role="button"
            tabindex="0"
            @click="store.selectAnalysis(result)"
            @keydown.enter="store.selectAnalysis(result)"
          >
            <div class="analysis-head">
              <n-tag size="small" :type="severityTagType(result.severity)">
                {{ analysisTypeLabel(result.type) }}
              </n-tag>
              <n-tag size="small" :bordered="false">
                {{ severityLabel(result.severity) }}
              </n-tag>
            </div>
            <strong>{{ analysisTitle(result) }}</strong>
            <p>{{ analysisSummary(result) }}</p>
            <div class="analysis-actions">
              <n-button size="small" type="primary" tertiary @click.stop="copyAIContext(result.id)">
                {{ t('analysis.copyContext') }}
              </n-button>
              <n-button size="small" tertiary @click.stop="exportAIContext(result.id)">
                {{ t('analysis.exportContext') }}
              </n-button>
            </div>
            <dl>
              <dt>{{ t('details.package') }}</dt>
              <dd>{{ result.packageName || '-' }}</dd>
              <dt>PID</dt>
              <dd>{{ result.pid || '-' }}</dd>
              <dt>{{ t('details.time') }}</dt>
              <dd>{{ result.timestamp || '-' }}</dd>
              <dt>{{ t('analysis.reason') }}</dt>
              <dd>{{ result.reason || result.exceptionType || result.signal || '-' }}</dd>
              <dt>{{ t('analysis.related') }}</dt>
              <dd>{{ result.relatedEntryIds?.join(', ') || '-' }}</dd>
            </dl>
            <h3>{{ t('analysis.keyFrames') }}</h3>
            <ul>
              <li v-for="frame in result.keyFrames || []" :key="frame">{{ frame }}</li>
              <li v-if="!result.keyFrames?.length">-</li>
            </ul>
            <h3>{{ t('analysis.suggestions') }}</h3>
            <ul>
              <li v-for="suggestion in analysisSuggestions(result)" :key="suggestion">{{ suggestion }}</li>
            </ul>
            <h3>{{ t('details.raw') }}</h3>
            <pre>{{ result.rawText || '-' }}</pre>
          </div>
        </div>
      </n-tab-pane>
    </n-tabs>
  </aside>
</template>
