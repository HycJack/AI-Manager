import { create } from "zustand";
import { LoadPrefs, SavePrefs } from "@bindings/ai-manager/internal/app/settingsservice";
import { Preferences as PreferencesModel } from "@bindings/ai-manager/internal/app/models";
import { DEFAULT_THEME_ID } from "@/modules/theme/types";

export type ThemePref = "system" | "light" | "dark";

export type Preferences = {
  /** Light/dark/system. "system" follows the OS preference. */
  theme: ThemePref;
  /** Id of the active theme (see modules/theme/themes). */
  themeId: string;
  /** UI zoom level, e.g. 1.0 === 100%. */
  zoomLevel: number;
  /** Show dot-prefixed files in the (future) file explorer. */
  showHidden: boolean;
  /** Open automatically at login. */
  launchAtLogin: boolean;
  /** UI language: "en" or "zh". */
  language: "en" | "zh";
};

export const DEFAULT_PREFERENCES: Preferences = {
  theme: "system",
  themeId: DEFAULT_THEME_ID,
  zoomLevel: 1,
  showHidden: false,
  launchAtLogin: false,
  language: "en",
};

/** Load preferences persisted by the backend SettingsService. */
export async function loadPreferences(): Promise<Preferences> {
  const p = await LoadPrefs();
  return {
    theme: (p.theme || DEFAULT_PREFERENCES.theme) as ThemePref,
    themeId: p.themeId || DEFAULT_PREFERENCES.themeId,
    zoomLevel: p.zoomLevel || DEFAULT_PREFERENCES.zoomLevel,
    showHidden: p.showHidden ?? DEFAULT_PREFERENCES.showHidden,
    launchAtLogin: p.launchAtLogin ?? DEFAULT_PREFERENCES.launchAtLogin,
    language: (p.language || DEFAULT_PREFERENCES.language) as "en" | "zh",
  };
}

type PreferencesStore = Preferences & {
  hydrated: boolean;
  /** Load from disk once (idempotent) and subscribe to persistence. */
  init: () => Promise<void>;
  setTheme: (v: ThemePref) => void;
  setThemeId: (v: string) => void;
  setZoomLevel: (v: number) => void;
  setShowHidden: (v: boolean) => void;
  setLaunchAtLogin: (v: boolean) => void;
  setLanguage: (v: "en" | "zh") => void;
};

let initPromise: Promise<void> | null = null;

// --- single-source-of-truth persistence -------------------------------
// The backend JSON bag is authoritative. A localStorage mirror is written
// alongside it only as a fast boot cache so the first paint matches the last
// session before the async backend load resolves (prevents theme flash).

const FAST_CACHE_KEY = "skeleton-prefs";

function readFastCache(): Partial<Preferences> | null {
  try {
    const raw = localStorage.getItem(FAST_CACHE_KEY);
    return raw ? (JSON.parse(raw) as Partial<Preferences>) : null;
  } catch {
    return null;
  }
}

function writeFastCache(prefs: Preferences): void {
  try {
    localStorage.setItem(FAST_CACHE_KEY, JSON.stringify(prefs));
  } catch {
    /* ignore quota / privacy errors */
  }
}

let flushTimer: ReturnType<typeof setTimeout> | null = null;
let pendingFlush: Preferences | null = null;
const FLUSH_DELAY_MS = 400;

function toModel(prefs: Preferences): PreferencesModel {
  return new PreferencesModel({
    theme: prefs.theme,
    themeId: prefs.themeId,
    zoomLevel: prefs.zoomLevel,
    showHidden: prefs.showHidden,
    launchAtLogin: prefs.launchAtLogin,
    language: prefs.language,
  });
}

/** Coalesce rapid setter calls (e.g. dragging the zoom slider) into a single
 *  debounced disk write. The latest state is always what gets flushed. */
function schedulePersist(prefs: Preferences): void {
  pendingFlush = prefs;
  writeFastCache(prefs); // mirror immediately (cheap, in-memory)
  if (flushTimer) clearTimeout(flushTimer);
  flushTimer = setTimeout(() => {
    flushTimer = null;
    void SavePrefs(toModel(pendingFlush!)).catch((e) => console.error("save prefs", e));
    pendingFlush = null;
  }, FLUSH_DELAY_MS);
}

/** Flush any pending write immediately (e.g. on window hide / quit). */
export function flushPreferencesNow(): void {
  if (flushTimer) {
    clearTimeout(flushTimer);
    flushTimer = null;
  }
  if (pendingFlush) {
    void SavePrefs(toModel(pendingFlush)).catch((e) => console.error("save prefs", e));
    pendingFlush = null;
  }
}

export const usePreferencesStore = create<PreferencesStore>((set, get) => ({
  ...DEFAULT_PREFERENCES,
  hydrated: false,
  init: () => {
    if (initPromise) return initPromise;
    initPromise = (async () => {
      // 1) Fast path: paint from localStorage cache before backend resolves.
      const cached = readFastCache();
      if (cached) set({ ...cached, hydrated: true });
      // 2) Authoritative: hydrate from backend, overwriting the cache.
      try {
        const prefs = await loadPreferences();
        set({ ...prefs, hydrated: true });
        writeFastCache(prefs);
      } catch (e) {
        initPromise = null;
        throw e;
      }
    })();
    return initPromise;
  },
  setTheme: (theme) => {
    set({ theme });
    schedulePersist(get());
  },
  setThemeId: (themeId) => {
    set({ themeId });
    schedulePersist(get());
  },
  setZoomLevel: (zoomLevel) => {
    set({ zoomLevel });
    schedulePersist(get());
  },
  setShowHidden: (showHidden) => {
    set({ showHidden });
    schedulePersist(get());
  },
  setLaunchAtLogin: (launchAtLogin) => {
    set({ launchAtLogin });
    schedulePersist(get());
  },
  setLanguage: (language) => {
    set({ language });
    schedulePersist(get());
  },
}));
