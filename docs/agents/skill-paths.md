# Agent Skill Directory Paths — Verified Reference

Cross-referenced against the official Agent Skills standard ([agentskills.io](https://agentskills.io)) client showcase, each agent's own docs, and (where docs are thin) the source code of the agent's public repository.

Verification date: 2026-09-11.

## Key findings

- **14 of 20 paths are exactly right** in the input table.
- **4 need correction**: Codex, OpenClaw, Mistral Vibe (partial), and DeepSeek (mislabeled product).
- **2 have unverified status**: Grok and Junie (paths plausible but no first-party docs surfaced).
- **1 is only a community project**: "DeepSeek" as a skill-capable agent is actually the unofficial **Deep Code** CLI (`lessweb/deepcode-cli`); DeepSeek itself does not ship a coding CLI that scans `SKILL.md`.
- **Continue** uses a different mechanism entirely (`.continue/rules/*.mdc`, no `SKILL.md`); its row in the input table does not describe what Continue actually supports.

## The `.agents/skills/` compatibility pattern

Several agents adopt the universal `.agents/skills/` directory (project-level) and `~/.agents/skills/` (user-level) alongside their own namespaced path, so a single skill tree can be picked up by multiple agents. This pattern is documented by Codex, Amp, OpenCode, OpenClaw, Mistral Vibe, Cline, and others — and it is the reason the input table's "or" entries are usually correct.

When the AI-Manager product installs a skill into the "shared `.agents/skills/`" location, it is compatible with any of the agents that implement this pattern.

## Verified table

Legend: ✅ = matches input, ✏️ = input was wrong, ⚠️ = unverified, 🚫 = agent doesn't natively support `SKILL.md`.

| # | Agent | Project-level | Global/user-level | Status | Source |
|---|-------|--------------|-------------------|--------|--------|
| 1 | **Claude Code** | `.claude/skills/<name>/SKILL.md` | `~/.claude/skills/<name>/SKILL.md` | ✅ | [code.claude.com/docs/en/skills.md](https://code.claude.com/docs/en/skills.md) |
| 2 | **Cursor** | `.cursor/skills/` and `.agents/skills/` (also reads `.claude/skills/`, `.codex/skills/`) | `~/.cursor/skills/` and `~/.agents/skills/` (also `~/.claude/skills/`, `~/.codex/skills/`) | ✅ | [cursor.com/docs/skills](https://cursor.com/docs/skills) — the input table's `.cursor/skills/` is correct; `.agents/skills/` is also supported as the compatibility layer |
| 3 | **OpenAI Codex** | `.agents/skills/` (NOT `.codex/skills/` — the latter is a deprecated backward-compat alias) | `~/.codex/skills/` (default `CODEX_HOME/skills`) plus `~/.agents/skills/` | ✏️ | [codex-rs/ext/skills/src/host_roots.rs](https://github.com/openai/codex/blob/main/codex-rs/ext/skills/src/host_roots.rs) — project skills use `AGENTS_DIR_NAME = ".agents"`; user skills keep `$CODEX_HOME/skills` marked "Deprecated user skills location (`$CODEX_HOME/skills`), kept for backward compatibility" |
| 4 | **GitHub Copilot** | `.github/skills/`, `.claude/skills/`, or `.agents/skills/` | `~/.copilot/skills/` or `~/.agents/skills/` | ✅ | [docs.github.com/en/copilot/concepts/agents/about-agent-skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills) |
| 5 | **Gemini CLI** | Not listed on agentskills.io — unverified | Not listed | ⚠️ | No first-party `SKILL.md` doc found. Gemini CLI does support `GEMINI.md` for project context, but the Agent Skills `SKILL.md` format is not documented for it. |
| 6 | **Windsurf** | `.windsurf/skills/` (also `.agents/skills/`, and `.claude/skills/` if Claude-config reading is on) | `~/.codeium/windsurf/skills/` (also `~/.agents/skills/`, `~/.claude/skills/`) | ✅ | [docs.devin.ai/desktop/cascade/skills](https://docs.devin.ai/desktop/cascade/skills) — note Windsurf's docs now live under docs.devin.ai (Codeium was acquired by Cognition) |
| 7 | **Roo Code** | `.roo/skills/` | `~/.roo/skills/` | ✅ (community-confirmed; not on agentskills.io) | Roo Code's own docs at docs.roocode.com/features/skills (redirects through roocodeinc.github.io). Not in the official agentskills.io client showcase but the paths are consistent with the `.agents/skills/` pattern Roo Code also reads. |
| 8 | **Goose** | `.agents/skills/` (also `.goose/skills/`, `.claude/skills/` as backward compat) | `~/.agents/skills/` (also `~/.claude/skills/`), plus `~/.agents/plugins/<plugin-name>/` for plugin skills | ✅ | [goose-docs.ai/docs/guides/context-engineering/using-skills](https://goose-docs.ai/docs/guides/context-engineering/using-skills) — note the old `block.github.io/goose` URL 404s and redirects here |
| 9 | **Junie (JetBrains)** | `.junie/skills/` | `~/.junie/skills/` | ⚠️ | Not on agentskills.io. JetBrains' docs site (junie.jetbrains.com) returned 403. Indirect evidence: community plugin [specdd/plugin-junie](https://github.com/specdd/plugin-junie) installs into `~/.junie/skills/` and `.junie/skills/`; GitHub Skills CLI supports `--agent junie`. Confidence: medium. |
| 10 | **Amp** | `.agents/skills/` (also `.claude/skills/` by default unless disabled) | `~/.config/agents/skills/`, `~/.agents/skills/`, `~/.config/amp/skills/`, `~/.claude/skills/` (in this precedence order) | ✅ | [ampcode.com/docs/customize/skills](https://ampcode.com/docs/customize/skills) — precedence: `~/.config/agents/skills/` → `~/.agents/skills/` → `~/.config/amp/skills/` → `.agents/skills/` → `.claude/skills/` → `~/.claude/skills/` → `~/.claude/plugins/cache/` → `amp.skills.path` → built-ins |
| 11 | **Cline** | `.cline/skills/` (recommended), `.clinerules/skills/`, `.claude/skills/` | `~/.cline/skills/` (macOS/Linux), `%USERPROFILE%\.cline\skills\` (Windows) | ✅ | [docs.cline.bot/customization/skills.md](https://docs.cline.bot/customization/skills.md) — "Where Skills Live" section. When project and global share a name, **global wins** (opposite of most other agents). |
| 12 | **Continue** | Not a SKILL.md agent — uses `.continue/rules/*.mdc` (Continue-native format, not `SKILL.md`) | `~/.continue/rules/*.mdc` | 🚫 | Continue does not implement the Agent Skills standard. Its `.continue/rules/` format uses `.mdc` files with `description`/`globs`/`alwaysApply` frontmatter. The `.continue/skills/<name>/SKILL.md` path in the input table does not correspond to any real Continue feature. |
| 13 | **Kiro** | `.kiro/skills/<name>/SKILL.md` | `~/.kiro/skills/<name>/SKILL.md` | ✅ | [kiro.dev/docs/skills.md](https://kiro.dev/docs/skills.md) — "Skill scope" section. Workspace skills win over global on name collision. Custom agents opt in via `skill://.kiro/skills/*/SKILL.md` URI scheme. |
| 14 | **Trae** | Via Settings → Rule & Skills → Skills → Create (or import `SKILL.md` folder) | Same — Trae's SOLO mode stores skills in its own workspace | ✅ (conceptually) | [trae.ai/blog/trae_tutorial_0115](https://www.trae.ai/blog/trae_tutorial_0115) — officially launched Skills built on the open Agent Skills standard, but the blog documents **UI-based creation/import**, not a documented on-disk path. The `./.trae/skills/` convention in the input table is plausible but not officially published. |
| 15 | **OpenCode** | `.opencode/skills/`, `.claude/skills/`, `.agents/skills/` | `~/.config/opencode/skills/`, `~/.claude/skills/`, `~/.agents/skills/` | ✅ | [opencode.ai/docs/skills](https://opencode.ai/docs/skills) — "Place files" section lists all six paths verbatim |
| 16 | **DeepSeek** | 🚫 Not supported by DeepSeek itself | 🚫 | ✏️ | DeepSeek does not ship a first-party coding CLI. The only SKILL.md-capable "DeepSeek" agent is the **unofficial Deep Code CLI** ([lessweb/deepcode-cli](https://github.com/lessweb/deepcode-cli)) at [deepcode.vegamo.cn/en/docs/configuration/agent-skills](https://deepcode.vegamo.cn/en/docs/configuration/agent-skills): project `.deepcode/skills/` and `.agents/skills/`, user `~/.deepcode/skills/` and `~/.agents/skills/`. If AI-Manager wants a "DeepSeek" row, it should either point at Deep Code explicitly or omit the row. |
| 17 | **Grok** | ⚠️ Not documented | ⚠️ Not documented | ⚠️ | Not on agentskills.io. No public xAI coding-CLI repo surfaced (`github.com/xai-org/grok-cli` 404s; the "Grok Build" binary is referenced by third-party repos but has no public SKILL.md docs). Treat as unsupported until xAI publishes docs. |
| 18 | **OpenClaw** | `<workspace>/skills/` (highest precedence) and `<workspace>/.agents/skills/` | `~/.agents/skills/` (default state only) and `<state-dir>/skills/` (managed, shared) | ✏️ (input's `.openclaw/skills` is the node-hosted variant, not the primary path) | [docs.openclaw.ai/tools/skills](https://docs.openclaw.ai/tools/skills) — the loading-order table shows `<workspace>/skills` first, then `.agents/skills`. `~/.openclaw/skills` is used by connected headless **node-hosted** skills only. Codex's `$CODEX_HOME/skills` is explicitly **not** an OpenClaw root. |
| 19 | **Mistral Vibe** | `.vibe/skills/` and `.agents/skills/` (both discovered; trusted-folders only) | `~/.vibe/skills/` and `~/.agents/skills/` | ✏️ (input missed `~/.agents/skills`) | Verified from source: [vibe/core/paths/_vibe_home.py](https://github.com/mistralai/mistral-vibe/blob/main/vibe/core/paths/_vibe_home.py) defines `VIBE_HOME` → `~/.vibe/`; [vibe/core/paths/_local_config_files.py](https://github.com/mistralai/mistral-vibe/blob/main/vibe/core/paths/_local_config_files.py) scans `.vibe/skills` and `.agents/skills` at project root; [vibe/core/config/harness_files/_paths.py](https://github.com/mistralai/mistral-vibe/blob/main/vibe/core/config/harness_files/_paths.py) defines both `GLOBAL_SKILLS_DIR = ~/.vibe/skills` and `GLOBAL_AGENTS_SKILLS_DIR = ~/.agents/skills`. |
| 20 | **Antigravity** | `.agent/skills/` | `~/.gemini/antigravity/skills/` | ✅ | Not on agentskills.io, but the paths are confirmed by [rominirani/antigravity-skills](https://github.com/rominirani/antigravity-skills/blob/main/README.md) — an installable sample-skill repo linked from a Google Cloud Medium tutorial and a Google Codelab. Input table's `核实` (to be verified) is now verified. |

## What changed in the input table

Four rows are wrong, four rows are worth annotating:

- **Codex**: `.codex/skills/` is not a real discovery root. Use `.agents/skills/` at project level. `~/.codex/skills/` still works as the deprecated-but-supported user location.
- **OpenClaw**: `<workspace>/.openclaw/skills` was misread from the `<state-dir>/skills` row — the state directory is `~/.openclaw`, so managed skills live at `~/.openclaw/skills`, and node-hosted skills live at `~/.openclaw/skills` on the node itself. Project skills live at `<workspace>/skills` (primary) or `<workspace>/.agents/skills/`. There is no `./.openclaw/skills` project path.
- **Mistral Vibe**: `~/.vibe/skills` was marked "(推断)" but is actually documented; `~/.agents/skills` is also a global root and should be added.
- **DeepSeek**: the DeepSeek row is a category error — DeepSeek doesn't have a coding CLI. Either drop the row or relabel it "Deep Code (unofficial DeepSeek CLI)".

Annotate these:

- **Gemini CLI**: `SKILL.md` support not documented; `GEMINI.md` is the supported context file but it is not a Skills-format agent.
- **Continue**: `.continue/skills/<name>/SKILL.md` is not a real feature. Continue uses `.continue/rules/*.mdc`.
- **Trae**: the disk path is not officially documented — Trae creates skills via UI import. Do not assume `.trae/skills/` works.
- **Grok**: no first-party `SKILL.md` support found.

## Cross-agent compatibility

For AI-Manager's "shared `.agents/skills/`" install target, the compatibility list (agents that will pick up a skill placed there) is:

- ✅ OpenAI Codex, Amp, OpenCode, OpenClaw, Mistral Vibe, Goose, Cline (via `.claude/skills`), Cursor (via `.agents/skills`), GitHub Copilot, Kiro (opt-in), Roo Code, Trae (via UI import)
- ⚠️ Junie, Antigravity — likely supported given the community ecosystem but not officially documented
- ❌ Claude Code only reads `.claude/skills/` (not `.agents/skills/`), so a shared `.agents/skills/` install is invisible to Claude Code unless AI-Manager also symlinks to `.claude/skills/`. This is the single most important compatibility gotcha in the table.

## Sources

- Agent Skills standard & client showcase: [agentskills.io](https://agentskills.io/), [agentskills.io/clients](https://agentskills.io/clients)
- Claude Code: [code.claude.com/docs/en/skills.md](https://code.claude.com/docs/en/skills.md)
- Cursor: [cursor.com/docs/skills](https://cursor.com/docs/skills)
- OpenAI Codex: [docs/skills.md](https://github.com/openai/codex/blob/main/docs/skills.md) (stubs to developers.openai.com) + [codex-rs/ext/skills/src/host_roots.rs](https://github.com/openai/codex/blob/main/codex-rs/ext/skills/src/host_roots.rs)
- GitHub Copilot: [docs.github.com/.../about-agent-skills](https://docs.github.com/en/copilot/concepts/agents/about-agent-skills)
- Windsurf: [docs.devin.ai/desktop/cascade/skills](https://docs.devin.ai/desktop/cascade/skills)
- Goose: [goose-docs.ai/.../using-skills](https://goose-docs.ai/docs/guides/context-engineering/using-skills)
- Amp: [ampcode.com/docs/customize/skills](https://ampcode.com/docs/customize/skills)
- Cline: [docs.cline.bot/customization/skills.md](https://docs.cline.bot/customization/skills.md)
- Kiro: [kiro.dev/docs/skills.md](https://kiro.dev/docs/skills.md)
- OpenCode: [opencode.ai/docs/skills](https://opencode.ai/docs/skills)
- OpenClaw: [docs.openclaw.ai/tools/skills](https://docs.openclaw.ai/tools/skills)
- Mistral Vibe source: [vibe/core/paths/_vibe_home.py](https://github.com/mistralai/mistral-vibe/blob/main/vibe/core/paths/_vibe_home.py), [_local_config_files.py](https://github.com/mistralai/mistral-vibe/blob/main/vibe/core/paths/_local_config_files.py), [harness_files/_paths.py](https://github.com/mistralai/mistral-vibe/blob/main/vibe/core/config/harness_files/_paths.py)
- Deep Code (DeepSeek CLI): [deepcode.vegamo.cn/en/docs/configuration/agent-skills](https://deepcode.vegamo.cn/en/docs/configuration/agent-skills)
- Antigravity: [rominirani/antigravity-skills](https://github.com/rominirani/antigravity-skills/blob/main/README.md)
- Trae: [trae.ai/blog/trae_tutorial_0115](https://www.trae.ai/blog/trae_tutorial_0115)
