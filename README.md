<p align="center">
  <img src="./docs/assets/logo.png" alt="CatScope Logo" width="120" />
</p>

<h1 align="center">CatScope</h1>

<p align="center">轻量级 Android Logcat 调试工作台</p>

<p align="center">
  <a href="https://github.com/HBLADEH/CatScope/releases"><img src="https://img.shields.io/github/v/release/HBLADEH/CatScope?include_prereleases" alt="GitHub Release" /></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License" /></a>
</p>

<p align="center">
  <strong>简体中文</strong> · <a href="./README.en.md">English</a>
</p>

<p align="center">
  <a href="#下载安装">下载安装</a> ·
  <a href="#快速上手">快速上手</a> ·
  <a href="#从源码运行">从源码运行</a> ·
  <a href="#文档">文档</a>
</p>

CatScope 是面向 Android 开发者的桌面日志调试工具。连接设备或打开离线日志后，可以筛选应用日志、定位 Crash 与 ANR，将排障现场保存为会话，或导出供后续分析的 Markdown。

项目以中文用户为主，同时提供英文界面与[英文说明](./README.en.md)。Windows 是主要支持平台，macOS 提供兼容 Intel 和 Apple Silicon 的 universal preview。

![CatScope 日志工作台](./docs/screenshots/log-workbench.jpg)

上图为 Windows 版打开本地示例日志的实际界面。

<details>
<summary>查看问题分析与 AI Context 截图</summary>

![CatScope 问题分析与 AI Context](./docs/screenshots/analysis-panel.jpg)

</details>

## 主要功能

- **日志查看**：实时 ADB Logcat，以及 `.txt`、`.log`、`.jsonl` 离线文件；虚拟滚动日志列表。
- **搜索过滤**：包名与 PID 追踪、level、tag、关键词、排除词和正则表达式。
- **问题分析**：用本地规则识别 Java Crash、ANR、native crash、JNI error 和安装失败。
- **AI Context**：将分析结果与相关日志整理为 Markdown，支持复制和导出。
- **会话与预设**：通过 `.catscope-session` 保存和恢复日志、过滤条件与分析结果；用 workspace 和过滤预设复用配置。
- **构建安装启动**：调用项目的 Gradle Wrapper（默认 `assembleDebug`），安装 APK 并启动指定应用。

工作台以日志为中心：顶部选择来源、设备、应用和过滤条件，右侧查看详情与问题分析，底部展开“构建与运行”。会话保存与导出在“文件”菜单，ADB、外观、工作区与更新集中在“设置”中。支持浅色、深色和跟随系统主题。

CatScope 专注于日志排障，不提供代码编辑、Gradle Project Sync 或断点调试。完整限制见 [已知问题](./docs/KNOWN_ISSUES.md)。

## 下载安装

从 [最新版本](https://github.com/HBLADEH/CatScope/releases/latest) 下载对应平台的文件；历史版本和预览版见 [Releases](https://github.com/HBLADEH/CatScope/releases)。

| 平台 | 文件 | 安装方式 |
| --- | --- | --- |
| Windows 10 / 11（x64） | `*-windows-amd64.exe` | 直接运行，需要 [Microsoft WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/) |
| macOS（Intel / Apple Silicon，预览版） | `*-macos-universal.dmg` | 打开 DMG，将 CatScope 拖入 Applications |

macOS 版本尚未经过 Apple notarization，首次启动可能被 Gatekeeper 拦截。确认下载来源后，可在“系统设置 → 隐私与安全性”中允许打开。Linux 暂无官方发行包。

Windows 支持在 EXE 所在目录可写时校验更新并重启；macOS 需手动更新。发行包均附带 `.sha256` 文件，校验方式见 [发布资产说明](./docs/RELEASE_ASSETS.md)。

## 快速上手

**实时日志：**

1. 安装 [Android SDK Platform Tools](https://developer.android.com/tools/releases/platform-tools)。CatScope 可从 `PATH`、`ANDROID_HOME` 或 `ANDROID_SDK_ROOT` 查找 ADB，也可在“设置 → 设备”中配置完整路径。
2. 连接已开启 USB 调试的 Android 设备并允许授权，或启动模拟器。
3. 在 CatScope 中刷新并选择设备，启动 Logcat，再选择包名或设置过滤条件缩小范围。

**离线使用：** 从来源菜单或空白工作台打开 `.txt`、`.log`、`.jsonl` 日志，或打开 `.catscope-session` 恢复调试现场；无需 ADB 或设备。

**定位问题：** 点击“问题”查看本地分析结果，再查看关联日志或复制 AI Context。关联日志被过滤条件隐藏时，可以临时查看，随后返回原过滤结果。

设备显示 `unauthorized` 时，请在手机上允许 USB 调试；显示 `offline` 时，尝试重新连接或重启 ADB。完整操作见 [用户指南](./docs/USER_GUIDE.zh-CN.md)。

## 从源码运行

需要 Go 1.22+、npm 和 [Wails 平台构建依赖](https://wails.io/docs/gettingstarted/installation/)。推荐使用 Node.js 22 LTS（22.12 或更新版本）。

```sh
git clone https://github.com/HBLADEH/CatScope.git
cd CatScope
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
npm ci --prefix frontend
npm run build --prefix frontend
wails doctor
wails dev
```

在仓库根目录按改动范围执行检查：

```sh
go test ./...
npm run build --prefix frontend
git diff --check
```

本机应用构建使用 `wails build`。Windows 发布前运行 `scripts\check.ps1`，打包使用 `scripts\build-windows.ps1`；macOS 打包见 [发布资产说明](./docs/RELEASE_ASSETS.md)。开发服务验证结束后，请关闭相关进程。

项目使用 Wails v2、Go、Vue 3、TypeScript、Pinia 和 Naive UI。模块职责、日志流向和源码入口见 [手绘架构导读](./docs/ARCHITECTURE_SKETCH.zh-CN.md)与[架构说明](./docs/ARCHITECTURE.md)。

## 文档

完整资料见 [文档索引](./docs/README.zh-CN.md)。

| 主题 | 文档 |
| --- | --- |
| 日志查看、分析、会话与工作区 | [用户指南](./docs/USER_GUIDE.zh-CN.md) |
| 模块设计与日志流向 | [架构说明](./docs/ARCHITECTURE.md) · [手绘架构导读](./docs/ARCHITECTURE_SKETCH.zh-CN.md) |
| 当前限制 | [已知问题](./docs/KNOWN_ISSUES.md) |
| 手工验收 | [QA 清单](./docs/QA_CHECKLIST.zh-CN.md) |
| 发布与校验 | [发布资产说明](./docs/RELEASE_ASSETS.md) |
| 版本变化与后续方向 | [变更记录](./CHANGELOG.md) · [路线图](./docs/ROADMAP.md) |

## 贡献与反馈

欢迎使用中文通过 [Issues](https://github.com/HBLADEH/CatScope/issues) 报告问题或提出建议，也欢迎通过 Pull Request 贡献代码和文档。提交前请阅读 [贡献指南](./CONTRIBUTING.md) 与 [行为准则](./CODE_OF_CONDUCT.md)；使用问题可先查看 [支持说明](./SUPPORT.md)。

报告问题时请提供版本、操作系统、复现步骤和脱敏日志。安全漏洞请按 [安全策略](./SECURITY.md) 私下报告。

## 隐私

CatScope 不上传日志，也不调用外部模型 API。默认开启的更新检查会访问 GitHub，可在“应用更新”中关闭。日志、Session 和 AI Context 可能包含敏感信息，分享前请先脱敏。

## 许可证

[MIT](./LICENSE) © 2026 HBLADEH。
