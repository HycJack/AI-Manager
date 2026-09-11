# Agent Provider Config / Session / Memory 接入

## Problem Statement

AI-Manager 目前只管「技能」一个维度：技能库 → 项目 → 安装目标。用户在自己的机器上跑多个 agent（Claude Code、Codex、Gemini CLI、OpenCode、Cursor、Windsurf、Cline、Amp、Goose、OpenClaw…），每个 agent 都有自己的：

1. **Provider 配置**：用哪个模型提供商、哪个模型、base URL、API key、`settings.json` / `config.toml` / `opencode.json` 之类的配置文件。
2. **Session 数据**：对话记录存在哪（JSONL / SQLite / 独立 JSON），怎么按项目分组、怎么按时间排序、怎么跳到某一条。
3. **Memory 数据**：agent 持久记忆（`MEMORY.md` + topic files、KiroCrew 的 `preferences.md`、OpenClaw 的 `memory/YYYY-MM-DD.md`、Codex 的 `memories_1.sqlite` 等）。

这些信息现在散落在 `~/.claude/`、`~/.codex/`、`~/.gemini/`、`~/Library/Application Support/Cursor/` 等十多个地方，用户没法在一个界面里看到「我所有 agent 现在用什么模型、最近 session 是什么、有哪些记忆」，也没法统一配置。Kitter 不管这个，AI-Manager 应该补上这一层——把 agent 的运行时状态和配置纳入桌面端统一管理。

参考文档：`docs/agents/skill-paths.md`（技能目录，已交叉验证）与 `docs/agents/memory-session-paths.md`（session / memory 路径，已交叉验证），验证日期 2026-09-11。本文档的所有路径都从那两份表的「Verified table」中取，不复制「unverified / community-confirmed」的行到默认实现里（那些走可选加载）。

## Solution

新增一个「Agent」维度，与现有「Skill」「Project」并列。AI-Manager 把每个 agent 的 provider 配置、session、memory 接入：

- **后端**：新增 `internal/agents/paths.go`（agent 路径常量表，从 docs 直接落地）、`internal/providers/`（provider 配置读写，按 agent 分适配器）、`internal/sessions/`（session 扫描 + 元数据提取）、`internal/memory/`（memory 扫描 + 读取）。一个 `AgentService` 通过 Wails3 bindings 暴露。
- **前端**：新增 `AgentsPage`（左侧 agent 列表，右侧三标签：Provider / Sessions / Memory）。SettingsPage 里加一个「Agents」入口。
- **默认只读 session 与 memory**：session/memory 提供浏览、搜索、按项目/时间过滤、在外部编辑器中打开；不写入（避免误改 agent 的运行状态）。Provider 配置可编辑（本来就是用户自己的设置）。
- **路径数据可验证**：路径常量从 `docs/agents/*.md` 的 Verified table 1:1 落地，不引入未验证路径；未验证/社区确认的路径作为「可选加载」开关，默认不启用。

## Scope

### In scope

按 `docs/agents/memory-session-paths.md` 的「Verified table」，本版本默认接入以下 agent（有本地 provider config + session + memory 路径，且路径已在文档中验证）：

| Agent | Provider config | Session | Memory |
|-------|-----------------|---------|--------|
| Claude Code | `~/.claude/settings.json` / `~/.claude.json` | `~/.claude/projects/<slug>/*.jsonl` | `~/.claude/projects/<slug>/memory/MEMORY.md` + topic files |
| Codex CLI | `~/.codex/config.toml` | `~/.codex/sessions/<YYYY>/<MM>/<DD>/rollout-*.jsonl` | `~/.codex/memories_1.sqlite` + `~/.codex/memories/` |
| Gemini CLI | `~/.gemini/settings.json` | `~/.gemini/tmp/<project_id>/chats/*.json` | `~/.gemini/tmp/<project_id>/memory/` |
| OpenCode | `~/.config/opencode/opencode.json` | `~/.local/share/opencode/opencode.db` (global SQLite) | 无（OpenCode 无 memory 特性） |
| Continue.dev | `~/.continue/config.yaml` | `~/.continue/sessions/*.json` + `sessions.json` | 无（用 Rules，非 SKILL.md） |
| Cline | 通过 VS Code 设置 | `~/.cline/data/sessions/`（未在测试机上确认，走可选加载） | 无（用 Rules） |
| Windsurf | `~/.codeium/windsurf/settings.json` | `~/Library/Application Support/Windsurf/User/workspaceStorage/<hash>/state.vscdb` | `~/.codeium/windsurf/memories/global_rules.md` |
| OpenClaw | `<state-dir>/config.toml`（按 agentId） | `~/.openclaw/agents/<id>/agent/openclaw-agent.sqlite` | `~/.openclaw/workspace/MEMORY.md` + `USER.md` + `AGENTS.md` + `SOUL.md` + `IDENTITY.md` + `memory/YYYY-MM-DD.md` |
| Kiro | `~/.kiro/config.yaml`（CLI） | `~/.kiro/sessions/cli/<id>.jsonl` + `<id>.json` | `~/.kiro/crew/workspace/memory/preferences.md` + `projects.md` + `history/{date}.md` + FTS5 `~/.kiro/crew/memory_index.db` |
| Amp | 服务端（`ampcode.com`）；本机镜像 `~/.local/share/amp/threads/T-*.json` 未验证 | 服务端；本地镜像未验证（走可选加载） | 服务端 |
| Goose | `<data_dir>/config.yaml` | `<data_dir>/sessions/sessions.db` (SQLite) | `~/.config/goose/memory/`（未在源码中确认，走可选加载） |
| Roo Code | 通过 VS Code 设置 | VS Code `globalStorage/rooveterinaryinc.roo-cline/tasks/<taskId>/` | `.roo/` / `~/.roo/`（未在源码中确认，走可选加载） |

### Out of scope（本版本不做，但保留扩展位）

- **Continue.dev**：用 `.continue/rules/*.mdc`，不是 SKILL.md 格式；provider config 和 session 接入，memory 不接入。
- **GitHub Copilot**：memory 在服务端（`github.com/settings/copilot/memory`），本机无路径；CLI session 路径社区确认但未在官方文档中公开，走可选加载。
- **Cursor**：无 native memory（persistent context 是 Rules）；session 路径有，provider config 走 VS Code 设置，本版本只读展示。
- **DeepSeek**：DeepSeek 无第一方 coding CLI；不接入（避免把 Deep Code 误标成 DeepSeek）。
- **Trae IDE**：路径未公开（`bytedance/trae-agent` 是研究用 CLI，不是 Trae IDE 产品）；不接入。
- **Grok**：无公开 SKILL.md 文档；不接入。
- **Junie**：memory 路径约定不明（`.junie/guidelines.md` vs `.junie/AGENTS.md`）；本版本不接入 memory，session 路径社区确认，走可选加载。
- **Sourcegraph Cody**：已 deprecated（被 Amp 取代）；不接入。
- **Tabnine / Replit AI**：memory 是 project-context 文件 / 云端管理；不接入本地路径。
- **Write 到 session / memory**：本版本只读。写 session 会破坏 agent 内部一致性；写 memory 需要按 agent 语义（KiroCrew 的 FTS5 索引、OpenClaw 的 per-agent SQLite）定制写入路径，超出本版本范围。

### 可选加载开关

对「路径未验证 / 社区确认 / 源码中没找到」的 agent 路径，在 `Preferences` 里加一个 `EnableUnverifiedAgentPaths` 开关，默认 `false`。用户手动开启后，AI-Manager 才扫描这些路径；默认不启用，避免把 unverified 路径当成事实。

## Data Model

### Agent 元数据（新增到 `internal/agents/`）

```go
// internal/agents/paths.go

// AgentPaths 描述一个 agent 在本机上的持久化路径。所有路径都带 {home} / {project} / {state_dir} 占位符，
// 由 ResolveAgentPaths 在运行时替换。
type AgentPaths struct {
    // ProviderConfigPath 是 agent 的 provider 配置入口（可能不存在，也可能有多个入口）。
    ProviderConfigPath string `json:"provider_config_path"`

    // SessionRootPath 是 session 文件的根目录（可能是一个目录，也可能是单个 SQLite 文件）。
    SessionRootPath string `json:"session_root_path"`

    // SessionFormat 描述 session 文件如何组织（"jsonl-per-project" / "sqlite-global" / "json-flat" / …）。
    SessionFormat SessionFormat `json:"session_format"`

    // MemoryRootPath 是 memory 数据的根目录（可能不存在）。
    MemoryRootPath string `json:"memory_root_path"`

    // MemoryFormat 描述 memory 数据格式（"markdown-index" / "sqlite" / "markdown-flat" / "rules-only" / …）。
    MemoryFormat MemoryFormat `json:"memory_format"`
}

type SessionFormat string
const (
    SessionFormatJSONLPerProject SessionFormat = "jsonl-per-project" // Claude Code, Codex
    SessionFormatSQLiteGlobal    SessionFormat = "sqlite-global"     // OpenCode, Windsurf, Goose, OpenClaw, Kiro (FTS5)
    SessionFormatJSONFlat        SessionFormat = "json-flat"         // Gemini CLI, Continue
    SessionFormatVSCodeStorage   SessionFormat = "vscode-storage"    // Cursor, Copilot, Cline, Roo
    SessionFormatNone            SessionFormat = "none"              // Amp (server-side), Replit AI
)

type MemoryFormat string
const (
    MemoryFormatMarkdownIndex  MemoryFormat = "markdown-index" // Claude Code, OpenClaw, Kiro, Gemini CLI
    MemoryFormatSQLite         MemoryFormat = "sqlite"         // Codex, Kiro (FTS5)
    MemoryFormatMarkdownFlat   MemoryFormat = "markdown-flat"  // Windsurf
    MemoryFormatRulesOnly      MemoryFormat = "rules-only"     // Cursor, Cline, Tabnine (persistent context 是 Rules，不是 memory)
    MemoryFormatServerSide     MemoryFormat = "server-side"    // Copilot, Amp
    MemoryFormatNone           MemoryFormat = "none"           // Continue, OpenCode, DeepSeek
)
```

路径表（从 `docs/agents/memory-session-paths.md` 的 Verified table 1:1 落地，路径占位符用 `{home}` / `{project}` / `{state_dir}`）：

```go
// DefaultAgentPaths returns the verified agent path table.
// Source: docs/agents/memory-session-paths.md (verified 2026-09-11).
func DefaultAgentPaths() map[AgentKind]AgentPaths {
    return map[AgentKind]AgentPaths{
        AgentClaudeCode: {
            ProviderConfigPath: "{home}/.claude/settings.json",
            SessionRootPath:    "{home}/.claude/projects",
            SessionFormat:      SessionFormatJSONLPerProject,
            MemoryRootPath:     "{home}/.claude/projects/{project}/memory",
            MemoryFormat:       MemoryFormatMarkdownIndex,
        },
        AgentCodex: {
            ProviderConfigPath: "{home}/.codex/config.toml",
            SessionRootPath:    "{home}/.codex/sessions",
            SessionFormat:      SessionFormatJSONLPerProject,
            MemoryRootPath:     "{home}/.codex/memories_1.sqlite",
            MemoryFormat:       MemoryFormatSQLite,
        },
        AgentOpenCode: {
            ProviderConfigPath: "{home}/.config/opencode/opencode.json",
            SessionRootPath:    "{home}/.local/share/opencode/opencode.db",
            SessionFormat:      SessionFormatSQLiteGlobal,
            MemoryRootPath:     "",
            MemoryFormat:       MemoryFormatNone,
        },
        // …其余 agent 同理
    }
}
```

### Provider 配置（`internal/providers/`）

每个 agent 的 provider 配置文件格式不同（JSON / TOML / YAML / 通过 VS Code 设置），用 adapter 模式：

```go
// ProviderConfig 是 provider 配置的通用视图。agent 特有字段放在 Extra 里。
type ProviderConfig struct {
    Agent    AgentKind        `json:"agent"`
    Path     string           `json:"path"`     // 实际配置文件路径（已替换 {home}）
    Format   string           `json:"format"`   // "json" / "toml" / "yaml" / "vscode-settings"
    Exists   bool             `json:"exists"`
    Provider string           `json:"provider"` // "openai" / "anthropic" / "deepseek" / "custom" / "unknown"
    Model    string           `json:"model"`
    BaseURL  string           `json:"base_url"`
    APIKey   string           `json:"api_key"`    // 已脱敏（只回显后 4 位）
    Extra    map[string]any   `json:"extra"`      // agent 特有字段
    Errors   []string         `json:"errors"`     // 读取失败的字段
}

// ProviderAdapter 按 agent 定制 provider 配置的读/写。
type ProviderAdapter interface {
    Agent() AgentKind
    Read() (*ProviderConfig, error)
    Write(cfg *ProviderConfig) error
}
```

按 agent 实现：

| Agent | Adapter | 配置文件 | 格式 |
|-------|---------|----------|------|
| Claude Code | `ClaudeCodeProviderAdapter` | `~/.claude/settings.json` + `~/.claude.json` | JSON |
| Codex CLI | `CodexProviderAdapter` | `~/.codex/config.toml` | TOML |
| Gemini CLI | `GeminiProviderAdapter` | `~/.gemini/settings.json` | JSON |
| OpenCode | `OpenCodeProviderAdapter` | `~/.config/opencode/opencode.json` | JSON |
| Continue | `ContinueProviderAdapter` | `~/.continue/config.yaml` | YAML |
| Windsurf | `WindsurfProviderAdapter` | `~/.codeium/windsurf/settings.json` | JSON |
| OpenClaw | `OpenClawProviderAdapter` | `<state-dir>/config.toml` | TOML |
| Kiro | `KiroProviderAdapter` | `~/.kiro/config.yaml` | YAML |
| 其他 | `UnimplementedProviderAdapter` | 返回 `ProviderConfig{Exists: false}` | — |

**API key 脱敏**：`Read()` 返回值中 `APIKey` 只显示后 4 位（`****abcd`）；`Write()` 时如果 `APIKey` 是 `****` 占位，保留原值不动。

### Session（`internal/sessions/`）

```go
// Session 描述一个 agent session 的元数据 + 内容入口。
type Session struct {
    Agent      AgentKind  `json:"agent"`
    ID         string     `json:"id"`
    Project    string     `json:"project"`  // 项目路径（按 agent 规则从 session 文件中提取）
    StartedAt  time.Time  `json:"started_at"`
    EndedAt    time.Time  `json:"ended_at,omitempty"`
    MessageCount int       `json:"message_count"`
    TokenEstimate int      `json:"token_estimate,omitempty"` // 估算（字符数 / 4）
    Path       string     `json:"path"`     // 原始文件路径（JSONL/SQLite row）
    Format     string     `json:"format"`   // "jsonl" / "sqlite" / "json"
}

// SessionScanner 按 agent 定制 session 扫描。
type SessionScanner interface {
    Agent() AgentKind
    Scan() ([]Session, error)
    // ReadTranscript 读取 session 原始内容（read-only）。
    ReadTranscript(id string) (string, error)
}
```

按 `SessionFormat` 分派到不同扫描器：

- `SessionFormatJSONLPerProject`（Claude Code / Codex）：遍历 `~/.claude/projects/<slug>/*.jsonl`，从 slug 反推项目路径（slash→dash 反向），每个 JSONL 是一个 session，从第一行提取 metadata。
- `SessionFormatSQLiteGlobal`（OpenCode / Windsurf / Goose / OpenClaw / Kiro）：打开 SQLite，按 agent 约定的 schema 查表。SQLite 只读打开（`file:...?mode=ro`），避免锁冲突。
- `SessionFormatJSONFlat`（Gemini CLI / Continue）：遍历 `<root>/*.json`，每个 JSON 是一个 session，提取 metadata。
- `SessionFormatVSCodeStorage`（Cursor / Copilot / Cline / Roo）：走可选加载；读取 `workspaceStorage/<hash>/state.vscdb`，查 `chatSessions` / `agent-transcripts` 表。

### Memory（`internal/memory/`）

```go
// MemoryEntry 描述一条 memory。
type MemoryEntry struct {
    Agent     AgentKind `json:"agent"`
    Path      string    `json:"path"`
    Title     string    `json:"title"`
    Kind      string    `json:"kind"` // "index" / "topic" / "daily" / "preference"
    Content   string    `json:"content"`
    SizeBytes int       `json:"size_bytes"`
    ModifiedAt time.Time `json:"modified_at"`
}

// MemoryScanner 按 agent 定制 memory 扫描。
type MemoryScanner interface {
    Agent() AgentKind
    Scan() ([]MemoryEntry, error)
    // ReadEntry 读取一条 memory 的完整内容（read-only）。
    ReadEntry(path string) (string, error)
}
```

按 `MemoryFormat` 分派：

- `MemoryFormatMarkdownIndex`（Claude Code / OpenClaw / Kiro / Gemini CLI）：读 `MEMORY.md` 作为索引，解析 `[[topic]]` 链接（Obsidian 风格）列出 topic files；OpenClaw 的 workspace 还有 `USER.md`、`AGENTS.md`、`SOUL.md`、`IDENTITY.md` 一并展示。
- `MemoryFormatSQLite`（Codex / Kiro FTS5）：只读打开 SQLite，按 schema 查表返回 memory 行。
- `MemoryFormatMarkdownFlat`（Windsurf）：读 `~/.codeium/windsurf/memories/global_rules.md` 作为单条 memory。
- `MemoryFormatRulesOnly`（Cursor / Cline / Tabnine）：返回空列表 + 提示「该 agent 无 native memory，persistent context 是 Rules」。
- `MemoryFormatServerSide`（Copilot / Amp）：返回空列表 + 提示「memory 在服务端」+ 服务端 URL（`https://github.com/settings/copilot/memory` / `https://ampcode.com/threads`）。
- `MemoryFormatNone`（Continue / OpenCode / DeepSeek）：返回空列表。

## Architecture

### 后端包结构

```
internal/agents/
  agents.go            # 已有：AgentKind、InstallTarget、AgentDir、TargetDir、DiscoverSkills
  config.go            # 已有：AgentConfigFile、LoadAgentConfig、SaveAgentConfig（agents.json）
  paths.go             # 新增：AgentPaths、SessionFormat、MemoryFormat、DefaultAgentPaths、ResolveAgentPaths
  paths_test.go        # 新增

internal/providers/    # 新增
  adapter.go           # ProviderConfig、ProviderAdapter 接口、注册表
  claude_code.go       # ClaudeCodeProviderAdapter
  codex.go             # CodexProviderAdapter（TOML）
  gemini.go            # GeminiProviderAdapter
  opencode.go          # OpenCodeProviderAdapter
  continue.go          # ContinueProviderAdapter（YAML）
  windsurf.go          # WindsurfProviderAdapter
  openclaw.go          # OpenClawProviderAdapter
  kiro.go              # KiroProviderAdapter
  adapter_test.go
  providers_test.go

internal/sessions/     # 新增
  scanner.go           # Session、SessionScanner 接口、注册表、格式分派
  jsonl.go             # JSONLPerProjectScanner
  sqlite.go            # SQLiteGlobalScanner
  json.go              # JSONFlatScanner
  vscode_storage.go    # VSCodeStorageScanner（可选加载）
  scanner_test.go

internal/memory/       # 新增
  scanner.go           # MemoryEntry、MemoryScanner 接口、注册表、格式分派
  markdown_index.go    # MarkdownIndexScanner（Obsidian 风格 topic files）
  sqlite.go            # SQLiteMemoryScanner
  markdown_flat.go     # MarkdownFlatScanner
  scanner_test.go

internal/app/          # 已有
  agent_service.go     # 新增：AgentService（Wails3 bindings）
  agent_service_test.go
```

### AgentService API（Wails3 bindings）

```go
// AgentService 暴露 agent 的 provider / session / memory 视图。
type AgentService struct { state *State }

// 元数据
func (s *AgentService) ListAgents() ([]AgentInfo, error)
func (s *AgentService) GetAgent(agent string) (*AgentInfo, error)

// Provider 配置
func (s *AgentService) GetProviderConfig(agent string) (*providers.ProviderConfig, error)
func (s *AgentService) UpdateProviderConfig(agent string, patch providers.ProviderConfig) (*providers.ProviderConfig, error)

// Session
func (s *AgentService) ListSessions(agent string, filter SessionFilter) ([]sessions.Session, error)
func (s *AgentService) GetSessionTranscript(agent, id string) (string, error)
func (s *AgentService) OpenSessionInEditor(agent, id string) error  // open -a 打开原始文件

// Memory
func (s *AgentService) ListMemory(agent string) ([]memory.MemoryEntry, error)
func (s *AgentService) GetMemoryEntry(agent, path string) (*memory.MemoryEntry, error)
func (s *AgentService) OpenMemoryInEditor(agent, path string) error

// 开关
func (s *AgentService) SetEnableUnverifiedPaths(enabled bool) error
```

`AgentInfo` 是前端 UI 需要的展示信息：

```go
type AgentInfo struct {
    Kind         AgentKind          `json:"kind"`
    Label        string             `json:"label"`
    IconType     string             `json:"icon_type"`
    Paths        agents.AgentPaths  `json:"paths"`
    SupportsProvider bool           `json:"supports_provider"`
    SupportsSession  bool           `json:"supports_session"`
    SupportsMemory   bool           `json:"supports_memory"`
    Status       string             `json:"status"` // "verified" / "unverified" / "not-supported"
}
```

### 前端结构

```
frontend/src/
  pages/
    AgentsPage.tsx               # 新增：左侧 agent 列表 + 右侧三标签（Provider / Sessions / Memory）
  modules/
    agents/
      store.ts                   # 新增：zustand store（agent 列表、当前选中、缓存）
      provider/
        ProviderForm.tsx         # Provider 配置表单（model、provider、base URL、API key）
      sessions/
        SessionList.tsx          # Session 列表（按项目/时间过滤，虚拟滚动）
        SessionViewer.tsx        # Session 详情（transcript 渲染）
      memory/
        MemoryList.tsx           # Memory 列表（topic tree + daily log）
        MemoryViewer.tsx         # Memory 详情（Markdown 渲染）
```

### UI 布局

```
┌─────────────────────────────────────────────────────────────┐
│  Sidebar (App)  │  AgentsPage                               │
│                  │  ┌─────────────────────────────────────┐  │
│  [Skills]       │  │ [Provider] [Sessions] [Memory]       │  │
│  [Projects]     │  ├──────────┬──────────────────────────┤  │
│  [Agents] ← 当前│  │ Agent    │  Detail                  │  │
│  [Settings]     │  │ List     │                          │  │
│                  │  │          │                          │  │
│                  │  │ ●Claude  │  Model: claude-3.5       │  │
│                  │  │ ●Codex   │  Provider: Anthropic     │  │
│                  │  │ ●Cursor  │  Base URL: https://...   │  │
│                  │  │ ○OpenCode│  API Key: ****abcd       │  │
│                  │  │ ○OpenClaw│                          │  │
│                  │  │ ○Kiro    │  [Save] [Revert]         │  │
│                  │  │ ○Goose   │                          │  │
│                  │  │ ○...     │                          │  │
│                  │  └──────────┴──────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

三个标签的交互：

- **Provider**：表单式，model / provider / base URL / API key 可编辑；保存时调 `UpdateProviderConfig`；API key 脱敏显示。
- **Sessions**：左侧 agent 列表（已选 agent 下的 session 列表），右侧 transcript viewer。按项目分组 + 按时间倒序。支持搜索 session ID / project / 内容（本地全文搜索）。
- **Memory**：左侧 topic tree（从 MEMORY.md 索引解析），右侧 Markdown viewer。OpenClaw 额外展示 USER.md / AGENTS.md / SOUL.md / IDENTITY.md 等 workspace 文件。

## User Stories

1. As a developer, I want to see all my agents and their current provider/model configuration in one place, so that I don't have to flip between `~/.claude/settings.json`, `~/.codex/config.toml`, and `~/.config/opencode/opencode.json` manually.
2. As a developer, I want to edit an agent's provider config (model, base URL, API key) from AI-Manager, so that I can switch providers without opening the raw config file.
3. As a developer, I want AI-Manager to mask my API keys when displaying them, so that I don't accidentally leak credentials in screenshots.
4. As a developer, I want to see my recent sessions for each agent, so that I can find a previous conversation without digging into `~/.claude/projects/`.
5. As a developer, I want to filter sessions by project, so that I can see only the sessions related to the project I'm working on.
6. As a developer, I want to view a session transcript (read-only) inside AI-Manager, so that I can review a conversation without leaving the app.
7. As a developer, I want to open a session file in my external editor, so that I can inspect the raw JSONL when I need to.
8. As a developer, I want to see the estimated token count for a session, so that I can spot unusually long conversations.
9. As a developer, I want to see all my agent memories in one place, so that I can check what my agents have remembered.
10. As a developer, I want to see OpenClaw's workspace files (MEMORY.md, USER.md, AGENTS.md, SOUL.md, IDENTITY.md, memory/*.md) as a topic tree, so that I can navigate the agent's persistent identity.
11. As a developer, I want to see a clear notice when an agent's memory is server-side (Copilot, Amp) or doesn't exist (Continue, OpenCode), so that I understand why the list is empty.
12. As a developer, I want unverified / community-confirmed paths to be opt-in via a settings toggle, so that I don't get false positives from paths that aren't actually used.
13. As a developer, I want the path table to be sourceable from `docs/agents/*.md`, so that when an agent changes its path, I can update the docs and the code in one place.

## Implementation Decisions

1. **路径常量直接来自 `docs/agents/memory-session-paths.md` 的 Verified table**：`internal/agents/paths.go` 的 `DefaultAgentPaths()` 与该文档的 Verified table 一一对应。任何新增 agent 都要先更新文档再改代码（文档优先）。

2. **未验证路径走 `EnableUnverifiedAgentPaths` 开关**：默认 `false`，不扫描 community-confirmed 但官方未文档化的路径。开启后，scanner 才把那些路径加进来。

3. **SQLite 只读打开**：所有 SQLite 操作都用 `file:...?mode=ro`，避免锁冲突和误写。Go 用 `modernc.org/sqlite`（纯 Go，无 CGO 依赖）。

4. **JSONL 解析用 `json.Decoder` 逐行**：Claude Code / Codex 的 JSONL 文件可能很大（几 MB），不能一次性读入内存。`ReadTranscript` 流式读取，返回原始字符串；前端做 Markdown 渲染。

5. **Provider config 用 adapter 模式**：每个 agent 一个 adapter，不共享 JSON 字段结构（agent 之间 schema 不同）。新增 agent 只需实现 adapter + 在注册表里加一行。

6. **API key 脱敏**：`Read()` 返回 `****<后4位>`；`Write()` 时如果传入的是 `****` 前缀，保留原值。脱敏在 adapter 内部完成，Service 层不处理。

7. **Session 元数据按格式分派**：JSONL / SQLite / JSON 各有自己的 metadata 提取逻辑，不试图统一 schema。`Session` 结构体是 UI 展示的公共视图，agent 特有字段走 `Extra`（暂不做，未来如需可扩展）。

8. **Memory 的 Markdown 索引解析**：Claude Code 的 `MEMORY.md` 用 Obsidian 风格 `[[topic]]` 链接；解析时用正则 `\[([^\]]+)\]` 提取 topic 名，再按 agent 约定的 topic 文件命名规则（如 `<topic>.md`）扫描。OpenClaw 的 workspace 还有一组固定文件（USER.md / AGENTS.md / SOUL.md / IDENTITY.md）作为顶层节点。

9. **新增 `internal/agents/paths.go` 而不是改 `config.go`**：`config.go` 目前只管 `~/.aimanager/agents.json`（用户自定义的 agent 显示配置）；`paths.go` 管的是「agent 在本机的实际持久化路径」，是产品内置的事实数据，不是用户配置。两个文件职责不同，不合并。

10. **Wails3 bindings 重新生成**：`internal/app/agent_service.go` 新增后，跑 `wails3 generate bindings` 重新生成 `frontend/bindings/`。

## Testing Decisions

1. **Go 侧全测**：
   - `internal/agents/paths_test.go`：验证 `DefaultAgentPaths()` 返回的每个 agent 路径都能解析出有效的绝对路径；验证 `ResolveAgentPaths` 正确替换 `{home}` / `{project}` / `{state_dir}` 占位符。
   - `internal/providers/providers_test.go`：每个 adapter 的 `Read()` 在临时目录上返回正确的 `ProviderConfig`；`Write()` 写回去后 `Read()` 能读回相同值；API key 脱敏逻辑正确。
   - `internal/sessions/scanner_test.go`：每种 `SessionFormat` 的 scanner 在临时目录上扫描出预期的 session 列表；SQLite scanner 只读打开、不报错。
   - `internal/memory/scanner_test.go`：每种 `MemoryFormat` 的 scanner 在临时目录上扫描出预期的 memory 列表；Markdown 索引解析 `[[topic]]` 链接正确。

2. **路径回归测试**：新增 `internal/agents/paths_verify_test.go`，从 `docs/agents/memory-session-paths.md` 提取 Verified table，断言 `DefaultAgentPaths()` 与文档一致。文档变更时测试会失败，提醒同步代码。

3. **API key 脱敏专项测试**：`internal/providers/adapter_test.go` 验证：写入真实 key → `Read()` 返回 `****<后4位>` → 用脱敏值 `Write()` 回去 → `Read()` 仍然是原值。

4. **SQLite 只读测试**：scanner 打开 SQLite 后用 `INSERT` / `UPDATE` 触发 error（验证只读模式生效）。

5. **前端第一版手动测试**：不引入 vitest，通过 Wails3 dev mode 手动验证。后续 ticket 可加。

6. **集成测试**：通过 Wails3 dev mode 端到端验证（Go 后端 + React 前端 + Wails3 bindings）。

## Out of Scope

1. **写 session / memory**：本版本只读。写 session 会破坏 agent 内部一致性；写 memory 需要按 agent 语义定制写入路径，超出本版本范围。
2. **跨 agent session 搜索**：本版本按 agent 分列，不做跨 agent 全文搜索（性能 + 实现复杂度）。后续可加全局搜索。
3. **Session 摘要 / token 精确计算**：token estimate 用「字符数 / 4」粗算，不调用 LLM 做摘要。精确 token 计数需要模型特定的 tokenizer，本版本不做。
4. **云端 memory 代理**：Copilot / Amp 的 memory 在服务端，本版本只展示「服务端 URL」+ 空列表，不做云端代理读写。
5. **Agent 安装 / 卸载 / 版本管理**：本版本只管「已安装的 agent 的配置和运行时数据」，不管 agent 本身。
6. **Agent 之间的 session/memory 迁移**：不做跨 agent 迁移。
7. **Agent 运行时状态（当前正在跑的 session）**：本版本只读已落盘的 session 文件，不接入 agent 的实时进程状态。

## Further Notes

### Ticket 切分（建议 8 张）

- **T12**：`internal/agents/paths.go` — AgentPaths 结构体 + `DefaultAgentPaths()` 表（从 docs 落地）+ `ResolveAgentPaths` + 路径回归测试
- **T13**：`internal/providers/` — ProviderConfig + ProviderAdapter 接口 + 5 个核心 adapter（Claude Code / Codex / OpenCode / Gemini / Continue）+ 测试
- **T14**：`internal/providers/` — 剩余 adapter（Windsurf / OpenClaw / Kiro / 其他 unimplemented）+ API key 脱敏测试
- **T15**：`internal/sessions/` — Session + SessionScanner 接口 + JSONL / SQLite / JSON 三种扫描器 + 测试
- **T16**：`internal/memory/` — MemoryEntry + MemoryScanner 接口 + MarkdownIndex / SQLite / MarkdownFlat 三种扫描器 + 测试
- **T17**：`internal/app/agent_service.go` — AgentService 暴露 ListAgents / GetProviderConfig / UpdateProviderConfig / ListSessions / GetSessionTranscript / ListMemory / GetMemoryEntry / OpenSessionInEditor / OpenMemoryInEditor / SetEnableUnverifiedPaths
- **T18**：`frontend/src/pages/AgentsPage.tsx` + `modules/agents/` — 三标签 UI（Provider / Sessions / Memory）+ zustand store + 接入 SettingsPage 入口
- **T19**：可选加载开关（`EnableUnverifiedAgentPaths`）+ VS Code storage scanner + Cursor / Copilot / Cline / Roo 接入 + 文档同步

### 依赖关系

- T12 是前置（无依赖）。
- T13 / T14 依赖 T12。
- T15 / T16 依赖 T12。
- T17 依赖 T12 + T13 + T14 + T15 + T16。
- T18 依赖 T17。
- T19 依赖 T17 + T18。

### 风险

1. **agent 路径变更**：agent 升级可能改路径。缓解：文档驱动（`docs/agents/*.md` 是单一真相），每次升级同步更新；路径回归测试在 CI 里跑。
2. **SQLite schema 变化**：OpenCode / Windsurf / OpenClaw / Kiro 的 SQLite schema 可能随版本变化。缓解：scanner 做 schema 探测，找不到表时返回空列表 + 提示「schema 未识别」，不报错。
3. **API key 泄露**：用户截图 / 共享 UI 时可能泄露脱敏后的 key。缓解：默认脱敏到后 4 位；`UpdateProviderConfig` 写入时不记录明文到日志。
4. **大文件性能**：Claude Code 的 JSONL session 文件可能很大。缓解：流式读取，UI 端虚拟滚动。
5. **文档与代码漂移**：`DefaultAgentPaths()` 与 `docs/agents/memory-session-paths.md` 不同步。缓解：路径回归测试断言两者一致。

### 不在本版本做的事（明确）

- 不做 CLI（见 `docs/adr/0002-desktop-only-no-cli.md`）。
- 不做像素级复刻 Kitter（见 `docs/adr/0001-keep-existing-shell.md`）。
- 不做云端 agent（Copilot / Amp）的 memory 代理。
- 不做跨 agent session 全文搜索。
- 不做 session / memory 写入。
