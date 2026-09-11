import { useState, useCallback, useEffect } from "react";

/** Supported language codes. */
export type Language = "en" | "zh";

/** Default language. */
export const DEFAULT_LANGUAGE: Language = "en";

const STORAGE_KEY = "ai-manager-language";

/**
 * Translation dictionary. Keys are stable identifiers; values are the
 * human-readable strings for each language. Add new keys as the app grows.
 */
const translations: Record<Language, Record<string, string>> = {
  en: {
    // App
    "app.title": "AI-Manager",
    "app.tagline": "Agent Skills Manager",
    // Navigation
    "nav.home": "Home",
    "nav.skills": "Skills",
    "nav.projects": "Projects",
    "nav.settings": "Settings",
    // Skills page
    "skills.title": "Skills",
    "skills.search": "Search skills...",
    "skills.empty.title": "No skills yet",
    "skills.empty.desc": "Add your first skill to get started. Import from local folders, npx packages, or Claude plugins.",
    "skills.empty.add": "Add your first skill",
    "skills.empty.discover": "Discover skills",
    "skills.discover": "Discover Skills",
    "skills.addLocal": "Add from Local Folder",
    "skills.noFolder": "No folder selected",
    "skills.browse": "Browse…",
    "skills.add": "Add",
    "skills.adding": "Adding…",
    "skills.added": "Skill added",
    "skills.searchSkillsSh": "Search skills.sh...",
    "skills.searchGithub": "Search GitHub...",
    "skills.searchHint": "Search for skills to get started",
    "skills.results": "results for",
    "skills.noResults": "No results found",
    "skills.installs": "installs",
    "skills.githubRepos": "repositories on GitHub",
    "skills.noDesc": "No description",
    "skills.browseOnly": "Browse Only",
    "skills.detail.select": "Select a skill to view details",
    "skills.context.install": "Install...",
    "skills.context.delete": "Delete...",
    "skills.context.copyPath": "Copy Path",
    "skills.context.copyName": "Copy Name",
    // Settings page
    "settings.title": "Settings",
    "settings.subtitle": "General, themes, and application preferences",
    "settings.tab.general": "General",
    "settings.tab.themes": "Themes",
    "settings.tab.library": "Library",
    "settings.tab.about": "About",
    // General section
    "settings.appearance": "Appearance Mode",
    "settings.appearance.hint": "Theme colors are selected in the Themes tab.",
    "settings.mode.system": "System",
    "settings.mode.light": "Light",
    "settings.mode.dark": "Dark",
    "settings.zoom": "Interface Zoom",
    "settings.language": "Language",
    "settings.language.hint": "Switch between English and Chinese.",
    "settings.fileBrowser": "File Browser",
    "settings.showHidden": "Show Hidden Files",
    "settings.showHidden.desc": "Display dot-prefixed files (.env, .gitignore, etc.) in file lists and search.",
    "settings.startup": "Startup",
    "settings.launchAtLogin": "Launch at Login",
    "settings.launchAtLogin.desc": "Open the application automatically when you sign in.",
    // Themes section
    "settings.themes.title": "Themes",
    "settings.themes.hint": "Click to select a theme and apply immediately.",
    "settings.themes.currentMode": "Current mode: {mode}",
    // Library section
    "settings.library.title": "Skill Library",
    "settings.library.path": "Library Path",
    "settings.library.path.hint": "Directory where skills are stored.",
    "settings.library.browse": "Browse…",
    "settings.library.open": "Open in Finder",
    "settings.library.rescan": "Rescan Library",
    "settings.library.rescan.hint": "Re-scan the library directory for new or changed skills.",
    "settings.library.notConfigured": "No library path configured",
    // About section
    "settings.about.title": "About",
    "settings.about.version": "Version",
    "settings.about.commit": "Commit",
    "settings.about.buildTime": "Build Time",
    "settings.about.updates": "Updates",
    "settings.about.checkForUpdates": "Check for Updates",
    "settings.about.checking": "Checking…",
    "settings.about.upToDate": "You're on the latest version",
    "settings.about.updateAvailable": "Update available",
    "settings.about.installUpdate": "Install Update",
    "settings.about.restartToUpdate": "Restart to Complete Update",
    // Install dialog
    "install.title": "Install Skills",
    "install.subtitle": "Select where to install the selected skills.",
    "install.skillsToInstall": "Skills to install",
    "install.location": "Install location",
    "install.project": "Project",
    "install.global": "Global (User)",
    "install.projectPath": "Project path",
    "install.projectPlaceholder": "Enter or browse for a project directory…",
    "install.browse": "Browse",
    "install.targets": "Targets",
    "install.summary": "Summary",
    "install.cancel": "Cancel",
    "install.installing": "Installing…",
    "install.confirm": "Install {count}",
    "install.notValidProject": "Please enter a valid project path or choose a project.",
    "install.noTargets": "Select at least one target.",
    // Toasts
    "toast.autostartOn": "Autostart enabled",
    "toast.autostartOff": "Autostart disabled",
    "toast.autostartFailed": "Failed to set autostart: {error}",
    "toast.languageChanged": "Language switched to {lang}",
    "toast.libraryChanged": "Library path updated",
    "toast.libraryScanComplete": "Library scan complete",
    "toast.libraryScanFailed": "Library scan failed: {error}",
    "toast.skillDeleted": "Deleted {count} skill(s)",
    "toast.skillDeleteFailed": "Failed to delete: {error}",
    "toast.installSuccess": "Installed {count} skill(s) to {targets} target(s)",
    "toast.installFailed": "Install failed: {error}",
    "toast.installBrowseFailed": "Failed to open folder picker",
    // Delete dialog
    "delete.title": "Delete Skills",
    "delete.subtitle": "You are about to delete the following skills. This action cannot be undone.",
    "delete.selectedCount": "{count} selected",
    "delete.builtinWarning": "Builtin skills cannot be deleted",
    "delete.confirm": "Delete",
    "delete.cancel": "Cancel",
    "delete.select": "Delete {count}",
  },
  zh: {
    // App
    "app.title": "AI 管理器",
    "app.tagline": "Agent 技能管理器",
    // Navigation
    "nav.home": "首页",
    "nav.skills": "技能",
    "nav.projects": "项目",
    "nav.settings": "设置",
    // Skills page
    "skills.title": "技能",
    "skills.search": "搜索技能...",
    "skills.empty.title": "暂无技能",
    "skills.empty.desc": "添加第一个技能开始使用。可从本地文件夹、npx 包或 Claude 插件导入。",
    "skills.empty.add": "添加第一个技能",
    "skills.empty.discover": "发现技能",
    "skills.discover": "发现技能",
    "skills.addLocal": "从本地文件夹添加",
    "skills.noFolder": "未选择文件夹",
    "skills.browse": "浏览…",
    "skills.add": "添加",
    "skills.adding": "添加中…",
    "skills.added": "技能已添加",
    "skills.searchSkillsSh": "搜索 skills.sh...",
    "skills.searchGithub": "搜索 GitHub...",
    "skills.searchHint": "搜索技能开始使用",
    "skills.results": "个结果，搜索",
    "skills.noResults": "未找到结果",
    "skills.installs": "次安装",
    "skills.githubRepos": "GitHub 仓库",
    "skills.noDesc": "无描述",
    "skills.browseOnly": "仅浏览",
    "skills.detail.select": "选择一个技能查看详情",
    "skills.context.install": "安装...",
    "skills.context.delete": "删除...",
    "skills.context.copyPath": "复制路径",
    "skills.context.copyName": "复制名称",
    // Settings page
    "settings.title": "设置",
    "settings.subtitle": "外观、主题与应用偏好",
    "settings.tab.general": "通用",
    "settings.tab.themes": "主题",
    "settings.tab.library": "技能库",
    "settings.tab.about": "关于",
    // General section
    "settings.appearance": "外观模式",
    "settings.appearance.hint": "主题颜色在「主题」页切换。",
    "settings.mode.system": "系统",
    "settings.mode.light": "浅色",
    "settings.mode.dark": "深色",
    "settings.zoom": "界面缩放",
    "settings.language": "语言",
    "settings.language.hint": "在英语和中文之间切换。",
    "settings.fileBrowser": "文件浏览器",
    "settings.showHidden": "显示隐藏文件",
    "settings.showHidden.desc": "在文件列表与搜索中显示点开头的文件（.env、.gitignore 等）。",
    "settings.startup": "启动",
    "settings.launchAtLogin": "开机自启",
    "settings.launchAtLogin.desc": "登录系统时自动打开应用。",
    // Themes section
    "settings.themes.title": "主题",
    "settings.themes.hint": "点击选择主题并立即应用。",
    "settings.themes.currentMode": "当前模式：{mode}",
    // Library section
    "settings.library.title": "技能库",
    "settings.library.path": "库路径",
    "settings.library.path.hint": "技能存储目录。",
    "settings.library.browse": "浏览…",
    "settings.library.open": "在 Finder 中打开",
    "settings.library.rescan": "重新扫描",
    "settings.library.rescan.hint": "重新扫描库目录中的新或变更的技能。",
    "settings.library.notConfigured": "未配置库路径",
    // About section
    "settings.about.title": "关于",
    "settings.about.version": "版本",
    "settings.about.commit": "提交",
    "settings.about.buildTime": "构建时间",
    "settings.about.updates": "更新",
    "settings.about.checkForUpdates": "检查更新",
    "settings.about.checking": "检查中…",
    "settings.about.upToDate": "已是最新版本",
    "settings.about.updateAvailable": "有可用更新",
    "settings.about.installUpdate": "安装更新",
    "settings.about.restartToUpdate": "重启以完成更新",
    // Install dialog
    "install.title": "安装技能",
    "install.subtitle": "选择已选技能的安装位置。",
    "install.skillsToInstall": "待安装技能",
    "install.location": "安装位置",
    "install.project": "项目",
    "install.global": "全局（用户级）",
    "install.projectPath": "项目路径",
    "install.projectPlaceholder": "输入或浏览选择项目目录…",
    "install.browse": "浏览",
    "install.targets": "目标",
    "install.summary": "摘要",
    "install.cancel": "取消",
    "install.installing": "安装中…",
    "install.confirm": "安装 {count} 个",
    "install.notValidProject": "请输入有效的项目路径或选择一个项目。",
    "install.noTargets": "请至少选择一个目标。",
    // Toasts
    "toast.autostartOn": "已设置开机自启",
    "toast.autostartOff": "已关闭开机自启",
    "toast.autostartFailed": "设置失败: {error}",
    "toast.languageChanged": "语言已切换为{lang}",
    "toast.libraryChanged": "库路径已更新",
    "toast.libraryScanComplete": "库扫描完成",
    "toast.libraryScanFailed": "库扫描失败: {error}",
    "toast.skillDeleted": "已删除 {count} 个技能",
    "toast.skillDeleteFailed": "删除失败: {error}",
    "toast.installSuccess": "已将 {count} 个技能安装到 {targets} 个目标",
    "toast.installFailed": "安装失败: {error}",
    "toast.installBrowseFailed": "打开文件夹选择器失败",
    // Delete dialog
    "delete.title": "删除技能",
    "delete.subtitle": "您即将删除以下技能。此操作不可撤销。",
    "delete.selectedCount": "已选择 {count} 个",
    "delete.builtinWarning": "内置技能无法删除",
    "delete.confirm": "删除",
    "delete.cancel": "取消",
    "delete.select": "删除 {count} 个",
  },
};

/** Returns the stored language, defaulting to DEFAULT_LANGUAGE. */
export function getStoredLang(): Language {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "en" || stored === "zh") return stored;
  } catch {
    // ignore
  }
  return DEFAULT_LANGUAGE;
}

/** Persist the language choice to localStorage. */
function storeLang(lang: Language): void {
  try {
    localStorage.setItem(STORAGE_KEY, lang);
  } catch {
    // ignore
  }
}

/** Translate a key with optional string interpolation. */
export function t(key: string, lang: Language, vars?: Record<string, string | number>): string {
  const value = translations[lang][key] ?? translations.en[key] ?? key;
  if (!vars) return value;
  return value.replace(/\{(\w+)\}/g, (_, name) => String(vars[name] ?? `{${name}}`));
}

/** i18n hook — provides the current language, a setter, and a translate function. */
export function useI18n() {
  const [lang, setLang] = useState<Language>(getStoredLang);

  // Sync with localStorage on mount (in case another tab changed it).
  useEffect(() => {
    const handler = (e: StorageEvent) => {
      if (e.key === STORAGE_KEY) {
        setLang((e.newValue as Language) || DEFAULT_LANGUAGE);
      }
    };
    window.addEventListener("storage", handler);
    return () => window.removeEventListener("storage", handler);
  }, []);

  const changeLang = useCallback((newLang: Language) => {
    setLang(newLang);
    storeLang(newLang);
  }, []);

  const translate = useCallback(
    (key: string, vars?: Record<string, string | number>) => t(key, lang, vars),
    [lang],
  );

  return { lang, setLang: changeLang, t: translate };
}
