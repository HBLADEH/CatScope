# CatScope Architecture

本文档描述 CatScope 的技术架构、模块边界和关键实现策略。

## 1. 技术选型

首选方案：

```text
桌面框架：Wails v2
后端语言：Go
前端框架：Vue 3 + TypeScript
构建工具：Vite
ADB 调用：Go exec.Command 调用本地 adb
构建调用：Go exec.Command 调用 gradlew / gradlew.bat
本地存储：SQLite + JSON 配置文件
日志渲染：虚拟滚动列表
```

## 2. 总体架构

```text
┌─────────────────────────────┐
│          Frontend            │
│ Vue 3 + TypeScript + Vite    │
│                              │
│ - 日志主界面                 │
│ - 过滤器面板                 │
│ - 设备选择器                 │
│ - 构建安装面板               │
│ - 崩溃分析面板               │
│ - AI 上下文生成面板          │
└──────────────┬──────────────┘
               │ Wails Bridge
┌──────────────▼──────────────┐
│            Go Core           │
│                              │
│ - ADB Manager                │
│ - Logcat Stream Manager      │
│ - Log Parser                 │
│ - Filter Engine              │
│ - Build Runner               │
│ - Install Runner             │
│ - Launch Runner              │
│ - Session Storage            │
│ - Exporter                   │
│ - Analyzer                   │
└──────────────┬──────────────┘
               │ exec.Command
┌──────────────▼──────────────┐
│       Android Toolchain      │
│                              │
│ - adb                        │
│ - gradlew / gradlew.bat      │
│ - Android SDK                │
└─────────────────────────────┘
```

## 前端框架选型

CatScope 前端使用 Vue 3 生态。建议组合如下：

```text
Vue 3：主前端框架
TypeScript：类型约束
Vite：开发和构建
Pinia：全局状态管理
Naive UI：主 UI 组件库
VueUse：常用组合式工具函数
Iconify / unplugin-icons：图标方案
@tanstack/vue-virtual 或 vue-virtual-scroller：日志虚拟滚动
```

选择 Naive UI 的原因：

- Vue 3 原生支持较好。
- 暗色主题和桌面工具风格适配较好。
- 表格、表单、弹窗、抽屉、标签页、下拉选择、通知组件比较完整。
- 适合快速实现设备选择、过滤器面板、日志详情面板、错误提示等界面。

日志主表格不要完全依赖普通 DataTable 渲染海量日志。大量日志场景下，应优先使用专门的虚拟滚动列表实现行渲染，再组合 Naive UI 的按钮、选择器、输入框、弹窗、通知等组件。

## 3. 后端模块

### 3.1 ADB Manager

职责：

- 查找 adb 路径
- 执行 adb 命令
- 获取设备列表
- 获取设备基础信息
- 管理多设备 serial

常用命令：

```bash
adb devices -l
adb -s <serial> get-state
adb -s <serial> shell getprop ro.product.model
adb -s <serial> shell getprop ro.build.version.release
adb -s <serial> shell getprop ro.build.version.sdk
adb -s <serial> shell getprop ro.product.cpu.abi
adb -s <serial> shell pm list packages
adb -s <serial> shell pidof <package>
```

设备结构：

```ts
type AndroidDevice = {
  serial: string
  state: "device" | "offline" | "unauthorized" | "unknown"
  model?: string
  brand?: string
  androidVersion?: string
  sdkVersion?: string
  abi?: string
  isEmulator?: boolean
}
```

### 3.2 Logcat Stream Manager

职责：

- 启动 Logcat 流
- 停止 Logcat 流
- 重启 Logcat 流
- 捕获 stdout / stderr
- 处理 ADB 断开
- 自动重连
- 支持 buffer 选择

推荐命令：

```bash
adb -s <serial> logcat -v threadtime -b main,system,crash
```

支持 buffer：

```text
main
system
crash
events
radio
all
```

### 3.3 Log Parser

职责：

- 解析 threadtime 格式
- 解析失败时保留 raw
- 多行日志合并
- Java stacktrace 合并
- Native crash 合并
- JSON 日志格式化
- 厂商 ROM 特殊格式兼容

日志结构：

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

多行归并策略：

```text
如果当前行符合标准 Logcat 头部，则创建新的 LogEntry。
如果当前行不符合标准头部，则追加到上一条 LogEntry 的 multiline 字段。
```

### 3.4 Filter Engine

过滤分三层：

1. 采集层过滤：减少 ADB 输出量。
2. 解析层过滤：Go 后端按字段过滤。
3. UI 层过滤：前端临时搜索和高亮。

过滤器结构：

```json
{
  "name": "Native Crash",
  "levels": ["E", "F"],
  "tags": ["AndroidRuntime", "DEBUG", "libc"],
  "messageRegex": "SIGSEGV|SIGABRT|tombstone|backtrace|JNI DETECTED ERROR",
  "packageName": "com.example.app",
  "excludeRegex": "Choreographer|OpenGLRenderer"
}
```

### 3.5 Ring Buffer

职责：

- 限制内存占用
- 保留最近 N 行日志
- 支持丢弃旧日志计数
- 支持前端增量拉取

建议默认：

```text
maxLogLines = 100000
```

### 3.6 Analyzer

内置分析器：

- Java Crash Analyzer
- Native Crash Analyzer
- ANR Analyzer
- Install Error Analyzer

Java Crash 关键词：

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

Native Crash 关键词：

```text
signal 11
SIGSEGV
SIGABRT
backtrace:
tombstone
libxxx.so
JNI DETECTED ERROR
```

ANR 关键词：

```text
ANR in
Input dispatching timed out
BroadcastQueue
executing service
Application Not Responding
```

### 3.7 Build Runner

职责：

- 选择项目根目录
- 检测 Gradle Wrapper
- 执行 assembleDebug
- 捕获构建输出
- 定位 APK 文件

Windows 命令：

```bat
gradlew.bat assembleDebug
```

Unix 命令：

```bash
./gradlew assembleDebug
```

### 3.8 Install Runner

职责：

- 安装 APK
- 捕获安装输出
- 识别安装错误

基础命令：

```bash
adb -s <serial> install -r <apk-path>
```

常见错误：

```text
INSTALL_FAILED_VERSION_DOWNGRADE
INSTALL_FAILED_UPDATE_INCOMPATIBLE
INSTALL_FAILED_NO_MATCHING_ABIS
INSTALL_FAILED_INVALID_APK
INSTALL_PARSE_FAILED_NO_CERTIFICATES
INSTALL_PARSE_FAILED_MANIFEST_MALFORMED
```

### 3.9 Launch Runner

MVP 阶段优先使用 monkey 启动：

```bash
adb -s <serial> shell monkey -p <package> 1
```

后续支持解析 launcher activity：

```bash
adb -s <serial> shell cmd package resolve-activity --brief <package>
```

### 3.10 AI Context Generator

职责：

- 提取选中日志
- 提取错误前后上下文
- 提取设备信息
- 提取 App 信息
- 提取 Build / Install 输出
- 生成 Markdown 分析材料

## 4. 前端模块

### 4.1 页面布局

```text
顶部工具栏：设备选择 | 包名选择 | Level | 搜索 | 开始/暂停 | 清屏 | 导出 | 构建安装
左侧面板：过滤器列表 | 已安装应用 | 历史会话
中间区域：日志表格
右侧面板：日志详情 | 崩溃分析 | AI 上下文
底部状态栏：设备状态 | 日志数量 | 丢弃数量 | ADB 状态 | 内存占用
```

### 4.2 日志表格字段

```text
时间 | Level | PID | TID | Package | Tag | Message
```

### 4.3 快捷键

```text
Ctrl + F：搜索
Ctrl + L：清屏
Ctrl + P：暂停 / 恢复
Ctrl + S：保存当前日志
Ctrl + E：导出日志
Ctrl + R：刷新设备
Ctrl + B：Build
Ctrl + I：Install
Ctrl + Enter：Build + Install + Launch
Esc：关闭当前详情面板或搜索框
```

## 5. 性能策略

- 后端持续读取 adb stdout。
- 后端解析 LogEntry。
- 后端写入 Ring Buffer。
- 前端每 100ms - 300ms 拉取新增批次。
- 前端使用虚拟滚动渲染。
- 超出上限时丢弃旧日志并显示丢弃数量。
- 用户主动保存时再落盘。

## 6. 数据存储

### 全局配置

```json
{
  "adbPath": "",
  "androidSdkPath": "",
  "theme": "system",
  "defaultBuffer": ["main", "system", "crash"],
  "defaultLogFormat": "threadtime",
  "maxLogLines": 100000,
  "autoReconnect": true
}
```

### 工作区配置

```json
{
  "projectName": "ExampleApp",
  "projectPath": "D:/workstation/android/ExampleApp",
  "packageName": "com.example.app",
  "gradleCommand": "gradlew.bat",
  "defaultTask": "assembleDebug",
  "defaultModule": "app",
  "defaultVariant": "debug",
  "lastDeviceSerial": "",
  "filters": []
}
```

## 7. 关键风险

### 包名过滤不能只靠 PID

App 重启后 PID 会变化。需要通过 packageName 定期刷新 `pidof`，发现 PID 变化后自动更新过滤条件。

### 多行日志合并

Java stacktrace、Native crash、ANR 日志通常是多行。非标准头部日志必须追加到上一条 LogEntry。

### 日志量过大

必须使用 Ring Buffer、批量刷新和虚拟滚动。

### Gradle 构建复杂度

MVP 阶段只支持 assembleDebug，不做完整 Gradle Project Sync。

### ADB 路径问题

查找顺序：

```text
用户配置 adbPath
ANDROID_HOME / ANDROID_SDK_ROOT
PATH
用户手动选择
```
