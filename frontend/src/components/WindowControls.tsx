import { useEffect, useState } from "react";
import { Minus, Square, X, Copy } from "lucide-react";
import { cn } from "@/lib/utils";
import { Window } from "@wailsio/runtime";

// Detects whether we're on macOS (native traffic lights handle min/max/close).
export const isMac =
  typeof navigator !== "undefined" &&
  /Mac|iPhone|iPad|iPod/.test(navigator.platform || navigator.userAgent || "");

/**
 * Window controls for the custom (frameless) title bar.
 *
 * - macOS: renders nothing — the hidden-inset title bar provides the native
 *   traffic-light buttons.
 * - Windows/Linux: renders min / max-restore / close buttons.
 */
export function WindowControls() {
  const [maximised, setMaximised] = useState(false);

  useEffect(() => {
    if (isMac) return;
    Window.IsMaximised?.().then(setMaximised).catch(() => {});
  }, []);

  if (isMac) return null;

  const onMin = () => Window.Minimise?.();
  const onMax = async () => {
    await Window.ToggleMaximise?.();
    Window.IsMaximised?.().then(setMaximised).catch(() => {});
  };
  const onClose = () => Window.Close?.();

  return (
    <div className="app-no-drag flex items-center">
      <button
        onClick={onMin}
        className="flex h-8 w-11 items-center justify-center text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        title="最小化"
      >
        <Minus className="h-4 w-4" />
      </button>
      <button
        onClick={onMax}
        className="flex h-8 w-11 items-center justify-center text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        title={maximised ? "向下还原" : "最大化"}
      >
        {maximised ? <Copy className="h-3.5 w-3.5 rotate-90" /> : <Square className="h-3.5 w-3.5" />}
      </button>
      <button
        onClick={onClose}
        className="flex h-8 w-11 items-center justify-center text-muted-foreground transition-colors hover:bg-destructive hover:text-destructive-foreground"
        title="关闭"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}

/** Left-side spacer that reserves room for macOS traffic lights. */
export function TrafficLightSpacer({ className }: { className?: string }) {
  return (
    <div
      className={cn(
        "app-no-drag hidden",
        isMac && "flex",
        isMac && (className ?? "w-[72px]")
      )}
    />
  );
}
