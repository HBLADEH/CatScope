# CatScope 手工 QA 检查清单

本清单用于当前候选发布版本的手工验收。记录中必须写明实际测试的 tag、commit 或安装包版本。CatScope 是轻量 Logcat 排障工作台；不要把完整 IDE、Gradle Project Sync 或外部 AI API 集成视为本阶段缺陷。

## 测试矩阵

- OS: Windows 10 / Windows 11；有条件时覆盖 Intel 或 Apple Silicon 上的 macOS universal preview。
- App build: 本地开发构建、发布版 `CatScope.exe`，或 preview DMG 中的 macOS `CatScope.app`。
- Device: 至少一台 Android 真机或模拟器。
- adb: Android SDK Platform Tools 可通过 `PATH`、`ANDROID_HOME`、`ANDROID_SDK_ROOT` 或 CatScope 配置找到。
- Test files: 一个 `.txt` 或 `.log` Logcat 文件，一个 CatScope `.jsonl` 导出文件，如有条件再准备一个 `.catscope-session` 文件。

## 启动

- [ ] 从发布版 exe 或 `wails dev` 启动 CatScope。
- [ ] macOS 上从 DMG 安装，将 `CatScope.app` 拖入 Applications，并记录 Gatekeeper 是否要求通过 Finder 右键菜单打开。
- [ ] 确认主窗口正常打开，没有白屏。
- [ ] adb 可用时，确认启动后没有错误 toast。
- [ ] adb 缺失或配置错误时，确认应用仍可操作，并显示可恢复的错误。
- [ ] 确认应用内或发布包版本与当前测试版本一致。

## ADB 设备连接

- [ ] 连接 Android 设备或启动模拟器。
- [ ] 刷新设备列表。
- [ ] 在可复现条件下，确认 `device`、`unauthorized`、`offline` 和未知状态显示清晰。
- [ ] 对已授权设备，确认 model、brand、Android version、SDK 和 ABI 信息可见。
- [ ] 断开并重新连接设备，刷新后确认 UI 能恢复。
- [ ] 运行 `adb version`，并在 QA 记录中保存版本信息。
- [ ] macOS 上确认 adb 以 `adb` 形式被发现；如果 CatScope 看不到 shell `PATH`，请配置完整 SDK 路径，例如 `/Users/<you>/Library/Android/sdk/platform-tools/adb`。

## Live Logcat

- [ ] 选择已授权设备并启动 Live Logcat。
- [ ] 确认日志从 `main`、`system` 和 `crash` buffer 以 threadtime 格式持续流入。
- [ ] 暂停和恢复视图，确认应用响应正常。
- [ ] 清空日志，确认表格重置。
- [ ] 使用关键词搜索，确认匹配行更新。
- [ ] 按 level、tag、exclude keyword 和 regex 过滤。
- [ ] 点击日志行，确认详情可读。
- [ ] 停止 Logcat，确认流进程干净退出。
- [ ] 如果有多台设备，切换设备并确认日志流正常。

## Package / PID Tracking

- [ ] 选择设备后打开 Package 选择器。
- [ ] 搜索一个已知已安装 package。
- [ ] 选择 package，确认 package 过滤生效。
- [ ] 在设备上启动或重启应用，确认进程变化后 PID Tracking 会更新。
- [ ] 清空 package，确认回到全部日志模式。
- [ ] 确认 package 和 level 过滤可组合使用。

## Build / Install / Launch

- [ ] 选择包含 `gradlew` 或 `gradlew.bat` 的有效 Android 项目目录。
- [ ] macOS 上确认同时存在两个 wrapper 文件时，CatScope 会优先使用 `gradlew`。
- [ ] 确认 CatScope 会校验 `settings.gradle` 或 `settings.gradle.kts`。
- [ ] 执行默认 `assembleDebug` 任务。
- [ ] 确认 CatScope 会在 `build/outputs/apk` 下找到最新 debug APK。
- [ ] 使用默认 `adb install -r` 行为安装 APK。
- [ ] 在相关场景下切换 `-d`、`-g` 或 `-t` 等安装选项。
- [ ] 启动配置的 package，确认 `adb shell monkey -p <package> 1` 能拉起应用。
- [ ] 触发或粘贴一次安装失败，确认 Install Error Analyzer 能接收并分析。
- [ ] 确认缺少 Gradle wrapper、无效项目、构建失败、安装失败和 package 缺失等错误清晰且可恢复。

## Analyzer

- [ ] 使用包含 `AndroidRuntime`、`FATAL EXCEPTION`、`Process:` 和 `Caused by:` 的日志验证 Java Crash 识别。
- [ ] 使用包含 `SIGSEGV`、`SIGABRT`、`backtrace:`、`tombstone` 或 `Abort message` 的日志验证 Native Crash 识别。
- [ ] 使用包含 `ANR in`、`Application Not Responding` 或 input timeout 的日志验证 ANR 识别。
- [ ] 使用包含 `JNI DETECTED ERROR IN APPLICATION` 或 `CheckJNI` 的日志验证 JNI Error 识别。
- [ ] 使用包含 `INSTALL_FAILED_*`、`INSTALL_PARSE_FAILED_*`、`Failure [INSTALL_FAILED...]` 或 `adb: failed to install` 的文本验证 Install Error 识别。
- [ ] 确认分析摘要、可能原因、关键日志和下一步建议可读。
- [ ] 确认 Analyzer 是本地规则，不依赖网络。

## AI Context

- [ ] 选择一个 Analysis 结果。
- [ ] 生成 AI Context Markdown。
- [ ] 确认 Markdown 在可用时包含设备信息、package/PID、分析摘要、相关日志、上下文日志、关键帧和建议。
- [ ] 将 AI Context 复制到剪贴板。
- [ ] 将 AI Context 导出为 `.md` 文件。
- [ ] 分享前检查输出中是否包含敏感信息。
- [ ] 确认 CatScope 不需要也不会调用外部 AI API。

## Offline Log File

- [ ] 打开 `.txt` 或 `.log` Logcat 文件。
- [ ] 确认 threadtime 行能正确解析，无法解析的 raw 行仍保留可见。
- [ ] 打开 CatScope 导出的 `.jsonl` 文件。
- [ ] 确认离线日志支持搜索、level、tag、exclude、regex、package 过滤、详情、Analyzer 和 AI Context。
- [ ] 确认 UI 在可用时显示 offline source mode、文件路径、文件名、日志条数和 raw 行解析数量。

## Session

- [ ] 将实时调试现场保存为 `.catscope-session` 文件。
- [ ] 将离线日志分析现场保存为 `.catscope-session` 文件。
- [ ] 重新打开 session，确认 CatScope 进入 Session 模式。
- [ ] 确认日志、raw 文本、多行 stacktrace、过滤条件、workspace 信息、Analysis 结果、AI Context options 和 notes 已恢复。
- [ ] 确认 session 名称、文件路径、日志数量、分析数量和创建时间可见。
- [ ] 确认大 session 在 preview 测试范围内仍基本可用，并记录任何性能问题。

## Workspace / Filter Presets

- [ ] 保存一个包含 project path、package name、selected device、log levels、search keyword、install options 和 AI Context options 的 workspace。
- [ ] 切换到其他状态后再恢复该 workspace。
- [ ] 更新并删除 workspace。
- [ ] 应用内置预设：All Logs、Errors Only、AndroidRuntime、Native Crash、Install Errors 和 Current Package。
- [ ] 保存、应用、重命名并删除一个自定义 filter preset。
- [ ] 确认 preset 能保留 level、package、keyword、regex、tags 和 exclude keyword。

## Export

- [ ] 将可见或选中的日志导出为 `.txt`。
- [ ] 将日志导出为 `.jsonl`。
- [ ] 在 Offline Log File Viewer 中重新打开导出的 `.jsonl` 文件。
- [ ] 将 AI Context 导出为 `.md`。
- [ ] 保存 `.catscope-session` 文件。
- [ ] 确认导出文件不依赖网络，并可在本地检查。

## macOS 发布包

- [ ] 运行 `scripts/build-macos.sh vX.Y.Z-preview`。
- [ ] 确认 `dist/CatScope-v<version>-macos-universal.dmg` 存在。
- [ ] 确认 `dist/CatScope-v<version>-macos-universal.dmg.sha256` 存在。
- [ ] 运行 `lipo -archs build/bin/CatScope.app/Contents/MacOS/CatScope` 并确认输出包含 `x86_64 arm64`。
- [ ] 运行 `codesign -dv build/bin/CatScope.app` 并确认 preview 签名为 ad-hoc/self-signed。
- [ ] 运行 `hdiutil verify dist/CatScope-v<version>-macos-universal.dmg`。
- [ ] 运行 `cd dist`，再运行 `shasum -a 256 -c CatScope-v<version>-macos-universal.dmg.sha256`。

## Return to Live Mode

- [ ] 打开离线日志文件后返回 Live mode。
- [ ] 打开 session 文件后返回 Live mode。
- [ ] 确认选中设备后可以再次启动 Live Logcat。
- [ ] 确认 live filters 要么来自所选 workspace，要么由当前 UI 状态清晰控制。
- [ ] 确认旧的 offline/session 标签不会残留在 Live mode header。
- [ ] 确认返回 Live mode 后 Analyzer、详情、导出和 AI Context 仍可使用。

## 回归记录

- [ ] 记录 app version、OS version、Android device、Android version、adb version 和测试日期。
- [ ] UI 回归请附截图。
- [ ] 报告 bug 时附加导出的日志、AI Context Markdown 或脱敏后的 `.catscope-session` 文件。
- [ ] 提交重复 issue 前先检查 [Known Issues](./KNOWN_ISSUES.md)。
