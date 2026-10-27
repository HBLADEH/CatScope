<script setup lang="ts">
import { computed, ref } from 'vue'
import { useElementSize } from '@vueuse/core'
import { NButton, NTag, NIcon, useMessage } from 'naive-ui'

import {
  analysisSuggestions,
  analysisSummary,
  analysisTitle,
  analysisTypeLabel,
  severityLabel,
  t
} from '@/i18n'
import { useLogStore } from '@/stores/logs'
import { CloseOutline, LocateOutline } from '@vicons/ionicons5'
import type { AnalysisResult } from '@/types/backend'

const props = withDefaults(defineProps<{
  defaultTab?: 'details' | 'analysis'
  relatedLogsShown?: boolean
}>(), {
  defaultTab: 'details',
  relatedLogsShown: false
})

const store = useLogStore()
const panelRef = ref<HTMLElement | null>(null)
const { height: panelHeight } = useElementSize(panelRef)
const compactIssuePicker = computed(() => store.analysisResults.length > 3 || (panelHeight.value > 0 && panelHeight.value < 400))
const message = useMessage()
const emit = defineEmits<{ 'tab-change': [tab: 'details' | 'analysis']; close: []; locate: [id: number]; 'show-related': [ids: number[]] }>()
const activeTab = computed({
  get: () => props.defaultTab,
  set: (tab: 'details' | 'analysis') => emit('tab-change', tab)
})
const selectedAnalysis = computed(() => store.selectedAnalysis)
const relatedIDs = computed(() => selectedAnalysis.value?.relatedEntryIds || [])
const hasRelatedEntries = computed(() => store.logs.some(entry => relatedIDs.value.includes(entry.id)))
const relatedHidden = computed(() => !props.relatedLogsShown && hasRelatedEntries.value && !store.filteredLogs.some(entry => relatedIDs.value.includes(entry.id)))
const issueOptions = computed(() => store.analysisResults.map(result => ({ label: analysisTitle(result), value: result.id })))
function locateSelected() {
  const visibleID = props.relatedLogsShown ? undefined : relatedIDs.value.find(value => store.filteredLogs.some(entry => entry.id === value))
  const id = visibleID ?? relatedIDs.value.find(value => store.logs.some(entry => entry.id === value))
  const entry = store.logs.find(item => item.id === id)
  if (entry) {
    store.selectLog(entry)
    emit('locate', entry.id)
  }
}
function selectIssue(result: AnalysisResult) {
  store.selectAnalysis(result)
  locateSelected()
}
function handleIssueChange(id: string | number | null) {
  const result = store.analysisResults.find(item => item.id === id)
  if (result) selectIssue(result)
}

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
  <aside ref="panelRef" class="details-panel">
    <n-tabs v-model:value="activeTab" type="line" animated>
      <template #suffix><n-button quaternary circle size="small" :aria-label="t('details.close')" @click="emit('close')"><n-icon :component="CloseOutline" /></n-button></template>
      <n-tab-pane name="details" :tab="t('details.title')">
        <template v-if="store.selectedLog">
          <dl class="log-metadata">
            <dt>{{ t('details.time') }}</dt><dd>{{ store.selectedLog.timestamp || '-' }}</dd>
            <dt>{{ t('details.level') }}</dt><dd>{{ store.selectedLog.level || '-' }}</dd>
            <dt>PID / TID</dt><dd>{{ store.selectedLog.pid || '-' }} / {{ store.selectedLog.tid || '-' }}</dd>
            <dt>{{ t('details.package') }}</dt><dd>{{ store.selectedLog.packageName || '-' }}</dd>
            <dt>{{ t('details.tag') }}</dt><dd>{{ store.selectedLog.tag || '-' }}</dd>
            <dt>{{ t('details.lines') }}</dt><dd>{{ 1 + (store.selectedLog.multiline?.length ?? 0) }}</dd>
          </dl>
          <h3>{{ t('workbench.rawLog') }}</h3>
          <pre class="log-raw">{{ fullText }}</pre>
        </template>
        <p v-else class="empty-copy">{{ t('details.selectLog') }}</p>
      </n-tab-pane>
      <n-tab-pane name="analysis" :tab="t('details.analysisTitle')">
        <div class="analysis-toolbar"><n-button size="small" tertiary :disabled="store.filteredLogs.length === 0" @click="store.analyzeCurrentLogs">{{ t('workbench.reanalyze') }}</n-button></div>
        <p v-if="store.analysisResults.length === 0" class="empty-copy">{{ t('analysis.empty') }}</p>
        <template v-else>
          <n-select v-if="compactIssuePicker" :value="selectedAnalysis?.id || null" :options="issueOptions" filterable :placeholder="t('workbench.chooseIssue')" :aria-label="t('workbench.chooseIssue')" @update:value="handleIssueChange" />
          <div v-else class="analysis-queue">
            <button v-for="result in store.analysisResults" :key="result.id" type="button" class="analysis-result" :class="{ selected: selectedAnalysis?.id === result.id }" :aria-pressed="selectedAnalysis?.id === result.id" @click="selectIssue(result)">
              <strong>{{ analysisTitle(result) }}</strong>
              <span class="analysis-result-meta"><n-tag size="small" :type="severityTagType(result.severity)" :bordered="false">{{ analysisTypeLabel(result.type) }} · {{ severityLabel(result.severity) }}</n-tag><span>{{ result.timestamp || '-' }}</span></span>
            </button>
          </div>
          <article v-if="selectedAnalysis" class="analysis-content">
            <n-alert v-if="relatedHidden" type="info" :show-icon="false" class="related-hidden-alert">{{ t('workbench.relatedHidden') }}<n-button size="small" text type="primary" @click="emit('show-related', relatedIDs)">{{ t('workbench.showRelated') }}</n-button></n-alert>
            <p class="analysis-cause">{{ selectedAnalysis.reason || analysisSummary(selectedAnalysis) }}</p>
            <div class="analysis-actions"><n-button size="small" type="primary" @click="copyAIContext(selectedAnalysis.id)">{{ t('analysis.copyContext') }}</n-button><n-button size="small" @click="exportAIContext(selectedAnalysis.id)">{{ t('analysis.exportContext') }}</n-button><n-button v-if="hasRelatedEntries && !relatedHidden" size="small" quaternary circle :title="t('workbench.locateRelated')" :aria-label="t('workbench.locateRelated')" @click="locateSelected"><n-icon :component="LocateOutline" /></n-button></div>
            <dl><dt>{{ t('details.package') }}</dt><dd>{{ selectedAnalysis.packageName || '-' }}</dd><dt>PID</dt><dd>{{ selectedAnalysis.pid || '-' }}</dd></dl>
            <h3>{{ t('analysis.keyFrames') }}</h3><ul><li v-for="frame in selectedAnalysis.keyFrames || []" :key="frame">{{ frame }}</li><li v-if="!selectedAnalysis.keyFrames?.length">-</li></ul>
            <h3>{{ t('analysis.suggestions') }}</h3><ul><li v-for="suggestion in analysisSuggestions(selectedAnalysis)" :key="suggestion">{{ suggestion }}</li></ul>
            <details><summary>{{ t('workbench.rawLog') }}</summary><pre>{{ selectedAnalysis.rawText || '-' }}</pre></details>
          </article>
          <p v-else class="empty-copy">{{ t('workbench.selectIssue') }}</p>
        </template>
      </n-tab-pane>
    </n-tabs>
  </aside>
</template>
