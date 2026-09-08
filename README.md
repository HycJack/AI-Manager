# AI-Manager

跨项目 Agent Skills 管理器 — 一个维护的技能库，每个项目按需获取，通过 managed symlink 保持同步。

A desktop application for managing Agent Skills across projects: one maintained library, each project gets only what it needs, updates propagate through managed links.

参考实现：[Kitter](https://github.com/what1f/kitter)（Rust/gpui）→ 复刻为 Go + React 版本。

## Features

### 技能库管理

- **4 种来源添加**：本地文件夹、Npx/skills.sh/GitHub、Claude plugin、现有安装采纳
- **虚拟列表**：`@tanstack/react-virtual` 高效渲染大量技能
- **搜索过滤**：按名称实时过滤
- **分组折叠**：可折叠的技能分组
- **详情面板**：描述（README）、安装目标、文件列表三个标签页
- **多选操作**：Shift+click 范围选、Cmd/Ctrl+click 切换选

### 项目与安装

- **10 个 Agent 目标**：Claude、Codex、Cursor、Cline、Continue、Aider、Antigravity、Trae、Windsurf、Shared（Universal）
- **Symlink 安装**：技能库 → 项目 agent 目录的符号链接（非复制）
- **全局/项目安装**：用户级或项目级安装
- **跨 agent 去重**：同一源目录在多个 agent 中只算一个技能（Kitter 模式）
- **Effective Skills**：扫描所有 agent 目录，解析 symlink 到原始源，去重后显示

### 标签与分组

- **两层标签树**：Parent → Child，拖拽分配技能
- **分组管理**：创建/编辑/删除分组，移动技能
- **@dnd-kit 拖拽**：技能 ↔ 标签、技能 ↔ 分组的拖放操作

### 设置与主题

- **15 个内置主题**：Claude、Sky、Violet、Forest、Ocean、Nord、Dracula、Monokai、Gruvbox、Solarized、Tokyo Night、Catppuccin、Rosé Pine、Midnight 等
- **Light / Dark / System** 外观模式
- **中英双语**：EN/ZH 语言切换，全局生效
- **库路径浏览**：原生文件夹选择器
- **系统托盘** + 全局快捷键（Cmd+Alt+M 显示窗口）

### 工程化

- **原子写 JSON 持久化**：tmp + rename，`sync.Mutex` 并发保护
- **跨平台路径解析**：macOS（`~/Library/Application Support/`）、Windows（`%APPDATA%`）、Linux（`$XDG_DATA_HOME`）
- **GitHub Actions CI**：Windows / macOS / Linux 三平台构建 + 自动发布

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.27+ · Wails3 v3.0.0-beta |
| Frontend | React 18 · TypeScript · Vite · Tailwind CSS v4 |
| UI | shadcn/ui · Radix Primitives · lucide-react |
| State | zustand (debounced persist to Go backend) |
| Virtual List | @tanstack/react-virtual |
| Drag & Drop | @dnd-kit/core · @dnd-kit/sortable |
| Animations | framer-motion |
| Build | Wails3 Taskfile · NSIS (Windows) · DMG (macOS) · AppImage (Linux) |

## Project Structure

```
main.go                      Wails3 entry (config → services → window → tray)
internal/
  agents/                    10 AgentKind + 10 InstallTarget + 目录映射
  app/                       9 个 Wails3 Services (34 methods)
  config/                    通用 JSON 配置 + Registry + 原子写
  effective/                 有效技能计算 (跨 agent 扫描 + symlink 去重 + token 估算)
  logger/                    文件日志
  platform/                  跨平台路径解析 (macOS/Windows/Linux + env override)
  project/                   项目 + symlink 安装/卸载/扫描
  skill/                     数据模型 + 4 种来源扫描器 + Library
  tags/                      两层标签树 (parent → children)
  version/                   版本信息
frontend/
  src/
    pages/
      SkillsPage.tsx         技能列表 (虚拟化) + 详情面板 (可调整大小)
      ProjectsPage.tsx       项目选择 + 10 agent 有效技能视图 + token 估算
      SettingsPage.tsx       语言切换 + 库路径 + 主题网格
    components/
      SkillAddDialog.tsx     添加技能 (4 种来源)
      SkillInstallDialog.tsx 安装 (10 targets + 全局/项目)
      SkillDeleteDialog.tsx  删除确认 (多选 + 内置保护)
      TagsPanel.tsx          标签/分组管理 (拖拽)
      ui/                    17 shadcn 组件
    modules/
      settings/store.ts      zustand 设置存储
      tags/store.ts          标签/分组状态管理
      theme/                 15 个主题引擎
      i18n/                  中英双语翻译
  bindings/                  Wails3 自动生成的 TypeScript 类型
build/                       跨平台构建资源 (Taskfiles, 图标, 打包)
.github/workflows/           GitHub Actions CI/CD
```

## Quick Start

### Prerequisites

- **Go** 1.27+
- **Node.js** 20+
- **wails3 CLI**: `go install github.com/wailsapp/wails/v3/cmd/wails3@latest`
- **C compiler** (CGO): GCC (Windows: mingw-w64, macOS: Xcode CLT, Linux: gcc)

### Build & Run

```bash
# 1. Install frontend dependencies
cd frontend && npm install && cd ..

# 2. Generate Wails3 bindings
wails3 generate bindings

# 3. Build production binary
wails3 build
# → bin/ai-manager (macOS/Linux) / bin/ai-manager.exe (Windows)

# 4. Development mode (hot reload)
wails3 dev
```

### Create data directory (macOS)

macOS sandbox may block automatic directory creation:

```bash
mkdir -p ~/Library/Application\ Support/AIManager
```

## How to Extend

### Add a new Service

```go
// internal/app/my_service.go
package app

type MyService struct{ state *State }

func NewMyService(s *State) *MyService { return &MyService{state: s} }

func (s *MyService) DoSomething(input string) (string, error) {
    return "result: " + input, nil
}
```

Register in `main.go`:
```go
Services: []application.Service{
    application.NewService(app.NewMyService(state)),
},
```

Then: `wails3 generate bindings`

### Add a new setting

1. Add field to `Preferences` struct in `internal/app/settings_service.go`
2. Add to `Preferences` type in `frontend/src/modules/settings/store.ts`
3. Add UI control in `SettingsPage.tsx`
4. `wails3 generate bindings`

### Add a new theme

Create `frontend/src/modules/theme/themes/my-theme.ts`:
```ts
export const myTheme: Theme = {
  id: "my-theme",
  name: "My Theme",
  variants: {
    dark: { colors: { /* ... */ } },
    light: { colors: { /* ... */ } },
  },
};
```

Register in `themes/index.ts`.

## CI/CD

GitHub Actions workflows in `.github/workflows/`:

| Workflow | Platform | Artifact |
|----------|----------|----------|
| `build-windows.yml` | Windows | `ai-manager.exe` + NSIS installer |
| `build-macos.yml` | macOS (universal) | `ai-manager.app` + DMG |
| `build-linux.yml` | Linux | `ai-manager` AppImage |

- Push to `main` → build + upload artifact
- Push a `v*` tag → build + create GitHub Release with binaries
- Manual dispatch supported

## Domain Model

See `CONTEXT.md` for 11 domain terms: Skill, Library, Project, Installation, Agent, Effective Skill, Source, Tag, Group, Install Target, Context Token Budget.

## License

MIT
