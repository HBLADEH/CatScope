<script setup lang="ts">
import { computed } from 'vue'

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
    <h2>Details</h2>
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
  </aside>
</template>
