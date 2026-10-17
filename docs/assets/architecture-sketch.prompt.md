# 架构图生成提示词

生成方式：内置 `image_gen`，`generate` 模式；不透明背景。

下面保留本次最终使用的完整提示词，便于架构变化后更新配图。

```text
Use case: infographic-diagram
Asset type: CatScope developer onboarding architecture poster, reusable in repository documentation.
Primary request: 请根据以下已核对的真实架构，绘制一张中文的手绘卡通架构图。让第一次接触项目的开发者一眼看懂：谁负责界面、谁干活、日志从哪里来、怎样变成可读结果。风格亲切，但技术关系必须准确。
Style/medium: 精致的手绘技术笔记、可爱猫咪工程师、圆润的铅笔和细马克笔线条、少量蜡笔涂色；有纸面质感，干净、专业、容易阅读。猫咪必须参与工作（操作屏幕、搬运日志、拿放大镜），不能遮挡标签。不要用一大堆猫或装饰填充页面。
Composition/framing: 一张横向 4:3 高分辨率海报，约 2400×1800。留白充足，中文大而清晰，模块标题和箭头方向明显。视觉层次从上往下：前端操作台 → Wails 桥梁 → Go 后端工坊 → 本地工具与文件。用自然的手绘分区、少量柔和的蓝、橙、绿点色。所有文字写成简体中文，保留指定英文名称。每个模块只保留下面给出的短标签，不增加大段说明。
Text and required layout:
TOP TITLE: "CatScope 架构图"
SUBTITLE: "跟着一条日志，认识整个工作台"

TOP LEVEL, a broad computer workstation labeled "① 前端操作台" and "Vue 3 + TypeScript".
Three simple sublabels inside:
"App.vue · 界面入口"
"logs.ts / Pinia · 状态与过滤"
"LogTable / LogDetails · 虚拟滚动与详情"
A small operator cat with a laptop.

BETWEEN FRONTEND AND BACKEND, an actual small cartoon bridge labeled "② Wails v2 桥梁".
Print "utils/wails.ts ↔ app.go".
Use arrows in both directions; annotate "调用方法 ↔ 返回数据".
Add a prominent small note "日志每 250ms 批量拉取".
This is an in-process desktop bridge, not HTTP, not a web server, not WebSocket; do not draw cloud/internet between the frontend and Go.

BACKEND ENCLOSURE: "③ Go 后端工坊", containing the dispatcher "app.go · 任务调度".
Inside this enclosure, use a clearly readable left-to-right data path:
LEFT, three small source icons with labels:
"Android 设备" with "internal/adb · 实时采集"
"离线文件" with ".txt / .log / .jsonl"
"Session" with "internal/storage · 恢复现场"
MIDDLE, the main log-processing station labeled "internal/logcat".
Inside show "解析与多行合并" → "Ring Buffer · 10万条".
Android live and offline text files feed the parser; Session restore feeds the buffer directly, bypassing parsing.
Draw a line from the buffer up toward the Wails bridge and frontend, labeled "日志批次". The frontend is responsible for filtering the displayed logs.
RIGHT, a separate magnifying-glass cat station labeled "规则分析".
Draw a branch from the log buffer to this station, not an obligatory step on the path to the frontend.
Short category label: "Crash / ANR / Native / JNI / 安装失败".
Then a downward short arrow to a paper sheet station labeled "internal/ai".
Short output label "本地 Markdown" and note "AI Context 不调用模型 API".
Do not depict an AI model or external service for log analysis or AI Context.

BOTTOM, four clean small tool drawers/support stations, visually subordinate to the core log route:
"internal/build" with "Gradle → APK → ADB 安装启动"
"internal/workspace" with "配置与预设 · JSON"
"internal/storage" with "保存与恢复 · Session"
"internal/update" with "GitHub Release 更新"
These are backend supporting functions, not four obligatory sequential stages.
Use a tiny wrench/build icon, folder, archive box, update arrow respectively.
Make the update station visually separate from the log/AI path; it may access GitHub, but no logs flow to GitHub.

BOTTOM CLOSING LINE: "看界面找 Vue，查流程找 app.go，改能力找 internal/"
Constraints: This must be a readable architecture diagram with concrete modules and causal arrows, not a marketing poster, not a dense UML graph. Keep arrows sparse, do not cross text, do not imply the analyzer runs before logs reach the UI. All three log sources share the viewing/analysis UI. No invented database, REST server, external AI API, microservices, Kotlin backend, Docker, IDE editor, or Android Studio. Render technical names exactly. No watermark.
```
