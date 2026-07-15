# CatScope User Guide

CatScope is a lightweight Android Logcat debugging workbench. It helps with live logs, offline logs, crash clues, sessions, and local Markdown AI context. It is not an Android Studio replacement, does not upload logs, and does not call external AI APIs.

## Connect an Android Device

1. Install Android SDK Platform Tools.
2. Make sure `adb` is available through `PATH`, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or CatScope configuration.
3. Enable Developer Options and USB debugging on the device.
4. Connect the device over USB.
5. Accept the USB debugging authorization prompt on first connection.
6. Refresh devices in CatScope and confirm the state is `device`.

If the state is `unauthorized`, check the phone authorization prompt. If it is `offline`, reconnect the device or restart adb with `adb kill-server`.

## Check and Install CatScope Updates

CatScope checks the latest stable GitHub Release after startup by default. In the App Updates section, you can disable automatic checks, check manually, or opt into Preview releases.

When a Windows portable build finds an update, Install and Restart downloads the matching EXE and verifies it against the release's `.sha256` asset. CatScope then exits, a temporary updater replaces the original EXE, and the new version starts. Distribution remains a single EXE. If the EXE directory is not writable, use View Release for a manual download. macOS currently detects updates but does not replace the app bundle in place.

## Live Logcat

Select a device and start Live Logcat. CatScope reads `main`, `system`, and `crash` buffers in `threadtime` format. Use search, level, tag, exclude, regex, package, and PID filters to narrow the stream. The search box also supports field queries such as `tag:ActivityManager`, `pid:1234`, `level:E`, and `-message:noise`.

## Package Selection and PID Tracking

Use the Package selector to search and choose an installed package. CatScope can then focus filters on that app and track the current PID when the app is running. If the app restarts, PID tracking attempts to follow the new process.

## Build / Install / Launch

Select an Android project directory with `gradlew` or `gradlew.bat`. CatScope validates `settings.gradle` or `settings.gradle.kts`, runs `assembleDebug` by default, finds the newest debug APK, installs it with adb, and launches the configured package with `adb shell monkey -p <package> 1`.

This workflow is intentionally small. Gradle Project Sync, code editing, and advanced module / variant management are not included in this preview.

## Analyzer

The local rule-based Analyzer recognizes Java crashes, native crashes, ANRs, JNI errors, and install errors. Open the Analysis Tab to inspect summaries, likely causes, key logs, and next steps.

## AI Context

For a selected analysis result, CatScope can generate local Markdown context with device metadata, package/PID details, summary, related logs, context logs, and suggestions. Copy it to the clipboard or export it as `.md`. Review sensitive data before sharing it.

## Offline Logs

Open `.txt`, `.log`, or `.jsonl` files in Offline Log File Viewer. CatScope reuses the parser, filters, Analyzer, and AI Context generator. JSONL exports from CatScope can be reopened.

## Sessions

Save live or offline work as a `.catscope-session` file. Sessions preserve logs, filters, workspace metadata, Analysis results, AI Context options, and notes. Reopen a session to continue the same debugging state.
