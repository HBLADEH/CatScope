# CatScope Manual QA Checklist

Use this checklist for v0.6.0-preview manual verification. CatScope is a lightweight Logcat troubleshooting workbench; do not treat missing full IDE behavior, Gradle Project Sync, or external AI API integration as regressions.

## Test Matrix

- OS: Windows 10 / Windows 11.
- App build: local dev build or release `CatScope.exe`.
- Device: at least one physical Android device or emulator.
- adb: Android SDK Platform Tools available through `PATH`, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or CatScope configuration.
- Test files: one `.txt` or `.log` Logcat file, one CatScope `.jsonl` export, and one `.catscope-session` file if available.

## Startup

- [ ] Launch CatScope from the release executable or `wails dev`.
- [ ] Confirm the main window opens without a blank screen.
- [ ] Confirm no startup error toast appears when adb is available.
- [ ] Confirm the app remains usable when adb is missing or misconfigured and shows a recoverable error.
- [ ] Confirm the version displayed in the app or release package matches the tested build.

## ADB Device Connection

- [ ] Connect an Android device or start an emulator.
- [ ] Refresh the device list.
- [ ] Confirm devices in `device`, `unauthorized`, `offline`, and unknown states are displayed clearly when reproducible.
- [ ] For an authorized device, confirm model, brand, Android version, SDK, and ABI are shown.
- [ ] Disconnect and reconnect the device, then refresh and confirm the UI recovers.
- [ ] Run `adb version` and record the version for the QA notes.

## Live Logcat

- [ ] Select an authorized device and start Live Logcat.
- [ ] Confirm logs stream from `main`, `system`, and `crash` buffers in threadtime format.
- [ ] Pause and resume the view without losing app responsiveness.
- [ ] Clear logs and confirm the visible table resets.
- [ ] Search by keyword and confirm matching rows update.
- [ ] Filter by level, tag, exclude keyword, and regex.
- [ ] Open a log row and confirm details are readable.
- [ ] Stop Logcat and confirm the stream process exits cleanly.
- [ ] Switch devices if more than one device is available.

## Package / PID Tracking

- [ ] Open the package selector after selecting a device.
- [ ] Search for a known installed package.
- [ ] Select the package and confirm package filtering is applied.
- [ ] Launch or restart the app on the device and confirm PID tracking updates when the process changes.
- [ ] Clear the selected package and confirm CatScope returns to all-log mode.
- [ ] Confirm package and level filters can be combined.

## Build / Install / Launch

- [ ] Select a valid Android project directory containing `gradlew` or `gradlew.bat`.
- [ ] Confirm CatScope validates `settings.gradle` or `settings.gradle.kts`.
- [ ] Run the default `assembleDebug` task.
- [ ] Confirm CatScope finds the newest debug APK under `build/outputs/apk`.
- [ ] Install the APK with the default `adb install -r` behavior.
- [ ] Toggle available install options such as `-d`, `-g`, or `-t` when relevant.
- [ ] Launch the configured package and confirm `adb shell monkey -p <package> 1` starts the app.
- [ ] Trigger or paste an install failure and confirm the Install Error Analyzer receives it.
- [ ] Confirm missing Gradle wrapper, invalid project, failed build, failed install, and missing package errors are visible and recoverable.

## Analyzer

- [ ] Verify Java crash detection with `AndroidRuntime`, `FATAL EXCEPTION`, `Process:`, and `Caused by:` logs.
- [ ] Verify Native Crash detection with `SIGSEGV`, `SIGABRT`, `backtrace:`, `tombstone`, or `Abort message` logs.
- [ ] Verify ANR detection with `ANR in`, `Application Not Responding`, or input timeout logs.
- [ ] Verify JNI Error detection with `JNI DETECTED ERROR IN APPLICATION` or `CheckJNI` logs.
- [ ] Verify Install Error detection with `INSTALL_FAILED_*`, `INSTALL_PARSE_FAILED_*`, `Failure [INSTALL_FAILED...]`, or `adb: failed to install`.
- [ ] Confirm analysis summaries, likely causes, key logs, and next steps are readable.
- [ ] Confirm Analyzer behavior is local and does not require network access.

## AI Context

- [ ] Select an analysis result.
- [ ] Generate AI Context Markdown.
- [ ] Confirm the Markdown includes device metadata, package/PID details, analysis summary, related logs, context logs, key frames, and suggestions when available.
- [ ] Copy AI Context to the clipboard.
- [ ] Export AI Context as a `.md` file.
- [ ] Review the output for sensitive data before sharing.
- [ ] Confirm no external AI API call is required or made by CatScope.

## Offline Log File

- [ ] Open a `.txt` or `.log` Logcat file.
- [ ] Confirm threadtime lines parse correctly and unparsed raw lines remain visible.
- [ ] Open a CatScope `.jsonl` export.
- [ ] Confirm offline logs support search, level, tag, exclude, regex, package filtering, details, Analyzer, and AI Context.
- [ ] Confirm the UI shows offline source mode, file path, file name, entry count, and raw-line parse count where available.

## Session

- [ ] Save a live debugging state as a `.catscope-session` file.
- [ ] Save an offline log analysis state as a `.catscope-session` file.
- [ ] Reopen the session and confirm CatScope enters Session mode.
- [ ] Confirm logs, raw text, multiline stacktraces, filters, workspace metadata, Analysis results, AI Context options, and notes are restored.
- [ ] Confirm session name, file path, log count, analysis count, and created time are shown.
- [ ] Confirm large sessions remain usable enough for preview testing, and note any performance issue.

## Workspace / Filter Presets

- [ ] Save a workspace with project path, package name, selected device, log levels, search keyword, install options, and AI Context options.
- [ ] Switch away from the workspace and then restore it.
- [ ] Update and delete a workspace.
- [ ] Apply built-in presets: All Logs, Errors Only, AndroidRuntime, Native Crash, Install Errors, and Current Package.
- [ ] Save, apply, rename, and delete a custom filter preset.
- [ ] Confirm presets preserve level, package, keyword, regex, tags, and exclude keyword.

## Export

- [ ] Export visible or selected logs as `.txt`.
- [ ] Export logs as `.jsonl`.
- [ ] Reopen the exported `.jsonl` file in Offline Log File Viewer.
- [ ] Export AI Context as `.md`.
- [ ] Save a `.catscope-session` file.
- [ ] Confirm exported files do not require network access and can be inspected locally.

## Return to Live Mode

- [ ] Open an offline log file, then return to Live mode.
- [ ] Open a session file, then return to Live mode.
- [ ] Confirm the selected device can start Live Logcat again.
- [ ] Confirm live filters are either restored from the chosen workspace or clearly controlled by the current UI state.
- [ ] Confirm old offline/session labels do not remain in the Live mode header.
- [ ] Confirm Analyzer, details, export, and AI Context still work after returning to Live mode.

## Regression Notes

- [ ] Record app version, OS version, Android device, Android version, adb version, and test date.
- [ ] Attach screenshots for UI regressions.
- [ ] Attach exported logs, AI Context Markdown, or a sanitized `.catscope-session` file when reporting a bug.
- [ ] Check [Known Issues](./KNOWN_ISSUES.md) before filing duplicate issues.
