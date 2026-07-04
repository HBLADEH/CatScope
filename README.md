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

CatScope is currently in the MVP stage. The core Logcat Viewer, rule-based Crash / ANR / Native / JNI Analyzer, and local AI Context Generator are available. Build / Install / Launch is planned or in progress.

## Features

### Available Now

- Wails v2 desktop app with a Go backend and a Vue 3 + TypeScript frontend.
- ADB discovery from user configuration, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, and `PATH`.
- Device list parsing with clear `device`, `offline`, `unauthorized`, and `unknown` states.
- Device information display: model, brand, Android version, SDK, and ABI.
- Live `adb logcat -v threadtime -b main,system,crash` streaming.
- Start, stop, restart, and device switching for Logcat streams.
- Continuous stdout / stderr reading with clear error reporting.
- Installed package listing for all packages and third-party packages.
- Package search, selection, clearing, and all-log mode.
- PID tracking for the selected package, including app restarts.
- threadtime parsing, raw line preservation, and multiline log merging.
- Java stacktrace and AndroidRuntime `FATAL EXCEPTION` grouping.
- Rule-based analyzer without external AI API calls:
  - Java Crash: `AndroidRuntime`, `FATAL EXCEPTION`, `Process:`, `Caused by:`, and common exception types.
  - Native Crash: `SIGSEGV`, `SIGABRT`, `backtrace:`, `tombstone`, `Abort message`, `fault addr`, and `libxxx.so`.
  - ANR: `ANR in`, `Application Not Responding`, `Input dispatching timed out`, and service / broadcast timeouts.
  - JNI Error: `JNI DETECTED ERROR IN APPLICATION`, `CheckJNI`, stale / deleted references, and pending exceptions.
- 100000-line ring buffer with batch reads and dropped-line counts.
- Virtualized frontend log table.
- Package / Level combined filtering and case-insensitive search.
- Pause, clear, detail / analysis panel, and txt export.
- Local AI Context Generator:
  - Generates Markdown context for the selected analysis result.
  - Includes device/package/PID metadata, analysis summary, related logs, context logs, key frames, and suggestions.
  - Copies the Markdown to the clipboard or exports it as a `.md` file.
  - Does not call OpenAI, Claude, Gemini, or any cloud model.

### Planned

- Build / Install / Launch: build APKs, push them to devices, install, and launch target apps.
- Install Error Analyzer: explain common install failures and suggest next steps.
- Offline historical log viewer.
- More export formats: jsonl, csv, zip.
- macOS and Linux support.

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

Live Logcat requires an Android device or emulator with USB debugging authorized.

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

CatScope focuses on Android troubleshooting workflows around Logcat. The first goal is an excellent Logcat Viewer; build, install, launch, crash analysis, and AI-ready context are adjacent workflows that support the same debugging loop.

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
3. Add Build / Install / Launch to complete the basic daily debugging loop.
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

CatScope reads device information and Logcat through local adb. The project does not require uploading logs to a remote service. The AI Context Generator only creates local Markdown for you to copy or export; it does not call any external AI API. When sharing exported logs or AI context, avoid leaking sensitive device information, user data, tokens, package names, or internal business logs.

## License

This repository does not include a license file yet. Before public distribution or accepting external contributions, add a `LICENSE` file and declare the license here.
