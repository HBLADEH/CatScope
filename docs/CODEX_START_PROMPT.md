# Codex Start Prompt for CatScope

你现在要实现一个名为 CatScope 的桌面工具项目。

CatScope 是一个脱离 Android Studio 的轻量级 Android Logcat 调试工作台。它不是完整 IDE，也不是 Android Studio 替代品。它的核心目标是提供比 `adb logcat` 更好用、比 Android Studio 更轻量、比普通日志查看器更懂 Android 的 Logcat 调试体验。

## 技术栈要求

请使用以下技术栈：

- 桌面框架：Wails v2
- 后端语言：Go
- 前端框架：Vue 3 + TypeScript
- 构建工具：Vite
- 状态管理：Pinia
- 主 UI 组件库：Naive UI
- 工具库：VueUse
- 图标方案：Iconify / unplugin-icons，或 Naive UI 兼容图标方案
- 日志虚拟滚动：优先 `@tanstack/vue-virtual`，也可以使用 `vue-virtual-scroller`
- ADB 调用：Go `exec.Command` 调用本地 adb
- 构建调用：Go `exec.Command` 调用 `gradlew` / `gradlew.bat`
- 本地配置：JSON
- 后续本地存储：SQLite
- 优先平台：Windows

## 项目核心边界

第一阶段不要做完整 IDE，不要实现代码编辑器、布局预览、Gradle Project Sync、断点调试、Profiler、签名管理、AAB 发布、复杂 Flavor 可视化管理、NDK/CMake 完整配置 UI。

请始终围绕以下核心目标开发：

1. 实时查看 Android Logcat。
2. 高性能显示大量日志。
3. 支持日志解析、过滤、搜索、高亮。
4. 支持包名过滤和 PID 自动追踪。
5. 支持常见崩溃识别。
6. 支持复制适合 AI Agent 分析的日志上下文。
7. 后续支持简单 Build / Install / Launch 调试闭环。

## 推荐目录结构

请按以下结构组织项目：

```text
catscope/
├─ frontend/
│  ├─ src/
│  │  ├─ components/
│  │  ├─ pages/
│  │  ├─ stores/
│  │  ├─ composables/
│  │  ├─ types/
│  │  ├─ utils/
│  │  ├─ App.vue
│  │  └─ main.ts
│  ├─ package.json
│  └─ vite.config.ts
│
├─ internal/
│  ├─ adb/
│  │  ├─ devices.go
│  │  ├─ logcat.go
│  │  ├─ install.go
│  │  ├─ launch.go
│  │  └─ shell.go
│  │
│  ├─ logcat/
│  │  ├─ entry.go
│  │  ├─ parser.go
│  │  ├─ filter.go
│  │  ├─ buffer.go
│  │  ├─ stream.go
│  │  └─ analyzer.go
│  │
│  ├─ build/
│  │  ├─ gradle.go
│  │  ├─ apk_finder.go
│  │  └─ runner.go
│  │
│  ├─ workspace/
│  │  ├─ config.go
│  │  ├─ project.go
│  │  └─ filters.go
│  │
│  ├─ storage/
│  │  ├─ session.go
│  │  ├─ export.go
│  │  └─ database.go
│  │
│  └─ ai/
│     ├─ context.go
│     └─ markdown.go
│
├─ docs/
├─ wails.json
├─ go.mod
├─ go.sum
└─ README.md
```

## 第一阶段实现目标

请优先实现 MVP 中最小可运行闭环，不要一次性实现所有功能。

第一轮请完成：

1. 初始化 Wails v2 + Vue 3 + TypeScript 项目，前端使用 Vite。
2. 建立后端目录：`internal/adb`、`internal/logcat`、`internal/workspace`、`internal/storage`、`internal/ai`。
3. 实现 ADB 路径查找：优先用户配置，其次 `ANDROID_HOME` / `ANDROID_SDK_ROOT`，再次 PATH。
4. 实现设备列表：调用 `adb devices -l` 并解析 serial 和 state。
5. 实现设备信息读取：model、Android version、SDK、ABI。
6. 实现 Logcat 启动和停止：调用 `adb -s <serial> logcat -v threadtime -b main,system,crash`。
7. 实现 Logcat threadtime parser。
8. 实现 Ring Buffer，默认保留最近 100000 行。
9. 前端使用 Vue 3 + Pinia + Naive UI 实现基础 UI：设备选择、Start、Stop、Pause、Clear、搜索框、日志表格、状态栏。
10. 日志表格必须使用虚拟滚动，避免大量 DOM 卡死。

## 后端核心数据结构

请实现类似结构：

```go
type AndroidDevice struct {
    Serial         string `json:"serial"`
    State          string `json:"state"`
    Model          string `json:"model,omitempty"`
    Brand          string `json:"brand,omitempty"`
    AndroidVersion string `json:"androidVersion,omitempty"`
    SDKVersion     string `json:"sdkVersion,omitempty"`
    ABI            string `json:"abi,omitempty"`
    IsEmulator     bool   `json:"isEmulator,omitempty"`
}

type LogEntry struct {
    ID          int64    `json:"id"`
    Timestamp   string   `json:"timestamp"`
    PID         int      `json:"pid"`
    TID         int      `json:"tid"`
    Level       string   `json:"level"`
    Tag         string   `json:"tag"`
    Message     string   `json:"message"`
    PackageName string   `json:"packageName,omitempty"`
    Raw         string   `json:"raw"`
    Multiline   []string `json:"multiline,omitempty"`
}
```

## Logcat 解析要求

优先解析 `threadtime` 格式。

规则：

- 标准 Logcat 行解析成新的 `LogEntry`。
- 解析失败的行保留 raw。
- 如果当前行不是标准头部，则追加到上一条日志的 `Multiline` 字段。
- 不要因为单行解析失败导致整个流中断。

## UI 要求

主界面布局：

```text
顶部工具栏：设备选择 | 刷新 | Start | Stop | Pause | Clear | 搜索 | Level
中间区域：日志表格
右侧区域：日志详情
底部状态栏：设备状态 | 日志数量 | 丢弃数量 | Logcat 状态
```

日志表格字段：

```text
时间 | Level | PID | TID | Tag | Message
```

交互要求：

- 点击日志行显示完整内容。
- Level 需要高亮。
- 支持搜索关键词。
- 支持暂停 UI 刷新。
- 支持清空当前日志。

## 性能要求

- 不要把每一行日志都作为单独事件立刻推给前端。
- 后端应批量缓存日志。
- 前端应批量刷新。
- 前端日志列表必须使用虚拟滚动。
- Ring Buffer 超出上限后丢弃旧日志，并记录丢弃数量。

## 验收标准

本轮完成后必须满足：

1. `wails dev` 可以启动。
2. 前端可以显示设备列表。
3. 可以选择设备启动 Logcat。
4. 可以看到实时日志。
5. 可以暂停、继续、清屏。
6. 可以搜索日志。
7. 日志 Level 有基本高亮。
8. 断开设备或 adb 报错时，不会导致程序崩溃。
9. 代码结构清晰，后续方便加入包名过滤、Crash Analyzer、Build / Install / Launch。

## 输出要求

请先检查当前仓库状态和目录结构，然后制定实现计划，再开始修改代码。

每完成一个阶段，请运行必要的构建或测试命令，并说明结果。

不要引入与目标无关的大型依赖，不要把项目做成完整 IDE。
