# CatScope User Guide

CatScope is a lightweight Android Logcat debugging workbench. It helps with live logs, offline logs, crash clues, sessions, and local Markdown AI context. It is not an Android Studio replacement, does not upload logs, and does not call external AI APIs.

## Workbench Layout

- The first row selects the workspace and log source, plus the device and app for live logs. The source menu opens log files, restores sessions, and returns to live mode.
- The second row contains search, presets, levels, and More filters. Active tag, exclusion, and regex filters remain visible and can be cleared individually.
- Click a log row to inspect its full raw text on the right. Issues opens a compact issue list; select a result for its stack frames, suggestions, and AI Context actions.
- Build & run opens the bottom task area. Drag its top edge, or focus the separator and use the arrow keys, to resize it. Collapsing it does not cancel ongoing work.
- File provides session saving, Save as, and log exports. Edit the session name and notes in the save dialog; failed saves retain the input.
- Settings groups appearance and language, device and ADB, workspaces, and app updates. Narrow windows place details below the logs.

## Connect an Android Device

1. Install Android SDK Platform Tools.
2. Make sure `adb` is available through `PATH`, `ANDROID_HOME`, `ANDROID_SDK_ROOT`, or CatScope configuration.
3. Enable Developer Options and USB debugging on the device.
4. Connect the device over USB.
5. Accept the USB debugging authorization prompt on first connection.
6. Refresh devices in CatScope and confirm the state is `device`.

If the state is `unauthorized`, check the phone authorization prompt. If it is `offline`, reconnect the device or restart adb with `adb kill-server`.

## Check and Install CatScope Updates

CatScope checks the latest stable GitHub Release after startup by default. Open Settings → App Updates to disable automatic checks, check manually, or opt into Preview releases.

When a Windows portable build finds an update, Install and Restart downloads the matching EXE and verifies it against the release's `.sha256` asset. CatScope then exits, a temporary updater replaces the original EXE, and the new version starts. Distribution remains a single EXE. If the EXE directory is not writable, use View Release for a manual download. macOS currently detects updates but does not replace the app bundle in place.

## Live Logcat

Select a device and start Live Logcat. CatScope reads `main`, `system`, and `crash` buffers in `threadtime` format. Use search, level, tag, exclude, regex, package, and PID filters to narrow the stream. The search box also supports field queries such as `tag:ActivityManager`, `pid:1234`, `level:E`, and `-message:noise`.

## Package Selection and PID Tracking

Use the Package selector to search and choose an installed package. CatScope can then focus filters on that app and track the current PID when the app is running. If the app restarts, PID tracking attempts to follow the new process.

## Build / Install / Launch

Open Build & run at the top right and select an Android project directory with `gradlew` or `gradlew.bat` in the bottom task area. CatScope validates `settings.gradle` or `settings.gradle.kts`, runs `assembleDebug` by default, finds the newest debug APK, installs it with adb, and launches the configured package with `adb shell monkey -p <package> 1`.

Build & run performs the three steps in sequence. Single step retains individual build, APK lookup, install, build-and-install, and launch actions. Install & logging options contains the install flags, automatic Logcat, and clear-on-launch settings.

This workflow is intentionally small. Gradle Project Sync, code editing, and advanced module / variant management are not included in this preview.

## Analyzer

The local rule-based Analyzer recognizes Java crashes, native crashes, ANRs, JNI errors, and install errors. Open Issues, select a result in the right panel, and inspect its summary, likely cause, key logs, and next steps. Selecting an issue locates the related log while keeping analysis open.

When current filters hide the issue's logs, use Temporarily show related logs to inspect them while preserving your search and filters. Return to filter results restores that view. Click the selected issue again or use Locate related logs to return to the fault location.

## AI Context

Open Issues, select an analysis result, then use Copy AI Context or Export AI Context below the cause. CatScope generates local Markdown with device metadata, package/PID details, summary, related logs, context logs, and suggestions. Export uses `.md`. Review sensitive data before sharing it.

## Offline Logs

Choose Open log file from the source menu or the empty log area's shortcut to open `.txt`, `.log`, or `.jsonl` files. CatScope reuses the parser, filters, Analyzer, and AI Context generator. JSONL exports from CatScope can be reopened.

## Sessions

Use File → Save session to save live or offline work as a `.catscope-session` file, or Save session as to choose another path. Sessions preserve logs, filters, workspace metadata, Analysis results, AI Context options, and notes. Open a session from the source menu to continue the same debugging state.
