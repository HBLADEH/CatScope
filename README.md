# CatScope

<p align="center">
  <img src="./Logo.png" alt="CatScope Logo" width="160" />
</p>

<p align="center">
  <strong>Logcat without Android Studio.</strong>
</p>

<p align="center">
  A lightweight Android debugging workbench for logs, crashes, installs, launches, and daily troubleshooting.
</p>

<p align="center">
  English · <a href="./README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <a href="#why-catscope">Why CatScope</a> ·
  <a href="#features">Features</a> ·
  <a href="#quick-start">Quick Start</a> ·
  <a href="#roadmap">Roadmap</a> ·
  <a href="#documentation">Documentation</a>
</p>

CatScope is a lightweight desktop workbench for Android Logcat. It is built for developers who need Android logs, crash clues, package filtering, device state, and routine debugging actions without opening the full Android Studio IDE.

> Better than raw `adb logcat`, lighter than Android Studio, and more Android-aware than a generic log viewer.

CatScope is not trying to replace Android Studio. The core idea is **Logcat without Android Studio**, then carefully add the nearby workflows that make daily Android troubleshooting smoother: build, install, launch, crash analysis, export, and AI-friendly context generation.

## Why CatScope

Android Studio is powerful, but it can be too heavy when the task is simply:

- watch Logcat for one device or app,
- filter logs by package, PID, level, tag, or keyword,
- inspect crash / ANR / native crash clues,
- export a focused log session,
- install and launch a debug build,
- collect context for a teammate or an AI agent.

CatScope keeps that workflow small and direct. The product boundary is intentionally narrow: it should become a great Android troubleshooting companion, not another full IDE.

## Status

CatScope is currently in the MVP stage. The core Logcat Viewer, Offline Log File Viewer, rule-based Crash / ANR / Native / JNI / Install Error Analyzer, local AI Context Generator, a minimal Build / Install / Launch workflow, and lightweight Workspace / Filter Presets are available.

## Features

### Feature Checklist

- [x] Desktop app foundation
  - [x] Wails v2 desktop app with a Go backend.
  - [x] Vue 3 + TypeScript frontend.
- [x] ADB and device management
  - [x] ADB discovery from user configuration, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, and `PATH`.
  - [x] Device list parsing with clear `device`, `offline`, `unauthorized`, and `unknown` states.
  - [x] Device information display: model, brand, Android version, SDK, and ABI.
- [x] Live Logcat Viewer
  - [x] Live `adb logcat -v threadtime -b main,system,crash` streaming.
  - [x] Start, stop, restart, and device switching for Logcat streams.
  - [x] Continuous stdout / stderr reading with clear error reporting.
  - [x] 100000-line ring buffer with batch reads and dropped-line counts.
  - [x] Virtualized frontend log table.
- [x] Package and PID filtering
  - [x] Installed package listing for all packages and third-party packages.
  - [x] Package search, selection, clearing, and all-log mode.
  - [x] PID tracking for the selected package, including app restarts.
  - [x] Package / Level combined filtering and case-insensitive search.
- [x] Log parsing and interaction
  - [x] threadtime parsing, raw line preservation, and multiline log merging.
  - [x] Java stacktrace and AndroidRuntime `FATAL EXCEPTION` grouping.
  - [x] Pause, clear, detail / analysis panel, txt export, and jsonl export.
- [x] Offline Log File Viewer
  - [x] Open `.txt`, `.log`, and `.jsonl` log files.
  - [x] Parse ordinary threadtime Logcat text files with the existing parser.
  - [x] Preserve unparsed raw lines and multiline stacktraces.
  - [x] Reuse search, level, tag, exclude, regex, package filtering, virtual scrolling, details, Analyzer, and AI Context generation.
  - [x] Show live/offline source mode, file path, file name, entry count, and raw-line parse count.
  - [x] Reopen CatScope JSONL exports as offline logs.
- [x] Rule-based Analyzer without external AI API calls
  - [x] Java Crash: `AndroidRuntime`, `FATAL EXCEPTION`, `Process:`, `Caused by:`, and common exception types.
  - [x] Native Crash: `SIGSEGV`, `SIGABRT`, `backtrace:`, `tombstone`, `Abort message`, `fault addr`, and `libxxx.so`.
  - [x] ANR: `ANR in`, `Application Not Responding`, `Input dispatching timed out`, and service / broadcast timeouts.
  - [x] JNI Error: `JNI DETECTED ERROR IN APPLICATION`, `CheckJNI`, stale / deleted references, and pending exceptions.
  - [x] Install Error: `INSTALL_FAILED_*`, `INSTALL_PARSE_FAILED_*`, `DELETE_FAILED_*`, `Failure [INSTALL_FAILED...]`, and `adb: failed to install`.
- [x] Install Error Analyzer
  - [x] Analyze install failure text or log output.
  - [x] Provide bilingual reasons and next-step suggestions.
  - [x] Prepare analyzer output for future Build / Install / Launch workflows.
- [x] Local AI Context Generator
  - [x] Generate Markdown context for the selected analysis result.
  - [x] Include device/package/PID metadata, analysis summary, related logs, context logs, key frames, and suggestions.
  - [x] Copy the Markdown to the clipboard or export it as a `.md` file.
  - [x] Avoid OpenAI, Claude, Gemini, or any cloud model calls.
- [x] Build / Install / Launch MVP
  - [x] Select an Android project directory and detect `gradlew` / `gradlew.bat`.
  - [x] Validate `settings.gradle` / `settings.gradle.kts`.
  - [x] Run `assembleDebug` by default.
  - [x] Find the newest debug APK under `build/outputs/apk`.
  - [x] Install APKs with `adb install -r` and optional `-d`, `-g`, `-t`.
  - [x] Launch the configured package with `adb shell monkey -p <package> 1`.
  - [x] Send install failures to the Install Error Analyzer and Analysis panel.
- [x] Workspace / Filter Presets
  - [x] Store multiple lightweight workspaces in the user's local CatScope config.
  - [x] Restore project path, package name, selected device, log levels, search keyword, install options, and AI context options.
  - [x] Save, select, update, and delete workspaces.
  - [x] Built-in presets: All Logs, Errors Only, AndroidRuntime, Native Crash, Install Errors, and Current Package.
  - [x] Save, apply, rename, and delete custom filter presets with level, package, keyword, regex, tags, and exclude keyword.
- [ ] More export formats: csv, zip.
- [ ] Module and variant selection for Build / Install / Launch.
- [ ] macOS and Linux support.

## Quick Start

### Requirements

- Go 1.22 or later.
- Node.js 20 or later and npm 10 or later.
- Wails v2 CLI.
- Microsoft WebView2 Runtime.
- Android SDK Platform Tools, with `adb` available through one of:
  - `ANDROID_HOME` or `ANDROID_SDK_ROOT`.
  - `platform-tools` added to `PATH`.
  - CatScope adb path configuration, with a more complete UI planned.

Install the Wails CLI:

```powershell
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

### Run Locally

```powershell
git clone <repository-url>
cd CatScope

go test ./...

cd frontend
npm install
npm run build

cd ..
wails doctor
wails dev
```

Live Logcat requires an Android device or emulator with USB debugging authorized. Offline Log File Viewer works without an Android device and can open `.txt`, `.log`, or `.jsonl` files for search, filtering, analysis, and AI Context generation.

If the device is `unauthorized`, approve the authorization prompt on the device and refresh. If it is `offline`, reconnect the device or restart adb server and refresh.

## Tech Stack

```text
Desktop framework: Wails v2
Backend: Go
Frontend: Vue 3 + TypeScript
Build tool: Vite
State management: Pinia
UI library: Naive UI
Virtual scrolling: Vue virtual scrolling utilities
ADB integration: Go exec.Command
Local storage: JSON configuration, SQLite planned
Primary platform: Windows
Planned platforms: macOS, Linux
```

Vue 3 keeps the desktop-tool UI easy to evolve. Naive UI fits dark themes, forms, drawers, tabs, notifications, and workbench-style layouts. Pinia owns device, log stream, filter, and session state. Virtual scrolling keeps large Logcat sessions responsive.

## Project Scope

CatScope focuses on Android troubleshooting workflows around Logcat. The first goal is an excellent live and offline Logcat Viewer; build, install, launch, crash analysis, AI-ready context, and saved workspace presets are adjacent workflows that support the same debugging loop. The current build runner is intentionally small: it runs Gradle wrapper tasks such as the default `assembleDebug`, but it is not Gradle Project Sync and it is not a full IDE. The multi-workspace support is also lightweight configuration, not a full IDE project system.

CatScope is not intended to provide:

- code editing,
- layout preview,
- Gradle Project Sync,
- breakpoint debugging,
- profiler,
- complete signing management,
- AAB publishing,
- complex Flavor visual management,
- full NDK / CMake configuration UI,
- a full Android Studio replacement.

## Target Users

- Android app developers.
- Android plugin developers.
- Native so / JNI debugging engineers.
- QA and automation engineers.
- Developers who often debug with AI agents.
- Anyone who needs Logcat and basic app troubleshooting without opening Android Studio.

## Roadmap

1. Make the Logcat Viewer fast, filterable, searchable, and comfortable for long sessions.
2. Keep improving the rule-based Analyzer and AI context generation.
3. Expand Build / Install / Launch with module and variant selection.
4. Improve cross-platform behavior and historical log analysis.

See [docs/ROADMAP.md](./docs/ROADMAP.md) for the full roadmap.

## Repository Layout

```text
CatScope/
├─ frontend/          # Vue 3 + TypeScript frontend
├─ internal/          # Go backend packages
├─ docs/              # Architecture, roadmap and development notes
├─ app.go             # Wails app bindings
├─ main.go            # Application entry point
├─ wails.json         # Wails project config
├─ go.mod
├─ go.sum
├─ Logo.png
├─ README.md
└─ README.zh-CN.md
```

## Documentation

- [Roadmap](./docs/ROADMAP.md)
- [Architecture](./docs/ARCHITECTURE.md)
- [MVP Tasks](./docs/MVP_TASKS.md)
- [Codex Start Prompt](./docs/CODEX_START_PROMPT.md)

## Contributing

Issues, suggestions, and pull requests are welcome. The project is still early, so high-impact contributions include:

- Logcat Viewer stability, performance, and interaction improvements.
- adb compatibility across devices, ROMs, and emulators.
- Crash / ANR / native crash recognition rules.
- Build / Install / Launch workflows.
- Documentation, screenshots, setup notes, and cross-platform validation.

Before submitting changes, please run:

```powershell
go test ./...

cd frontend
npm install
npm run build
```

## Privacy

CatScope reads device information and live Logcat through local adb. Offline log files are read from local disk only. Workspace and preset settings are saved as JSON in the user's local CatScope configuration directory, such as `%APPDATA%/CatScope/config.json` on Windows, and are not written into Android project directories. The project does not require uploading logs to a remote service. The AI Context Generator only creates local Markdown for you to copy or export; it does not call any external AI API. When sharing exported logs or AI context, avoid leaking sensitive device information, user data, tokens, package names, or internal business logs.

## License

This repository does not include a license file yet. Before public distribution or accepting external contributions, add a `LICENSE` file and declare the license here.
