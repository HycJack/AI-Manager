import { useState, useEffect, useMemo } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Checkbox } from "@/components/ui/checkbox";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Alert, AlertTitle, AlertDescription } from "@/components/ui/alert";
import { AlertTriangle } from "lucide-react";
import { RemoveSkill } from "@bindings/ai-manager/internal/app/skillservice";
import type { Summary } from "@bindings/ai-manager/internal/skill/models";
import { useI18n } from "@/modules/i18n";
import { toast } from "@/lib/toast";
import { cn } from "@/lib/utils";

export interface SkillDeleteDialogProps {
  open: boolean;
  skills: Summary[];
  builtinIds: Set<string>;
  onConfirm: () => void;
  onCancel: () => void;
}

export function SkillDeleteDialog({
  open,
  skills,
  builtinIds,
  onConfirm,
  onCancel,
}: SkillDeleteDialogProps) {
  const { t } = useI18n();
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [deleting, setDeleting] = useState(false);

  // Reset selection when dialog opens
  useEffect(() => {
    if (open) {
      // Select all non-builtin skills by default
      const initial = new Set<string>();
      for (const s of skills) {
        if (!builtinIds.has(s.id)) {
          initial.add(s.id);
        }
      }
      setSelected(initial);
      setDeleting(false);
    }
  }, [open, skills, builtinIds]);

  const isBuiltin = useMemo(
    () => (skillId: string) => builtinIds.has(skillId),
    [builtinIds],
  );

  const hasBuiltinSelected = useMemo(
    () => Array.from(selected).some((id) => builtinIds.has(id)),
    [selected, builtinIds],
  );

  const hasBuiltinInList = useMemo(
    () => skills.some((s) => builtinIds.has(s.id)),
    [skills, builtinIds],
  );

  const selectedCount = selected.size;
  const canConfirm = selectedCount > 0 && !hasBuiltinSelected;

  const toggleSkill = (skillId: string, builtin: boolean) => {
    if (builtin) return; // builtin skills can't be toggled
    setSelected((prev) => {
      const next = new Set(prev);
      if (next.has(skillId)) {
        next.delete(skillId);
      } else {
        next.add(skillId);
      }
      return next;
    });
  };

  const handleConfirm = async () => {
    if (!canConfirm) return;
    setDeleting(true);
    try {
      for (const skillId of selected) {
        // Find the skill by id to get its name/slug
        const skill = skills.find((s) => s.id === skillId);
        if (skill) {
          await RemoveSkill(skill.name);
        }
      }
      toast(
        t("toast.skillDeleted", { count: selectedCount }),
        false,
      );
      onConfirm();
    } catch (e) {
      toast(t("toast.skillDeleteFailed", { error: String(e) }), true);
    } finally {
      setDeleting(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={(v) => !v && onCancel()}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t("delete.title")}</DialogTitle>
          <DialogDescription>{t("delete.subtitle")}</DialogDescription>
        </DialogHeader>

        <div className="max-h-64 overflow-y-auto rounded-md border">
          {skills.map((skill) => {
            const builtin = isBuiltin(skill.id);
            const checked = selected.has(skill.id);
            return (
              <label
                key={skill.id}
                className={cn(
                  "flex items-center gap-3 px-3 py-2 text-sm cursor-pointer hover:bg-accent transition-colors",
                  builtin && "opacity-50 cursor-not-allowed hover:bg-transparent",
                )}
              >
                <Checkbox
                  checked={checked}
                  disabled={builtin}
                  onCheckedChange={() => toggleSkill(skill.id, builtin)}
                />
                <span className="flex-1 truncate">{skill.name}</span>
                <span className="text-muted-foreground text-xs">
                  v{skill.version}
                </span>
                {builtin && (
                  <Badge variant="secondary" className="text-xs">
                    BUILTIN
                  </Badge>
                )}
              </label>
            );
          })}
        </div>

        {hasBuiltinInList && (
          <Alert variant="destructive">
            <AlertTriangle className="h-4 w-4" />
            <AlertTitle>{t("delete.builtinWarning")}</AlertTitle>
          </Alert>
        )}

        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={deleting}>
            {t("delete.cancel")}
          </Button>
          <Button
            variant="destructive"
            onClick={handleConfirm}
            disabled={!canConfirm || deleting}
          >
            {deleting
              ? "…"
              : t("delete.select", { count: selectedCount })}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
