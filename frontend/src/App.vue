<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { darkTheme, NIcon, type DropdownOption, type SelectOption } from 'naive-ui'
import {
  AnalyticsOutline,
  ChevronBackOutline,
  ChevronDownOutline,
  ChevronForwardOutline,
  CloseOutline,
  DownloadOutline,
  FolderOpenOutline,
  HammerOutline,
  PauseOutline,
  PlayOutline,
  RefreshOutline,
  SaveOutline,
  StopOutline,
  TrashOutline
} from '@vicons/ionicons5'

import LogDetails from '@/components/LogDetails.vue'
import LogTable from '@/components/LogTable.vue'
import { useLogStore } from '@/stores/logs'

const store = useLogStore()
const sidebarCollapsed = ref(false)
const detailsPanelOpen = ref(false)
const detailsPanelTab = ref<'details' | 'analysis'>('details')
const allLogLevels = ['V', 'D', 'I', 'W', 'E', 'F']
type ThemeMode = 'system' | 'light' | 'dark'

function readThemeMode(): ThemeMode {
  const value = localStorage.getItem('catscope.themeMode')
  return value === 'light' || value === 'dark' || value === 'system' ? value : 'system'
}

const themeMode = ref<ThemeMode>(readThemeMode())
const systemPrefersDark = ref(window.matchMedia?.('(prefers-color-scheme: dark)').matches ?? true)

const activeThemeName = computed(() => {
  if (themeMode.value === 'system') {
    return systemPrefersDark.value ? 'dark' : 'light'
  }
  return themeMode.value
})

const naiveTheme = computed(() => (activeThemeName.value === 'dark' ? darkTheme : null))

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

const themeOptions: SelectOption[] = [
  { label: 'System', value: 'system' },
  { label: 'Light', value: 'light' },
  { label: 'Dark', value: 'dark' }
]

const exportOptions: DropdownOption[] = [
  { label: 'Export TXT', key: 'txt' },
  { label: 'Export JSONL', key: 'jsonl' }
]

const allLevelsSelected = computed(() => store.levels.length === allLogLevels.length)

const levelSummary = computed(() => {
  if (store.levels.length === allLogLevels.length) {
    return 'All Levels'
  }
  if (store.levels.length === 0) {
    return 'No Levels'
  }
  return `Levels: ${allLogLevels.filter((level) => store.levels.includes(level)).join(' ')}`
})

const sourceLabel = computed(() => {
  if (store.isSession) {
    return 'Session'
  }
  if (store.isOffline) {
    return 'Offline Log File'
  }
  return 'Live Device Logcat'
})

const sourceTagType = computed(() => {
  if (store.isSession) {
    return 'info'
  }
  if (store.isOffline) {
    return 'warning'
  }
  return 'success'
})

function handleDeviceChange(value: string | number | null) {
  void store.selectDevice(typeof value === 'string' ? value : null)
}

function handlePackageChange(value: string | number | null) {
  void store.selectPackage(typeof value === 'string' ? value : null)
}

function handlePackageModeChange(value: string | number | null) {
  void store.setPackageMode(value === 'all' ? 'all' : 'thirdParty')
}

function handleWorkspaceChange(value: string | number | null) {
  void store.selectWorkspace(value)
}

function handlePresetChange(value: string | number | null) {
  void store.applyPreset(value)
}

function handleExportSelect(key: string | number) {
  if (key === 'jsonl') {
    void store.exportFilteredJSONL()
    return
  }
  void store.exportFiltered()
}

function setAllLevels() {
  store.levels = [...allLogLevels]
}

function clearLevels() {
  store.levels = []
}

function toggleLevel(level: string, checked: boolean) {
  if (checked && !store.levels.includes(level)) {
    store.levels = [...store.levels, level]
    return
  }
  if (!checked) {
    store.levels = store.levels.filter((item) => item !== level)
  }
}

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

function setThemeMode(mode: ThemeMode) {
  themeMode.value = mode
}

function handleThemeModeChange(value: string | number | null) {
  if (value === 'light' || value === 'dark' || value === 'system') {
    setThemeMode(value)
  }
}

function openDetailsPanel(tab: 'details' | 'analysis' = 'details') {
  detailsPanelTab.value = tab
  detailsPanelOpen.value = true
}

function closeDetailsPanel() {
  detailsPanelOpen.value = false
}

function handleSystemThemeChange(event: MediaQueryListEvent) {
  systemPrefersDark.value = event.matches
}

watch(
  themeMode,
  (mode) => {
    localStorage.setItem('catscope.themeMode', mode)
  },
  { immediate: true }
)

watch(
  activeThemeName,
  (theme) => {
    document.documentElement.dataset.theme = theme
  },
  { immediate: true }
)

watch(
  () => store.selectedLog?.id,
  (id) => {
    if (id !== undefined) {
      openDetailsPanel('details')
    }
  }
)

onMounted(() => {
  const media = window.matchMedia?.('(prefers-color-scheme: dark)')
  if (media) {
    systemPrefersDark.value = media.matches
    media.addEventListener('change', handleSystemThemeChange)
  }
  void store.loadConfig()
  void store.refreshDevices()
})

onUnmounted(() => {
  window.matchMedia?.('(prefers-color-scheme: dark)').removeEventListener('change', handleSystemThemeChange)
  store.stopPolling()
})
</script>

<template>
  <n-config-provider :theme="naiveTheme">
    <n-dialog-provider>
      <n-message-provider>
        <main class="app-shell">
          <header class="toolbar">
            <div class="brand">
              <div class="brand-mark">C</div>
              <h1>CatScope</h1>
            </div>

            <div class="toolbar-controls">
              <div class="toolbar-device-group">
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
              </div>

              <div class="toolbar-action-group">
                <n-button type="primary" :disabled="!store.canStart" @click="store.start">
                  <template #icon>
                    <n-icon :component="PlayOutline" />
                  </template>
                  Start
                </n-button>

                <n-button :disabled="!store.running || store.isStaticSource" @click="store.stop">
                  <template #icon>
                    <n-icon :component="StopOutline" />
                  </template>
                  Stop
                </n-button>

                <n-button class="pause-button" :disabled="!store.running || store.isStaticSource" @click="store.togglePause">
                  <template #icon>
                    <n-icon :component="PauseOutline" />
                  </template>
                  {{ store.paused ? 'Resume' : 'Pause' }}
                </n-button>

                <n-input v-model:value="store.search" class="toolbar-search" clearable placeholder="Search log text" />

                <n-popover trigger="click" placement="bottom-end" raw>
                  <template #trigger>
                    <n-button class="level-filter-button">
                      <span>{{ levelSummary }}</span>
                      <n-icon class="level-filter-caret" :component="ChevronDownOutline" />
                    </n-button>
                  </template>

                  <div class="level-filter-popover">
                    <div class="level-filter-actions">
                      <n-button size="tiny" tertiary :disabled="allLevelsSelected" @click="setAllLevels">
                        All
                      </n-button>
                      <n-button size="tiny" tertiary :disabled="store.levels.length === 0" @click="clearLevels">
                        None
                      </n-button>
                    </div>
                    <div class="level-filter-grid">
                      <n-checkbox
                        v-for="level in allLogLevels"
                        :key="level"
                        :checked="store.levels.includes(level)"
                        @update:checked="(checked: boolean) => toggleLevel(level, checked)"
                      >
                        <span :class="`level level-${level}`">{{ level }}</span>
                      </n-checkbox>
                    </div>
                  </div>
                </n-popover>

                <n-button tertiary @click="store.clear">
                  <template #icon>
                    <n-icon :component="TrashOutline" />
                  </template>
                  Clear
                </n-button>

                <n-button tertiary :disabled="store.filteredLogs.length === 0" @click="store.analyzeCurrentLogs">
                  <template #icon>
                    <n-icon :component="AnalyticsOutline" />
                  </template>
                  Analyze
                </n-button>

                <n-dropdown
                  trigger="click"
                  :options="exportOptions"
                  :disabled="store.filteredLogs.length === 0"
                  @select="handleExportSelect"
                >
                  <n-button class="export-button" tertiary :disabled="store.filteredLogs.length === 0">
                    <template #icon>
                      <n-icon :component="DownloadOutline" />
                    </template>
                    Export
                  </n-button>
                </n-dropdown>
              </div>
            </div>
          </header>

          <section class="workspace" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
            <aside class="device-panel">
              <div class="sidebar-panel-bar">
                <span v-if="!sidebarCollapsed">Controls</span>
                <n-button tertiary circle :title="sidebarCollapsed ? '展开左侧栏' : '收起左侧栏'" @click="toggleSidebar">
                  <template #icon>
                    <n-icon :component="sidebarCollapsed ? ChevronForwardOutline : ChevronBackOutline" />
                  </template>
                </n-button>
              </div>

              <template v-if="!sidebarCollapsed">
              <section class="sidebar-settings-panel">
                <h2>Appearance</h2>
                <n-select
                  :value="themeMode"
                  :options="themeOptions"
                  size="small"
                  @update:value="handleThemeModeChange"
                />
              </section>

              <section class="source-panel">
                <h2>Log Source</h2>
                <div class="source-mode-row">
                  <n-tag :type="sourceTagType" size="small">
                    {{ sourceLabel }}
                  </n-tag>
                  <n-button v-if="store.isStaticSource" size="small" tertiary @click="store.returnToLiveMode">
                    Return to Live Mode
                  </n-button>
                </div>
                <h2 class="subsection-title">ADB Engine</h2>
                <div class="project-path-row">
                  <n-input
                    v-model:value="store.adbPathInput"
                    placeholder="Custom adb.exe path, or leave empty for auto"
                    @keyup.enter="store.useADBPath()"
                  />
                  <n-button :loading="store.loading" tertiary @click="store.chooseADBExecutable">
                    <template #icon>
                      <n-icon :component="FolderOpenOutline" />
                    </template>
                  </n-button>
                </div>
                <div class="project-actions">
                  <n-button size="small" type="primary" tertiary :loading="store.loading" @click="store.useADBPath()">
                    Use ADB
                  </n-button>
                  <n-button size="small" tertiary :disabled="!store.adbPathInput && !store.resolvedADBPath" @click="store.useADBPath('')">
                    Auto Detect
                  </n-button>
                </div>
                <p class="adb-path-hint">
                  {{ store.resolvedADBPath ? `Current: ${store.resolvedADBPath}` : 'Current: auto-detect on refresh' }}
                </p>
                <div class="project-path-row">
                  <n-input
                    v-model:value="store.offlinePathInput"
                    placeholder="Path to .txt, .log, or .jsonl"
                    @keyup.enter="store.openLogFile()"
                  />
                  <n-button :loading="store.offlineLoading" tertiary @click="store.openLogFile()">
                    <template #icon>
                      <n-icon :component="FolderOpenOutline" />
                    </template>
                  </n-button>
                </div>
                <div class="project-path-row">
                  <n-input
                    v-model:value="store.sessionPathInput"
                    placeholder="Path to .catscope-session"
                    @keyup.enter="store.openSession()"
                  />
                  <n-button :loading="store.sessionLoading" tertiary @click="store.openSession()">
                    <template #icon>
                      <n-icon :component="FolderOpenOutline" />
                    </template>
                  </n-button>
                </div>
                <n-input
                  v-model:value="store.sessionNameInput"
                  placeholder="Session name"
                />
                <n-input
                  v-model:value="store.sessionNotes"
                  type="textarea"
                  :autosize="{ minRows: 2, maxRows: 4 }"
                  placeholder="Session notes"
                />
                <div class="project-actions">
                  <n-button
                    size="small"
                    type="primary"
                    tertiary
                    :loading="store.sessionLoading"
                    :disabled="store.logs.length === 0"
                    @click="store.saveSession()"
                  >
                    <template #icon>
                      <n-icon :component="SaveOutline" />
                    </template>
                    Save Session
                  </n-button>
                  <n-button size="small" tertiary :loading="store.sessionLoading" @click="store.openSession()">
                    <template #icon>
                      <n-icon :component="FolderOpenOutline" />
                    </template>
                    Open Session
                  </n-button>
                </div>
                <dl v-if="store.isOffline" class="source-summary">
                  <dt>File</dt>
                  <dd>{{ store.status.offlineFileName || '-' }}</dd>
                  <dt>Path</dt>
                  <dd>{{ store.status.offlineFilePath || '-' }}</dd>
                  <dt>Entries</dt>
                  <dd>{{ store.status.count }}</dd>
                  <dt>Raw Lines</dt>
                  <dd>{{ store.status.offlineParseFailedCount || 0 }}</dd>
                </dl>
                <dl v-if="store.isSession || store.currentSession" class="source-summary">
                  <dt>Name</dt>
                  <dd>{{ store.currentSession?.name || store.status.sessionName || '-' }}</dd>
                  <dt>Path</dt>
                  <dd>{{ store.currentSession?.filePath || store.status.sessionFilePath || '-' }}</dd>
                  <dt>Logs</dt>
                  <dd>{{ store.currentSession?.logCount ?? store.status.count }}</dd>
                  <dt>Analysis</dt>
                  <dd>{{ store.currentSession?.analysisCount ?? store.analysisResults.length }}</dd>
                  <dt>Created</dt>
                  <dd>{{ store.currentSession?.createdAt || '-' }}</dd>
                </dl>
              </section>

              <section class="project-panel">
                <h2>Workspace</h2>
                <n-select
                  :value="store.activeWorkspaceID"
                  :options="store.workspaceOptions"
                  placeholder="Workspace"
                  @update:value="handleWorkspaceChange"
                />
                <n-input
                  v-model:value="store.workspaceName"
                  placeholder="Workspace name"
                  @blur="store.saveCurrentWorkspace"
                />
                <div class="project-actions">
                  <n-button size="small" type="primary" tertiary @click="store.saveCurrentWorkspace">
                    Save Workspace
                  </n-button>
                  <n-button size="small" tertiary @click="store.createWorkspace">
                    New
                  </n-button>
                  <n-button size="small" tertiary :disabled="store.workspaces.length <= 1" @click="store.deleteCurrentWorkspace">
                    Delete
                  </n-button>
                  <n-button size="small" tertiary @click="store.resetConfig">
                    Reset Config
                  </n-button>
                </div>

                <h2>Project</h2>
                <div class="project-path-row">
                  <n-input
                    v-model:value="store.projectConfig.projectPath"
                    placeholder="Android project path"
                    @blur="store.saveProjectConfig"
                  />
                  <n-button tertiary @click="store.chooseProjectDirectory">
                    <template #icon>
                      <n-icon :component="FolderOpenOutline" />
                    </template>
                  </n-button>
                </div>
                <n-input
                  v-model:value="store.projectConfig.packageName"
                  class="project-input"
                  placeholder="Package name"
                  @blur="store.saveProjectConfig"
                />
                <n-input
                  v-model:value="store.projectConfig.lastApkPath"
                  class="project-input"
                  placeholder="APK path"
                  @blur="store.saveProjectConfig"
                />

                <div class="install-options">
                  <n-checkbox v-model:checked="store.projectConfig.installOptions.allowDowngrade" @update:checked="store.saveProjectConfig">
                    -d
                  </n-checkbox>
                  <n-checkbox v-model:checked="store.projectConfig.installOptions.grantPermissions" @update:checked="store.saveProjectConfig">
                    -g
                  </n-checkbox>
                  <n-checkbox v-model:checked="store.projectConfig.installOptions.allowTestOnly" @update:checked="store.saveProjectConfig">
                    -t
                  </n-checkbox>
                </div>
                <div class="install-options">
                  <n-checkbox v-model:checked="store.autoStartLogcat" @update:checked="store.saveCurrentWorkspace">
                    Auto Logcat
                  </n-checkbox>
                  <n-checkbox v-model:checked="store.autoClearOnLaunch" @update:checked="store.saveCurrentWorkspace">
                    Clear on Launch
                  </n-checkbox>
                </div>

                <div class="project-actions">
                  <n-button size="small" :loading="store.buildLoading" :disabled="!store.canBuildProject" @click="store.buildDebug">
                    <template #icon>
                      <n-icon :component="HammerOutline" />
                    </template>
                    Build Debug
                  </n-button>
                  <n-button size="small" tertiary :disabled="!store.canBuildProject" @click="store.findLatestAPK">
                    Find APK
                  </n-button>
                  <n-button size="small" :loading="store.installLoading" :disabled="!store.canInstallAPK" @click="store.installAPK()">
                    Install APK
                  </n-button>
                  <n-button size="small" :loading="store.buildLoading || store.installLoading" :disabled="!store.canBuildProject || !store.canUseDeviceActions" @click="store.buildAndInstall">
                    Build + Install
                  </n-button>
                  <n-button size="small" :loading="store.launchLoading" :disabled="!store.canLaunchProject" @click="store.launchApp">
                    <template #icon>
                      <n-icon :component="PlayOutline" />
                    </template>
                    Launch App
                  </n-button>
                  <n-button size="small" type="primary" :loading="store.buildLoading || store.installLoading || store.launchLoading" :disabled="!store.canBuildProject || !store.canUseDeviceActions || !store.launchPackageName" @click="store.buildInstallLaunch">
                    Build + Install + Launch
                  </n-button>
                </div>

                <dl class="project-summary">
                  <dt>Task</dt>
                  <dd>{{ store.projectConfig.defaultBuildTask }}</dd>
                  <dt>APK</dt>
                  <dd>{{ store.latestAPK?.fileName || store.projectConfig.lastApkPath || '-' }}</dd>
                  <dt>Build</dt>
                  <dd>{{ store.buildResult ? (store.buildResult.success ? 'success' : 'failed') : '-' }}</dd>
                  <dt>Install</dt>
                  <dd>{{ store.installResult ? (store.installResult.success ? 'success' : 'failed') : '-' }}</dd>
                </dl>

                <details v-if="store.buildOutput || store.installOutput" class="project-output">
                  <summary>Output</summary>
                  <pre v-if="store.buildOutput">{{ store.buildOutput }}</pre>
                  <pre v-if="store.installOutput">{{ store.installOutput }}</pre>
                </details>
              </section>

              <section class="filter-panel">
                <h2>Filter Preset</h2>
                <n-select
                  :value="store.selectedPresetID || null"
                  :options="store.presetOptions"
                  clearable
                  placeholder="Apply preset"
                  @update:value="handlePresetChange"
                />
                <n-input
                  v-model:value="store.presetDraftName"
                  class="project-input"
                  placeholder="Preset name"
                />
                <div class="project-actions">
                  <n-button size="small" type="primary" tertiary @click="store.saveCurrentFilter">
                    Save Current Filter
                  </n-button>
                  <n-button size="small" tertiary @click="store.presetManagerOpen = true">
                    Manage Presets
                  </n-button>
                </div>
                <n-input
                  v-model:value="store.tagFilter"
                  class="project-input"
                  placeholder="Tags, comma separated"
                  @blur="store.saveCurrentWorkspace"
                />
                <n-input
                  v-model:value="store.excludeKeyword"
                  class="project-input"
                  placeholder="Exclude keyword"
                  @blur="store.saveCurrentWorkspace"
                />
                <n-checkbox v-model:checked="store.regexEnabled" @update:checked="store.saveCurrentWorkspace">
                  Regex search
                </n-checkbox>
              </section>

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
              </template>
              <div v-else class="sidebar-collapsed-rail">
                <span title="Log Source">SRC</span>
                <span title="Workspace">WK</span>
                <span title="Filter Preset">FLT</span>
              </div>
            </aside>

            <LogTable />

            <transition name="details-slide">
              <aside v-if="detailsPanelOpen" class="details-drawer">
                <div class="details-drawer-header">
                  <strong>{{ detailsPanelTab === 'analysis' ? 'Analysis' : 'Log Details' }}</strong>
                  <n-button tertiary circle size="small" title="关闭右侧栏" @click="closeDetailsPanel">
                    <template #icon>
                      <n-icon :component="CloseOutline" />
                    </template>
                  </n-button>
                </div>
                <LogDetails :default-tab="detailsPanelTab" />
              </aside>
            </transition>
          </section>

          <footer class="statusbar">
            <div class="statusbar-info">
              <span>Device: {{ store.selectedSerial || 'none' }}</span>
              <span>Source: {{ store.logSource }}</span>
              <span>Visible: {{ store.filteredLogs.length }}</span>
              <span>Package: {{ store.selectedPackage || 'all' }}</span>
              <span>PID: {{ store.currentPIDs.length ? store.currentPIDs.join(',') : 'none' }}</span>
              <span>Issues: {{ store.analysisResults.length }}</span>
              <span>Buffered: {{ store.status.count }}</span>
              <span>Dropped: {{ store.status.discardedCount }}</span>
              <span>Status: {{ store.running ? (store.paused ? 'paused' : 'streaming') : 'stopped' }}</span>
              <span v-if="store.error" class="status-error">{{ store.error }}</span>
              <span v-else-if="store.notice" class="status-notice">{{ store.notice }}</span>
            </div>
            <n-button
              class="details-toggle-button"
              size="small"
              tertiary
              :type="detailsPanelOpen ? 'primary' : 'default'"
              title="打开右侧分析栏"
              @click="openDetailsPanel('analysis')"
            >
              <template #icon>
                <n-icon :component="AnalyticsOutline" />
              </template>
              Details / Analysis
            </n-button>
          </footer>

          <n-drawer v-model:show="store.presetManagerOpen" :width="360" placement="right">
            <n-drawer-content title="Filter Presets">
              <div class="preset-manager">
                <div v-for="preset in store.filterPresets" :key="preset.id" class="preset-row">
                  <n-input
                    v-model:value="preset.name"
                    size="small"
                    :disabled="preset.builtIn"
                    @blur="store.renamePreset(preset, preset.name)"
                  />
                  <n-button size="small" tertiary @click="store.applyPreset(preset.id)">
                    Apply
                  </n-button>
                  <n-button size="small" tertiary :disabled="preset.builtIn" @click="store.deletePreset(preset.id)">
                    Delete
                  </n-button>
                </div>
              </div>
            </n-drawer-content>
          </n-drawer>
        </main>
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>
