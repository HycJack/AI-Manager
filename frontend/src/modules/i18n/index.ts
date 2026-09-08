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
    // Toasts
    "toast.autostartOn": "Autostart enabled",
    "toast.autostartOff": "Autostart disabled",
    "toast.autostartFailed": "Failed to set autostart: {error}",
    "toast.languageChanged": "Language switched to {lang}",
    "toast.libraryChanged": "Library path updated",
    "toast.libraryScanComplete": "Library scan complete",
    "toast.libraryScanFailed": "Library scan failed: {error}",
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
    // Toasts
    "toast.autostartOn": "已设置开机自启",
    "toast.autostartOff": "已关闭开机自启",
    "toast.autostartFailed": "设置失败: {error}",
    "toast.languageChanged": "语言已切换为{lang}",
    "toast.libraryChanged": "库路径已更新",
    "toast.libraryScanComplete": "库扫描完成",
    "toast.libraryScanFailed": "库扫描失败: {error}",
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
