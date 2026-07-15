# CatScope Roadmap

本文档描述 CatScope 的版本路线图。项目应以 MVP 为核心，避免过早变成完整 IDE。

## 总体路线

```text
v0.1：基础 Logcat Viewer
v0.2：增强日志分析
v0.3：Build / Install / Launch 调试闭环
v0.4：AI Agent 友好能力
v1.0：正式可用版本
```

## v0.1：基础 Logcat Viewer

目标：实现一个可以替代基础 `adb logcat` 的图形化工具。

### 功能范围

- 设备列表
- 刷新设备
- 启动 / 停止 Logcat
- 实时日志显示
- Level 高亮
- Tag 显示
- PID / TID 显示
- 关键词搜索
- 暂停 / 恢复
- 清屏
- 导出 txt
- 基础包名过滤

### 技术重点

- ADB 路径查找
- ADB 子进程管理
- stdout 流式读取
- threadtime 格式解析
- Ring Buffer
- 前端虚拟滚动
- 基础过滤器
- 基础 UI 布局

### 验收标准

- 可以列出连接设备。
- 可以选择设备并启动 Logcat。
- 日志可以实时刷新。
- 10 万行日志下 UI 不明显卡顿。
- 可以暂停、清屏、搜索和导出。
- 可以按 Level 和包名做基础过滤。

---

## v0.2：增强日志分析

目标：让 CatScope 真正比普通 ADB wrapper 好用。

### 功能范围

- 多 buffer 选择：main、system、crash、events、radio
- 保存过滤器
- 离线打开日志
- 右键复制当前行
- 右键复制上下文
- Java Crash 折叠
- Native Crash 识别
- ANR 识别
- 日志详情面板
- 搜索结果跳转
- 正则过滤

### 技术重点

- 多行日志合并
- Java stacktrace 归并
- Native crash 规则识别
- ANR 规则识别
- 过滤器持久化
- 离线日志解析
- 日志详情面板

### 验收标准

- FATAL EXCEPTION 可以被自动识别。
- SIGSEGV / SIGABRT 可以被自动识别。
- ANR 相关日志可以被标记。
- 可以保存、加载过滤器。
- 可以打开历史 logcat 文件重新分析。

---

## v0.3：Build / Install / Launch 调试闭环

目标：实现简单 Android 调试闭环，但不做完整 IDE。

### 功能范围

- 选择 Android 项目目录
- 识别 `gradlew` / `gradlew.bat`
- 识别 `settings.gradle` / `settings.gradle.kts`
- 选择 module
- 选择 variant
- 执行 `assembleDebug`
- 捕获构建输出
- 自动查找 APK
- 安装到当前设备
- 启动 App
- 安装后自动切换到对应包名 Logcat

### 技术重点

- Gradle Wrapper 调用
- 构建输出解析
- APK 文件定位
- adb install 错误识别
- monkey 启动 App
- 工作区配置持久化

### 验收标准

- 可以在 Windows 下执行 `gradlew.bat assembleDebug`。
- 构建成功后可以找到 APK。
- 可以安装 APK 到当前设备。
- 可以启动目标 App。
- 安装或启动失败时有明确错误提示。

---

## v0.4：AI Agent 友好能力

目标：增强 AI 辅助开发工作流。

### 功能范围

- 复制 AI 分析上下文
- 生成 Markdown Bug Report
- 生成 GitHub Issue 内容
- 导出 crash session
- 内置问题分析模板
- 复制最近 N 行日志
- 复制错误前后上下文
- 附带设备信息、App 信息、Build 输出、Install 输出

### 技术重点

- 上下文窗口提取
- Crash Session 聚合
- Markdown 模板生成
- Build / Install / Logcat 关联
- 右侧 AI 上下文面板

### 验收标准

- 用户可以选中错误日志并一键生成 AI 分析文本。
- 生成内容包含设备、包名、错误摘要、关键日志、上下文日志。
- 内容可以直接复制给 Codex、ChatGPT、Claude 或其他 Agent。

---

## v1.0：正式可用版本

目标：形成可长期使用的轻量 Android 调试工作台。

### 功能范围

- Windows / macOS / Linux 打包
- 主题系统
- 项目工作区
- 插件化 Analyzer
- 自动更新（Windows 单 EXE 已支持；macOS 自动替换待签名和 notarization 后接入）
- 配置导入导出
- 多设备并行日志
- 性能优化
- 完整错误识别库

### 验收标准

- 普通 Android 开发者可以把 CatScope 作为日常 Logcat 工具。
- 在多数常见 Android 项目中可以完成 Build / Install / Launch。
- 大日志场景下稳定可用。
- 常见崩溃可以快速定位。

## 长期方向

- 插件市场
- 团队过滤器共享
- GitHub Issue 自动生成
- Native 符号解析辅助
- 多项目工作区管理

> 当前路线不包含上传日志、云端日志分析或外部 AI API 直连。若未来讨论这些方向，必须先单独评估隐私、数据处理和产品边界。
