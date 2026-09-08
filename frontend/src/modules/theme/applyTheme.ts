import type { Theme, ThemeColors, ThemeMode } from "./types";

const COLOR_VAR: Record<keyof ThemeColors, string> = {
  background: "--background",
  foreground: "--foreground",
  card: "--card",
  cardForeground: "--card-foreground",
  popover: "--popover",
  popoverForeground: "--popover-foreground",
  primary: "--primary",
  primaryForeground: "--primary-foreground",
  secondary: "--secondary",
  secondaryForeground: "--secondary-foreground",
  muted: "--muted",
  mutedForeground: "--muted-foreground",
  accent: "--accent",
  accentForeground: "--accent-foreground",
  destructive: "--destructive",
  border: "--border",
  input: "--input",
  ring: "--ring",
  sidebar: "--sidebar",
  sidebarForeground: "--sidebar-foreground",
  sidebarPrimary: "--sidebar-primary",
  sidebarPrimaryForeground: "--sidebar-primary-foreground",
  sidebarAccent: "--sidebar-accent",
  sidebarAccentForeground: "--sidebar-accent-foreground",
  sidebarBorder: "--sidebar-border",
  sidebarRing: "--sidebar-ring",
  radius: "--radius",
};

const ALL_VARS = Object.values(COLOR_VAR);
let lastApplied: string | null = null;

/**
 * Apply a theme to the document by overriding CSS variables (--background,
 * --primary, …) on <html>. Always uses the dark variant (or the first
 * available variant if dark is absent).
 */
export function applyTheme(theme: Theme, mode: ThemeMode): void {
  const root = document.documentElement;
  const variant =
    theme.variants[mode] ?? theme.variants.dark ?? theme.variants.light;

  // Always start from a clean slate so a switch fully reverts prior overrides.
  for (const v of ALL_VARS) root.style.removeProperty(v);

  const colors = variant?.colors;
  if (!colors) {
    lastApplied = theme.id;
    return;
  }
  for (const k of Object.keys(colors) as (keyof ThemeColors)[]) {
    const v = colors[k];
    if (v) root.style.setProperty(COLOR_VAR[k], v);
  }
  lastApplied = theme.id;
}

/** Revert any applied theme overrides back to the base stylesheet. */
export function clearTheme(): void {
  if (lastApplied === null) return;
  const root = document.documentElement;
  for (const v of ALL_VARS) root.style.removeProperty(v);
  lastApplied = null;
}
