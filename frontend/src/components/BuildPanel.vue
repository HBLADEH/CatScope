<script setup lang="ts">
import { computed } from 'vue'
import { ChevronDownOutline, FolderOpenOutline, PlayOutline } from '@vicons/ionicons5'
import { NIcon, type DropdownOption } from 'naive-ui'
import { t } from '@/i18n'
import { useLogStore } from '@/stores/logs'

const store = useLogStore()
const busy = computed(() => store.buildLoading || store.installLoading || store.launchLoading)
const canRun = computed(() => store.canBuildProject && store.canUseDeviceActions && Boolean(store.launchPackageName))
const runHint = computed(() => !store.canBuildProject ? t('workbench.needProject') : !store.canUseDeviceActions ? store.deviceHint || t('error.selectDevice') : !store.launchPackageName ? t('workbench.needPackage') : '')
const phase = computed(() => store.buildLoading ? t('workbench.building') : store.installLoading ? t('workbench.installing') : store.launchLoading ? t('workbench.launching') : store.launchResult?.success ? t('workbench.appLaunched') : '')
const actions = computed<DropdownOption[]>(() => [
  { label: t('project.buildDebug'), key: 'build', disabled: busy.value || !store.canBuildProject },
  { label: t('project.findApk'), key: 'find', disabled: busy.value || !store.canBuildProject },
  { label: t('project.installApk'), key: 'install', disabled: busy.value || !store.canInstallAPK },
  { label: t('project.buildInstall'), key: 'build-install', disabled: busy.value || !store.canBuildProject || !store.canUseDeviceActions },
  { label: t('project.launchApp'), key: 'launch', disabled: busy.value || !store.canLaunchProject }
])
function runAction(key: string | number) {
  const handlers: Record<string, () => unknown> = {
    build: () => store.buildDebug(), find: () => store.findLatestAPK(),
    install: () => store.installAPK(), 'build-install': () => store.buildAndInstall(),
    launch: () => store.launchApp()
  }
  handlers[String(key)]?.()
}
</script>

<template>
  <div class="build-panel-content">
    <div class="build-fields">
      <label class="build-project-field">{{ t('project.pathPlaceholder') }}
        <div class="project-path-row"><n-input v-model:value="store.projectConfig.projectPath" @blur="store.saveProjectConfig" /><n-button :aria-label="t('workbench.chooseProject')" @click="store.chooseProjectDirectory"><n-icon :component="FolderOpenOutline" /></n-button></div>
      </label>
      <label>{{ t('project.packagePlaceholder') }}<n-input v-model:value="store.projectConfig.packageName" :placeholder="store.selectedPackage || t('workbench.needPackage')" @blur="store.saveProjectConfig" /></label>
      <label>{{ t('project.apkPlaceholder') }}<n-input v-model:value="store.projectConfig.lastApkPath" @blur="store.saveProjectConfig" /></label>
    </div>
    <div class="build-action-row">
      <n-button type="primary" :loading="busy" :disabled="!canRun" :title="runHint" @click="store.buildInstallLaunch"><template #icon><n-icon :component="PlayOutline" /></template>{{ t('workbench.runPipeline') }}</n-button>
      <n-dropdown trigger="click" :options="actions" @select="runAction"><n-button>{{ t('workbench.moreActions') }}<n-icon :component="ChevronDownOutline" /></n-button></n-dropdown>
      <span v-if="store.selectedDevice" class="settings-hint" :title="store.selectedDevice.serial">{{ t('workbench.targetDevice', { name: store.selectedDevice.model || store.selectedDevice.serial }) }}</span>
      <span v-if="phase" class="build-phase" aria-live="polite">{{ phase }}</span>
      <span v-else-if="runHint" class="settings-hint">{{ runHint }}</span>
      <span v-else class="settings-hint">{{ t('workbench.pipelineHint') }}</span>
    </div>
    <details class="build-options">
      <summary>{{ t('workbench.installOptions') }}</summary>
      <div class="install-options">
        <n-checkbox v-model:checked="store.projectConfig.installOptions.allowDowngrade" @update:checked="store.saveProjectConfig">{{ t('workbench.allowDowngrade') }} (-d)</n-checkbox>
        <n-checkbox v-model:checked="store.projectConfig.installOptions.grantPermissions" @update:checked="store.saveProjectConfig">{{ t('workbench.grantPermissions') }} (-g)</n-checkbox>
        <n-checkbox v-model:checked="store.projectConfig.installOptions.allowTestOnly" @update:checked="store.saveProjectConfig">{{ t('workbench.allowTestOnly') }} (-t)</n-checkbox>
        <n-checkbox v-model:checked="store.autoStartLogcat" @update:checked="store.saveCurrentWorkspace">{{ t('project.autoLogcat') }}</n-checkbox>
        <n-checkbox v-model:checked="store.autoClearOnLaunch" @update:checked="store.saveCurrentWorkspace">{{ t('project.clearOnLaunch') }}</n-checkbox>
      </div>
    </details>
    <div v-if="store.buildResult || store.installResult || store.latestAPK" class="build-result-line">
      <span v-if="store.buildResult">{{ t('project.build') }}: {{ t(store.buildResult.success ? 'common.success' : 'common.failed') }}</span>
      <span v-if="store.installResult">{{ t('project.install') }}: {{ t(store.installResult.success ? 'common.success' : 'common.failed') }}</span>
      <span v-if="store.latestAPK" :title="store.projectConfig.lastApkPath">{{ store.latestAPK.fileName }}</span>
    </div>
    <details v-if="store.buildOutput || store.installOutput" class="project-output" open><summary>{{ t('project.output') }}</summary><pre v-if="store.buildOutput">{{ store.buildOutput }}</pre><pre v-if="store.installOutput">{{ store.installOutput }}</pre></details>
  </div>
</template>
