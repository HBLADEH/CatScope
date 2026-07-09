# 贡献指南

感谢你愿意改进 CatScope。项目希望保持为轻量、可靠的 Android Logcat 排障工具；请优先贡献能改善日志查看、诊断、ADB 兼容性和文档体验的改动。

## 开始前

1. 搜索现有 Issue 和 Pull Request，避免重复工作。
2. 影响较大、会改变产品边界或交互方式的改动，请先创建 Feature request 讨论方案。
3. 不要把真实日志、token、个人信息、内部包名、`.catscope-session` 或私有 APK 提交到仓库。

CatScope 不计划成为 Android Studio 的替代品。代码编辑、Gradle Project Sync、完整 IDE 能力和外部 AI API 不在当前范围内。

## 本地开发

要求：Go 1.22+、Node.js 20+、npm 10+，以及 Wails v2 CLI。实时 Logcat 开发还需要可用的 Android SDK Platform Tools。

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

前端开发服务或 `wails dev` 验证结束后，请主动关闭进程，避免遗留后台服务或端口占用。

## 提交改动

- 保持改动小而聚焦；不要顺手重构无关模块。
- Go 代码优先使用 table-driven tests；前端使用 Vue 3 Composition API 和 TypeScript。
- 大量日志列表必须保持虚拟滚动，默认 ring buffer 上限为 100000 条。
- 外部命令必须使用 `exec.Command`/`exec.CommandContext` 显式传参，不能拼接 shell 命令。
- AI Context 只能在本地生成 Markdown，不能上传日志或调用云端模型。
- 用户可见行为、架构边界或发布流程改变时，同步更新相应文档。

建议使用 Conventional Commit 风格，例如 `fix: 修复离线日志解析` 或 `docs: 更新发布流程`。

## 验证清单

按改动范围执行必要检查：

```powershell
# Go 后端
go test ./...
go vet ./...

# 前端（在 frontend/ 目录）
npm run build

# Windows 发布相关改动
scripts\check.ps1

# 所有文档改动
git diff --check
```

若修改 Wails 暴露的方法或前后端共享结构，请同时检查 `frontend/wailsjs` 绑定、Go 测试和前端构建。请在 Pull Request 中说明未运行的检查及原因。

## Pull Request 要求

请使用仓库的 PR 模板，并说明：

- 用户会看到什么变化，以及为什么；
- 覆盖过的自动化测试与手工验证；
- UI 改动的截图或录屏；
- 日志、截图、AI Context 与 session 文件已经脱敏。

维护者会优先检查功能边界、可恢复的错误处理、跨平台影响、测试覆盖和文档一致性。

## 报告问题与提出建议

- 可复现的缺陷：使用 Bug report 模板。
- 新工作流或功能想法：使用 Feature request 模板。
- 安全问题：请遵循 [SECURITY.md](./SECURITY.md)，不要公开提交漏洞细节。
