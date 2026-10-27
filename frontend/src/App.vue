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
  type SelectOption,
  type GlobalThemeOverrides
} from 'naive-ui'
import {
  AnalyticsOutline,
  ChevronDownOutline,
  CloseOutline,
  FolderOpenOutline,
  HammerOutline,
  PauseOutline,
  PlayOutline,
  RefreshOutline,
  StopOutline,
  TrashOutline,
  SettingsOutline,
  OptionsOutline
} from '@vicons/ionicons5'

import BuildPanel from '@/components/BuildPanel.vue'
import logoURL from '@/assets/logo.svg'
import LogDetails from '@/components/LogDetails.vue'
import LogTable from '@/components/LogTable.vue'
import { currentLocale, languageOptions, localizePresetName, setLocale, t } from '@/i18n'
import { useLogStore } from '@/stores/logs'
import type { UpdateInfo } from '@/types/backend'
import { backend } from '@/utils/wails'

const store = useLogStore()
const settingsOpen = ref(false)
const settingsTab = ref('general')
const buildPanelOpen = ref(false)
const buildPanelHeight = ref(220)
const logTableRef = ref<InstanceType<typeof LogTable> | null>(null)
const relatedFocusIDs = ref<number[]>([])
const saveSessionOpen = ref(false)
const saveSessionError = ref('')
const saveAsNewPath = ref(false)
let stopBuildResize: (() => void) | undefined
const detailsPanelOpen = ref(false)
const detailsPanelTab = ref<'details' | 'analysis'>('details')
const allLogLevels = ['V', 'D', 'I', 'W', 'E', 'F']
type ThemeMode = 'system' | 'light' | 'dark'

function readBooleanSetting(key: string, defaultValue: boolean) {
  const value = localStorage.getItem(key)
  return value === null ? defaultValue : value === 'true'
}

const appVersion = ref('')
const autoCheckUpdates = ref(readBooleanSetting('catscope.autoCheckUpdates', true))
const includePrereleaseUpdates = ref(readBooleanSetting('catscope.includePrereleaseUpdates', false))
const checkingUpdates = ref(false)
const installingUpdate = ref(false)
const updateInfo = ref<UpdateInfo | null>(null)
const updateError = ref('')
const updateModalOpen = ref(false)
let updateCheckTimer: number | undefined

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

async function checkForUpdates(manual = true) {
  if (checkingUpdates.value) {
    return
  }
  checkingUpdates.value = true
  updateError.value = ''
  try {
    const info = await backend.checkForUpdates(includePrereleaseUpdates.value)
    updateInfo.value = info
    appVersion.value = info.currentVersion
    if (info.available) {
      updateModalOpen.value = true
    } else if (manual) {
      updateModalOpen.value = true
    }
  } catch (error) {
    updateError.value = error instanceof Error ? error.message : String(error)
    if (manual) {
      updateModalOpen.value = true
    }
  } finally {
    checkingUpdates.value = false
  }
}

async function installAvailableUpdate() {
  if (installingUpdate.value) {
    return
  }
  installingUpdate.value = true
  updateError.value = ''
  try {
    await backend.installUpdate(includePrereleaseUpdates.value)
  } catch (error) {
    updateError.value = error instanceof Error ? error.message : String(error)
    installingUpdate.value = false
  }
}

function openReleasePage() {
  void backend.openReleasePage(updateInfo.value?.releaseUrl || '')
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
  if (window.innerWidth < 1100) buildPanelOpen.value = false
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

watch(autoCheckUpdates, (enabled) => {
  localStorage.setItem('catscope.autoCheckUpdates', String(enabled))
})

watch(includePrereleaseUpdates, (enabled) => {
  localStorage.setItem('catscope.includePrereleaseUpdates', String(enabled))
})

onMounted(() => {
  const media = window.matchMedia?.('(prefers-color-scheme: dark)')
  if (media) {
    systemPrefersDark.value = media.matches
    media.addEventListener('change', handleSystemThemeChange)
  }
  window.addEventListener('keydown', handleWorkbenchKeyDown)
  window.addEventListener('resize', handleViewportResize)
  void store.loadConfig()
  void store.refreshDevices()
  void backend.getAppVersion().then((version) => {
    appVersion.value = version
  }).catch(() => { appVersion.value = '' })
  if (autoCheckUpdates.value) {
    updateCheckTimer = window.setTimeout(() => {
      void checkForUpdates(false)
    }, 1500)
  }
})

onUnmounted(() => {
  window.matchMedia?.('(prefers-color-scheme: dark)').removeEventListener('change', handleSystemThemeChange)
  window.removeEventListener('keydown', handleWorkbenchKeyDown)
  window.removeEventListener('resize', handleViewportResize)
  stopBuildResize?.()
  store.stopPolling()
  if (updateCheckTimer !== undefined) {
    window.clearTimeout(updateCheckTimer)
  }
})

const themeOverrides = computed<GlobalThemeOverrides>(() => {
  const dark = activeThemeName.value === 'dark'
  return {
    common: {
      primaryColor: dark ? '#00C4DA' : '#007F95',
      primaryColorHover: dark ? '#38D5E5' : '#0095AC',
      primaryColorPressed: dark ? '#00A6BA' : '#00697C',
      primaryColorSuppl: dark ? '#00C4DA' : '#007F95',
      bodyColor: dark ? '#0B1420' : '#F3F7FB',
      cardColor: dark ? '#122131' : '#FFFFFF',
      modalColor: dark ? '#122131' : '#FFFFFF',
      popoverColor: dark ? '#162A3D' : '#FFFFFF',
      inputColor: dark ? '#0E1B29' : '#FFFFFF',
      borderColor: dark ? '#294054' : '#D7E2EC',
      textColorBase: dark ? '#DCE8F2' : '#203247',
      textColor1: dark ? '#EDF5FC' : '#082B49',
      textColor2: dark ? '#CBD9E5' : '#334B60',
      textColor3: dark ? '#8EA6BB' : '#60758A',
      borderRadius: '6px',
      fontSize: '13px',
      heightMedium: '32px',
      heightSmall: '28px'
    },
    Button: {
      textColorPrimary: dark ? '#06283D' : '#FFFFFF',
      textColorHoverPrimary: dark ? '#06283D' : '#FFFFFF',
      textColorPressedPrimary: dark ? '#06283D' : '#FFFFFF'
    },
    Tag: {
      textColorError: dark ? '#FFC2CC' : '#A61F38',
      colorError: dark ? '#4A2536' : '#FBE7EC'
    }
  }
})
const sourceOptions = computed<DropdownOption[]>(() => [
  { label: t('source.live'), key: 'live' },
  { label: t('workbench.openFile'), key: 'file' },
  { label: t('source.openSession'), key: 'session' }
])
const fileOptions = computed<DropdownOption[]>(() => [
  { label: t('source.saveSession'), key: 'save', disabled: store.logs.length === 0 },
  { label: t('workbench.saveAs'), key: 'save-as', disabled: store.logs.length === 0 },
  { type: 'divider', key: 'file-divider' },
  ...exportOptions.value.map(option => ({ ...option, disabled: store.filteredLogs.length === 0 }))
])
const activeFilterCount = computed(() => Number(Boolean(store.tagFilter.trim())) + Number(Boolean(store.excludeKeyword.trim())) + Number(store.regexEnabled))
const sourceFileName = computed(() => store.isSession
  ? store.currentSession?.name || store.status.sessionName || t('source.session')
  : store.status.offlineFileName || t('source.offline'))
const sourceFilePath = computed(() => store.isSession
  ? store.sessionPathInput || store.status.sessionFilePath
  : store.status.offlineFilePath)
const sourceCaption = computed(() => store.isStaticSource ? sourceLabel.value : t('workbench.liveShort'))
const captureState = computed(() => store.running
  ? (store.paused ? t('status.paused') : t('status.streaming'))
  : t('status.stopped'))

function handleSourceSelect(key: string | number) {
  relatedFocusIDs.value = []
  if (key === 'file') void store.openLogFile('')
  else if (key === 'session') void store.openSession('')
  else if (store.isStaticSource) void store.returnToLiveMode()
}
function handleFileSelect(key: string | number) {
  if (key === 'save' || key === 'save-as') {
    saveAsNewPath.value = key === 'save-as'
    saveSessionError.value = ''
    saveSessionOpen.value = true
  } else handleExportSelect(key)
}
async function saveSessionAndClose() {
  saveSessionError.value = ''
  await store.saveSession(saveAsNewPath.value ? '' : store.sessionPathInput)
  saveSessionError.value = store.error
  if (!saveSessionError.value) saveSessionOpen.value = false
}
function toggleBuildPanel() {
  buildPanelOpen.value = !buildPanelOpen.value
  if (buildPanelOpen.value && window.innerWidth < 1100) detailsPanelOpen.value = false
}
function openSettings(tab = 'general') {
  settingsTab.value = tab
  settingsOpen.value = true
}
function clearAdvancedFilters() {
  store.tagFilter = ''
  store.excludeKeyword = ''
  store.regexEnabled = false
  void store.saveCurrentWorkspace()
}
function resizeBuildPanel(delta: number) {
  const maximum = Math.max(140, Math.min(440, window.innerHeight - 330))
  buildPanelHeight.value = Math.max(140, Math.min(maximum, buildPanelHeight.value + delta))
}
function startBuildResize(event: PointerEvent) {
  if (event.button !== 0) return
  event.preventDefault()
  stopBuildResize?.()
  let lastY = event.clientY
  const move = (current: PointerEvent) => {
    resizeBuildPanel(lastY - current.clientY)
    lastY = current.clientY
  }
  const stop = () => {
    window.removeEventListener('pointermove', move)
    window.removeEventListener('pointerup', stop)
    window.removeEventListener('pointercancel', stop)
    stopBuildResize = undefined
  }
  stopBuildResize = stop
  window.addEventListener('pointermove', move)
  window.addEventListener('pointerup', stop, { once: true })
  window.addEventListener('pointercancel', stop, { once: true })
}
const visibleLogCount = computed(() => relatedFocusIDs.value.length
  ? store.logs.filter(entry => relatedFocusIDs.value.includes(entry.id)).length
  : store.filteredLogs.length)
watch(
  [() => store.search, () => store.levels.join(','), () => store.tagFilter, () => store.excludeKeyword, () => store.regexEnabled, () => store.selectedPackage, () => store.selectedPresetID, () => store.status.source, () => store.logs.length === 0],
  () => { relatedFocusIDs.value = [] }
)
async function locateRelatedLog(id: number) {
  if (relatedFocusIDs.value.length) relatedFocusIDs.value = store.selectedAnalysis?.relatedEntryIds || [id]
  await logTableRef.value?.scrollToLog(id)
}
async function showRelatedLogs(ids: number[]) {
  relatedFocusIDs.value = [...ids]
  if (ids[0] !== undefined) await logTableRef.value?.scrollToLog(ids[0])
}
function handleViewportResize() {
  resizeBuildPanel(0)
  if (window.innerWidth < 1100 && detailsPanelOpen.value && buildPanelOpen.value) buildPanelOpen.value = false
}
function handleWorkbenchKeyDown(event: KeyboardEvent) {
  if (event.key !== 'Escape' || event.defaultPrevented || settingsOpen.value || saveSessionOpen.value || updateModalOpen.value || store.presetManagerOpen) return
  if (detailsPanelOpen.value) closeDetailsPanel()
  else buildPanelOpen.value = false
}
</script>

<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="themeOverrides" :locale="naiveLocale" :date-locale="naiveDateLocale">
    <n-dialog-provider>
      <n-message-provider>
        <main class="app-shell">
          <header class="workbench-header">
            <div class="context-bar">
              <div class="brand"><img class="brand-logo" :src="logoURL" alt="" /><h1>CatScope</h1></div>
              <n-select class="workspace-select" :value="store.activeWorkspaceID" :options="store.workspaceOptions" :placeholder="t('workspace.placeholder')" :aria-label="t('workspace.title')" @update:value="handleWorkspaceChange" />
              <span class="context-divider"></span>
              <n-dropdown trigger="click" :options="sourceOptions" @select="handleSourceSelect">
                <n-button class="source-switch" secondary><n-icon :component="FolderOpenOutline" /><span class="source-label">{{ sourceCaption }}</span><n-icon :component="ChevronDownOutline" /></n-button>
              </n-dropdown>
              <template v-if="!store.isStaticSource">
                <n-select class="device-select" :value="store.selectedSerial || null" :options="deviceOptions" :loading="store.loading" :placeholder="t('toolbar.noDevice')" :aria-label="t('device.title')" @update:value="handleDeviceChange" />
                <n-select class="package-select" :value="store.selectedPackage || null" :options="packageOptions" :loading="store.packageLoading" :disabled="!store.canSelectPackage" clearable filterable :placeholder="t('toolbar.allPackages')" :aria-label="t('details.package')" @update:value="handlePackageChange" />
                <n-button quaternary circle :loading="store.loading" :aria-label="t('workbench.refreshDevices')" :title="t('workbench.refreshDevices')" @click="store.refreshDevices"><n-icon :component="RefreshOutline" /></n-button>
              </template>
              <span v-else class="source-file-name" :title="sourceFilePath || sourceFileName">{{ sourceFileName }}</span>
              <div class="context-spacer"></div>
              <n-button class="build-toggle" :secondary="buildPanelOpen" :type="buildPanelOpen ? 'primary' : 'default'" :aria-pressed="buildPanelOpen" aria-controls="build-dock" @click="toggleBuildPanel"><template #icon><n-icon :component="HammerOutline" /></template>{{ t('workbench.buildRun') }}<span v-if="store.buildLoading || store.installLoading || store.launchLoading" class="running-dot"></span></n-button>
              <n-dropdown trigger="click" :options="fileOptions" @select="handleFileSelect"><n-button quaternary>{{ t('workbench.fileMenu') }}<n-icon :component="ChevronDownOutline" /></n-button></n-dropdown>
              <n-button quaternary circle :aria-label="t('workbench.settings')" :title="t('workbench.settings')" @click="openSettings()"><n-icon :component="SettingsOutline" /></n-button>
            </div>
            <div class="query-bar">
              <div v-if="!store.isStaticSource" class="capture-controls">
                <n-button v-if="!store.running" type="primary" :disabled="!store.canStart" :title="!store.canStart ? store.deviceHint || t('error.selectDevice') : ''" @click="store.start"><template #icon><n-icon :component="PlayOutline" /></template>{{ t('toolbar.start') }}</n-button>
                <template v-else>
                  <n-button type="primary" @click="store.togglePause"><template #icon><n-icon :component="store.paused ? PlayOutline : PauseOutline" /></template>{{ store.paused ? t('toolbar.resume') : t('toolbar.pause') }}</n-button>
                  <n-button quaternary circle :aria-label="t('toolbar.stop')" :title="t('toolbar.stop')" @click="store.stop"><n-icon :component="StopOutline" /></n-button>
                </template>
              </div>
              <n-input v-model:value="store.search" class="toolbar-search" clearable :aria-label="t('workbench.search')" :placeholder="t('workbench.searchPlaceholder')" />
              <n-select class="preset-select" :value="store.selectedPresetID || null" :options="store.presetOptions" clearable :placeholder="t('filter.applyPlaceholder')" :aria-label="t('filter.presets')" @update:value="handlePresetChange" />
              <n-popover trigger="click" placement="bottom-end" raw>
                <template #trigger><n-button class="level-filter-button"><span>{{ levelSummary }}</span><n-icon :component="ChevronDownOutline" /></n-button></template>
                <div class="level-filter-popover">
                  <div class="level-filter-actions"><n-button size="tiny" tertiary :disabled="allLevelsSelected" @click="setAllLevels">{{ t('common.all') }}</n-button><n-button size="tiny" tertiary :disabled="store.levels.length === 0" @click="clearLevels">{{ t('common.none') }}</n-button></div>
                  <div class="level-filter-grid"><n-checkbox v-for="level in allLogLevels" :key="level" :checked="store.levels.includes(level)" @update:checked="(checked: boolean) => toggleLevel(level, checked)"><span :class="`level level-${level}`">{{ level }}</span></n-checkbox></div>
                </div>
              </n-popover>
              <n-popover trigger="click" placement="bottom-end" raw>
                <template #trigger><n-button :type="activeFilterCount ? 'primary' : 'default'" :secondary="activeFilterCount > 0"><template #icon><n-icon :component="OptionsOutline" /></template>{{ t('workbench.filters') }}<span v-if="activeFilterCount" class="filter-count">{{ activeFilterCount }}</span></n-button></template>
                <div class="advanced-filters">
                  <label>{{ t('filter.tagsPlaceholder') }}<n-input v-model:value="store.tagFilter" clearable @blur="store.saveCurrentWorkspace" /></label>
                  <label>{{ t('filter.excludePlaceholder') }}<n-input v-model:value="store.excludeKeyword" clearable @blur="store.saveCurrentWorkspace" /></label>
                  <n-checkbox v-model:checked="store.regexEnabled" @update:checked="store.saveCurrentWorkspace">{{ t('filter.regex') }}</n-checkbox>
                  <label v-if="!store.isStaticSource">{{ t('workbench.packageScope') }}<n-select :value="store.packageMode" :options="packageModeOptions" :disabled="!store.canSelectPackage" @update:value="handlePackageModeChange" /></label>
                  <div class="panel-inline-actions"><n-button size="small" @click="store.presetManagerOpen = true">{{ t('filter.manage') }}</n-button><n-button size="small" :disabled="activeFilterCount === 0" @click="clearAdvancedFilters">{{ t('workbench.clearAdvanced') }}</n-button></div>
                </div>
              </n-popover>
              <n-button class="issues-toggle" :type="detailsPanelOpen && detailsPanelTab === 'analysis' ? 'primary' : 'default'" :secondary="detailsPanelOpen && detailsPanelTab === 'analysis'" @click="openDetailsPanel('analysis')"><template #icon><n-icon :component="AnalyticsOutline" /></template>{{ t('workbench.issues', { count: store.analysisResults.length }) }}</n-button>
              <n-button quaternary circle :disabled="store.logs.length === 0" :aria-label="t('toolbar.clear')" :title="t('toolbar.clear')" @click="store.clear"><n-icon :component="TrashOutline" /></n-button>
            </div>
            <div v-if="activeFilterCount" class="active-filters">
              <span>{{ t('workbench.activeFilters') }}</span>
              <n-tag v-if="store.tagFilter" size="small" closable @close="store.tagFilter = ''; store.saveCurrentWorkspace()">Tag: {{ store.tagFilter }}</n-tag>
              <n-tag v-if="store.excludeKeyword" size="small" closable @close="store.excludeKeyword = ''; store.saveCurrentWorkspace()">{{ t('workbench.exclude') }}: {{ store.excludeKeyword }}</n-tag>
              <n-tag v-if="store.regexEnabled" size="small" closable @close="store.regexEnabled = false; store.saveCurrentWorkspace()">Regex</n-tag>
            </div>
          </header>

          <section class="workbench-body">
            <section class="workspace" :class="{ 'details-open': detailsPanelOpen }">
              <LogTable ref="logTableRef" :focus-entry-ids="relatedFocusIDs" @select="openDetailsPanel('details')" @return-filter="relatedFocusIDs = []" />
              <transition name="details-slide">
                <aside v-if="detailsPanelOpen" class="details-drawer">
                  <LogDetails :default-tab="detailsPanelTab" :related-logs-shown="relatedFocusIDs.length > 0" @tab-change="detailsPanelTab = $event" @close="closeDetailsPanel" @locate="locateRelatedLog" @show-related="showRelatedLogs" />
                </aside>
              </transition>
            </section>
            <transition name="build-slide">
              <section v-if="buildPanelOpen" id="build-dock" class="build-dock" :style="{ height: `${buildPanelHeight}px` }">
                <div class="dock-resize" role="separator" aria-orientation="horizontal" :aria-label="t('workbench.resizeBuild')" tabindex="0" @pointerdown="startBuildResize" @keydown.up.prevent="resizeBuildPanel(20)" @keydown.down.prevent="resizeBuildPanel(-20)"></div>
                <div class="dock-header"><strong><n-icon :component="HammerOutline" />{{ t('workbench.buildRun') }}</strong><span>{{ store.projectConfig.defaultBuildTask }}</span><n-button quaternary circle size="small" :aria-label="t('workbench.closeBuild')" @click="buildPanelOpen = false"><n-icon :component="CloseOutline" /></n-button></div>
                <BuildPanel />
              </section>
            </transition>
          </section>

          <footer class="statusbar">
            <div class="statusbar-info">
              <span class="capture-status" :class="{ active: store.running && !store.paused }"><i></i>{{ store.isStaticSource ? sourceLabel : captureState }}</span>
              <span>{{ t('status.visible', { value: visibleLogCount }) }} / {{ store.status.count }}</span>
              <span v-if="store.currentPIDs.length">PID {{ store.currentPIDs.join(', ') }}</span>
              <span v-if="store.status.discardedCount" class="status-warning">{{ t('status.dropped', { value: store.status.discardedCount }) }}</span>
              <span v-if="store.error" class="status-error" :title="store.error">{{ store.error }}</span>
              <span v-else-if="store.notice" class="status-notice" :title="store.notice">{{ store.notice }}</span>
            </div>
            <button v-if="!store.isStaticSource && store.deviceHint" class="link-button connection-link" @click="openSettings('connection')">{{ t('workbench.connectionSettings') }}</button>
          </footer>

          <n-drawer v-model:show="settingsOpen" :width="520" placement="right">
            <n-drawer-content :title="t('workbench.settings')" closable>
              <n-tabs v-model:value="settingsTab" type="line" animated>
                <n-tab-pane name="general" :tab="t('workbench.general')">
                  <div class="settings-form">
                    <label>{{ t('theme.appearance') }}<n-select :value="themeMode" :options="themeOptions" @update:value="handleThemeModeChange" /></label>
                    <label>{{ t('language.label') }}<n-select :value="currentLocale" :options="languageOptions" @update:value="handleLocaleChange" /></label>
                  </div>
                </n-tab-pane>
                <n-tab-pane name="connection" :tab="t('device.title')">
                  <div class="settings-form">
                    <label>{{ t('source.adbEngine') }}<div class="project-path-row"><n-input v-model:value="store.adbPathInput" :placeholder="t('source.adbPathPlaceholder')" @keyup.enter="store.useADBPath()" /><n-button :loading="store.loading" :aria-label="t('workbench.chooseADB')" @click="store.chooseADBExecutable"><n-icon :component="FolderOpenOutline" /></n-button></div></label>
                    <div class="panel-inline-actions"><n-button type="primary" :loading="store.loading" @click="store.useADBPath()">{{ t('source.useAdb') }}</n-button><n-button :loading="store.loading" @click="store.useADBPath('')">{{ t('source.autoDetect') }}</n-button></div>
                    <p class="settings-hint">{{ store.resolvedADBPath || t('common.currentAutoDetect') }}</p>
                    <n-alert v-if="store.deviceHint" type="warning" :show-icon="false">{{ store.deviceHint }}</n-alert>
                    <n-button @click="store.refreshDevices">{{ t('workbench.refreshDevices') }}</n-button>
                    <dl v-if="store.selectedDevice" class="settings-device-info">
                      <dt>{{ t('device.serial') }}</dt><dd>{{ store.selectedDevice.serial }}</dd>
                      <dt>{{ t('device.model') }}</dt><dd>{{ store.selectedDevice.model || '-' }}</dd>
                      <dt>{{ t('device.android') }}</dt><dd>{{ store.selectedDevice.androidVersion || '-' }}</dd>
                      <dt>{{ t('device.sdk') }}</dt><dd>{{ store.selectedDevice.sdkVersion || '-' }}</dd>
                      <dt>{{ t('device.abi') }}</dt><dd>{{ store.selectedDevice.abi || '-' }}</dd>
                      <dt>{{ t('device.currentPid') }}</dt><dd>{{ store.currentPIDs.join(', ') || '-' }}</dd>
                    </dl>
                  </div>
                </n-tab-pane>
                <n-tab-pane name="workspace" :tab="t('workspace.title')">
                  <div class="settings-form">
                    <label>{{ t('workspace.title') }}<n-select :value="store.activeWorkspaceID" :options="store.workspaceOptions" @update:value="handleWorkspaceChange" /></label>
                    <label>{{ t('workspace.namePlaceholder') }}<n-input v-model:value="store.workspaceName" @blur="store.saveCurrentWorkspace" /></label>
                    <div class="panel-inline-actions"><n-button type="primary" @click="store.saveCurrentWorkspace">{{ t('workspace.save') }}</n-button><n-button @click="store.createWorkspace">{{ t('workspace.new') }}</n-button><n-button :disabled="store.workspaces.length <= 1" @click="store.deleteCurrentWorkspace">{{ t('workspace.delete') }}</n-button></div>
                    <details class="settings-advanced"><summary>{{ t('workbench.advanced') }}</summary><n-button tertiary @click="store.resetConfig">{{ t('workspace.reset') }}</n-button></details>
                  </div>
                </n-tab-pane>
                <n-tab-pane name="updates" :tab="t('update.title')">
                  <div class="settings-form"><span>{{ t('update.currentVersion', { version: appVersion || '-' }) }}</span><n-checkbox v-model:checked="autoCheckUpdates">{{ t('update.autoCheck') }}</n-checkbox><n-checkbox v-model:checked="includePrereleaseUpdates">{{ t('update.includePrerelease') }}</n-checkbox><n-button :loading="checkingUpdates" @click="checkForUpdates(true)">{{ t('update.checkNow') }}</n-button></div>
                </n-tab-pane>
              </n-tabs>
            </n-drawer-content>
          </n-drawer>

          <n-drawer v-model:show="store.presetManagerOpen" :width="380" placement="right">
            <n-drawer-content :title="t('filter.presets')" closable>
              <div class="settings-form"><label>{{ t('filter.namePlaceholder') }}<n-input v-model:value="store.presetDraftName" /></label><n-button type="primary" @click="store.saveCurrentFilter">{{ t('filter.saveCurrent') }}</n-button></div>
              <div class="preset-manager">
                <div v-for="preset in store.filterPresets" :key="preset.id" class="preset-row">
                  <n-input :value="preset.builtIn ? displayPresetName(preset.id, preset.name) : preset.name" size="small" :disabled="preset.builtIn" @update:value="handlePresetNameInput(preset.id, $event)" @blur="store.renamePreset(preset, preset.name)" />
                  <n-button size="small" tertiary @click="store.applyPreset(preset.id)">{{ t('common.apply') }}</n-button>
                  <n-button size="small" tertiary :disabled="preset.builtIn" @click="store.deletePreset(preset.id)">{{ t('common.delete') }}</n-button>
                </div>
              </div>
            </n-drawer-content>
          </n-drawer>

          <n-modal v-model:show="saveSessionOpen" preset="card" class="session-modal" :title="t('source.saveSession')">
            <div class="settings-form"><label>{{ t('source.sessionNamePlaceholder') }}<n-input v-model:value="store.sessionNameInput" /></label><label>{{ t('source.sessionNotesPlaceholder') }}<n-input v-model:value="store.sessionNotes" type="textarea" :autosize="{ minRows: 3, maxRows: 6 }" /></label><n-alert v-if="saveSessionError" type="error" :show-icon="false">{{ saveSessionError }}</n-alert></div>
            <template #footer><div class="panel-inline-actions"><n-button @click="saveSessionOpen = false">{{ t('workbench.cancel') }}</n-button><n-button type="primary" :loading="store.sessionLoading" @click="saveSessionAndClose">{{ t('source.saveSession') }}</n-button></div></template>
          </n-modal>
          <n-modal v-model:show="updateModalOpen" preset="card" class="update-modal" :title="t('update.dialogTitle')">
            <n-alert v-if="updateError" type="error" :show-icon="false">
              {{ updateError }}
            </n-alert>

            <template v-else-if="updateInfo?.available">
              <p class="update-version-line">
                {{ t('update.available', { current: updateInfo.currentVersion, latest: updateInfo.latestVersion }) }}
              </p>
              <n-tag v-if="updateInfo.prerelease" type="warning" size="small">Preview</n-tag>
              <pre v-if="updateInfo.releaseNotes" class="update-release-notes">{{ updateInfo.releaseNotes }}</pre>
              <p v-if="updateInfo.autoInstallHint" class="update-hint">{{ updateInfo.autoInstallHint }}</p>
            </template>

            <p v-else-if="updateInfo">{{ t('update.upToDate', { version: updateInfo.currentVersion }) }}</p>
            <p v-else-if="checkingUpdates">{{ t('update.checking') }}</p>

            <template #footer>
              <div class="update-dialog-actions">
                <n-button @click="updateModalOpen = false">{{ t('update.later') }}</n-button>
                <n-button v-if="updateInfo?.available" tertiary @click="openReleasePage">
                  {{ t('update.viewRelease') }}
                </n-button>
                <n-button
                  v-if="updateInfo?.canAutoInstall"
                  type="primary"
                  :loading="installingUpdate"
                  @click="installAvailableUpdate"
                >
                  {{ installingUpdate ? t('update.installing') : t('update.installAndRestart') }}
                </n-button>
              </div>
            </template>
          </n-modal>
        </main>
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>
