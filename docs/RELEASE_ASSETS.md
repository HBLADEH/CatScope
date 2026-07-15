# 发布资产说明

准备 CatScope 预览版或正式版发布时，用这份清单检查产物是否齐全。

## 必传产物

- Windows 预览版或正式版：
  - `CatScope-v<version>-windows-amd64.exe`
  - `CatScope-v<version>-windows-amd64.exe.sha256`
- macOS universal 包：
  - `CatScope-v<version>-macos-universal.dmg`
  - `CatScope-v<version>-macos-universal.dmg.sha256`
- README 文件：
  - `README.md`
  - `README.zh-CN.md`
- Release Notes
- SHA256 校验文件
- 真实截图，有就上传；没有就不要放占位图

## GitHub Actions 发版流程

`Release` workflow 会同时构建 Windows 和 macOS 产物。

Windows 应用内升级依赖版本化 EXE 与同名 `.sha256` 文件，两个资产缺一不可。构建前 workflow 会运行版本一致性检查，确保 tag、运行时版本、Windows 资源版本和前端包版本相同。

- 推送预览版 tag，例如 `vX.Y.Z-preview`，会创建 draft prerelease。
- 推送正式版 tag，例如 `vX.Y.Z`，会创建正式 GitHub Release。
- 如果想在 GitHub Actions 页面手动重跑，可以用 `workflow_dispatch`；打开 `draft_release` 可以先生成草稿，检查无误后再发布。

示例：

```sh
git tag vX.Y.Z-preview
git push origin vX.Y.Z-preview

git tag vX.Y.Z
git push origin vX.Y.Z
```

## 校验值

本地构建 Windows 可执行文件后，在仓库根目录运行：

```powershell
Get-FileHash build/bin/CatScope.exe -Algorithm SHA256
```

把得到的 SHA256 写进 Release Notes，或放到发布资产说明里。

本地构建 macOS 包时运行：

```sh
scripts/build-macos.sh vX.Y.Z-preview
cd dist
shasum -a 256 -c CatScope-vX.Y.Z-preview-macos-universal.dmg.sha256
```

macOS 脚本会生成同时支持 Intel Mac 和 Apple Silicon Mac 的 universal DMG，并自动写好 `.sha256` 文件。

## macOS 预览版说明

- 当前预览版 DMG 是 self-signed/ad-hoc signed，还没有做 Apple notarization。
- 用户第一次打开时，macOS 可能会拦截。可以在 Finder 里右键打开，或到系统设置里允许该应用。
- 真正面向普通用户的 macOS 正式版，建议补上 Developer ID 签名、Apple notarization、stapling，以及 GitHub Actions secrets 配置说明。

## Logo

发布页面请使用 `Logo.png` 或 `docs/assets/logo.png`。不要用 `Logo.candidate.png`，它是带白底的原始候选图。

## 截图

建议截图命名：

- `live-logcat.png`
- `analysis-tab.png`
- `ai-context.png`
- `build-install-launch.png`
- `offline-log.png`
- `session.png`
- `workspace-presets.png`

只上传从 CatScope 实际截取的截图。不要为了凑数创建占位图。
