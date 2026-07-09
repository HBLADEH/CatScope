<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import {
  darkTheme,
  dateEnUS,
  dateZhCN,
  enUS,
  NIcon,
  zhCN,
  type DropdownOption,
  type SelectOption
} from 'naive-ui'
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
import { currentLocale, languageOptions, localizePresetName, setLocale, t } from '@/i18n'
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
const naiveLocale = computed(() => (currentLocale.value === 'zh-CN' ? zhCN : enUS))
const naiveDateLocale = computed(() => (currentLocale.value === 'zh-CN' ? dateZhCN : dateEnUS))

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

const packageModeOptions = computed<SelectOption[]>(() => [
  { label: t('toolbar.thirdParty'), value: 'thirdParty' },
  { label: t('toolbar.allPackages'), value: 'all' }
])

const themeOptions = computed<SelectOption[]>(() => [
  { label: t('theme.system'), value: 'system' },
  { label: t('theme.light'), value: 'light' },
  { label: t('theme.dark'), value: 'dark' }
])

const exportOptions = computed<DropdownOption[]>(() => [
  { label: t('toolbar.exportTxt'), key: 'txt' },
  { label: t('toolbar.exportJsonl'), key: 'jsonl' }
])

const allLevelsSelected = computed(() => store.levels.length === allLogLevels.length)

const levelSummary = computed(() => {
  if (store.levels.length === allLogLevels.length) {
    return t('toolbar.allLevels')
  }
  if (store.levels.length === 0) {
    return t('toolbar.noLevels')
  }
  return t('toolbar.levels', { levels: allLogLevels.filter((level) => store.levels.includes(level)).join(' ') })
})

const sourceLabel = computed(() => {
  if (store.isSession) {
    return t('source.session')
  }
  if (store.isOffline) {
    return t('source.offline')
  }
  return t('source.live')
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

function handleLocaleChange(value: string | number | null) {
  if (value === 'zh-CN' || value === 'en-US') {
    setLocale(value)
  }
}

function buildStateLabel(success?: boolean) {
  if (success === undefined) {
    return '-'
  }
  return success ? t('common.success') : t('common.failed')
}

function displayPresetName(id: string, name: string) {
  return localizePresetName(id, name)
}

function updatePresetName(presetID: string, value: string | number | null) {
  const preset = store.filterPresets.find((item) => item.id === presetID)
  if (!preset || preset.builtIn) {
    return
  }
  preset.name = typeof value === 'string' ? value : ''
}

function handlePresetNameInput(presetID: string, value: string) {
  updatePresetName(presetID, value)
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
  currentLocale,
  (locale) => {
    store.aiContextOptions.language = locale
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
  <n-config-provider :theme="naiveTheme" :locale="naiveLocale" :date-locale="naiveDateLocale">
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
                  :placeholder="t('toolbar.noDevice')"
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
                  :placeholder="t('toolbar.allPackages')"
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
                  {{ t('toolbar.start') }}
                </n-button>

                <n-button :disabled="!store.running || store.isStaticSource" @click="store.stop">
                  <template #icon>
                    <n-icon :component="StopOutline" />
                  </template>
                  {{ t('toolbar.stop') }}
                </n-button>

                <n-button class="pause-button" :disabled="!store.running || store.isStaticSource" @click="store.togglePause">
                  <template #icon>
                    <n-icon :component="PauseOutline" />
                  </template>
                  {{ store.paused ? t('toolbar.resume') : t('toolbar.pause') }}
                </n-button>

                <div class="toolbar-filter-group">
                  <n-input v-model:value="store.search" class="toolbar-search" clearable :placeholder="t('toolbar.searchPlaceholder')" />

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
                          {{ t('common.all') }}
                        </n-button>
                        <n-button size="tiny" tertiary :disabled="store.levels.length === 0" @click="clearLevels">
                          {{ t('common.none') }}
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
                </div>

                <n-button tertiary @click="store.clear">
                  <template #icon>
                    <n-icon :component="TrashOutline" />
                  </template>
                  {{ t('toolbar.clear') }}
                </n-button>

                <n-button tertiary :disabled="store.filteredLogs.length === 0" @click="store.analyzeCurrentLogs">
                  <template #icon>
                    <n-icon :component="AnalyticsOutline" />
                  </template>
                  {{ t('toolbar.analyze') }}
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
                    {{ t('toolbar.export') }}
                  </n-button>
                </n-dropdown>
              </div>
            </div>
          </header>

          <section class="workspace" :class="{ 'sidebar-collapsed': sidebarCollapsed }">
            <aside class="device-panel">
              <div class="sidebar-panel-bar">
                <span v-if="!sidebarCollapsed">{{ t('source.controls') }}</span>
                <n-button tertiary circle :title="sidebarCollapsed ? t('toolbar.expandSidebar') : t('toolbar.collapseSidebar')" @click="toggleSidebar">
                  <template #icon>
                    <n-icon :component="sidebarCollapsed ? ChevronForwardOutline : ChevronBackOutline" />
                  </template>
                </n-button>
              </div>

              <template v-if="!sidebarCollapsed">
              <section class="sidebar-settings-panel">
                <h2>{{ t('theme.appearance') }}</h2>
                <n-select
                  :value="themeMode"
                  :options="themeOptions"
                  size="small"
                  @update:value="handleThemeModeChange"
                />
                <h2>{{ t('language.label') }}</h2>
                <n-select
                  :value="currentLocale"
                  :options="languageOptions"
                  size="small"
                  @update:value="handleLocaleChange"
                />
              </section>

              <section class="source-panel">
                <h2>{{ t('source.logSource') }}</h2>
                <div class="source-mode-row">
                  <n-tag :type="sourceTagType" size="small">
                    {{ sourceLabel }}
                  </n-tag>
                  <n-button v-if="store.isStaticSource" size="small" tertiary @click="store.returnToLiveMode">
                    {{ t('source.returnLive') }}
                  </n-button>
                </div>
                <h2 class="subsection-title">{{ t('source.adbEngine') }}</h2>
                <div class="project-path-row">
                  <n-input
                    v-model:value="store.adbPathInput"
                    :placeholder="t('source.adbPathPlaceholder')"
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
                    {{ t('source.useAdb') }}
                  </n-button>
                  <n-button size="small" tertiary :disabled="!store.adbPathInput && !store.resolvedADBPath" @click="store.useADBPath('')">
                    {{ t('source.autoDetect') }}
                  </n-button>
                </div>
                <p class="adb-path-hint">
                  {{ store.resolvedADBPath ? t('common.current', { value: store.resolvedADBPath }) : t('common.currentAutoDetect') }}
                </p>
                <div class="project-path-row">
                  <n-input
                    v-model:value="store.offlinePathInput"
                    :placeholder="t('source.offlinePathPlaceholder')"
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
                    :placeholder="t('source.sessionPathPlaceholder')"
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
                  :placeholder="t('source.sessionNamePlaceholder')"
                />
                <n-input
                  v-model:value="store.sessionNotes"
                  type="textarea"
                  :autosize="{ minRows: 2, maxRows: 4 }"
                  :placeholder="t('source.sessionNotesPlaceholder')"
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
                    {{ t('source.saveSession') }}
                  </n-button>
                  <n-button size="small" tertiary :loading="store.sessionLoading" @click="store.openSession()">
                    <template #icon>
                      <n-icon :component="FolderOpenOutline" />
                    </template>
                    {{ t('source.openSession') }}
                  </n-button>
                </div>
                <dl v-if="store.isOffline" class="source-summary">
                  <dt>{{ t('source.file') }}</dt>
                  <dd>{{ store.status.offlineFileName || '-' }}</dd>
                  <dt>{{ t('source.path') }}</dt>
                  <dd>{{ store.status.offlineFilePath || '-' }}</dd>
                  <dt>{{ t('source.entries') }}</dt>
                  <dd>{{ store.status.count }}</dd>
                  <dt>{{ t('source.rawLines') }}</dt>
                  <dd>{{ store.status.offlineParseFailedCount || 0 }}</dd>
                </dl>
                <dl v-if="store.isSession || store.currentSession" class="source-summary">
                  <dt>{{ t('source.name') }}</dt>
                  <dd>{{ store.currentSession?.name || store.status.sessionName || '-' }}</dd>
                  <dt>{{ t('source.path') }}</dt>
                  <dd>{{ store.currentSession?.filePath || store.status.sessionFilePath || '-' }}</dd>
                  <dt>{{ t('source.logs') }}</dt>
                  <dd>{{ store.currentSession?.logCount ?? store.status.count }}</dd>
                  <dt>{{ t('source.analysis') }}</dt>
                  <dd>{{ store.currentSession?.analysisCount ?? store.analysisResults.length }}</dd>
                  <dt>{{ t('source.created') }}</dt>
                  <dd>{{ store.currentSession?.createdAt || '-' }}</dd>
                </dl>
              </section>

              <section class="project-panel">
                <h2>{{ t('workspace.title') }}</h2>
                <n-select
                  :value="store.activeWorkspaceID"
                  :options="store.workspaceOptions"
                  :placeholder="t('workspace.placeholder')"
                  @update:value="handleWorkspaceChange"
                />
                <n-input
                  v-model:value="store.workspaceName"
                  :placeholder="t('workspace.namePlaceholder')"
                  @blur="store.saveCurrentWorkspace"
                />
                <div class="project-actions">
                  <n-button size="small" type="primary" tertiary @click="store.saveCurrentWorkspace">
                    {{ t('workspace.save') }}
                  </n-button>
                  <n-button size="small" tertiary @click="store.createWorkspace">
                    {{ t('workspace.new') }}
                  </n-button>
                  <n-button size="small" tertiary :disabled="store.workspaces.length <= 1" @click="store.deleteCurrentWorkspace">
                    {{ t('workspace.delete') }}
                  </n-button>
                  <n-button size="small" tertiary @click="store.resetConfig">
                    {{ t('workspace.reset') }}
                  </n-button>
                </div>

                <h2>{{ t('project.title') }}</h2>
                <div class="project-path-row">
                  <n-input
                    v-model:value="store.projectConfig.projectPath"
                    :placeholder="t('project.pathPlaceholder')"
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
                  :placeholder="t('project.packagePlaceholder')"
                  @blur="store.saveProjectConfig"
                />
                <n-input
                  v-model:value="store.projectConfig.lastApkPath"
                  class="project-input"
                  :placeholder="t('project.apkPlaceholder')"
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
                    {{ t('project.autoLogcat') }}
                  </n-checkbox>
                  <n-checkbox v-model:checked="store.autoClearOnLaunch" @update:checked="store.saveCurrentWorkspace">
                    {{ t('project.clearOnLaunch') }}
                  </n-checkbox>
                </div>

                <div class="project-actions">
                  <n-button size="small" :loading="store.buildLoading" :disabled="!store.canBuildProject" @click="store.buildDebug">
                    <template #icon>
                      <n-icon :component="HammerOutline" />
                    </template>
                    {{ t('project.buildDebug') }}
                  </n-button>
                  <n-button size="small" tertiary :disabled="!store.canBuildProject" @click="store.findLatestAPK">
                    {{ t('project.findApk') }}
                  </n-button>
                  <n-button size="small" :loading="store.installLoading" :disabled="!store.canInstallAPK" @click="store.installAPK()">
                    {{ t('project.installApk') }}
                  </n-button>
                  <n-button size="small" :loading="store.buildLoading || store.installLoading" :disabled="!store.canBuildProject || !store.canUseDeviceActions" @click="store.buildAndInstall">
                    {{ t('project.buildInstall') }}
                  </n-button>
                  <n-button size="small" :loading="store.launchLoading" :disabled="!store.canLaunchProject" @click="store.launchApp">
                    <template #icon>
                      <n-icon :component="PlayOutline" />
                    </template>
                    {{ t('project.launchApp') }}
                  </n-button>
                  <n-button size="small" type="primary" :loading="store.buildLoading || store.installLoading || store.launchLoading" :disabled="!store.canBuildProject || !store.canUseDeviceActions || !store.launchPackageName" @click="store.buildInstallLaunch">
                    {{ t('project.buildInstallLaunch') }}
                  </n-button>
                </div>

                <dl class="project-summary">
                  <dt>{{ t('project.task') }}</dt>
                  <dd>{{ store.projectConfig.defaultBuildTask }}</dd>
                  <dt>{{ t('project.apk') }}</dt>
                  <dd>{{ store.latestAPK?.fileName || store.projectConfig.lastApkPath || '-' }}</dd>
                  <dt>{{ t('project.build') }}</dt>
                  <dd>{{ store.buildResult ? buildStateLabel(store.buildResult.success) : '-' }}</dd>
                  <dt>{{ t('project.install') }}</dt>
                  <dd>{{ store.installResult ? buildStateLabel(store.installResult.success) : '-' }}</dd>
                </dl>

                <details v-if="store.buildOutput || store.installOutput" class="project-output">
                  <summary>{{ t('project.output') }}</summary>
                  <pre v-if="store.buildOutput">{{ store.buildOutput }}</pre>
                  <pre v-if="store.installOutput">{{ store.installOutput }}</pre>
                </details>
              </section>

              <section class="filter-panel">
                <h2>{{ t('filter.title') }}</h2>
                <n-select
                  :value="store.selectedPresetID || null"
                  :options="store.presetOptions"
                  clearable
                  :placeholder="t('filter.applyPlaceholder')"
                  @update:value="handlePresetChange"
                />
                <n-input
                  v-model:value="store.presetDraftName"
                  class="project-input"
                  :placeholder="t('filter.namePlaceholder')"
                />
                <div class="project-actions">
                  <n-button size="small" type="primary" tertiary @click="store.saveCurrentFilter">
                    {{ t('filter.saveCurrent') }}
                  </n-button>
                  <n-button size="small" tertiary @click="store.presetManagerOpen = true">
                    {{ t('filter.manage') }}
                  </n-button>
                </div>
                <n-input
                  v-model:value="store.tagFilter"
                  class="project-input"
                  :placeholder="t('filter.tagsPlaceholder')"
                  @blur="store.saveCurrentWorkspace"
                />
                <n-input
                  v-model:value="store.excludeKeyword"
                  class="project-input"
                  :placeholder="t('filter.excludePlaceholder')"
                  @blur="store.saveCurrentWorkspace"
                />
                <n-checkbox v-model:checked="store.regexEnabled" @update:checked="store.saveCurrentWorkspace">
                  {{ t('filter.regex') }}
                </n-checkbox>
              </section>

              <h2>{{ t('device.title') }}</h2>
              <dl v-if="store.selectedDevice">
                <dt>{{ t('device.serial') }}</dt>
                <dd>{{ store.selectedDevice.serial }}</dd>
                <dt>{{ t('device.state') }}</dt>
                <dd>{{ store.selectedDevice.state }}</dd>
                <dt>{{ t('device.model') }}</dt>
                <dd>{{ store.selectedDevice.model || t('common.unknown') }}</dd>
                <dt>{{ t('device.android') }}</dt>
                <dd>{{ store.selectedDevice.androidVersion || t('common.unknown') }}</dd>
                <dt>{{ t('device.sdk') }}</dt>
                <dd>{{ store.selectedDevice.sdkVersion || t('common.unknown') }}</dd>
                <dt>{{ t('device.abi') }}</dt>
                <dd>{{ store.selectedDevice.abi || t('common.unknown') }}</dd>
                <dt>{{ t('device.package') }}</dt>
                <dd>{{ store.selectedPackage || t('device.allLogs') }}</dd>
                <dt>{{ t('device.currentPid') }}</dt>
                <dd>{{ store.currentPIDs.length ? store.currentPIDs.join(', ') : t('common.noRunningProcess') }}</dd>
                <dt>{{ t('device.knownPid') }}</dt>
                <dd>{{ store.knownPIDs.length ? store.knownPIDs.join(', ') : '-' }}</dd>
              </dl>
              <p v-else class="empty-copy">{{ t('device.connectAndRefresh') }}</p>
              <n-alert v-if="store.deviceHint" class="device-alert" type="warning" :show-icon="false">
                {{ store.deviceHint }}
              </n-alert>
              <n-alert v-if="store.packageHint" class="device-alert" type="info" :show-icon="false">
                {{ store.packageHint }}
              </n-alert>
              </template>
              <div v-else class="sidebar-collapsed-rail">
                <span :title="t('source.logSource')">SRC</span>
                <span :title="t('workspace.title')">WK</span>
                <span :title="t('filter.title')">FLT</span>
              </div>
            </aside>

            <LogTable />

            <transition name="details-slide">
              <aside v-if="detailsPanelOpen" class="details-drawer">
                <div class="details-drawer-header">
                  <strong>{{ detailsPanelTab === 'analysis' ? t('details.analysisTitle') : t('details.title') }}</strong>
                  <n-button tertiary circle size="small" :title="t('details.close')" @click="closeDetailsPanel">
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
              <span>{{ t('status.device', { value: store.selectedSerial || t('status.none') }) }}</span>
              <span>{{ t('status.source', { value: sourceLabel }) }}</span>
              <span>{{ t('status.visible', { value: store.filteredLogs.length }) }}</span>
              <span>{{ t('status.package', { value: store.selectedPackage || t('status.all') }) }}</span>
              <span>{{ t('status.pid', { value: store.currentPIDs.length ? store.currentPIDs.join(',') : t('status.none') }) }}</span>
              <span>{{ t('status.issues', { value: store.analysisResults.length }) }}</span>
              <span>{{ t('status.buffered', { value: store.status.count }) }}</span>
              <span>{{ t('status.dropped', { value: store.status.discardedCount }) }}</span>
              <span>{{ t('status.state', { value: store.running ? (store.paused ? t('status.paused') : t('status.streaming')) : t('status.stopped') }) }}</span>
              <span v-if="store.error" class="status-error">{{ store.error }}</span>
              <span v-else-if="store.notice" class="status-notice">{{ store.notice }}</span>
            </div>
            <n-button
              class="details-toggle-button"
              size="small"
              tertiary
              :type="detailsPanelOpen ? 'primary' : 'default'"
              :title="t('details.openAnalysis')"
              @click="openDetailsPanel('analysis')"
            >
              <template #icon>
                <n-icon :component="AnalyticsOutline" />
              </template>
              {{ t('details.detailsAnalysis') }}
            </n-button>
          </footer>

          <n-drawer v-model:show="store.presetManagerOpen" :width="360" placement="right">
            <n-drawer-content :title="t('filter.presets')">
              <div class="preset-manager">
                <div v-for="preset in store.filterPresets" :key="preset.id" class="preset-row">
                  <n-input
                    :value="preset.builtIn ? displayPresetName(preset.id, preset.name) : preset.name"
                    size="small"
                    :disabled="preset.builtIn"
                    @update:value="handlePresetNameInput(preset.id, $event)"
                    @blur="store.renamePreset(preset, preset.name)"
                  />
                  <n-button size="small" tertiary @click="store.applyPreset(preset.id)">
                    {{ t('common.apply') }}
                  </n-button>
                  <n-button size="small" tertiary :disabled="preset.builtIn" @click="store.deletePreset(preset.id)">
                    {{ t('common.delete') }}
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
