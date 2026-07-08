# CatScope 用户指南

CatScope 是一个轻量 Android Logcat 调试工作台。它适合查看实时日志、分析崩溃线索、打开离线日志、保存调试现场，并生成本地 Markdown 形式的 AI Context。它不是 Android Studio 的替代品，也不会上传日志或调用外部 AI API。

## 连接 Android 设备

1. 安装 Android SDK Platform Tools。
2. 确保 `adb` 可通过 `PATH`、`ANDROID_HOME`、`ANDROID_SDK_ROOT` 或 CatScope 配置找到。
3. 在手机上开启开发者选项和 USB 调试。
4. 用 USB 连接手机。
5. 第一次连接时，在手机弹窗中允许 USB 调试授权。
6. 在 CatScope 中刷新设备列表，确认设备状态为 `device`。

如果状态是 `unauthorized`，请检查手机授权弹窗。如果状态是 `offline`，请重新插拔 USB，或运行 `adb kill-server` 后再刷新。

## 查看 Live Logcat

1. 在设备列表中选择目标设备。
2. 点击启动 Live Logcat。
3. CatScope 会读取 `main`、`system` 和 `crash` buffer，并按 `threadtime` 格式解析。
4. 可使用搜索、Level、Tag、Exclude、Regex 和 Package 过滤缩小日志范围。搜索框也支持字段查询，例如 `tag:ActivityManager`、`pid:1234`、`level:E`、`-message:noise`。
5. 点击日志行可查看详情和相关分析。

## 选择 Package

1. 连接设备后打开 Package 选择器。
2. 搜索应用包名。
3. 选择目标 package 后，日志过滤会聚焦该应用。
4. 清空 package 可回到全部日志模式。

## 使用 PID Tracking

选择 package 后，CatScope 会尝试追踪该应用当前 PID。应用重启后，PID Tracking 会尽量更新到新的进程 ID，帮助过滤当前进程的日志。若应用未运行，先启动应用或使用 Build / Install / Launch 面板启动。

## Build / Install / Launch

1. 选择一个 Android 项目目录。
2. CatScope 会检测 `gradlew` 或 `gradlew.bat`，并校验 `settings.gradle` 或 `settings.gradle.kts`。
3. 默认构建任务是 `assembleDebug`。
4. 构建完成后，CatScope 会在 `build/outputs/apk` 下查找最新 debug APK。
5. 安装使用 `adb install -r`，并支持 `-d`、`-g`、`-t` 等常用选项。
6. 启动当前 package 时，CatScope 使用 `adb shell monkey -p <package> 1`。

当前 Build / Install / Launch 仍是 MVP：module / variant 选择较基础，不支持 Gradle Project Sync。

## 查看 Analyzer

Analyzer 会基于本地规则识别常见问题：

- Java Crash。
- Native Crash。
- ANR。
- JNI Error。
- Install Error。

选择相关日志或安装失败输出后，在 Analysis Tab 查看摘要、原因、关键日志和建议。Analyzer 不调用外部 AI API。

## 复制 AI Context

1. 在 Analysis Tab 选择一个分析结果。
2. 打开 AI Context 区域。
3. CatScope 会在本地生成 Markdown，包含设备、package、PID、分析摘要、相关日志、上下文日志和建议。
4. 点击复制，或导出为 `.md` 文件。

AI Context 只是本地 Markdown。是否粘贴给外部工具由你决定，请先检查敏感信息。

## 打开离线日志

1. 使用 Offline Log File Viewer 打开 `.txt`、`.log` 或 `.jsonl` 文件。
2. CatScope 会复用 threadtime parser，并保留无法解析的 raw 行。
3. 离线日志同样支持搜索、过滤、Analyzer 和 AI Context。
4. CatScope 导出的 JSONL 文件可以再次打开。

## 保存和打开 Session

1. 在实时日志或离线日志分析过程中保存 `.catscope-session` 文件。
2. Session 会保存日志、过滤条件、workspace 信息、Analysis 结果、AI Context options 和 notes。
3. 之后可以打开 session 文件，恢复到 Session 模式继续查看和分析。

大型 session 文件当前没有压缩或分片，保存和分享前请留意文件大小与敏感信息。
