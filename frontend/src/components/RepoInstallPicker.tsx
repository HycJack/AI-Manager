import { useEffect, useMemo, useState } from "react";
import { Check, Loader2, Package, X } from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { toast } from "@/lib/toast";
import { useI18n } from "@/modules/i18n";
import { ListRepoSkills, AddSkill } from "@bindings/ai-manager/internal/app/skillservice";
import type { AvailableSkill } from "@bindings/ai-manager/internal/app/models";

interface RepoInstallPickerProps {
  open: boolean;
  /** "owner/repo" or "owner/repo/subdir" */
  source: string;
  /** Skill id/name from the discovery hit; preselects a matching row */
  skillKey: string;
  onClose: () => void;
  onInstalled: () => void;
}

/** normaliseSlug makes "grill_me" and "grill-me" comparable. */
function normaliseSlug(s: string): string {
  return s.toLowerCase().replace(/[_\s]+/g, "-");
}

/**
 * RepoInstallPicker resolves a repository into the skills it actually ships
 * and lets the caller choose which to install. Repositories such as
 * mattpocock/skills hold dozens of skills, so installing them all silently
 * would be surprising.
 */
export function RepoInstallPicker({ open, source, skillKey, onClose, onInstalled }: RepoInstallPickerProps) {
  const { t } = useI18n();
  const [loading, setLoading] = useState(false);
  const [addError, setAddError] = useState("");
  const [skills, setSkills] = useState<AvailableSkill[] | null>(null);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [adding, setAdding] = useState(false);

  useEffect(() => {
    if (!open || !source) {
      setSkills(null);
      setSelected(new Set());
      setAddError("");
      return;
    }
    let cancelled = false;
    setLoading(true);
    setAddError("");
    ListRepoSkills(source)
      .then((list) => {
        if (cancelled) return;
        setSkills(list);
        setSelected(new Set(pickDefault(list, skillKey)));
      })
      .catch((e) => {
        if (!cancelled) toast(String(e), true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [open, source, skillKey]);

  const toggle = (slug: string) => {
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(slug)) next.delete(slug);
      else next.add(slug);
      return next;
    });
  };

  const allSelected = skills !== null && skills.length > 0 && selected.size === skills.length;

  const install = async () => {
    if (!source || selected.size === 0) return;
    setAdding(true);
    try {
      await AddSkill("npx", source, "", Array.from(selected));
      toast(t("skills.added"));
      onInstalled();
      onClose();
    } catch (e) {
      setAddError(String(e));
    } finally {
      setAdding(false);
    }
  };

  const slugList = useMemo(() => skills?.map((s) => s.slug) ?? [], [skills]);

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-[60] flex items-center justify-center bg-black/50" onClick={onClose}>
      <div
        className="w-full max-w-md max-h-[80vh] rounded-xl border bg-background shadow-lg flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center gap-3 border-b px-5 py-3">
          <Package className="h-4 w-4 shrink-0 text-muted-foreground" />
          <h3 className="truncate text-base font-semibold">{t("skills.repoPicker.title")}</h3>
          <span className="ml-auto truncate text-xs text-muted-foreground">{source}</span>
          <button
            onClick={onClose}
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:bg-muted hover:text-foreground transition-colors"
          >
            <X className="h-4 w-4" />
          </button>
        </div>

        <div className="flex-1 overflow-auto p-3">
          {loading ? (
            <div className="flex h-40 items-center justify-center">
              <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
            </div>
          ) : skills && skills.length > 0 ? (
            <div className="space-y-2">
              <div className="flex items-center gap-2 px-1">
                <Badge variant="secondary">{skills.length}</Badge>
                <span className="text-xs text-muted-foreground">{t("skills.repoPicker.subtitle")}</span>
                <button
                  className="ml-auto text-xs text-primary hover:underline"
                  onClick={() => setSelected(new Set(allSelected ? [] : slugList))}
                >
                  {allSelected ? t("skills.repoPicker.none") : t("skills.repoPicker.all")}
                </button>
              </div>
              {skills.map((s) => {
                const on = selected.has(s.slug);
                return (
                  <button
                    key={s.slug}
                    onClick={() => toggle(s.slug)}
                    className={cn(
                      "flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-colors",
                      on ? "border-primary/40 bg-muted/50" : "hover:bg-muted/30",
                    )}
                  >
                    <span
                      className={cn(
                        "flex h-4 w-4 shrink-0 items-center justify-center rounded border",
                        on ? "border-primary bg-primary text-primary-foreground" : "border-muted-foreground/50",
                      )}
                    >
                      {on && <Check className="h-3 w-3" />}
                    </span>
                    <span className="min-w-0 flex-1">
                      <span className="block truncate text-sm font-medium">{s.name}</span>
                      <span className="block truncate text-xs text-muted-foreground">{s.slug}</span>
                    </span>
                  </button>
                );
              })}
            </div>
          ) : (
            <p className="py-8 text-center text-sm text-muted-foreground">{t("skills.noResults")}</p>
          )}
        </div>

        <div className="border-t px-5 py-3">
          {addError && <p className="mb-2 text-xs text-destructive">{addError}</p>}
          <div className="flex items-center gap-2">
            <span className="text-xs text-muted-foreground">
              {selected.size} / {skills?.length ?? 0} · {t("skills.repoPicker.selected")}
            </span>
            <Button variant="outline" size="sm" className="ml-auto" onClick={onClose} disabled={adding}>
              {t("delete.cancel")}
            </Button>
            <Button size="sm" onClick={install} disabled={adding || selected.size === 0 || loading}>
              {adding ? <Loader2 className="mr-1 h-4 w-4 animate-spin" /> : null}
              {adding ? t("skills.adding") : t("skills.repoPicker.install")}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}

/** Preselect the row the discovery hit points at, else everything. */
function pickDefault(skills: AvailableSkill[], skillKey: string): string[] {
  const key = skillKey.trim();
  if (key) {
    const normal = normaliseSlug(key);
    const hit = skills.find((s) => {
      if (s.slug === key || s.name === key) return true;
      return normaliseSlug(s.slug) === normal || normaliseSlug(s.name) === normal;
    });
    if (hit) return [hit.slug];
  }
  return skills.map((s) => s.slug);
}
