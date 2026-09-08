import { useState, useEffect } from "react";
import {
  FolderOpen,
  Loader2,
  PackageCheck,
  Monitor,
  Globe,
} from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { toast } from "@/lib/toast";
import { useI18n } from "@/modules/i18n";
import { cn } from "@/lib/utils";
import { InstallSkills, BrowseProject, ListProjects } from "@bindings/ai-manager/internal/app/projectservice";
import { InstallTarget } from "@bindings/ai-manager/internal/agents/models";
import type { Summary } from "@bindings/ai-manager/internal/skill/models";

// ---------------------------------------------------------------------------
// Target definitions
// ---------------------------------------------------------------------------

const ALL_TARGETS: { value: string; label: string }[] = [
  { value: "shared", label: "Shared" },
  { value: "claude", label: "Claude" },
  { value: "codex", label: "Codex" },
  { value: "cursor", label: "Cursor" },
  { value: "cline", label: "Cline" },
  { value: "continue", label: "Continue" },
  { value: "aider", label: "Aider" },
  { value: "antigravity", label: "Antigravity" },
  { value: "trae", label: "Trae" },
  { value: "windsurf", label: "Windsurf" },
];

// ---------------------------------------------------------------------------
// Props
// ---------------------------------------------------------------------------

interface SkillInstallDialogProps {
  open: boolean;
  skills: Summary[];
  onConfirm: () => void;
  onCancel: () => void;
  /** Pre-populated projects list (optional). If empty, we fetch on open. */
  projects?: { path: string; name: string }[];
}

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

export function SkillInstallDialog({
  open,
  skills,
  onConfirm,
  onCancel,
  projects = [],
}: SkillInstallDialogProps) {
  const { t } = useI18n();
  const [location, setLocation] = useState<"project" | "global">("project");
  const [projectPath, setProjectPath] = useState("");
  const [selectedTargets, setSelectedTargets] = useState<Set<string>>(
    new Set(["shared"]),
  );
  const [installing, setInstalling] = useState(false);
  const [picking, setPicking] = useState(false);
  const [loadingProjects, setLoadingProjects] = useState(false);
  const [fetchedProjects, setFetchedProjects] = useState<{
    path: string;
    name: string;
  }[]>([]);

  // Reset state when dialog opens
  useEffect(() => {
    if (open) {
      setLocation("project");
      setProjectPath("");
      setSelectedTargets(new Set(["shared"]));
      setInstalling(false);
      setPicking(false);
      setFetchedProjects([]);
    }
  }, [open]);

  // Fetch projects when opened if none provided
  useEffect(() => {
    if (open && projects.length === 0) {
      setLoadingProjects(true);
      ListProjects()
        .then((result) => {
          setFetchedProjects(
            result.map((p) => ({ path: p.path, name: p.name })),
          );
        })
        .catch(() => {})
        .finally(() => setLoadingProjects(false));
    }
  }, [open, projects]);

  const allProjects = projects.length > 0 ? projects : fetchedProjects;

  // When switching to global, force only "shared" selected
  useEffect(() => {
    if (location === "global") {
      setSelectedTargets(new Set(["shared"]));
    }
  }, [location]);

  const toggleTarget = (target: string, isDisabled: boolean) => {
    if (isDisabled) return;
    setSelectedTargets((prev) => {
      const next = new Set(prev);
      if (next.has(target)) {
        next.delete(target);
      } else {
        next.add(target);
      }
      return next;
    });
  };

  async function handleBrowseProject() {
    setPicking(true);
    try {
      const path = await BrowseProject();
      if (path) {
        setProjectPath(path);
      }
    } catch {
      toast(t("toast.installBrowseFailed"), true);
    } finally {
      setPicking(false);
    }
  }

  function handleLocationChange(newLocation: "project" | "global") {
    setLocation(newLocation);
    if (newLocation === "project") {
      // Restore shared as default if nothing selected
      setSelectedTargets((prev) => {
        if (prev.size === 0) return new Set(["shared"]);
        return prev;
      });
    }
  }

  const canConfirm =
    skills.length > 0 &&
    selectedTargets.size > 0 &&
    (location === "global" || projectPath.trim() !== "") &&
    !installing;

  const summaryCount = `${skills.length} skill${skills.length !== 1 ? "s" : ""} × ${selectedTargets.size} target${selectedTargets.size !== 1 ? "s" : ""}`;

  async function handleInstall() {
    if (!canConfirm) return;
    setInstalling(true);
    try {
      const names = skills.map((s) => s.name);
      const targets = Array.from(selectedTargets) as InstallTarget[];
      await InstallSkills(names, projectPath, targets, location === "global");
      toast(
        t("toast.installSuccess", {
          count: skills.length,
          targets: selectedTargets.size,
        }),
        false,
      );
      onConfirm();
    } catch (err: any) {
      toast(t("toast.installFailed", { error: err?.message ?? String(err) }), true);
    } finally {
      setInstalling(false);
    }
  }

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onCancel()}>
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t("install.title")}</DialogTitle>
          <DialogDescription>{t("install.subtitle")}</DialogDescription>
        </DialogHeader>

        <div className="grid gap-5 max-h-[65vh] overflow-y-auto pr-1">
          {/* Skills to install */}
          <div>
            <Label className="text-sm font-medium">
              {t("install.skillsToInstall")}
            </Label>
            <div className="mt-2 max-h-32 overflow-y-auto rounded-md border">
              {skills.map((skill) => (
                <div
                  key={skill.id}
                  className="flex items-center justify-between gap-3 px-3 py-2 text-sm border-b last:border-b-0"
                >
                  <span className="truncate font-medium">{skill.name}</span>
                  <span className="text-muted-foreground text-xs shrink-0">
                    v{skill.version}
                  </span>
                </div>
              ))}
            </div>
          </div>

          <Separator />

          {/* Install location */}
          <div className="grid gap-2">
            <Label>{t("install.location")}</Label>
            <div className="grid grid-cols-2 gap-2">
              <button
                type="button"
                onClick={() => handleLocationChange("project")}
                disabled={installing}
                className={cn(
                  "flex items-center gap-2 rounded-lg border px-3 py-2.5 text-left text-sm transition-colors",
                  location === "project"
                    ? "border-primary bg-primary/5 font-medium"
                    : "hover:bg-secondary",
                )}
              >
                <Monitor className="h-4 w-4 shrink-0" />
                <span>{t("install.project")}</span>
              </button>
              <button
                type="button"
                onClick={() => handleLocationChange("global")}
                disabled={installing}
                className={cn(
                  "flex items-center gap-2 rounded-lg border px-3 py-2.5 text-left text-sm transition-colors",
                  location === "global"
                    ? "border-primary bg-primary/5 font-medium"
                    : "hover:bg-secondary",
                )}
              >
                <Globe className="h-4 w-4 shrink-0" />
                <span>{t("install.global")}</span>
              </button>
            </div>
          </div>

          {/* Project path (only in project mode) */}
          {location === "project" && (
            <div className="grid gap-2">
              <Label>{t("install.projectPath")}</Label>
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <FolderOpen className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                  <input
                    type="text"
                    value={projectPath}
                    onChange={(e) => setProjectPath(e.target.value)}
                    placeholder={t("install.projectPlaceholder")}
                    className="w-full h-9 rounded-md border border-input bg-transparent pl-9 pr-3 text-sm shadow-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                    disabled={installing}
                  />
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleBrowseProject}
                  disabled={picking || installing}
                >
                  {picking ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <FolderOpen className="h-4 w-4" />
                  )}
                  {t("install.browse")}
                </Button>
              </div>

              {/* Project picker from registered projects */}
              {allProjects.length > 0 && (
                <div className="flex flex-wrap gap-1.5 mt-1">
                  {allProjects.map((p) => (
                    <button
                      key={p.path}
                      type="button"
                      onClick={() => setProjectPath(p.path)}
                      disabled={installing}
                      className={cn(
                        "rounded-md border px-2.5 py-1 text-xs transition-colors",
                        projectPath === p.path
                          ? "border-primary bg-primary/10 font-medium"
                          : "hover:bg-secondary",
                      )}
                    >
                      {p.name || p.path.split("/").pop()}
                    </button>
                  ))}
                </div>
              )}
            </div>
          )}

          <Separator />

          {/* Targets grid */}
          <div className="grid gap-2">
            <Label>{t("install.targets")}</Label>
            <div className="grid grid-cols-3 sm:grid-cols-5 gap-2">
              {ALL_TARGETS.map((target) => {
                const isDisabled = location === "global" && target.value !== "shared";
                const checked = selectedTargets.has(target.value);
                return (
                  <label
                    key={target.value}
                    className={cn(
                      "flex items-center gap-2 rounded-md border px-2.5 py-2 text-sm cursor-pointer transition-colors",
                      checked && !isDisabled
                        ? "border-primary bg-primary/5"
                        : "hover:bg-secondary",
                      isDisabled && "opacity-40 cursor-not-allowed hover:bg-transparent",
                    )}
                  >
                    <Checkbox
                      checked={checked}
                      disabled={isDisabled || installing}
                      onCheckedChange={() => toggleTarget(target.value, isDisabled)}
                    />
                    <span className="text-xs font-medium">{target.label}</span>
                  </label>
                );
              })}
            </div>
          </div>

          <Separator />

          {/* Summary */}
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">
              {t("install.summary")}
            </span>
            <span className="font-semibold flex items-center gap-1.5">
              <PackageCheck className="h-4 w-4" />
              {summaryCount}
            </span>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={onCancel} disabled={installing}>
            {t("install.cancel")}
          </Button>
          <Button
            onClick={handleInstall}
            disabled={!canConfirm}
          >
            {installing ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                {t("install.installing")}
              </>
            ) : (
              <>
                <PackageCheck className="h-4 w-4" />
                {t("install.confirm", { count: skills.length })}
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
