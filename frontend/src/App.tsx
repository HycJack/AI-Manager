import { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "framer-motion";
import { Layers, Settings, FolderTree, PanelLeftClose, PanelLeftOpen, Sparkles, Bot } from "lucide-react";
import type { Layout, LayoutChangedMeta } from "react-resizable-panels";
import { cn } from "@/lib/utils";
import { useToasts } from "@/lib/toast";
import {
  ResizableHandle,
  ResizablePanel,
  ResizablePanelGroup,
} from "@/components/ui/resizable";
import { Separator } from "@/components/ui/separator";
import {
  TrafficLightSpacer,
  WindowControls,
  isMac,
} from "@/components/WindowControls";
import { useSidebarPanel, SIDEBAR_COLLAPSED_WIDTH } from "@/modules/sidebar";
import { usePreferencesStore } from "@/modules/settings/store";
import SkillsPage from "@/pages/SkillsPage";
import SettingsPage from "@/pages/SettingsPage";
import ProjectsPage from "@/pages/ProjectsPage";
import AgentsPage from "@/pages/AgentsPage";

const NAV = [
  { key: "skills", label: "技能", icon: Layers },
  { key: "projects", label: "项目", icon: FolderTree },
  { key: "agents", label: "Agents", icon: Bot },
] as const;

type TabKey = (typeof NAV)[number]["key"] | "settings";

export default function App() {
  const [tab, setTab] = useState<TabKey>("skills");
  const toasts = useToasts();
  const {
    sidebarRef,
    sidebarWidthRef,
    collapsed,
    setCollapsed,
    persistCollapsed,
    persistWidth,
    toggleSidebar,
  } = useSidebarPanel();

  // UI zoom preference. Applied to the content area only — the sidebar and
  // title chrome stay at 100% so navigation never scales with the content.
  const zoomLevel = usePreferencesStore((s) => s.zoomLevel);

  // Select a page. The collapsed state is an icon rail where every nav icon is
  // visible, so selecting must NOT auto-expand — otherwise clicking an icon in
  // the rail would pop the whole sidebar open instead of just switching pages.
  const select = (key: TabKey) => {
    setTab(key);
  };

  // Global keyboard shortcuts
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      const mod = e.metaKey || e.ctrlKey;
      if (mod && e.key === "1") { e.preventDefault(); select("skills"); }
      else if (mod && e.key === "2") { e.preventDefault(); select("projects"); }
      else if (mod && e.key === "3") { e.preventDefault(); select("agents"); }
      else if (mod && e.key === "4") { e.preventDefault(); select("settings"); }
      else if (mod && e.key === "k") {
        e.preventDefault();
        const input = document.querySelector<HTMLInputElement>('input[type="search"], input[placeholder*="搜索"], input[placeholder*="Search"]');
        input?.focus();
      } else if (mod && e.key === "a" && tab === "skills") {
        e.preventDefault();
        document.dispatchEvent(new CustomEvent("select-all-skills"));
      } else if ((e.key === "Delete" || e.key === "Backspace") && tab === "skills" && !mod) {
        // Only trigger if focus is not on an input
        const target = e.target as HTMLElement;
        if (target.tagName !== "INPUT" && target.tagName !== "TEXTAREA") {
          document.dispatchEvent(new CustomEvent("delete-selected-skills"));
        }
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [tab]);

  return (
    <div className="relative flex h-full">
      {/* The whole window is one resizable split: the sidebar runs edge-to-edge
          to the very top (no separate header bar). On macOS the sidebar's top
          padding reserves room for the native traffic lights. */}
      <ResizablePanelGroup
        orientation="horizontal"
        className="min-h-0 flex-1"
        onLayoutChanged={(_: Layout, meta: LayoutChangedMeta) => {
          const width = sidebarRef.current?.getSize().inPixels ?? 0;
          if (meta.isUserInteraction && width > 0) persistWidth(width);
        }}
      >
        <ResizablePanel
          id="sidebar"
          panelRef={sidebarRef}
          defaultSize={collapsed ? `${SIDEBAR_COLLAPSED_WIDTH}px` : `${sidebarWidthRef.current}px`}
          minSize={`160px`}
          maxSize={`400px`}
          collapsible
          collapsedSize={SIDEBAR_COLLAPSED_WIDTH}
          onResize={(size) => {
            const isCollapsed = size.inPixels <= SIDEBAR_COLLAPSED_WIDTH;
            setCollapsed(isCollapsed);
            persistCollapsed(isCollapsed);
          }}
        >
          <aside className="flex h-full min-h-0 flex-col border-r border-border/60 bg-muted/40">
            {collapsed ? (
              /* Collapsed icon rail — keeps the brand mark plus a grab handle
                 at the top edge and the expand toggle below, so the narrow
                 strip is still usable on its own. */
              <div className="app-drag flex h-full min-h-0 flex-col items-center px-2 pt-3">
                <button
                  type="button"
                  title="展开侧边栏"
                  onClick={toggleSidebar}
                  className="app-no-drag flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                >
                  <PanelLeftOpen className="h-4 w-4" />
                </button>

                <Separator className="my-3 w-8" />

                {/* Primary nav icons stay at the top, right after the expand
                    toggle — matching the expanded layout where nav is topmost. */}
                <div className="flex flex-col items-center gap-1">
                  {NAV.map(({ key, label, icon: Icon }) => (
                    <IconButton
                      key={key}
                      active={tab === key}
                      icon={Icon}
                      label={label}
                      onClick={() => select(key)}
                    />
                  ))}
                </div>

                {/* Settings pinned to the bottom, mirroring the expanded rail. */}
                <div className="mt-auto flex flex-col items-center gap-1 pb-4">
                  <IconButton
                    active={tab === "settings"}
                    icon={Settings}
                    label="设置"
                    onClick={() => select("settings")}
                  />
                </div>
              </div>
            ) : (
              <>
                {/* Brand row — stretches to the window top. On macOS the
                    traffic-light spacer reserves room for the native controls. */}
                <div className="app-drag flex h-14 shrink-0 items-center gap-2 px-3">
                  <div className="app-no-drag flex min-w-0 items-center gap-2">
                    <TrafficLightSpacer />
                    <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-accent text-primary-foreground">
                      <Sparkles className="h-4 w-4" />
                    </div>
                    <span className="truncate text-sm font-semibold">
                      AI-Manager
                    </span>
                  </div>
                  <button
                    type="button"
                    title="收起侧边栏"
                    onClick={toggleSidebar}
                    className="app-no-drag flex h-7 w-7 shrink-0 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground"
                  >
                    <PanelLeftClose className="h-4 w-4" />
                  </button>
                </div>

                <div className="flex min-h-0 flex-1 flex-col overflow-y-auto px-3 py-3">
                  {/* Nav (main pages). */}
                  <nav className="flex flex-col gap-1">
                    {NAV.map(({ key, label, icon: Icon }) => (
                      <NavButton
                        key={key}
                        active={tab === key}
                        icon={Icon}
                        label={label}
                        onClick={() => select(key)}
                      />
                    ))}
                  </nav>

                  <Separator className="my-3" />

                  {/* Bottom: settings. */}
                  <div className="mt-auto flex flex-col gap-1">
                    <NavButton
                      active={tab === "settings"}
                      icon={Settings}
                      label="设置"
                      onClick={() => select("settings")}
                    />
                  </div>
                </div>
              </>
            )}
          </aside>
        </ResizablePanel>

        <ResizableHandle withHandle />

        {/* Content area. */}
        <ResizablePanel>
          <div className="relative flex h-full min-h-0 flex-col">
            {/* Content-area drag strip — the frameless window can be dragged
                from anywhere along the content area's top edge, not just the
                sidebar's brand row. The floating window controls overlap its
                top-right (they're `app-no-drag`), so buttons stay clickable.
                bg-background keeps it opaque so WebView2's non-client hit test
                registers it as a real caption (drag) region. */}
            <div className="app-drag h-12 shrink-0 bg-background" />
            <main
              className={cn(
                "min-h-0 flex-1 overflow-y-auto px-7 py-6",
                // On Windows/Linux the floating window controls sit in the top
                // right corner — reserve that strip so content never hides under
                // them. macOS uses native traffic lights, so no offset needed.
                !isMac && "pr-24"
              )}
              style={{
                zoom:
                  Number.isFinite(zoomLevel) && zoomLevel > 0
                    ? `${zoomLevel * 100}%`
                    : "100%",
              }}
            >
              {tab === "skills" ? <SkillsPage /> : tab === "projects" ? <ProjectsPage /> : tab === "agents" ? <AgentsPage /> : <SettingsPage />}
            </main>
          </div>
        </ResizablePanel>
      </ResizablePanelGroup>

      {/* Window controls (Windows/Linux) — pinned to the top-right corner,
          floating over the content so no header bar is needed. macOS renders
          native traffic lights instead (WindowControls returns null). */}
      <div className="absolute right-0 top-0 z-40 flex h-12 items-center pr-1">
        <WindowControls />
      </div>

      {/* Toasts. */}
      <div className="pointer-events-none fixed bottom-6 left-1/2 z-50 flex -translate-x-1/2 flex-col gap-2">
        <AnimatePresence>
          {toasts.map((t) => (
            <motion.div
              key={t.id}
              initial={{ opacity: 0, y: 20, scale: 0.95 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -10, scale: 0.95 }}
              transition={{ duration: 0.2, ease: "easeOut" }}
              className={cn(
                "pointer-events-auto rounded-lg border bg-popover px-4 py-2.5 text-sm shadow-lg",
                t.error ? "border-destructive text-destructive" : "border-border"
              )}
            >
              {t.message}
            </motion.div>
          ))}
        </AnimatePresence>
      </div>
    </div>
  );
}

function NavButton({
  active,
  icon: Icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: typeof Layers;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "group relative flex h-9 items-center gap-2.5 rounded-lg px-3 text-sm text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground",
        active &&
          "bg-secondary font-medium text-foreground shadow-[inset_2px_0_0_0_var(--primary)]"
      )}
    >
      <Icon className="h-4 w-4 shrink-0" />
      {label}
    </button>
  );
}

// Icon-only nav item for the collapsed rail.
function IconButton({
  active,
  icon: Icon,
  label,
  onClick,
}: {
  active: boolean;
  icon: typeof Layers;
  label: string;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      title={label}
      onClick={onClick}
      className={cn(
        "group relative flex h-9 w-9 items-center justify-center rounded-lg text-muted-foreground transition-colors hover:bg-secondary hover:text-foreground",
        active &&
          "bg-secondary text-foreground shadow-[inset_2px_0_0_0_var(--primary)]"
      )}
    >
      <Icon className="h-4 w-4 shrink-0" />
    </button>
  );
}
