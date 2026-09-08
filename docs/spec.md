# AI-Manager: Kitter 功能移植

## Problem Statement

AI-Manager 当前是 Wails3 模板（托盘、设置、版本、更新、自启动），没有自己的领域功能。用户想跨项目管理 Agent Skills（一个技能库、每个项目只装需要的技能、更新一次全生效），但 AI-Manager 目前做不到——要么手动管 symlink，要么装另一个应用（Kitter，Rust/gpui 桌面端）。

## Solution

把 Kitter 的完整功能集移植到 AI-Manager 现有的 Wails3 + React 技术栈：
- 保留 AI-Manager 现有的壳（侧边栏 + 主区域 + 自定义标题栏 + 15 主题）
- 复刻 Kitter 的领域模型（Skill、Library、Project、Installation、Agent、Effective Skill、Source、Tag、Group、Install Target、Context Token Budget）
- 复刻 Kitter 的 3 个页面（Skills、Projects、Settings）和所有 flow（添加、安装、组织、标签、删除、确认）
- Go 后端用 Go 惯用写法改写（不镜像 Kitter 的 Rust 字段结构），JSON 持久化到 AI-Manager 自己的路径
- 前端用 React 推荐库（@dnd-kit、@tanstack/react-virtual、framer-motion；react-resizable-panels 已装）
- 中英双语

## User Stories

### 技能库管理

1. As a developer, I want to add a skill from a local folder, so that I can manage skills I've written or downloaded locally.
2. As a developer, I want to add a skill from a GitHub repository (via npx/skills.sh), so that I can install community skills without cloning manually.
3. As a developer, I want to add a skill from a Claude plugin source, so that I can use skills distributed through Claude's plugin system.
4. As a developer, I want to adopt existing skill installations from my projects, so that I can bring scattered skills under centralized management without moving their source directories.
5. As a developer, I want to see the list of all skills in my library, so that I can find and manage them at a glance.
6. As a developer, I want to search and filter skills by name, so that I can find a specific skill quickly in a large library.
7. As a developer, I want to see each skill's description, source, and installation count, so that I understand what it does and where it's used.
8. As a developer, I want to view a skill's files (SKILL.md, agents/*, references/*), so that I can inspect its contents without leaving the app.
9. As a developer, I want to remove a skill from the library, so that I can clean up skills I no longer need.
10. As a developer, I want a confirmation dialog before deleting a skill, so that I don't accidentally delete something important.
11. As a developer, I want to check for skill updates, so that I can keep my skills up to date.
12. As a developer, I want to update a specific skill to its latest version, so that I get bug fixes and new features.

### 项目与安装

13. As a developer, I want to add a project (by browsing for a folder), so that I can install skills into it.
14. As a developer, I want to see my recent projects, so that I can quickly re-select one I used recently.
15. As a developer, I want to install selected skills into a project, so that agents working in that project can discover them.
16. As a developer, I want to choose which agents to install skills for (Universal, Claude Code, Codex, Cursor, etc.), so that I can target specific agent runtimes.
17. As a developer, I want to install skills globally (user-level) instead of per-project, so that I can use certain skills everywhere.
18. As a developer, I want to uninstall a skill from a project, so that I can remove it when it's no longer needed.
19. As a developer, I want to see where each skill is installed (which projects, which targets), so that I can understand its reach.
20. As a developer, I want installations to be symlinks (not copies), so that one update propagates to all linked projects.

### 有效技能视图

21. As a developer, I want to see the effective skill set for each agent in a project, so that I know what's actually active.
22. As a developer, I want to see both managed installations and unmanaged skills (outside AI-Manager), so that I have the full picture.
23. As a developer, I want to see where each effective skill comes from (managed, unmanaged, built-in, plugin), so that I can distinguish sources.
24. As a developer, I want to see an estimated context token cost per agent, so that I can spot skills adding unnecessary overhead.
25. As a developer, I want warnings when context token cost exceeds thresholds, so that I'm alerted to potential performance issues.

### 标签与分组

26. As a developer, I want to create tags to organize and filter skills, so that I can find related skills easily.
27. As a developer, I want to create parent-child tag hierarchies (two levels), so that I can categorize skills hierarchically.
28. As a developer, I want to assign and unassign tags to skills, so that I can categorize them dynamically.
29. As a developer, I want to drag and drop skills onto tags to assign them, so that I can organize quickly.
30. As a developer, I want to create groups to install multiple skills at once, so that I can batch-install related skills.
31. As a developer, I want to move skills between groups, so that I can reorganize my library.
32. As a developer, I want to delete groups (optionally moving their skills to another group), so that I can clean up.

### 多选与批量操作

33. As a developer, I want to select multiple skills (click, Shift+click for range, Cmd/Ctrl+click for toggle), so that I can batch-operate.
34. As a developer, I want to install multiple selected skills at once, so that I save time.
35. As a developer, I want to delete multiple selected skills at once, so that I can clean up efficiently.
36. As a developer, I want multi-select to protect built-in skills from accidental deletion, so that I don't break the app.

### 设置

37. As a developer, I want to switch between English and Chinese, so that I can use the app in my preferred language.
38. As a developer, I want to switch between light and dark themes, so that I can match my system preference.
39. As a developer, I want to choose from 15 built-in themes, so that I can customize the app's appearance.
40. As a developer, I want to change the library directory path, so that I can store skills where I want.
41. As a developer, I want to browse for the library directory via a folder picker, so that I don't have to type paths manually.

### UI 交互

42. As a developer, I want resizable panes (sidebar, list, detail), so that I can adjust the layout to my preference.
43. As a developer, I want context menus on right-click, so that I can access actions without hunting through menus.
44. As a developer, I want keyboard shortcuts, so that I can work efficiently without a mouse.
45. As a developer, I want toast notifications for actions (added, removed, installed, updated), so that I get feedback.
46. As a developer, I want smooth animations for dialogs and transitions, so that the app feels polished.
47. As a developer, I want a search input on the skills page, so that I can filter the list instantly.
48. As a developer, I want collapsible groups and content directories, so that I can focus on what matters.
49. As a developer, I want a detail pane with tabs (description, installs, files), so that I can view different aspects of a skill.

## Implementation Decisions

### 后端（Go）

1. **包结构按模块组织**：`internal/skill/`、`internal/library/`、`internal/project/`、`internal/source/`、`internal/tags/`、`internal/effective/`、`internal/agents/`、`internal/config/`、`internal/platform/`。每个模块是独立 package，Go 惯用写法（struct + method，不镜像 Kitter 的 Rust 字段）。

2. **持久化格式：JSON 文件**。`config.json`（配置：语言、主题、库路径、最近项目、项目活动）、`registry.json`（注册表：技能、来源、分组）、`tags.json`（标签：两层树 + 分配）、`skills/<storage_name>/`（技能文件）。原子写：tmp 文件 + rename。

3. **持久化路径：AI-Manager 自己的**。macOS: `~/Library/Application Support/AIManager/`，Windows: `%LOCALAPPDATA%\AIManager\`，Linux: `~/.local/share/AIManager/`。支持环境变量覆盖。

4. **API 暴露 22 个方法**（Wails3 bindings）：
   - 技能库：`ListSkills`、`GetSkill`、`AddSkill`、`RemoveSkill`、`UpdateSkill`、`CheckUpdates`
   - 安装：`InstallSkills`、`UninstallSkill`、`ListInstallations`
   - 项目：`ListProjects`、`AddProject`、`RemoveProject`、`BrowseProject`
   - 有效技能：`GetEffectiveSkills`
   - 标签/分组：`ListTags`、`AddTag`、`RenameTag`、`DeleteTag`、`AssignTag`、`ListGroups`、`AddGroup`、`DeleteGroup`、`MoveSkillToGroup`
   - 配置：`GetConfig`、`UpdateConfig`、`BrowseLibrary`

5. **InstallTarget 10 个变体**：Universal（`.agents/skills`）、Codex、ClaudeCode、Cursor、OpenCode、Pi、Grok、Antigravity、Droid、Copilot。每个对应不同的 symlink 路径。

6. **AgentKind 10 个变体**：Codex、ClaudeCode、Cursor、OpenCode、Copilot、Antigravity、Amp、Droid、Pi、Grok。每个有发现根目录和跨 agent 读取支持。

7. **4 种来源扫描**：Local（文件夹扫描）、Npx/skills.sh/GitHub（npx add）、Claude plugin（claude plugin install）、Existing installations（adoption 扫描）。

8. **Wails3 bindings 包名从 `skeleton` 改成 `ai-manager`**：改 `main.go` 的 module path + 前端所有 `@bindings/skeleton/...` import。

### 前端（React）

9. **保留现有壳**：`App.tsx` 的 ResizablePanelGroup（sidebar + content）不变。不引入 Kitter 的 `layout.rs` 可调整大小面板。

10. **页面优先**：Skills → Projects → Settings。替换现有 `HomePage.tsx`。

11. **新增 shadcn 组件**：dialog、popover、dropdown-menu、tooltip、input、checkbox、badge、card、select、alert（Radix 包已在 deps，需补 shadcn 封装）。

12. **新增 React 库**：@dnd-kit/core（标签拖拽）、@tanstack/react-virtual（技能列表虚拟化）、framer-motion（动画）。react-resizable-panels 已装。

13. **状态管理**：zustand，每个领域一个 store（`modules/skills/store.ts`、`modules/projects/store.ts`、`modules/tags/store.ts`、`modules/i18n/store.ts`）。

14. **i18n**：简单字符串选择器（跟 Kitter 的 `uses_english()` 一致），不用 i18next。一个 i18n store 存当前语言，一个 `t(en, zh)` 函数做切换。

15. **模块组织**：`pages/`（SkillsPage、ProjectsPage、SettingsPage）、`modules/skills/`（列表、详情、添加 flow、安装 flow）、`modules/projects/`（项目列表、有效技能视图）、`modules/tags/`（标签管理、拖拽）、`modules/flows/`（跨页面共享 flow）、`modules/i18n/`（中英切换）。

### 架构决策记录

- `docs/adr/0001-keep-existing-shell.md` — 保留现有壳，不复刻 Kitter 的可调整大小面板
- `docs/adr/0002-desktop-only-no-cli.md` — 只写桌面端，不做 CLI

### 领域术语

见 `CONTEXT.md`，11 个术语：Skill、Library、Project、Installation、Agent、Effective Skill、Source、Tag、Group、Install Target、Context Token Budget。

## Testing Decisions

1. **Go 侧全测**：每个新模块（`skill/`、`library/`、`project/`、`source/`、`tags/`、`effective/`、`agents/`）用 table-driven tests。已有先例：`config_test.go`、`settings_service_test.go`、`search_service_test.go`、`update_service_test.go`。

2. **测试外部行为**：测试 API 方法的输入输出（如 `ListSkills` 返回正确的技能列表），不测试内部实现细节。

3. **关键测试点**：
   - 技能库：添加/移除/更新/检查更新
   - 安装：symlink 创建/删除、多 target 安装
   - 有效技能：跨 agent 扫描、token 估算
   - 标签：两层树操作（添加/重命名/删除/分配/移动）
   - 来源：4 种扫描逻辑（本地文件夹、npx、Claude plugin、adoption）

4. **前端第一版手动测试**：不引入 vitest，通过 Wails3 dev mode 手动验证。后续 ticket 可加 vitest。

5. **集成测试**：通过 Wails3 dev mode 端到端验证（Go 后端 + React 前端 + Wails3 bindings）。

## Out of Scope

1. **CLI**：不做命令行工具（见 ADR-0002）。
2. **像素级复刻 Kitter**：不复刻 Kitter 的精确 token（row 30、control 28、radii 10-25），用 AI-Manager 自己的 Tailwind token + 15 主题系统。
3. **Kitter 的 gpui layout API**：不复刻 `h_resizable`，用 AI-Manager 已有的 react-resizable-panels。
4. **前端单元测试**：第一版不做（无 vitest），后续可加。
5. **Kitter 的 `_kitter-builtin` 内置技能**：不内置，AI-Manager 有自己的技能库，不预装 Kitter 的技能。
6. **多语言扩展**：只支持 EN + ZH，不支持其他语言。

## Further Notes

1. **Ticket 切分（12 张）**：
   - T0: shadcn 组件补齐 + bindings 包名改 `ai-manager`
   - T1: Go 后端数据模型 + 持久化
   - T2: Go 后端技能库 API（4 种来源扫描）
   - T3: 前端 Skills 页壳（列表 + 详情，空内容）
   - T4: 前端添加 flow（4 种来源）
   - T5: 前端删除确认
   - T6: Go 后端项目 + 安装（InstallTarget 10 变体、agent 扫描）
   - T7: 前端安装 flow
   - T8: 前端 Projects 页（有效技能视图 + token 估算）
   - T9: 前端组织/标签 flow（分组 + 标签拖拽）
   - T10: 前端 Settings 页扩展
   - T11: i18n + 快捷键 + 上下文菜单 + 完善

2. **依赖关系**：T0 和 T1 是前置（无依赖）；T2 依赖 T1；T3 依赖 T2；T4/T5 依赖 T3；T6 依赖 T1；T7 依赖 T3+T6；T8 依赖 T6；T9 依赖 T3+T8；T10 无依赖；T11 依赖 T3+T8+T10。

3. **每张 ticket 是 tracer-bullet**：做完就能跑、能看到进展。

4. **Wails3 bindings 生成**：每次 Go 后端改动后，运行 `wails generate bindings` 重新生成 `frontend/bindings/`。
