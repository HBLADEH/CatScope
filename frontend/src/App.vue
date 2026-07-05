<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { darkTheme, NIcon, type SelectOption } from 'naive-ui'
import {
  AnalyticsOutline,
  DownloadOutline,
  FolderOpenOutline,
  HammerOutline,
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

function handleWorkspaceChange(value: string | number | null) {
  void store.selectWorkspace(value)
}

function handlePresetChange(value: string | number | null) {
  void store.applyPreset(value)
}

onMounted(() => {
  void store.loadConfig()
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
