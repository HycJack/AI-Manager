import { useCallback, useEffect, useRef, useState } from "react";
import type { PanelImperativeHandle } from "react-resizable-panels";

export const SIDEBAR_DEFAULT_WIDTH = 224;
export const SIDEBAR_MIN_WIDTH = 160;
export const SIDEBAR_MAX_WIDTH = 400;
// Collapsed rail keeps a slim icon strip (wide enough for icons + the expand
// toggle) rather than vanishing entirely.
export const SIDEBAR_COLLAPSED_WIDTH = 56;

const SIDEBAR_WIDTH_STORAGE_KEY = "ai-manager.sidebar.width";
const SIDEBAR_COLLAPSED_STORAGE_KEY = "ai-manager.sidebar.collapsed";

function clampSidebarWidth(width: number): number {
  return Math.min(
    SIDEBAR_MAX_WIDTH,
    Math.max(SIDEBAR_MIN_WIDTH, Math.round(width)),
  );
}

function readSidebarWidth(): number {
  try {
    const stored = window.localStorage.getItem(SIDEBAR_WIDTH_STORAGE_KEY);
    const parsed = stored ? Number.parseInt(stored, 10) : NaN;
    return Number.isFinite(parsed)
      ? clampSidebarWidth(parsed)
      : SIDEBAR_DEFAULT_WIDTH;
  } catch {
    return SIDEBAR_DEFAULT_WIDTH;
  }
}

function readSidebarCollapsed(): boolean {
  try {
    return window.localStorage.getItem(SIDEBAR_COLLAPSED_STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

export function useSidebarPanel() {
  const sidebarRef = useRef<PanelImperativeHandle | null>(null);
  const sidebarWidthRef = useRef(readSidebarWidth());
  const writeTimerRef = useRef(0);

  // Live collapsed state, driven by the panel's onResize events. Used to
  // show the floating "expand" affordance when the sidebar is fully hidden.
  const [collapsed, setCollapsed] = useState(readSidebarCollapsed);

  const persistCollapsed = useCallback((collapsed: boolean) => {
    try {
      window.localStorage.setItem(
        SIDEBAR_COLLAPSED_STORAGE_KEY,
        collapsed ? "1" : "0",
      );
    } catch {
      /* storage may fail in private mode */
    }
  }, []);

  const persistWidth = useCallback((width: number) => {
    if (width <= SIDEBAR_COLLAPSED_WIDTH) return; // collapsed — keep last width
    sidebarWidthRef.current = clampSidebarWidth(width);
    if (writeTimerRef.current) window.clearTimeout(writeTimerRef.current);
    writeTimerRef.current = window.setTimeout(() => {
      writeTimerRef.current = 0;
      try {
        window.localStorage.setItem(
          SIDEBAR_WIDTH_STORAGE_KEY,
          String(sidebarWidthRef.current),
        );
      } catch {
        /* ignore */
      }
    }, 200);
  }, []);

  useEffect(() => {
    return () => {
      if (writeTimerRef.current) window.clearTimeout(writeTimerRef.current);
    };
  }, []);

  const toggleSidebar = useCallback(() => {
    const p = sidebarRef.current;
    if (!p) return;
    if (p.getSize().inPixels <= SIDEBAR_COLLAPSED_WIDTH) {
      p.resize(`${sidebarWidthRef.current}px`);
    } else {
      p.collapse();
    }
  }, []);

  const expandSidebar = useCallback(() => {
    sidebarRef.current?.resize(`${sidebarWidthRef.current}px`);
  }, []);

  return {
    sidebarRef,
    sidebarWidthRef,
    collapsed,
    setCollapsed,
    persistCollapsed,
    persistWidth,
    toggleSidebar,
    expandSidebar,
  };
}
