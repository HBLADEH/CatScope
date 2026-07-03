<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NTag } from 'naive-ui'

import { useLogStore } from '@/stores/logs'

const store = useLogStore()

const fullText = computed(() => {
  const entry = store.selectedLog
  if (!entry) {
    return ''
  }
  return [entry.raw, ...(entry.multiline ?? [])].filter(Boolean).join('\n')
})
</script>

<template>
  <aside class="details-panel">
    <n-tabs type="line" animated>
      <n-tab-pane name="details" tab="Details">
        <template v-if="store.selectedLog">
          <dl>
            <dt>ID</dt>
            <dd>{{ store.selectedLog.id }}</dd>
            <dt>Time</dt>
            <dd>{{ store.selectedLog.timestamp || '-' }}</dd>
            <dt>Level</dt>
            <dd>{{ store.selectedLog.level || '-' }}</dd>
            <dt>PID / TID</dt>
            <dd>{{ store.selectedLog.pid || '-' }} / {{ store.selectedLog.tid || '-' }}</dd>
            <dt>Package</dt>
            <dd>{{ store.selectedLog.packageName || '-' }}</dd>
            <dt>Tag</dt>
            <dd>{{ store.selectedLog.tag || '-' }}</dd>
            <dt>Message</dt>
            <dd>{{ store.selectedLog.message || '-' }}</dd>
            <dt>Raw</dt>
            <dd>{{ store.selectedLog.raw || '-' }}</dd>
            <dt>Lines</dt>
            <dd>{{ 1 + (store.selectedLog.multiline?.length ?? 0) }}</dd>
          </dl>
          <section class="multiline-block">
            <h3>Multiline</h3>
            <pre>{{ store.selectedLog.multiline?.join('\n') || '-' }}</pre>
          </section>
          <h3>Full Raw</h3>
          <pre>{{ fullText }}</pre>
        </template>
        <p v-else class="empty-copy">Select a log row to inspect the full entry.</p>
      </n-tab-pane>

      <n-tab-pane name="analysis" tab="Analysis">
        <div class="analysis-toolbar">
          <span>{{ store.analysisResults.length }} issue(s)</span>
          <n-button size="small" tertiary :disabled="store.filteredLogs.length === 0" @click="store.analyzeCurrentLogs">
            Analyze Current Logs
          </n-button>
        </div>

        <div v-if="store.analysisResults.length === 0" class="empty-copy">
          No crash, ANR, native crash, or JNI issue detected yet.
        </div>

        <div v-else class="analysis-list">
          <button
            v-for="result in store.analysisResults"
            :key="result.id"
            class="analysis-item"
            :class="{ selected: store.selectedAnalysis?.id === result.id }"
            type="button"
            @click="store.selectAnalysis(result)"
          >
            <div class="analysis-head">
              <n-tag size="small" :type="result.severity === 'fatal' ? 'error' : 'warning'">
                {{ result.type }}
              </n-tag>
              <n-tag size="small" :bordered="false">
                {{ result.severity }}
              </n-tag>
            </div>
            <strong>{{ result.title }}</strong>
            <p>{{ result.summary }}</p>
            <dl>
              <dt>Package</dt>
              <dd>{{ result.packageName || '-' }}</dd>
              <dt>PID</dt>
              <dd>{{ result.pid || '-' }}</dd>
              <dt>Time</dt>
              <dd>{{ result.timestamp || '-' }}</dd>
              <dt>Reason</dt>
              <dd>{{ result.reason || result.exceptionType || result.signal || '-' }}</dd>
              <dt>Related</dt>
              <dd>{{ result.relatedEntryIds?.join(', ') || '-' }}</dd>
            </dl>
            <h3>Key Frames</h3>
            <ul>
              <li v-for="frame in result.keyFrames || []" :key="frame">{{ frame }}</li>
              <li v-if="!result.keyFrames?.length">-</li>
            </ul>
            <h3>Suggestions</h3>
            <ul>
              <li v-for="suggestion in result.suggestions || []" :key="suggestion">{{ suggestion }}</li>
            </ul>
            <h3>Raw</h3>
            <pre>{{ result.rawText || '-' }}</pre>
          </button>
        </div>
      </n-tab-pane>
    </n-tabs>
  </aside>
</template>
