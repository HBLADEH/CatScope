# CatScope MVP Tasks

本文档用于指导 Codex 或其他 AI Agent 按阶段实现 CatScope MVP。

## 开发总原则

- 先实现可运行的最小闭环。
- 后端优先保证 ADB、Logcat、Parser 稳定。
- 前端优先保证日志显示、搜索、暂停、清屏、导出可用。
- 不做完整 IDE。
- 不做复杂 Gradle Sync。
- 不在第一阶段引入过多抽象。
- 每完成一个任务都要能构建、运行、测试。

---

## Phase 0：项目初始化

### 0.1 初始化 Wails 项目

目标：创建 Go + Wails + Vue 3 + TypeScript 项目骨架，并引入主流 Vue UI 开发栈。

任务：

- 初始化 Wails v2 项目。
- 前端选择 Vue 3 + TypeScript + Vite。
- 引入 Pinia 作为状态管理。
- 引入 Naive UI 作为主 UI 组件库。
- 引入 VueUse 作为组合式工具库。
- 引入 Iconify / unplugin-icons 或兼容方案作为图标方案。
- 引入 `@tanstack/vue-virtual` 或 `vue-virtual-scroller` 用于日志虚拟滚动。
- 确认 Windows 环境可以运行 `wails dev`。
- 整理目录结构。
- 添加基础 README。

验收：

- `wails dev` 可以启动应用。
- 前端页面能调用 Go 后端方法。

---

### 0.2 建立基础目录

创建目录：

```text
internal/adb
internal/logcat
internal/build
internal/workspace
internal/storage
internal/ai
docs
```

验收：

- 目录结构与 ARCHITECTURE.md 一致。
- 项目可以正常编译。

---

## Phase 1：ADB 设备管理

### 1.1 实现 ADB 路径查找

查找顺序：

1. 用户配置 adbPath
2. ANDROID_HOME / ANDROID_SDK_ROOT
3. PATH
4. 用户手动选择

任务：

- 实现 `FindADB()`。
- 支持 Windows 路径。
- 支持校验 `adb version`。

验收：

- 能找到本机 adb。
- 找不到时返回明确错误。

---

### 1.2 实现设备列表

命令：

```bash
adb devices -l
```

任务：

- 解析设备 serial。
- 解析设备 state。
- 返回设备列表给前端。

验收：

- UI 可以显示设备列表。
- 支持刷新设备。

---

### 1.3 获取设备信息

命令：

```bash
adb -s <serial> shell getprop ro.product.model
adb -s <serial> shell getprop ro.build.version.release
adb -s <serial> shell getprop ro.build.version.sdk
adb -s <serial> shell getprop ro.product.cpu.abi
```

任务：

- 实现 `GetDeviceInfo(serial)`。
- 在 UI 中展示 model、Android version、SDK、ABI。

验收：

- 选择设备后可以展示设备详情。

---

## Phase 2：Logcat 流和解析

### 2.1 启动 Logcat 流

命令：

```bash
adb -s <serial> logcat -v threadtime -b main,system,crash
```

任务：

- 实现 Logcat 进程启动。
- 捕获 stdout。
- 捕获 stderr。
- 支持停止进程。
- 支持重复启动前自动停止旧进程。

验收：

- 点击 Start 可以看到实时日志。
- 点击 Stop 可以停止日志。
- ADB 报错时 UI 有提示。

---

### 2.2 实现 threadtime Parser

目标结构：

```ts
type LogEntry = {
  id: number
  timestamp: string
  pid: number
  tid: number
  level: "V" | "D" | "I" | "W" | "E" | "F"
  tag: string
  message: string
  packageName?: string
  raw: string
  multiline?: string[]
}
```

任务：

- 解析标准 threadtime 行。
- 解析失败时保留 raw。
- 非标准行追加到上一条日志 multiline。

验收：

- 普通日志可正确解析字段。
- Java stacktrace 不会散成完全无关的多条日志。

---

### 2.3 实现 Ring Buffer

任务：

- 默认最多保留 100000 行。
- 超过后丢弃旧日志。
- 记录 discardedCount。
- 提供按范围读取。
- 提供读取新增日志。

验收：

- 大量日志下内存不会无限增长。
- UI 可以显示已丢弃日志数量。

---

## Phase 3：前端日志 UI

### 3.1 基础布局

实现：

```text
顶部工具栏
左侧过滤器区
中间日志表格
右侧详情面板
底部状态栏
```

验收：

- UI 基础结构完整。
- 可以选择设备、开始、停止、暂停、清屏。

---

### 3.2 日志表格

字段：

```text
时间 | Level | PID | TID | Tag | Message
```

任务：

- 使用虚拟滚动。
- 不直接渲染所有日志 DOM。
- Level 使用不同样式高亮。
- 点击日志行显示详情。

验收：

- 10 万行日志下 UI 仍可操作。
- 日志可以自动滚动到底部。
- 暂停后不自动滚动。

---

### 3.3 搜索和过滤

任务：

- 关键词搜索。
- Level 过滤。
- Tag 过滤。
- 正则开关。
- 搜索结果高亮。

验收：

- 可以快速查找关键词。
- Level 过滤结果正确。
- 正则错误时 UI 给出提示，不崩溃。

---

### 3.4 清屏、暂停、导出

任务：

- Pause：暂停 UI 自动刷新，但后台可以继续缓存。
- Clear：清空当前 UI / 当前 session。
- Export：导出 txt。

验收：

- 暂停后日志不继续滚动。
- 恢复后可以继续显示。
- 导出的 txt 内容正确。

---

## Phase 4：包名过滤和 PID 追踪

### 4.1 获取包列表

命令：

```bash
adb -s <serial> shell pm list packages
```

任务：

- 拉取已安装包名。
- 支持前端搜索包名。

验收：

- UI 可以选择目标包名。

---

### 4.2 PID 追踪

命令：

```bash
adb -s <serial> shell pidof <package>
```

任务：

- 根据包名获取当前 PID。
- 定期刷新 PID。
- App 重启后自动更新 PID。
- 历史旧 PID 日志保留。

验收：

- App 重启后仍能继续过滤目标包日志。

---

## Phase 5：基础分析器

### 5.1 Java Crash Analyzer

识别：

```text
FATAL EXCEPTION
AndroidRuntime
Caused by:
NullPointerException
ClassNotFoundException
NoClassDefFoundError
SecurityException
UnsatisfiedLinkError
```

任务：

- 标记 Java Crash。
- 提取异常类型。
- 提取崩溃线程。
- 提取第一条业务栈。

验收：

- Java 崩溃日志可以在 UI 中突出显示。

---

### 5.2 Native Crash Analyzer

识别：

```text
signal 11
SIGSEGV
SIGABRT
backtrace:
tombstone
libxxx.so
JNI DETECTED ERROR
```

任务：

- 标记 Native Crash。
- 提取 signal。
- 提取 so 名称。
- 提取 backtrace。

验收：

- Native 崩溃日志可以在 UI 中突出显示。

---

### 5.3 Install Error Analyzer

识别：

```text
INSTALL_FAILED_VERSION_DOWNGRADE
INSTALL_FAILED_UPDATE_INCOMPATIBLE
INSTALL_FAILED_NO_MATCHING_ABIS
INSTALL_FAILED_INVALID_APK
INSTALL_PARSE_FAILED_NO_CERTIFICATES
INSTALL_PARSE_FAILED_MANIFEST_MALFORMED
```

任务：

- 对安装错误给出中文解释。
- 给出建议操作。

验收：

- 安装失败时用户可以看到明确原因和建议。

---

## Phase 6：AI Context Generator

### 6.1 复制上下文

任务：

- 选中一条日志。
- 提取前后 N 行。
- 生成 Markdown。
- 复制到剪贴板。

验收：

- 用户可以直接粘贴给 AI Agent。

---

### 6.2 生成 Bug Report

模板：

```markdown
# Android Logcat 问题分析

## 设备信息

- Device:
- Android Version:
- SDK:
- ABI:
- Serial:

## App 信息

- Package:
- PID:
- Build Variant:

## 错误摘要

## 关键日志

```log
...
```

## 上下文日志

```log
...
```
```

验收：

- 生成内容结构完整。
- 包含设备信息、包名、关键日志和上下文。

---

## Phase 7：Build / Install / Launch

### 7.1 Build Debug

任务：

- 选择 Android 项目目录。
- 检测 `gradlew.bat` / `gradlew`。
- 执行 `assembleDebug`。
- 捕获输出。

验收：

- 可以在 Windows 项目中执行 Debug 构建。

---

### 7.2 APK Finder

任务：

- 在 `build/outputs/apk` 下查找最新 APK。
- 支持 module 路径。
- 返回 APK 路径。

验收：

- 构建成功后可以找到正确 APK。

---

### 7.3 Install APK

命令：

```bash
adb -s <serial> install -r <apk-path>
```

验收：

- 可以把 APK 安装到当前设备。
- 安装失败时展示错误分析。

---

### 7.4 Launch App

MVP 命令：

```bash
adb -s <serial> shell monkey -p <package> 1
```

验收：

- 可以启动目标 App。
- 启动后自动切换到该包名过滤。

---

## 最小可交付版本定义

MVP 必须至少包含：

- 设备选择
- 实时 Logcat
- 日志解析
- 日志高亮
- 基础搜索
- 基础过滤
- 暂停 / 清屏
- 虚拟滚动
- 导出日志
- 包名过滤
- PID 自动追踪
- Java Crash 识别
- AI 上下文复制

当以上能力稳定后，再进入 Build / Install / Launch 阶段。
