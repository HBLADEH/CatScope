# AGENTS.md

面向 CatScope 仓库中 AI coding agents 和贡献者的协作说明。

## 语言和表达

- 本项目的主要协作语言是中文。今后提交说明、Issue、PR、Code Review、Release Notes、QA 记录、用户向文档和日常协作沟通，都默认优先使用中文。
- 文字风格要口语化、通俗易懂，但保持专业准确。少用生硬直译，少堆术语；必要术语可以保留英文，例如 Logcat、ADB、Release、Workflow、Crash、ANR。
- 面向用户的说明先讲“怎么做”和“会发生什么”，再补充原因。不要写成营销文案，也不要把简单流程写得太绕。
- 如果已有英文文档需要继续维护，可以保留英文版本；但新增或改动共享产品说明时，应优先补齐中文版本，并让中英文语义基本一致。
- Git commit message 可以使用简洁中文，或使用项目已有的 Conventional Commit 风格加中文说明，例如 `docs: 更新发布流程说明`。
- GitHub Release 标题、Release Notes、Issue 模板回复、PR 描述和合并说明，默认使用中文；只有面向外部英文用户或上游依赖时再补英文。

## 项目概览

CatScope 是一个轻量级 Android Logcat 调试工作台。它是 Wails v2 桌面应用，后端使用 Go，前端使用 Vue 3 + TypeScript。

产品重点要始终围绕 Android 排障工作流：

- 实时和离线 Logcat 查看。
- package、PID、level、tag、关键词、排除词和 regex 过滤。
- Java crash、ANR、native crash、JNI error 和安装失败分析。
- 本地 AI Context Markdown 生成，不调用外部 AI API。
- 基础构建、安装、启动、workspace、preset 和 session 工作流。

CatScope 不是 Android Studio 替代品。除非任务明确要求，不要加入完整 IDE 能力，例如代码编辑、Gradle Project Sync、可视化布局编辑、Profiler、断点调试、签名管理、AAB 发布或复杂 flavor 管理。

当前主目标平台仍是 Windows；macOS 已进入 universal preview 阶段，需要继续保持兼容 Intel Mac 和 Apple Silicon Mac；Linux 暂时不是核心目标。

## 技术栈

- 桌面框架：Wails v2。
- 后端：Go 1.22+。
- 前端：Vue 3、TypeScript、Vite。
- 状态管理：Pinia。
- UI：Naive UI。
- 工具库：VueUse。
- 虚拟滚动：`@tanstack/vue-virtual`。
- Android 工具：本机 `adb`，由 Go 通过 `exec.Command` 调用。
- Android 构建：本机 `gradlew` / `gradlew.bat`，由 Go 通过 `exec.Command` 调用。
- 本地配置：JSON。未来扩展存储可以考虑 SQLite。

## 仓库结构

- `main.go`、`app.go`：Wails 入口和暴露给前端的后端方法。
- `internal/adb`：ADB 发现、设备/package 列表、安装、启动、实时 Logcat 进程管理。
- `internal/logcat`：日志条目、解析器、ring buffer、离线加载、导出、PID 跟踪和分析器。
- `internal/ai`：本地 AI Context Markdown 生成。这里不能调用外部 AI API。
- `internal/build`：Gradle runner 和 APK 查找。
- `internal/storage`：`.catscope-session` 保存和加载。
- `internal/workspace`：本地配置、workspace、过滤预设和项目配置。
- `internal/process`：平台相关的进程辅助逻辑。
- `frontend/src`：Vue 前端源码。
- `frontend/src/components`：日志表格、详情面板等 UI 组件。
- `frontend/src/stores`：Pinia store。`logs.ts` 是主要 UI 状态和工作流 store。
- `frontend/src/types/backend.ts`：前端使用的后端类型。
- `frontend/src/utils/wails.ts`：Wails bridge wrapper。
- `frontend/wailsjs`：Wails 生成绑定。后端 API 变化时才重新生成。
- `docs`：架构、用户指南、QA 清单、Release Notes、Roadmap 等文档。
- `scripts`：构建和检查脚本。

## 常用命令

仓库根目录：

```sh
go test ./...
```

前端目录：

```sh
npm install
npm run build
npm run dev
```

仓库根目录，已安装 Wails CLI 时：

```sh
wails doctor
wails dev
wails build
```

Windows 发布相关检查：

```powershell
scripts\check.ps1
scripts\build-windows.ps1
```

macOS universal preview 构建：

```sh
scripts/build-macos.sh
```

## 验证要求

根据改动范围选择检查：

- Go 后端改动：运行 `go test ./...`。
- 前端改动：在 `frontend/` 下运行 `npm run build`。
- 如果为了验证改动启动了前端开发服务（例如 `npm run dev`、`wails dev` 或其他本地预览服务），测试结束后要主动关闭相关服务，避免端口占用和后台进程遗留。
- Wails 暴露 API 或导出结构变化：重新生成/检查 `frontend/wailsjs`，并运行 Go 测试和前端构建。
- 发布或构建流程改动：Windows 可用时运行 `scripts/check.ps1`；macOS 发布包改动时运行 `scripts/build-macos.sh`。
- 仅文档改动：至少运行 `git diff --check`。

如果本地缺少工具导致命令无法运行，最终说明里要讲清楚。

## 后端规范

- 保持模块边界清楚。ADB 相关逻辑放在 `internal/adb`，解析、缓存、分析逻辑放在 `internal/logcat`，session 文件放在 `internal/storage`，workspace/config 放在 `internal/workspace`。
- 外部进程和长时间运行操作要使用 `context.Context`。
- 用户输入的路径、包名、任务名等字符串，在用于进程调用前要 trim 和校验。
- 不要拼 shell 命令。使用 `exec.Command` 并显式传参。
- 尽量保留原始 Logcat 文本。解析失败不能中断日志流。
- live、offline、session 三类日志来源的行为要尽量对齐。
- 默认日志 ring buffer 上限是 100000 条，除非任务明确要求修改。
- AI Context 只能生成本地 Markdown，不能上传日志，也不能调用云端模型 API。
- parser、analyzer、ADB 输出解析、storage、config 行为优先写 table-driven Go tests。

## Logcat 规则

主要支持格式：

```text
adb logcat -v threadtime -b main,system,crash
```

解析行为：

- 标准 `threadtime` 行解析为结构化 `LogEntry`。
- `Raw` 中保留原始行。
- 无法单独解析的 continuation line，在合适时追加到上一条日志。
- Java stacktrace 和 AndroidRuntime fatal exception 上下文要合并，不能丢原文。
- 单行格式异常不能终止整个日志流。

过滤和分析要保持 Android 语义：package name、PID tracking、level、tag、安装失败、ANR、native crash、JNI error 和 Java exception 都是一等概念。

## 前端规范

- 使用 Vue 3 Composition API 和 TypeScript。
- 共享工作流状态放在 Pinia，尤其是 `frontend/src/stores/logs.ts`。
- 控件、弹窗、表单、通知和布局优先使用 Naive UI。
- 大量日志列表必须使用虚拟滚动，不要直接渲染海量 DOM 行。
- UI 文案要实用、简短、偏工作台风格，不要写成营销页。
- 项目主要语言为中文；保留必要英文术语，例如 Logcat、ADB、Crash、ANR。
- 已有中英文混排习惯可以保留，但新功能的用户可见文案要优先保证中文自然。
- Wails 生成调用尽量放在现有 bridge utilities 后面。
- 新增后端-facing 类型时，TypeScript 类型要和 Go JSON 字段保持一致。

## Wails API 说明

- 暴露给前端的后端方法在 `app.go` 的 `App` 上。
- 新增、重命名或删除暴露方法，或者修改前端使用的导出结构时，要更新/重新生成 `frontend/wailsjs`。
- JSON tag 要稳定。前端状态、session 文件和配置文件可能依赖它们。
- 不要破坏已有 `.catscope-session` 文件；必要时提供迁移或兼容路径。

## Android / ADB 注意事项

- ADB 发现要尊重配置路径、`ANDROID_HOME`、`ANDROID_SDK_ROOT` 和 `PATH`。
- 设备状态要明确处理：`device`、`offline`、`unauthorized`、`unknown`。
- 实时 Logcat 需要已授权的真机或模拟器；离线日志查看不应该依赖设备。
- 设备断开、未授权、ADB 报错要表现为可恢复的用户错误，不能导致应用崩溃。
- 安装失败输出要尽量流入 install-error analyzer。
- macOS 上 adb 通常叫 `adb`，不是 `adb.exe`；从 Finder 启动的 GUI 应用可能拿不到 shell `PATH`，因此手动配置完整 adb 路径要可靠。

## 文档规范

行为变化时同步更新文档：

- 用户可见功能变化：更新 `README.md`、`README.zh-CN.md` 和 `docs/` 下相关文档。
- 架构或模块边界变化：更新 `docs/ARCHITECTURE.md`。
- QA 或发布流程变化：更新 `docs/QA_CHECKLIST.md`、`docs/QA_CHECKLIST.zh-CN.md`、`docs/RELEASE_ASSETS.md` 或 Release Notes。

中文文档优先。英文文档可以保留，但共享产品描述要尽量保持中英文语义对齐。

## 发布规范

- 预览版 tag 使用 `vX.Y.Z-preview`，例如 `v0.6.3-preview`。
- 正式版 tag 使用 `vX.Y.Z`，例如 `v0.6.3`。
- GitHub Actions 的 `Release` workflow 会同时构建 Windows 和 macOS 产物。
- Release Notes 默认使用中文，写清楚“新增什么、怎么安装、已知限制、校验方式”。
- macOS preview 如果还没有 Apple notarization，要明确说明 Gatekeeper 首次打开可能拦截。

## 变更纪律

- 改动要小而聚焦，跟随现有代码风格。
- 不要引入大型依赖，除非确实必要。
- 做功能或修 bug 时避免顺手大重构。
- 不要修改生成文件，除非确实需要重新生成。
- 不要提交构建产物、本地 session、私有日志或机器相关配置。
- 尊重工作区里已有的用户改动。不要回滚你没做的改动，除非用户明确要求。
