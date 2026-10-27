<p align="center">
  <img src="./docs/assets/logo.png" alt="CatScope Logo" width="120" />
</p>

<h1 align="center">CatScope</h1>

<p align="center">A lightweight Android Logcat debugging workbench</p>

<p align="center">
  <a href="https://github.com/HBLADEH/CatScope/releases"><img src="https://img.shields.io/github/v/release/HBLADEH/CatScope?include_prereleases" alt="GitHub Release" /></a>
  <a href="./LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="MIT License" /></a>
</p>

<p align="center">
  English · <a href="./README.zh-CN.md">简体中文</a>
</p>

<p align="center">
  <a href="#installation">Installation</a> ·
  <a href="#getting-started">Getting Started</a> ·
  <a href="#architecture">Architecture</a> ·
  <a href="#development">Development</a> ·
  <a href="#documentation">Documentation</a>
</p>

CatScope lets you inspect live and offline Android Logcat, filter app logs, analyze crashes and ANRs, and save debugging sessions you can reopen later.

![CatScope live Logcat interface](./docs/screenshots/live-logcat.png)

<details>
<summary>More screenshots: analysis and build workflows</summary>

| Analysis | Build, install, and launch |
| --- | --- |
| <img src="./docs/screenshots/analysis-tab.png" alt="CatScope analysis panel" width="420" /> | <img src="./docs/screenshots/build-install-launch.png" alt="CatScope build, install, and launch panel" width="420" /> |

</details>

## Features

- **Log viewing:** Live ADB Logcat and offline `.txt`, `.log`, and `.jsonl` files, with a virtualized log list.
- **Filtering:** Package and PID tracking, levels, tags, keywords, exclusions, and regular expressions.
- **Analysis:** Local rules for Java crashes, ANRs, native crashes, JNI errors, and install failures.
- **AI Context:** Copy or export Markdown containing analysis results and related logs.
- **Sessions and presets:** Save and restore logs, filters, and analysis results in `.catscope-session` files; reuse configuration through workspaces and filter presets.
- **Build, install, and launch:** Run the project's Gradle Wrapper (`assembleDebug` by default), install the APK, and launch the configured app.

The workbench keeps logs in the main area: source, device, app, and filters sit at the top; details and analysis open on the right; Build & run opens below the logs. File contains session saving and exports, while Settings groups ADB, appearance, workspaces, and updates. Light, dark, and system themes use the logo's navy and cyan palette.

See [Known Issues](./docs/KNOWN_ISSUES.md) for scope and current limitations.

## Installation

Download the file for your platform from [GitHub Releases](https://github.com/HBLADEH/CatScope/releases). The project is currently in Preview.

| Platform | File | Installation |
| --- | --- | --- |
| Windows 10 / 11 (x64) | `*-windows-amd64.exe` | Run directly; requires [Microsoft WebView2 Runtime](https://developer.microsoft.com/microsoft-edge/webview2/) |
| macOS (Intel / Apple Silicon, preview) | `*-macos-universal.dmg` | Open the DMG and drag CatScope into Applications |

The macOS build is not yet Apple-notarized, so Gatekeeper may block the first launch. After verifying the download source, allow it in System Settings → Privacy & Security. Official Linux builds are not available.

Windows supports verified in-app updates and restart when the EXE directory is writable; macOS requires manual updates. Releases include `.sha256` files; see [Release Assets](./docs/RELEASE_ASSETS.md) for verification instructions.

## Getting Started

**Live logs:**

1. Install [Android SDK Platform Tools](https://developer.android.com/tools/releases/platform-tools). CatScope can find ADB through `PATH`, `ANDROID_HOME`, or `ANDROID_SDK_ROOT`, or you can set its full path in the app.
2. Connect an Android device with USB debugging enabled and approve the authorization prompt, or start an emulator.
3. Refresh and select the device in CatScope, start Logcat, then choose a package or set filters to narrow the results.

**Offline use:** Open a `.txt`, `.log`, or `.jsonl` log file, or restore a `.catscope-session` file. No device is required.

For `unauthorized` devices, approve USB debugging on the phone. For `offline` devices, try reconnecting or restarting ADB. See the [User Guide](./docs/USER_GUIDE.md) for full instructions.

## Architecture

Wails v2 connects the Vue 3 + TypeScript UI to the Go backend. `app.go` coordinates ADB, log parsing, rule-based analysis, builds, and local storage.

![Hand-drawn CatScope architecture: frontend workstation, Wails bridge, Go backend workshop, and local tools and files](./docs/assets/architecture-sketch.png)

The [Hand-drawn Architecture Guide (Chinese)](./docs/ARCHITECTURE_SKETCH.zh-CN.md) covers module responsibilities, log flow, and source entry points. See [Architecture](./docs/ARCHITECTURE.md) for the detailed design.

## Development

Requires Go 1.22+, npm, and the [Wails platform build dependencies](https://wails.io/docs/gettingstarted/installation/). Node.js 22 LTS (22.12 or later) is recommended.

```sh
git clone https://github.com/HBLADEH/CatScope.git
cd CatScope
go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
npm ci --prefix frontend
npm run build --prefix frontend
wails doctor
wails dev
```

Run the checks relevant to your changes from the repository root:

```sh
go test ./...
npm run build --prefix frontend
git diff --check
```

Use `wails build` to build the app for your current platform. See [Release Assets](./docs/RELEASE_ASSETS.md) for packaging and verification, and the [contribution guide](./CONTRIBUTING.md) for submission requirements.

## Documentation

Browse the [documentation index](./docs/README.md) for the full set.

| Topic | Document |
| --- | --- |
| Log viewing, analysis, sessions, and workspaces | [User Guide](./docs/USER_GUIDE.md) |
| Current limitations | [Known Issues](./docs/KNOWN_ISSUES.md) |
| Manual verification | [QA Checklist](./docs/QA_CHECKLIST.md) |
| Releases and checksums | [Release Assets](./docs/RELEASE_ASSETS.md) |
| Changes and planned work | [Changelog](./CHANGELOG.md) · [Roadmap](./docs/ROADMAP.md) |

## Contributing and Feedback

Use [Issues](https://github.com/HBLADEH/CatScope/issues) to report bugs or suggest improvements, and pull requests to contribute code. Read the [contribution guide](./CONTRIBUTING.md) and [Code of Conduct](./CODE_OF_CONDUCT.md) first.

Include the version, operating system, reproduction steps, and sanitized logs in bug reports. Report security vulnerabilities privately as described in the [security policy](./SECURITY.md).

## Privacy

CatScope does not upload logs or call external model APIs. Update checks contact GitHub and are enabled by default; you can disable them in App Updates. Logs, sessions, and AI Context may contain sensitive information, so redact them before sharing.

## License

[MIT](./LICENSE) © 2026 HBLADEH.
