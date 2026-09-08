import type { Theme } from "../types";

export const dracula: Theme = {
  id: "dracula",
  name: "Dracula",
  description: "The classic high-contrast purple dark theme.",
  variants: {
    dark: {
      colors: {
        background: "#282a36",
        foreground: "#f8f8f2",
        card: "#21222c",
        cardForeground: "#f8f8f2",
        popover: "#21222c",
        popoverForeground: "#f8f8f2",
        primary: "#bd93f9",
        primaryForeground: "#282a36",
        secondary: "#343746",
        secondaryForeground: "#f8f8f2",
        muted: "#343746",
        mutedForeground: "#a9acc0",
        accent: "#44475a",
        accentForeground: "#f8f8f2",
        destructive: "#ff5555",
        border: "rgba(248,248,242,0.08)",
        input: "rgba(248,248,242,0.12)",
        ring: "#bd93f9",
        sidebar: "#21222c",
        sidebarForeground: "#f8f8f2",
        sidebarPrimary: "#bd93f9",
        sidebarPrimaryForeground: "#282a36",
        sidebarAccent: "#44475a",
        sidebarAccentForeground: "#f8f8f2",
        sidebarBorder: "rgba(248,248,242,0.08)",
        sidebarRing: "#bd93f9",
      },
    },
  },
};
