# CatScope v0.6.0-preview

## Highlights

- Lightweight Android Logcat debugging workbench for Windows.
- Live Logcat, offline log files, and `.catscope-session` restore flow.
- Rule-based crash, ANR, native crash, JNI error, and install error analysis.
- Local AI Context Markdown generation without external AI API calls.
- Preview release packaging and pre-release check scripts.

## Features

- Device discovery through local adb and Android SDK Platform Tools.
- Live `threadtime` Logcat streaming with large ring buffer support.
- Search, level, tag, exclude, regex, package, and PID tracking filters.
- Offline `.txt`, `.log`, `.jsonl`, and CatScope session workflows.
- TXT / JSONL export and local Markdown AI context export.
- Basic Build / Install / Launch workflow for Gradle wrapper projects.
- Workspace and filter preset persistence in local `config.json`.

## Known Issues

- CatScope is not a complete IDE and is not an Android Studio replacement.
- Gradle Project Sync is not supported.
- Code editing is not supported.
- Module and variant selection is still basic.
- Launcher activity automatic parsing is not complete.
- External AI APIs are not integrated.
- Large session files are not compressed or split.
- Vite may show a chunk size warning; this does not currently block use.

## Requirements

- Windows with Microsoft WebView2 Runtime.
- Android SDK Platform Tools with `adb` available through `PATH`, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or CatScope configuration.
- USB debugging enabled and authorized for live device logging.

## Installation

1. Download the Windows release artifact from GitHub Releases.
2. Extract the portable archive or run the installer if one is provided.
3. Start `CatScope.exe`.
4. Connect an Android device or open an offline log/session file.

## Checksums

```text
<artifact-name>  <sha256>
```

## Changelog

- Prepared v0.6.0-preview release metadata.
- Added Windows build and pre-release check scripts.
- Added release notes, known issues, screenshot placeholders, and user guide docs.
- Expanded README setup, Windows, privacy, and FAQ notes.
