# Release Assets

Use this checklist when preparing a CatScope preview release.

## Required Uploads

- Windows preview:
  - `CatScope.exe`
- macOS universal preview:
  - `CatScope-v0.6.3-preview-macos-universal.dmg`
  - `CatScope-v0.6.3-preview-macos-universal.dmg.sha256`
- README files:
  - `README.md`
  - `README.zh-CN.md`
- Release notes
- SHA256 checksum
- Screenshots, when real screenshots are available

## Checksum

Run this from the repository root after building the Windows executable:

```powershell
Get-FileHash build/bin/CatScope.exe -Algorithm SHA256
```

Include the resulting SHA256 value in the release notes or release asset description.

For macOS preview builds, run:

```sh
scripts/build-macos.sh
cd dist
shasum -a 256 -c CatScope-v0.6.3-preview-macos-universal.dmg.sha256
```

The macOS script creates a universal DMG for Intel and Apple Silicon Macs and writes the checksum asset automatically.

## macOS Preview Notes

- The preview DMG is self-signed/ad-hoc signed and not Apple-notarized yet.
- Users may need to open `CatScope.app` through Finder's context menu or allow it in System Settings on first launch.
- A formal macOS release should add Developer ID signing, Apple notarization, stapling, and CI secret documentation.

## Logo

Use `Logo.png` or `docs/assets/logo.png` for release presentation. Do not use `Logo.candidate.png`; it is the raw candidate image with a white background.

## Screenshots

Recommended screenshot names:

- `live-logcat.png`
- `analysis-tab.png`
- `ai-context.png`
- `build-install-launch.png`
- `offline-log.png`
- `session.png`
- `workspace-presets.png`

Only upload real screenshots captured from CatScope. Do not create placeholder image files for release assets.
