import type { Theme } from "../types";

export const nord: Theme = {
  id: "nord",
  name: "Nord",
  description: "Arctic, north-bluish clean and minimal theme.",
  variants: {
    dark: {
      colors: {
        background: "#2e3440",
        foreground: "#d8dee9",
        card: "#3b4252",
        cardForeground: "#eceff4",
        popover: "#3b4252",
        popoverForeground: "#eceff4",
        primary: "#88c0d0",
        primaryForeground: "#2e3440",
        secondary: "#434c5e",
        secondaryForeground: "#eceff4",
        muted: "#4c566a",
        mutedForeground: "#d8dee9",
        accent: "#b48ead",
        accentForeground: "#2e3440",
        destructive: "#bf616a",
        border: "#434c5e",
        input: "#434c5e",
        ring: "#88c0d0",
        sidebar: "#3b4252",
        sidebarForeground: "#d8dee9",
        sidebarPrimary: "#88c0d0",
        sidebarPrimaryForeground: "#2e3440",
        sidebarAccent: "#434c5e",
        sidebarAccentForeground: "#eceff4",
        sidebarBorder: "#434c5e",
        sidebarRing: "#88c0d0",
      },
    },
  },
};
