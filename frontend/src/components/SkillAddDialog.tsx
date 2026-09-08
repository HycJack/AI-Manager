import { useState } from "react";
import {
  FolderOpen,
  Globe,
  Plug,
  PackageCheck,
  Loader2,
  ChevronDown,
  Folder,
  Plus,
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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { cn } from "@/lib/utils";
import { toast } from "@/lib/toast";
import { AddSkill } from "@bindings/ai-manager/internal/app/skillservice";
import { BrowseLibrary } from "@bindings/ai-manager/internal/app/configservice";

export type SourceKind = "local" | "npx" | "claude" | "existing";

const SOURCE_OPTIONS: {
  value: SourceKind;
  label: string;
  icon: typeof FolderOpen;
  placeholder: string;
  description: string;
}[] = [
  {
    value: "local",
    label: "Local Folder",
    icon: FolderOpen,
    placeholder: "Select a folder containing skills…",
    description: "Scan a local directory for SKILL.md or README.md files.",
  },
  {
    value: "npx",
    label: "Npx / GitHub / skills.sh",
    icon: Globe,
    placeholder: "e.g. github.com/owner/repo or skills.sh/packagename",
    description: "Add from an npm package, GitHub repo, or skills.sh endpoint.",
  },
  {
    value: "claude",
    label: "Claude Plugin",
    icon: Plug,
    placeholder: "Plugin name or ID…",
    description: "Import a Claude Code plugin as a skill.",
  },
  {
    value: "existing",
    label: "Existing Installations",
    icon: PackageCheck,
    placeholder: "Comma-separated paths…",
    description: "Adopt skills already installed in agent directories.",
  },
];

interface SkillAddDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onAdded?: () => void;
}

export default function SkillAddDialog({
  open,
  onOpenChange,
  onAdded,
}: SkillAddDialogProps) {
  const [sourceKind, setSourceKind] = useState<SourceKind>("local");
  const [inputPath, setInputPath] = useState("");
  const [groupName, setGroupName] = useState("");
  const [loading, setLoading] = useState(false);
  const [picking, setPicking] = useState(false);

  const currentOption = SOURCE_OPTIONS.find((o) => o.value === sourceKind)!;

  async function handleBrowseFolder() {
    setPicking(true);
    try {
      const path = await BrowseLibrary();
      if (path) {
        setInputPath(path);
      }
    } catch {
      toast("Failed to open folder picker", true);
    } finally {
      setPicking(false);
    }
  }

  async function handleAdd() {
    if (!inputPath.trim()) {
      toast(
        sourceKind === "local" ? "Please select a folder" : "Please enter a source",
        true,
      );
      return;
    }

    setLoading(true);
    try {
      await AddSkill(sourceKind, inputPath.trim(), groupName.trim(), []);
      toast(`Skill added from ${currentOption.label}`);
      onOpenChange(false);
      onAdded?.();
      // Reset state
      setSourceKind("local");
      setInputPath("");
      setGroupName("");
    } catch (err: any) {
      toast(`Failed to add skill: ${err?.message ?? err}`, true);
    } finally {
      setLoading(false);
    }
  }

  function handleSourceChange(kind: string) {
    setSourceKind(kind as SourceKind);
    setInputPath("");
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-xl">
        <DialogHeader>
          <DialogTitle>Add Skill</DialogTitle>
          <DialogDescription>
            Import a skill from one of four source types.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-4">
          {/* Source type selector */}
          <div className="grid gap-2">
            <Label>Source</Label>
            <div className="grid grid-cols-2 gap-2">
              {SOURCE_OPTIONS.map((opt) => (
                <button
                  key={opt.value}
                  type="button"
                  onClick={() => handleSourceChange(opt.value)}
                  className={cn(
                    "flex items-center gap-2 rounded-lg border px-3 py-2.5 text-left text-sm transition-colors",
                    sourceKind === opt.value
                      ? "border-primary bg-primary/5 font-medium"
                      : "hover:bg-secondary",
                  )}
                >
                  <opt.icon className="h-4 w-4 shrink-0" />
                  <span>{opt.label}</span>
                </button>
              ))}
            </div>
            <p className="text-xs text-muted-foreground">
              {currentOption.description}
            </p>
          </div>

          <Separator />

          {/* Source-specific input */}
          <div className="grid gap-2">
            <Label>Path / Input</Label>
            {sourceKind === "local" ? (
              <div className="flex gap-2">
                <div className="relative flex-1">
                  <Folder className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                  <Input
                    value={inputPath}
                    onChange={(e) => setInputPath(e.target.value)}
                    placeholder={currentOption.placeholder}
                    className="pl-9"
                    disabled={picking}
                  />
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleBrowseFolder}
                  disabled={picking || loading}
                >
                  {picking ? (
                    <Loader2 className="h-4 w-4 animate-spin" />
                  ) : (
                    <FolderOpen className="h-4 w-4" />
                  )}
                  Browse
                </Button>
              </div>
            ) : (
              <div className="relative">
                <currentOption.icon className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <Input
                  value={inputPath}
                  onChange={(e) => setInputPath(e.target.value)}
                  placeholder={currentOption.placeholder}
                  className="pl-9"
                  disabled={loading}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && !loading) handleAdd();
                  }}
                />
              </div>
            )}
          </div>

          {/* Group selector */}
          <div className="grid gap-2">
            <Label>
              Group <span className="text-muted-foreground font-normal">(optional)</span>
            </Label>
            <div className="relative">
              <Input
                value={groupName}
                onChange={(e) => setGroupName(e.target.value)}
                placeholder="e.g. frontend, tools, testing…"
                disabled={loading}
                className="pr-9"
              />
              <ChevronDown className="absolute right-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
            </div>
          </div>
        </div>

        <DialogFooter className="gap-2 sm:gap-0">
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={loading}>
            Cancel
          </Button>
          <Button onClick={handleAdd} disabled={loading || !inputPath.trim()}>
            {loading ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Adding…
              </>
            ) : (
              <>
                <Plus className="h-4 w-4" />
                Add Skill
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
