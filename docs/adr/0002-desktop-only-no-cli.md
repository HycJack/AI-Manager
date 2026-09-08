# 只写桌面端，不做 CLI

Kitter 同时发布桌面 app 和独立 CLI（`kitter add/install/project/update`），共享同一套核心库。本决策选择 AI-Manager 只写桌面端，不做 CLI。

理由：CLI 是独立 artifact，需要额外的入口、参数解析、输出格式化；桌面端的 Wails3 bindings 已经覆盖全部操作面。先保证桌面端跑通 Kitter 的全部功能，后续如果需要 CLI，再把核心逻辑抽成可复用包加 CLI 入口——但那是独立 effort，不阻塞桌面端。

后果：AI-Manager 没有 `ai-manager` 命令行工具；所有操作通过 UI 完成。如果用户想批量操作（如脚本化添加技能），需要绕过 UI 直接操作 JSON 文件。这个取舍是可接受的——Kitter 的 CLI 是 nice-to-have，不是核心价值。
