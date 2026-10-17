# 用一张手绘图认识 CatScope

先记住三个角色：**Vue 是操作台，Wails 是桥梁，Go 是干活的工坊。** 用户在界面上选择设备、查看日志；Go 调用本机 Android 工具，整理日志并分析问题，再把结果交给界面。

CatScope 是一个桌面应用。Vue 界面和 Go 后端通过 Wails 通信；理解这条主线，就能找到大多数功能的代码。

![CatScope 手绘架构图：前端操作台通过 Wails 调用 Go 后端，实时、离线和 Session 日志进入共享缓存，供前端查看及本地规则分析使用。](./assets/architecture-sketch.png)

## 先看懂图里的四个区域

### ① 前端操作台：接收操作，展示结果

- `App.vue` 把工具栏、日志表格、详情和分析面板组织起来。
- `stores/logs.ts` 是主要工作流入口，用 Pinia 保存设备、日志、过滤条件和分析结果，也负责调用后端。
- `LogTable.vue` 用虚拟滚动显示当前可见的行；`LogDetails.vue` 展示日志原文、分析结果和 AI Context。

**包名、level、tag、关键词、排除词和 regex 等显示过滤，主要在 `logs.ts` 的 `filteredLogs` 中完成。** 要改搜索行为，先找这里。

### ② Wails 桥梁：让 TypeScript 调用 Go

前端通过 `frontend/src/utils/wails.ts` 中的 `backend` 调用 `app.go` 上的公开方法。例如：

```text
点击开始 → logs.ts → backend.startLogcat(serial) → App.StartLogcat(serial)
```

可以把它理解成“前台给工坊派任务”。Wails 负责连接两端，前端拿到返回值后更新状态。`main.go` 创建桌面窗口、加载打包后的前端资源，并把 `App` 绑定给 Wails。

### ③ Go 后端工坊：采集、整理、分析

`app.go` 是调度员，负责组合各模块、维护当前日志来源和运行状态。具体能力放在 `internal/`：

| 模块 | 图里的角色 | 实际负责什么 |
| --- | --- | --- |
| `internal/adb` | 设备小助手 | 查找 ADB、读取设备和应用列表、启动 Logcat、安装 APK、启动应用 |
| `internal/logcat` | 日志加工站 | 解析 threadtime、保留原文、合并多行、追踪 PID、缓存、离线读取、导出和规则分析 |
| `internal/ai` | 材料整理员 | 把分析结果、设备信息和上下文日志整理成本地 Markdown |

日志进入后，先整理成 `LogEntry`，再放进 Ring Buffer。后端缓存默认保留最近 **100000 条日志条目**，超出上限会丢弃旧条目并记录数量。

**日志展示和规则分析是两条分支。** 查看日志不需要先等分析完成；规则分析识别 Java Crash、ANR、native crash、JNI error 和安装失败，AI Context 再按选中的结果整理材料。

### ④ 本地工具与文件：让排障流程接得上

| 模块 | 简单理解 |
| --- | --- |
| `internal/build` | 调用项目自己的 Gradle Wrapper 构建，默认任务为 `assembleDebug`，并查找 APK；安装和启动由 `internal/adb` 完成 |
| `internal/workspace` | 用本地 JSON 保存工作区、项目配置、ADB 路径和过滤预设 |
| `internal/storage` | 用 `.catscope-session` 保存和恢复日志、过滤条件、分析结果等调试现场 |
| `internal/update` | 检查 GitHub Release 更新；Windows 支持下载校验后替换并重启，macOS 提供检测和 Release 页面入口 |

`internal/process` 补充平台相关的进程处理，`internal/appversion` 提供应用版本信息。它们是辅助模块，图中只突出主要工作流。

## 跟着一条日志走一遍

### 实时日志：从手机到屏幕

```text
Android 设备
  → internal/adb 启动本机 adb logcat
  → app.go 的 ingestLogLine
  → 解析、关联包名/PID、合并多行
  → Ring Buffer
  → logs.ts 经 Wails 调用 GetLogBatch，批量取回新增条目
  → filteredLogs 筛选
  → LogTable 虚拟滚动展示
```

当前实时轮询间隔是 **250ms**，每次最多取 **2000 条**。日志原文仍由 Go 保存，前端负责显示过滤。

新批次还会由前端调用 `AnalyzeLogs` 交给 Go 规则分析器，分析结果回到 Pinia；用户也可以主动分析当前筛选后的日志。选择某个分析结果后，`GenerateAIContext` 会把它与后端日志快照等信息一起交给 `internal/ai` 生成 Markdown。

### 离线文件和 Session：换入口，继续用同一套工作台

```text
.txt / .log / .jsonl → internal/logcat 离线读取 → 共享缓存与查看流程
.catscope-session   → internal/storage 恢复   → 共享缓存与查看流程
```

离线文本会复用日志解析能力，JSONL 会读取已有的结构化条目。Session 恢复已经保存的条目和分析结果，不需要把它们重新变成 ADB 输出。两种入口都不依赖连接设备，界面继续使用相同的搜索、详情、分析和 AI Context 功能。

### 构建安装启动：一条短流程

```text
logs.ts → app.go 的 BuildInstallLaunch
        → internal/build：Gradle 构建，找到 APK
        → internal/adb：安装 APK
        → internal/adb：启动 package
```

构建或安装失败会停在对应步骤。安装失败输出会交给规则分析器，让用户在 Analysis 面板中继续排查。

## 想改功能，从哪里开始看

| 你要做的事 | 优先阅读的源码 |
| --- | --- |
| 调整窗口、工具栏、面板布局 | [main.go](../main.go)、[App.vue](../frontend/src/App.vue) |
| 改搜索过滤、暂停、设备选择等工作流 | [logs.ts](../frontend/src/stores/logs.ts) |
| 改日志列表或详情显示 | [LogTable.vue](../frontend/src/components/LogTable.vue)、[LogDetails.vue](../frontend/src/components/LogDetails.vue) |
| 追踪前后端一次调用 | [utils/wails.ts](../frontend/src/utils/wails.ts) → [app.go](../app.go) |
| 改采集或设备操作 | [internal/adb](../internal/adb/) |
| 改解析、多行合并、PID 关联或分析规则 | [internal/logcat](../internal/logcat/) |
| 改 AI Context 内容 | [context.go](../internal/ai/context.go) |
| 改保存与恢复 | [session.go](../internal/storage/session.go)、[internal/workspace](../internal/workspace/) |
| 改构建或更新流程 | [runner.go](../internal/build/runner.go)、[internal/update](../internal/update/) |

## 修改时记住这几件事

- **暂停显示和停止采集是两个动作。** 当前暂停会停掉前端轮询，后端仍可继续采集；停止 Logcat 才会关闭采集进程。
- **原文要留下来。** 解析失败不能打断整条日志流，多行堆栈也要保留，排障时经常需要回看。
- **AI Context 是本地文档生成。** 规则分析和 Markdown 生成不调用模型 API；更新检查访问 GitHub 是单独的能力。
- **改接口要照顾两端。** 新增或修改 Wails API 时，对齐 Go JSON 字段、`frontend/src/types/backend.ts` 和 `frontend/wailsjs` 绑定，并检查已有 Session 的兼容性。

本文按当前源码梳理，图中的猫咪和工坊是帮助理解的比喻。更详细的模块设计见 [架构说明](./ARCHITECTURE.md)。配图使用内置 imagegen 生成，完整提示词保存在 [architecture-sketch.prompt.md](./assets/architecture-sketch.prompt.md)。
