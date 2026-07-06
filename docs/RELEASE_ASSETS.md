# Release Assets

Use this checklist when preparing a CatScope preview release.

## Required Uploads

- `CatScope.exe`
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
