# 变更记录

本文件记录面向用户的重要变化。版本发布后会在此处和对应的 GitHub Release Notes 中同步说明。

格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)，版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### 文档与协作

- 补充贡献、安全、行为准则和支持入口。
- 建立中文文档索引，并明确 PR、验证与发布资料的入口。

## [0.6.6] - 2026-07-16

### 新增

- 启动后自动检查 GitHub Release，支持正式版和 Preview 更新通道。
- Windows 单 EXE 支持下载新版本、校验 SHA256、退出替换并重新启动。

### 改进

- 统一运行时、Wails 资源、前端包与 Release tag 的版本号，并在发版 workflow 中强制检查一致性。
- 前端生产构建会清理旧的 dist 资源，避免历史文件被重复打进 EXE。
- 更新中英文用户指南、发布资产说明和 QA 检查项。

完整发布说明见 [v0.6.6](./docs/releases/v0.6.6.md)。

## [0.6.5]

### 新增

- 日志 Tag 支持通过右键快捷追加包含或排除过滤。

## [0.6.4]

### 新增

- 搜索框支持 `tag:`、`pid:`、`level:`、`message:` 和负向字段查询。

### 改进

- 日志表格按实际行高测量虚拟列表，长 message 最多显示四行预览。
- Release workflow 在构建前安装前端依赖。

完整发布说明见 [v0.6.4](./docs/releases/v0.6.4.md)。
