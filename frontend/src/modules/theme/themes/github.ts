import type { Theme } from "../types";

export const github: Theme = {
  id: "github",
  name: "GitHub Light",
  description: "GitHub's familiar light interface colors.",
  variants: {
    light: {
      colors: {
        background: "#ffffff",
        foreground: "#1f2328",
        card: "#f6f8fa",
        cardForeground: "#1f2328",
        popover: "#ffffff",
        popoverForeground: "#1f2328",
        primary: "#0969da",
        primaryForeground: "#ffffff",
        secondary: "#f6f8fa",
        secondaryForeground: "#1f2328",
        muted: "#f6f8fa",
        mutedForeground: "#656d76",
        accent: "#ddf4ff",
        accentForeground: "#0969da",
        destructive: "#d1242f",
        border: "#d0d7de",
        input: "#d0d7de",
        ring: "#0969da",
        sidebar: "#f6f8fa",
        sidebarForeground: "#1f2328",
        sidebarPrimary: "#0969da",
        sidebarPrimaryForeground: "#ffffff",
        sidebarAccent: "#eaeef2",
        sidebarAccentForeground: "#1f2328",
        sidebarBorder: "#d0d7de",
        sidebarRing: "#0969da",
      },
    },
  },
};
