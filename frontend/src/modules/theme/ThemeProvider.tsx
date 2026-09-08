import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { usePreferencesStore } from "@/modules/settings/store";
import { applyTheme } from "./applyTheme";
import { getBuiltinTheme, getDefaultTheme } from "./themes";
import { type Theme, type ThemeMode } from "./types";

export type { Theme };
export type ThemeModePref = "system" | "light" | "dark";

type ThemeProviderProps = {
  children: React.ReactNode;
};

type ThemeProviderState = {
  mode: ThemeModePref;
  resolvedMode: "dark" | "light";
  themeId: string;
  activeTheme: Theme;
  setMode: (mode: ThemeModePref) => void;
  setThemeId: (id: string) => void;
};

const ThemeProviderContext = createContext<ThemeProviderState | null>(null);

function resolveTheme(id: string): Theme {
  return getBuiltinTheme(id) ?? getDefaultTheme();
}

export function ThemeProvider({ children }: ThemeProviderProps) {
  const prefsTheme = usePreferencesStore((s) => s.theme);
  const prefsThemeId = usePreferencesStore((s) => s.themeId);
  const setThemePref = usePreferencesStore((s) => s.setTheme);
  const setThemeIdPref = usePreferencesStore((s) => s.setThemeId);

  const [systemDark, setSystemDark] = useState<boolean>(() =>
    typeof window === "undefined"
      ? true
      : window.matchMedia("(prefers-color-scheme: dark)").matches,
  );

  useEffect(() => {
    const mq = window.matchMedia("(prefers-color-scheme: dark)");
    const onChange = (e: MediaQueryListEvent) => setSystemDark(e.matches);
    mq.addEventListener("change", onChange);
    return () => mq.removeEventListener("change", onChange);
  }, []);

  const mode: ThemeModePref = prefsTheme;
  const resolvedMode: ThemeMode =
    mode === "system" ? (systemDark ? "dark" : "light") : mode;

  // Toggle the .dark / .light class for Tailwind + the base CSS vars.
  useEffect(() => {
    const root = document.documentElement;
    root.classList.remove("light", "dark");
    root.classList.add(resolvedMode);
  }, [resolvedMode]);

  const activeTheme = useMemo(() => resolveTheme(prefsThemeId), [prefsThemeId]);
  useEffect(() => {
    applyTheme(activeTheme, resolvedMode);
  }, [prefsThemeId, activeTheme, resolvedMode]);

  const setMode = useCallback(
    (next: ThemeModePref) => setThemePref(next),
    [setThemePref],
  );
  const setThemeId = useCallback(
    (id: string) => setThemeIdPref(id),
    [setThemeIdPref],
  );

  const value = useMemo<ThemeProviderState>(
    () => ({ mode, resolvedMode, themeId: prefsThemeId, activeTheme, setMode, setThemeId }),
    [mode, resolvedMode, prefsThemeId, activeTheme, setMode, setThemeId],
  );

  return (
    <ThemeProviderContext.Provider value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

export function useTheme(): ThemeProviderState {
  const ctx = useContext(ThemeProviderContext);
  if (!ctx) throw new Error("useTheme must be used within a <ThemeProvider>");
  return ctx;
}
