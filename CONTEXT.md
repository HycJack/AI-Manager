# AI-Manager

A desktop application for managing Agent Skills across projects: one maintained library, each project gets only what it needs, updates propagate through managed links.

## Language

**Skill**:
A single agent capability — a self-contained unit of instructions, tools, or resources an agent can discover and use. A Skill has one maintained source and can be installed into one or more projects.
_Avoid_: plugin, extension, capability, feature

**Library**:
The maintained collection of Skills on this machine. Every Skill in the Library is the single source of truth for its installations.
_Avoid_: repository, collection, store

**Project**:
A directory on disk that one or more Skills can be installed into, so that agents working in that directory can discover them.
_Avoid_: workspace, folder, repo (too generic)

**Installation**:
A managed link from a Skill to a specific location within a Project. Installations are links, not copies — they point back to the Skill's source in the Library.
_Avoid_: copy, deployment, binding

**Agent**:
A runtime (Claude, Codex, Openclaw, AMP, Antigravity, Trae, …) that can discover Skills from the project, user, or built-in locations. Each agent has its own discovery roots.
_Avoid_: bot, runner, client

**Effective Skill**:
What an agent actually sees when it scans its discovery roots — the union of managed installations, unmanaged Skills (outside Kitter), built-in Skills, and plugin-provided Skills.
_Avoid_: active skill, installed skill (too narrow — misses unmanaged)

**Source**:
Where a Skill originally came from: a local folder, a GitHub repository, a skills.sh-compatible endpoint, or a Claude plugin source.
_Avoid_: origin, upstream, remote

**Tag**:
A user-assigned label for organizing and filtering Skills. Tags are scoped (library-wide or project-specific).
_Avoid_: label, category, mark

**Group**:
A named collection of Skills used to install several at once (e.g. "all the frontend skills"). Distinct from Tag: groups drive the install flow, tags drive the view/filter.
_Avoid_: bundle, set

**Install Target**:
Where within a Project a Skill gets installed: the shared `.agents/skills` directory (universal) or an agent-specific directory (e.g. `.claude/skills`).
_Avoid_: destination, path

**Context Token Budget**:
The estimated number of context tokens a set of Skills will consume when an agent loads them. Surfaces as a per-agent estimate in the Projects view, with warning and danger thresholds.
_Avoid_: token cost, context cost (use "context token budget" in full)
