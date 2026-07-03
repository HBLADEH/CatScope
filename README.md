# CatScope

> Lightweight Android Logcat Workbench.

CatScope 是一个脱离 Android Studio 的轻量级 Android Logcat 调试工作台。当前版本专注于 Android Logcat 实时展示、过滤、高亮、搜索和日志导出。

本项目不是 Android Studio 的替代品，也不是完整 IDE。它的目标是：

> 比 `adb logcat` 好用，比 Android Studio 更轻量，比普通日志查看器更懂 Android。

## 核心目标

- 提供一个专注 Logcat 的桌面工具。
- 支持设备发现、设备状态识别和基础设备信息展示。
- 支持实时 Logcat、threadtime 解析、多行日志归并和大容量 ring buffer。
- 支持第三方/全部 package 列表、package 过滤和 PID 自动追踪。
- 支持日志搜索、Level 高亮、暂停、清屏、详情查看和 txt 导出。
- 后续支持崩溃识别和 AI 上下文生成。
- 后续支持简单 APK 构建、推送、安装、启动，完成基础调试闭环。

## 本地开发

### 前置要求

Windows 下建议准备：

- Go 1.22 或更新版本。
- Node.js 20 或更新版本，npm 10 或更新版本。
- Wails v2 CLI：`go install github.com/wailsapp/wails/v2/cmd/wails@latest`。
- WebView2 Runtime。
- Android SDK Platform Tools，并确保 `adb.exe` 可通过以下任一方式找到：
  - 配置 `ANDROID_HOME` 或 `ANDROID_SDK_ROOT`。
  - 将 `platform-tools` 加入 `PATH`。
  - 后续版本会提供 UI 配置 adb 路径。

### 运行命令

```powershell
cd C:\Users\10125\OneDrive\文档\CatScope
go test ./...

cd frontend
npm install
npm run build

cd ..
wails doctor
wails dev
```

实时 Logcat 需要连接 Android 真机或模拟器，并在手机上允许 USB 调试授权。若设备显示 `unauthorized`，请在手机弹窗中允许授权后刷新设备；若显示 `offline`，请重新连接设备或重启 adb server 后刷新。

## 当前已实现

- Wails v2 + Go 后端，Vue 3 + TypeScript + Vite + Pinia + Naive UI 前端。
- ADB 查找：用户配置入口、`ANDROID_HOME`、`ANDROID_SDK_ROOT`、PATH。
- 设备列表解析，明确区分 `device` / `offline` / `unauthorized` / `unknown`。
- 设备信息读取：model、brand、Android version、SDK、ABI。
- Logcat 启动/停止/重启，Stop 后可重新 Start，切换设备会停止旧流。
- `adb logcat -v threadtime -b main,system,crash` 实时读取。
- stdout/stderr 持续读取，异常退出会显示明确错误。
- 已安装包列表读取：`pm list packages` 和 `pm list packages -3`。
- Package 选择器：支持搜索、清空选择并回到全部日志模式。
- PID 自动追踪：通过 `pidof <package>` 定时刷新当前 PID，App 重启后自动更新。
- PID 到 package 的旧映射会保留，旧日志不会因为 PID 变化失去 package 归属。
- threadtime parser，raw 保留，无法解析行追加为 multiline。
- Java stacktrace 和 AndroidRuntime `FATAL EXCEPTION` 基础归并。
- 100000 行 ring buffer，支持批量拉取和丢弃计数。
- 前端虚拟滚动日志表，支持 Package / Level 组合过滤、大小写不敏感搜索、暂停、清屏和详情面板。
- 导出当前过滤后的日志为 txt。

## 已知限制

- Build / Install / Launch 尚未实现。
- 当前不是完整 IDE，不包含代码编辑、Gradle Sync、断点调试、Profiler、签名管理等能力。
- Crash Analyzer、AI Context Generator 尚未实现。
- App 未运行时不会立即出现 PID；Logcat 可继续运行，CatScope 会等待 `pidof` 返回新 PID。
- 某些 ROM 对 `pidof` 或 `pm list packages` 的行为可能不同，可能需要后续兼容。
- 实时 Logcat 验证需要连接 Android 设备；无设备时只显示空状态。
- Package Filter 和 PID Tracking 仍需真机人工验收。

## 目标用户

- Android 应用开发者
- Android 插件开发者
- Native so / JNI 调试人员
- 自动化测试与调试人员
- 经常使用 AI Agent 辅助开发的工程师
- 只需要 Logcat 和基础安装调试能力、不想打开完整 Android Studio 的用户

## 技术栈

首选技术方案：

```text
桌面框架：Wails v2
后端语言：Go
前端框架：Vue 3 + TypeScript
构建工具：Vite
状态管理：Pinia
UI 组件库：Naive UI
工具组合：VueUse + Iconify / unplugin-icons
虚拟滚动：@tanstack/vue-virtual 或 vue-virtual-scroller
ADB 调用：Go exec.Command 调用本地 adb
构建调用：Go exec.Command 调用 gradlew / gradlew.bat
本地存储：SQLite + JSON 配置文件
优先平台：Windows
后续平台：macOS / Linux
```

前端选择 Vue 3 是为了降低工具型桌面应用的 UI 开发复杂度。Naive UI 适合暗色主题、表单、弹窗、抽屉、标签页、通知和桌面工具风格；Pinia 负责设备、日志流、过滤器和会话状态；虚拟滚动库专门处理大量 Logcat 行渲染。

## 核心功能

### Logcat Viewer

- 设备选择
- 实时日志流
- 暂停 / 恢复
- 清屏
- Level 高亮
- Tag 高亮
- 关键词搜索
- 正则搜索
- 包名过滤
- PID / TID 显示
- 多 buffer 支持：main、system、crash、events、radio
- 日志导出：txt、jsonl、csv、zip
- 离线打开历史日志

### Crash Analyzer

- Java Crash 识别
- Native Crash 识别
- ANR 识别
- JNI 错误识别
- 常见 Install Error 解释
- 崩溃上下文提取
- 关键日志折叠与展开

### Build / Install / Launch

尚未实现，属于后续阶段。

### AI Context Generator

- 复制当前日志行
- 复制错误前后上下文
- 复制崩溃会话
- 生成 Markdown Bug Report
- 生成适合 AI Agent 分析的完整上下文

## 产品边界

第一阶段不做：

- 代码编辑器
- 布局预览
- Gradle Project Sync
- 断点调试
- Profiler
- 完整签名管理
- AAB 发布
- 复杂 Flavor 可视化管理
- NDK / CMake 完整配置 UI
- 替代 Android Studio 的完整 IDE 功能

## 推荐目录结构

```text
catscope/
├─ frontend/
│  ├─ src/
│  │  ├─ components/
│  │  ├─ pages/
│  │  ├─ stores/
│  │  ├─ types/
│  │  ├─ utils/
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

## MVP 开发原则

1. 先做极致好用的 Logcat Viewer。
2. 再做崩溃识别和 AI 上下文。
3. 最后补 Build / Install / Launch 调试闭环。
4. 不做完整 IDE，不复制 Android Studio。
5. 所有功能都围绕“日志调试效率”展开。

## 文档索引

- [ROADMAP.md](./ROADMAP.md)：版本路线图
- [ARCHITECTURE.md](./ARCHITECTURE.md)：技术架构设计
- [MVP_TASKS.md](./MVP_TASKS.md)：MVP 开发任务清单
- [CODEX_START_PROMPT.md](./CODEX_START_PROMPT.md)：Codex 启动提示词
