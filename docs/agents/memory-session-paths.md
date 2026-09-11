# Agent Memory & Session Paths — Verified Reference

Cross-referenced against official documentation, public source code, and — where possible — actual on-disk layout on the local macOS machine.

Verification date: 2026-09-11.

## TL;DR

- **~5 of 20 rows fully match** the input table.
- **~6 need correction** (path is wrong, or the memory feature doesn't exist locally).
- **~5 are unverified** — the paths in the input table have no public backing.
- **1 is deprecated** (Sourcegraph Cody → replaced by Amp).
- **1 has changed meaning** (Trae: the `bytedance/trae-agent` public repo is a research CLI, not the Trae IDE product).
- **Cross-cutting pattern**: IDE-based agents (Cursor, Copilot, Windsurf, Roo, Cline) store session transcripts in VSCode-style `workspaceStorage/<hash>/` SQLite or JSONL; CLI agents (Codex, Claude Code, Gemini CLI, Aider, Amp) use home-directory tree structures.

## Legend

✅ = matches input exactly, ✏️ = input was wrong, ⚠️ = unverified, 🚫 = agent no longer has this feature or is deprecated.

## Verified table

| # | Agent | Memory | Session | Status | Source |
|---|-------|--------|---------|--------|--------|
| 1 | **Claude Code** | `~/.claude/projects/<encoded-cwd>/memory/MEMORY.md` (auto memory, v2.1.234+; index + topic files) | `~/.claude/projects/<encoded-cwd>/<uuid>.jsonl` + `sessions-index.json` (global history: `~/.claude/history.jsonl`) | ✅ (path matches) — but note: `<encoded-cwd>` is a **path-slug with slashes→dashes** (e.g. `-Users-yicaohuang-Downloads-front-proj`), not a hash | [code.claude.com/docs/en/memory.md](https://code.claude.com/docs/en/memory.md); verified on disk at `~/.claude/projects/-Users-yicaohuang-Downloads-front-proj-codex-test/aa98c6cd-...jsonl` |
| 2 | **Cursor** | None natively. Persistent context is **Rules**: `.cursor/rules/*.mdc`, `.cursorrules`, `AGENTS.md` | `~/.cursor/projects/<encoded-path>/agent-transcripts/<uuid>/<uuid>.jsonl` (one level deeper than claimed); also `~/Library/Application Support/Cursor/User/workspaceStorage/<hash>/{state.vscdb,workspace.json}` | ✅ (memory: no native) / ✏️ (session: one level deeper) | [cursor.com/docs/rules](https://cursor.com/docs/rules) — no `~/.cursor/memories/` on disk |
| 3 | **GitHub Copilot** | **Server-side** on GitHub (`github.com/settings/copilot/memory`), not local disk. Repo-level facts + user-level preferences | Copilot CLI: `~/.copilot/session-state/` (documents only `/config.json`, `/settings.json`, `/mcp-config.json` — CLI session dir undocumented but widely reported); VS Code: `workspaceStorage/<hash>/chatSessions/<uuid>.jsonl` + `GitHub.copilot-chat/transcripts/<uuid>.jsonl` | ✏️ (memory: server-side, not `/memories/repo/`) / ⚠️ (CLI session path: community-confirmed, not in official docs) | [docs.github.com/.../copilot-memory](https://docs.github.com/en/copilot/concepts/agents/copilot-memory) — "owned by the billing entity"; [chronicle.md](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/chronicle) |
| 4 | **Windsurf** | `~/.codeium/windsurf/memories/` (verified on disk; contains `global_rules.md`) | No independent session directory. Transcripts inside `~/Library/Application Support/Windsurf/User/workspaceStorage/<hash>/state.vscdb` (SQLite) | ✅ (memory path correct; feature renamed "Context Awareness" post-Cognition acquisition) | [docs.devin.ai/desktop/cascade/skills](https://docs.devin.ai/desktop/cascade/skills) — Windsurf docs now at docs.devin.ai |
| 5 | **Cline** | Not documented as a separate concept — persistent context is **Rules** (`.clinerules/`, `.cline/rules/`) | VS Code globalStorage: `saoudrizwan.claude-dev/tasks/` (verified via extension source pattern); CLI sessions: `~/.cline/data/sessions/` (not present in the tested install; `~/.cline/data/` contains `globalState.json`, `secrets.json`, `workspaces/`) | ⚠️ (session paths community-confirmed; the "conversation_history.json" in globalStorage tasks is a legacy claim — modern Cline stores transcript differently) | [docs.cline.bot/customization/skills.md](https://docs.cline.bot/customization/skills.md); ~/.cline/data/ on disk shows `globalState.json` + `workspaces/<id>/workspaceState.json` only |
| 6 | **Aider** | No native memory mechanism (as claimed); `--read <markdown>` is the community pattern | `.aider.chat.history.md` (project dir) + `.aider.input.history` (shell readline) + optional `--llm-history-file` | ✅ | [aider.chat/docs/config/options.html#history-files](https://aider.chat/docs/config/options.html) — defaults in `aider/args.py:271-296` |
| 7 | **Gemini CLI** | `~/.gemini/tmp/<project_id>/memory/` (verified from source) | `~/.gemini/tmp/<project_id>/chats/` (per-project JSON files) | ✅ (with note: `<project_id>` in the source is a resolved ID, not a hash — the mode is `cmd`, `project`, or a user-scoped token; on-disk sample: `~/.gemini/tmp/yicaohuang/chats/session-2026-04-14T07-43-2a83abdb.json`) | [google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli) — `packages/core/src/config/storage.ts`, `packages/core/src/utils/paths.ts` |
| 8 | **Codex CLI** | `~/.codex/memories/` (empty dir + backing SQLite: `~/.codex/memories_1.sqlite` — verified on disk) | `~/.codex/sessions/<YYYY>/<MM>/<DD>/rollout-*.jsonl` (verified) + `~/.codex/archived_sessions/rollout-*.jsonl` (flat, older sessions). Session DB: `~/.codex/state_5.sqlite`, `~/.codex/thread_history_1.sqlite` | ✅ (all paths verified on disk) | [openai/codex](https://github.com/openai/codex) — `codex-rs/state/` migrations; on-disk `~/.codex/` listing |
| 9 | **OpenCode** | Not documented — no memory directory found in source | `~/.local/share/opencode/opencode.db` (global SQLite, all projects share) | ✏️ (memory: not a real feature — the `.opencode/memory/` path is fabricated; session path is correct) | [anomalyco/opencode](https://github.com/anomalyco/opencode) — `packages/core/src/database/database.ts` |
| 10 | **Continue.dev** | No native memory directory (as claimed); persistent context via **Rules** (`.continue/rules/*.mdc`) | `~/.continue/sessions/` (per-session JSON files + `sessions.json` index) | ✅ | [continuedev/continue](https://github.com/continuedev/continue) — `core/util/paths.ts`, `core/util/history.ts` |
| 11 | **Roo Code** | `.roo/` (project) and `~/.roo/` (global) — **not verified in source**; the source has no memory module | VS Code globalStorage: `rooveterinaryinc.roo-cline/tasks/<taskId>/` | ✅ (session) / ⚠️ (memory: unverified) | [RooVetGit/Roo-Code](https://github.com/RooVetGit/Roo-Code) — `src/utils/storage.ts` (session) confirmed |
| 12 | **Amp** | Server-side only (as claimed) | Server-side: `https://ampcode.com/threads/T-...` (verified from docs). Local thread mirror at `~/.local/share/amp/threads/T-*.json` — **not verifiable** (no public source) | ✅ (server-side) / ⚠️ (local mirror unverified) | [ampcode.com/docs/threads](https://ampcode.com/docs/threads), [ampcode.com/docs/cli/settings](https://ampcode.com/docs/cli/settings) — settings paths documented, thread local storage is not |
| 13 | **Goose** | `~/.config/goose/memory/` — **not found in source** (memory module exists but path not confirmed); data dir is `~/.local/share/goose/` on Linux / `~/Library/Application Support/Block/goose/` on macOS | `<data_dir>/sessions/sessions.db` (SQLite) — verified from source; on macOS: `~/Library/Application Support/Block/goose/sessions/sessions.db` | ⚠️ (memory unverified) / ✅ (session DB) | [aaif-goose/goose](https://github.com/aaif-goose/goose) (formerly Block/goose) — `crates/goose/src/session/session_manager.rs`, `crates/goose/src/config/paths.rs` |
| 14 | **OpenClaw** | `~/.openclaw/workspace/MEMORY.md` + `USER.md` + `AGENTS.md` + `SOUL.md` + `IDENTITY.md` + `memory/YYYY-MM-DD.md` | `~/.openclaw/agents/<agentId>/agent/openclaw-agent.sqlite` (per-agent SQLite: session rows, transcripts, memory index, standing intents) + `~/.openclaw/agents/<agentId>/sessions/` (legacy migration sources) | ✅ (with detail: workspace is a *single* shared workspace, not just `MEMORY.md`; sessions are owned per-agent) | [docs.openclaw.ai/concepts/agent-workspace](https://docs.openclaw.ai/concepts/agent-workspace), [/concepts/session](https://docs.openclaw.ai/concepts/session), [/concepts/memory](https://docs.openclaw.ai/concepts/memory) |
| 15 | **Kiro** | `~/.kiro/crew/workspace/memory/preferences.md` + `projects.md` + `history/{date}.md` + FTS5 index `~/.kiro/crew/memory_index.db` | CLI: `~/.kiro/sessions/cli/<id>.jsonl` (+ `<id>.json` companion); IDE: `~/.kiro/sessions/<workspace-hash>/sess_<id>/` — **unverified** (no public source) | ✅ (KiroCrew memory verified verbatim in source; CLI sessions community-confirmed) / ⚠️ (IDE sessions) | [kirodotdev/KiroCrew](https://github.com/kirodotdev/KiroCrew/blob/main/src/kiro_crew/memory.py) — memory paths in `memory.py`; CLI path via [Observal/Observal](https://github.com/Observal/Observal/blob/main/observal_cli/sessions/kiro.py) |
| 16 | **Trae** | Not documented publicly | Not documented publicly | ⚠️ (all paths in input table unverified) | The public `bytedance/trae-agent` repo is a **research CLI**, not the Trae IDE product; it uses `trajectories/trajectory_YYYYMMDD_HHMMSS.json` in cwd. Trae IDE (www.trae.ai) is closed-source; docs.trae.ai returns JS-rendered empty content. |
| 17 | **Junie** | `~/.junie/skills/` and `.junie/skills/` are verified but they are *skill directories*, not memory. Junie has no documented memory directory. For persistent instructions the file is `.junie/guidelines.md` (most common) — **not** `.junie/AGENTS.md` as claimed. | `~/.junie/sessions/<session-id>/events.jsonl` (community-confirmed via tokscale) | ✏️ (memory: `.junie/AGENTS.md` is not the convention; `guidelines.md` is) / ✅ (session, community-confirmed) | [JetBrains/intellij-community](https://github.com/JetBrains/intellij-community/blob/main/plugins/mcp-server/src/com/intellij/mcpserver/impl/McpClientDetector.kt) (skills/mcp paths); [jetbrains/klibs-io/skills/README.md](https://github.com/JetBrains/klibs-io/blob/main/skills/README.md) |
| 18 | **Sourcegraph Cody** | 🚫 Deprecated — no documented memory path; sourcegraph/cody repo returns 404 | 🚫 Deprecated — no documented session path | 🚫 | Sourcegraph deprecated Cody in favor of **Amp** ([ampcode.com](https://ampcode.com)); forks explicitly say "Sourcegraph has deprecated Cody in favour of Amp". If AI-Manager lists Cody, mark it as end-of-life. |
| 19 | **Tabnine** | `TABNINE.md` (project context) — verified via `github/spec-kit` (`extensions/agent-context/agent-context-defaults.json` declares `"tabnine": "TABNINE.md"`) | No long-term memory, no session storage documented | ⚠️ (memory: `TABNINE.md` is not memory, it's a project-context file; session: not documented) | Tabnine was acquired by **Tricentis**; current docs at [docs.tabnine.com](https://docs.tabnine.com) — minimal GitBook; the `~/.tabnine/agent/settings.json` (MCP config) is documented via elara-labs/code-context-engine |
| 20 | **Replit AI** | `replit.md` (project-level persistent context) — convention confirmed by Replit docs | Checkpoint system (managed by Replit's UI, not a local file path) | ✅ (conceptually) | Replit AI docs: [replit.com/replit-ai](https://replit.com/replit-ai) — the memory/session model is managed by Replit's platform, not by local directories. AI-Manager should probably omit Replit from the local-path table or mark it "cloud-managed". |

## What changed in the input table

Eight rows need correction, four are unverified, one is deprecated:

- **Claude Code** (partial): `<encoded-cwd>` is a path-slug with slashes→dashes, NOT a hash. Session path is `~/.claude/projects/<slug>/<uuid>.jsonl`.
- **Cursor** (partial): agent-transcripts path is `~/.cursor/projects/<encoded-path>/agent-transcripts/<uuid>/<uuid>.jsonl` (one level deeper — there's a `<uuid>/` directory). Memory: no `~/.cursor/memories/`; persistent context is Rules.
- **GitHub Copilot**: `/memories/repo/` and `/memories/session/` do not exist. Copilot Memory is a real feature but is **server-side** on github.com (`github.com/settings/copilot/memory`), owned by the billing entity, managed via the web UI.
- **Cline**: `~/.cline/data/sessions/` not confirmed in the tested install; VS Code extension uses `saoudrizwan.claude-dev/tasks/` (not `saoudrizwan.claude-dev/globalStorage/`).
- **OpenCode**: `<sessionDirectory>/.opencode/memory/` is not a real path. OpenCode has no memory feature.
- **Amp**: local `~/.local/share/amp/threads/T-*.json` unverified — Amp's thread storage is server-side at `ampcode.com/threads/T-...`.
- **Goose**: `~/.config/goose/memory/` unverified in source. Session DB is at `<data_dir>/sessions/sessions.db` where `<data_dir>` is `~/.local/share/goose/` on Linux or `~/Library/Application Support/Block/goose/` on macOS (not the flat path in the input table).
- **Trae IDE**: all paths (`~/.trae/memory/...`, `~/.trae/cli/sessions/`) unverified — product is closed-source; the public `bytedance/trae-agent` is a research CLI with different paths.
- **Junie**: memory path `.junie/AGENTS.md` is not the convention — `.junie/guidelines.md` is more common (multiple variants exist). No dedicated memory directory documented.
- **Sourcegraph Cody**: 🚫 deprecated — replaced by Amp.
- **Tabnine**: `TABNINE.md` is a project-context file, not memory per se. Session storage is not documented.

## Cross-cutting patterns

### IDE-based agents use VSCode workspaceStorage

Cursor, Windsurf, VS Code Copilot, Roo Code, and Cline all store per-workspace state under:

```
~/Library/Application Support/<IDE>/User/workspaceStorage/<hash>/{state.vscdb, workspace.json}
```

The `<hash>` is a deterministic identifier for the workspace root (VSCode uses a workspace hash; Cursor's is a path-encoded slug, per disk evidence).

- **Cursor, VS Code Copilot**: writes session transcripts as plaintext JSONL under `agent-transcripts/<uuid>/` or `chatSessions/<uuid>.jsonl`.
- **Windsurf**: stores agent transcripts as SQLite rows inside `state.vscdb` (no plaintext files — harder to grep/review).
- **Roo Code, Cline**: session tasks under `globalStorage/<extensionId>/tasks/<taskId>/`.

### CLI agents use home-directory trees

Claude Code, Codex, Gemini CLI, Aider, OpenCode, Continue, Goose, and OpenClaw all use `~/.<agent>/` with project-level subdirectories.

- **Claude Code** (`~/.claude/projects/<slug>/`), **Codex** (`~/.codex/sessions/<date>/rollout-*.jsonl`), **Gemini CLI** (`~/.gemini/tmp/<mode>/chats/`), **Aider** (`.aider.chat.history.md` in project), **OpenCode** (`~/.local/share/opencode/opencode.db` global SQLite), **Continue** (`~/.continue/sessions/`), **Goose** (`<data_dir>/sessions/sessions.db`), **OpenClaw** (`~/.openclaw/agents/<id>/agent/openclaw-agent.sqlite`).

### Memory is often a distinct feature from session storage

- **Session storage** = the raw transcript of a conversation (JSONL or SQLite).
- **Memory** = curated durable facts extracted from sessions (usually a `MEMORY.md` index + topic files, or an indexed DB).

Not every agent has both:

- Both: Claude Code (auto memory + session JSONL), OpenClaw (`MEMORY.md` + per-agent SQLite), Kiro (KiroCrew memory + CLI JSONL), Gemini CLI (`~/.gemini/tmp/<pid>/memory/` + `chats/`), Codex (`~/.codex/memories/` + `sessions/`).
- Session only: Aider, Continue, OpenCode, Goose (memory unverified), Roo, Amp (server-side).
- Memory only (or memory-as-rules): Cursor (Rules), Tabnine (`TABNINE.md`), Trae (undocumented).
- Cloud-only memory: GitHub Copilot (memories stored on github.com, not local disk).

## Sources

- Claude Code: [code.claude.com/docs/en/memory.md](https://code.claude.com/docs/en/memory.md), [docs/en/sessions](https://code.claude.com/docs/en/sessions)
- Cursor: [cursor.com/docs/rules](https://cursor.com/docs/rules) (Rules is the memory replacement)
- GitHub Copilot: [docs.github.com/.../copilot-memory](https://docs.github.com/en/copilot/concepts/agents/copilot-memory), [chronicle.md](https://docs.github.com/en/copilot/concepts/agents/copilot-cli/chronicle)
- Windsurf: [docs.devin.ai/desktop/cascade/skills](https://docs.devin.ai/desktop/cascade/skills) (Windsurf docs now at docs.devin.ai post-Cognition)
- Aider: [aider.chat/docs/config/options.html#history-files](https://aider.chat/docs/config/options.html)
- Codex CLI: [openai/codex](https://github.com/openai/codex) — `codex-rs/state/`, verified on disk at `~/.codex/`
- Gemini CLI: [google-gemini/gemini-cli](https://github.com/google-gemini/gemini-cli) — `packages/core/src/config/storage.ts`, `packages/core/src/utils/paths.ts`
- OpenCode: [anomalyco/opencode](https://github.com/anomalyco/opencode) — `packages/core/src/database/database.ts`, `packages/core/src/global.ts`
- Continue.dev: [continuedev/continue](https://github.com/continuedev/continue) — `core/util/paths.ts`, `core/util/history.ts`
- Amp: [ampcode.com/docs/threads](https://ampcode.com/docs/threads), [ampcode.com/docs/cli/settings](https://ampcode.com/docs/cli/settings)
- Goose: [aaif-goose/goose](https://github.com/aaif-goose/goose) — `crates/goose/src/session/session_manager.rs`, `crates/goose/src/config/paths.rs`
- OpenClaw: [docs.openclaw.ai/concepts/agent-workspace](https://docs.openclaw.ai/concepts/agent-workspace), [/concepts/session](https://docs.openclaw.ai/concepts/session), [/concepts/memory](https://docs.openclaw.ai/concepts/memory)
- Kiro: [kirodotdev/KiroCrew](https://github.com/kirodotdev/KiroCrew) — `src/kiro_crew/memory.py`, `src/kiro_crew/config/paths.py`
- Junie: [JetBrains/intellij-community](https://github.com/JetBrains/intellij-community/blob/main/plugins/mcp-server/src/com/intellij/mcpserver/impl/McpClientDetector.kt), [JetBrains/klibs-io/skills/README.md](https://github.com/JetBrains/klibs-io/blob/main/skills/README.md)
- Sourcegraph Cody: deprecated (sourcegraph/cody → 404); successor is [ampcode.com](https://ampcode.com)
- Tabnine: [github/spec-kit/extensions/agent-context/agent-context-defaults.json](https://github.com/github/spec-kit/blob/main/extensions/agent-context/agent-context-defaults.json) (declares `TABNINE.md`)
