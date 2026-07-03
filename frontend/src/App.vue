<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { darkTheme, NIcon, type SelectOption } from 'naive-ui'
import {
  AnalyticsOutline,
  DownloadOutline,
  PauseOutline,
  PlayOutline,
  RefreshOutline,
  StopOutline,
  TrashOutline
} from '@vicons/ionicons5'

import LogDetails from '@/components/LogDetails.vue'
import LogTable from '@/components/LogTable.vue'
import { useLogStore } from '@/stores/logs'

const store = useLogStore()

const deviceOptions = computed<SelectOption[]>(() =>
  store.devices.map((device) => ({
    label: [
      device.model || device.serial,
      device.state !== 'device' ? `(${device.state})` : '',
      device.androidVersion ? `Android ${device.androidVersion}` : ''
    ]
      .filter(Boolean)
      .join(' '),
    value: device.serial
  }))
)

const levelOptions: SelectOption[] = ['V', 'D', 'I', 'W', 'E', 'F'].map((level) => ({
  label: level,
  value: level
}))

const packageOptions = computed<SelectOption[]>(() =>
  store.packages.map((item) => ({
    label: item.label ? `${item.label} (${item.packageName})` : item.packageName,
    value: item.packageName
  }))
)

const packageModeOptions: SelectOption[] = [
  { label: '3rd party', value: 'thirdParty' },
  { label: 'All packages', value: 'all' }
]

function handleDeviceChange(value: string | number | null) {
  void store.selectDevice(typeof value === 'string' ? value : null)
}

function handlePackageChange(value: string | number | null) {
  void store.selectPackage(typeof value === 'string' ? value : null)
}

function handlePackageModeChange(value: string | number | null) {
  void store.setPackageMode(value === 'all' ? 'all' : 'thirdParty')
}

onMounted(() => {
  void store.refreshDevices()
})

onUnmounted(() => {
  store.stopPolling()
})
</script>

<template>
  <n-config-provider :theme="darkTheme">
    <n-dialog-provider>
      <n-message-provider>
        <main class="app-shell">
          <header class="toolbar">
            <div class="brand">
              <div class="brand-mark">C</div>
              <div>
                <h1>CatScope</h1>
                <span>Android Logcat Workbench</span>
              </div>
            </div>

            <n-select
              :value="store.selectedSerial"
              class="device-select"
              :options="deviceOptions"
              :loading="store.loading"
              placeholder="No Android device"
              @update:value="handleDeviceChange"
            />

            <n-select
              :value="store.packageMode"
              class="package-mode-select"
              :options="packageModeOptions"
              :disabled="!store.canSelectPackage"
              @update:value="handlePackageModeChange"
            />

            <n-select
              :value="store.selectedPackage || null"
              class="package-select"
              :options="packageOptions"
              :loading="store.packageLoading"
              :disabled="!store.canSelectPackage"
              clearable
              filterable
              placeholder="All packages"
              @update:value="handlePackageChange"
            />

            <n-button :loading="store.loading" tertiary @click="store.refreshDevices">
              <template #icon>
                <n-icon :component="RefreshOutline" />
              </template>
            </n-button>

            <n-button type="primary" :disabled="!store.canStart" @click="store.start">
              <template #icon>
                <n-icon :component="PlayOutline" />
              </template>
              Start
            </n-button>

            <n-button :disabled="!store.running" @click="store.stop">
              <template #icon>
                <n-icon :component="StopOutline" />
              </template>
              Stop
            </n-button>

            <n-button :disabled="!store.running" @click="store.togglePause">
              <template #icon>
                <n-icon :component="PauseOutline" />
              </template>
              {{ store.paused ? 'Resume' : 'Pause' }}
            </n-button>

            <n-button tertiary @click="store.clear">
              <template #icon>
                <n-icon :component="TrashOutline" />
              </template>
              Clear
            </n-button>

            <n-button tertiary :disabled="store.filteredLogs.length === 0" @click="store.exportFiltered">
              <template #icon>
                <n-icon :component="DownloadOutline" />
              </template>
              Export
            </n-button>

            <n-button tertiary :disabled="store.filteredLogs.length === 0" @click="store.analyzeCurrentLogs">
              <template #icon>
                <n-icon :component="AnalyticsOutline" />
              </template>
              Analyze
            </n-button>

            <n-input v-model:value="store.search" clearable placeholder="Search log text" />

            <n-select
              v-model:value="store.levels"
              class="level-select"
              multiple
              :options="levelOptions"
              placeholder="Level"
            />
          </header>

          <section class="workspace">
            <aside class="device-panel">
              <h2>Device</h2>
              <dl v-if="store.selectedDevice">
                <dt>Serial</dt>
                <dd>{{ store.selectedDevice.serial }}</dd>
                <dt>State</dt>
                <dd>{{ store.selectedDevice.state }}</dd>
                <dt>Model</dt>
                <dd>{{ store.selectedDevice.model || 'Unknown' }}</dd>
                <dt>Android</dt>
                <dd>{{ store.selectedDevice.androidVersion || 'Unknown' }}</dd>
                <dt>SDK</dt>
                <dd>{{ store.selectedDevice.sdkVersion || 'Unknown' }}</dd>
                <dt>ABI</dt>
                <dd>{{ store.selectedDevice.abi || 'Unknown' }}</dd>
                <dt>Package</dt>
                <dd>{{ store.selectedPackage || 'All logs' }}</dd>
                <dt>Current PID</dt>
                <dd>{{ store.currentPIDs.length ? store.currentPIDs.join(', ') : '未检测到运行中的进程' }}</dd>
                <dt>Known PID</dt>
                <dd>{{ store.knownPIDs.length ? store.knownPIDs.join(', ') : '-' }}</dd>
              </dl>
              <p v-else class="empty-copy">Connect a device and refresh.</p>
              <n-alert v-if="store.deviceHint" class="device-alert" type="warning" :show-icon="false">
                {{ store.deviceHint }}
              </n-alert>
              <n-alert v-if="store.packageHint" class="device-alert" type="info" :show-icon="false">
                {{ store.packageHint }}
              </n-alert>
            </aside>

            <LogTable />

            <LogDetails />
          </section>

          <footer class="statusbar">
            <span>Device: {{ store.selectedSerial || 'none' }}</span>
            <span>Visible: {{ store.filteredLogs.length }}</span>
            <span>Package: {{ store.selectedPackage || 'all' }}</span>
            <span>PID: {{ store.currentPIDs.length ? store.currentPIDs.join(',') : 'none' }}</span>
            <span>Issues: {{ store.analysisResults.length }}</span>
            <span>Buffered: {{ store.status.count }}</span>
            <span>Dropped: {{ store.status.discardedCount }}</span>
            <span>Status: {{ store.running ? (store.paused ? 'paused' : 'streaming') : 'stopped' }}</span>
            <span v-if="store.error" class="status-error">{{ store.error }}</span>
            <span v-else-if="store.notice" class="status-notice">{{ store.notice }}</span>
          </footer>
        </main>
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>
