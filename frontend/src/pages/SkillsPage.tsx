import { useEffect, useState, useMemo, useCallback, useRef } from "react";
import {
  Search,
  Plus,
  ChevronRight,
  ChevronDown,
  FileText,
  Folder,
  Layers,
  Trash2,
  Package,
  Download,
  X,
  MoreVertical,
} from "lucide-react";
import { useVirtualizer } from "@tanstack/react-virtual";
import { cn } from "@/lib/utils";
import { toast } from "@/lib/toast";
import { useI18n } from "@/modules/i18n";
import {
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle,
} from "@/components/ui/resizable";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Checkbox } from "@/components/ui/checkbox";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ListSkills, GetSkill, AddSkill, RemoveSkill } from "@bindings/ai-manager/internal/app/skillservice";
import type { Summary } from "@bindings/ai-manager/internal/skill/models.js";
import type { SkillDetail } from "@bindings/ai-manager/internal/app/models.js";
import { SkillInstallDialog } from "@/components/SkillInstallDialog";

type GroupedSkills = {
  group: string;
  skills: Summary[];
};

/** Group skills alphabetically. */
function groupSkills(skills: Summary[]): GroupedSkills[] {
  const sorted = [...skills].sort((a, b) => a.name.localeCompare(b.name));
  return [{ group: "All Skills", skills: sorted }];
}

type FlatItem =
  | { type: "group-header"; group: string; skills: Summary[] }
  | { type: "skill"; skill: Summary };

function flattenGroups(groups: GroupedSkills[], collapsed: Set<string>): FlatItem[] {
  const items: FlatItem[] = [];
  for (const g of groups) {
    items.push({ type: "group-header", group: g.group, skills: g.skills });
    if (!collapsed.has(g.group)) {
      for (const s of g.skills) {
        items.push({ type: "skill", skill: s });
      }
    }
  }
  return items;
}

export default function SkillsPage() {
  const { t } = useI18n();
  const [search, setSearch] = useState("");
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [collapsedGroups, setCollapsedGroups] = useState<Set<string>>(new Set());
  const [skills, setSkills] = useState<Summary[]>([]);
  const [detail, setDetail] = useState<SkillDetail | null>(null);
  const [loadingDetail, setLoadingDetail] = useState(false);
  const [activeTab, setActiveTab] = useState("description");
  const [showAddDialog, setShowAddDialog] = useState(false);
  const [addKind, setAddKind] = useState("local");
  const [addInput, setAddInput] = useState("");
  const [addGroup, setAddGroup] = useState("");
  const [adding, setAdding] = useState(false);
  const [showDeleteDialog, setShowDeleteDialog] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Summary | null>(null);
  const [contextMenuSkill, setContextMenuSkill] = useState<Summary | null>(null);
  const [showInstallDialog, setShowInstallDialog] = useState(false);
  const [skillsToInstall, setSkillsToInstall] = useState<Summary[]>([]);
  const [deleting, setDeleting] = useState(false);

  const parentRef = useRef<HTMLDivElement>(null);
  const searchRef = useRef<HTMLInputElement>(null);

  // Fetch skills on mount
  useEffect(() => {
    ListSkills()
      .then(setSkills)
      .catch((e) => toast(String(e), true));
  }, []);

  // Filter by search
  const filtered = useMemo(() => {
    if (!search.trim()) return skills;
    const q = search.toLowerCase();
    return skills.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.slug.toLowerCase().includes(q) ||
        s.version.toLowerCase().includes(q),
    );
  }, [skills, search]);

  // Group skills
  const grouped = useMemo(() => groupSkills(filtered), [filtered]);

  // Flatten for virtualization
  const flatItems = useMemo(
    () => flattenGroups(grouped, collapsedGroups),
    [grouped, collapsedGroups],
  );

  // Virtualizer
  const virtualizer = useVirtualizer({
    count: flatItems.length,
    getScrollElement: () => parentRef.current,
    estimateSize: (index) => {
      const item = flatItems[index];
      if (!item) return 36;
      return item.type === "group-header" ? 36 : 44;
    },
    overscan: 20,
  });

  // Toggle group collapse
  const toggleGroup = useCallback((group: string) => {
    setCollapsedGroups((prev) => {
      const next = new Set(prev);
      if (next.has(group)) next.delete(group);
      else next.add(group);
      return next;
    });
  }, []);

  // Handle skill selection
  const handleSelect = useCallback(
    (skill: Summary, e: React.MouseEvent) => {
      const newSet = new Set(selectedIds);
      if (e.metaKey || e.ctrlKey) {
        if (newSet.has(skill.id)) newSet.delete(skill.id);
        else newSet.add(skill.id);
      } else if (e.shiftKey && selectedIds.size > 0) {
        newSet.clear();
        newSet.add(skill.id);
      } else {
        newSet.clear();
        newSet.add(skill.id);
      }
      setSelectedIds(newSet);

      // Load detail
      setLoadingDetail(true);
      GetSkill(skill.name)
        .then((d) => {
          setDetail(d);
          setActiveTab("description");
        })
        .catch((e) => toast(String(e), true))
        .finally(() => setLoadingDetail(false));
    },
    [selectedIds],
  );

  // Handle add skill
  const handleAdd = async () => {
    if (!addInput.trim()) return;
    setAdding(true);
    try {
      await AddSkill(addKind, addInput.trim(), addGroup.trim());
      toast("Skill added");
      const updated = await ListSkills();
      setSkills(updated);
      setShowAddDialog(false);
      setAddInput("");
      setAddGroup("");
    } catch (err) {
      toast(String(err), true);
    } finally {
      setAdding(false);
    }
  };

  // Handle delete skill
  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await RemoveSkill(deleteTarget.name);
      toast("Skill removed");
      setSelectedIds((prev) => {
        const next = new Set(prev);
        next.delete(deleteTarget.id);
        return next;
      });
      if (detail?.record?.name === deleteTarget.name) {
        setDetail(null);
      }
      const updated = await ListSkills();
      setSkills(updated);
      setShowDeleteDialog(false);
      setDeleteTarget(null);
    } catch (err) {
      toast(String(err), true);
    } finally {
      setDeleting(false);
    }
  };

  // Keyboard handler: Cmd+K (focus search), Cmd+A (select all), Cmd+N (add), Delete (delete)
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        e.preventDefault();
        searchRef.current?.focus();
        return;
      }
      if ((e.metaKey || e.ctrlKey) && e.key === "a") {
        e.preventDefault();
        setSelectedIds(new Set(filtered.map((s) => s.id)));
        return;
      }
      if ((e.metaKey || e.ctrlKey) && e.key === "n") {
        e.preventDefault();
        setShowAddDialog(true);
        return;
      }
      if ((e.key === "Delete" || e.key === "Backspace") && selectedIds.size > 0) {
        e.preventDefault();
        const first = filtered.find((s) => selectedIds.has(s.id));
        if (first) {
          setDeleteTarget(first);
          setShowDeleteDialog(true);
        }
      }
      if (e.key === "Escape") {
        setShowAddDialog(false);
        setShowDeleteDialog(false);
        setContextMenuSkill(null);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [selectedIds, filtered]);

  const selectedCount = selectedIds.size;

  return (
    <div className="flex h-full min-h-0 flex-col">
      {/* Header */}
      <div className="flex shrink-0 items-center gap-3 px-5 py-3">
        <h1 className="text-lg font-semibold">{t("skills.title")}</h1>
        <div className="relative ml-2 flex-1 max-w-md">
          <Search className="absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            ref={searchRef}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t("skills.search")}
            className="pl-9"
          />
          {search && (
            <button
              type="button"
              onClick={() => setSearch("")}
              className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
            >
              <X className="h-4 w-4" />
            </button>
          )}
        </div>

        <div className="ml-auto flex items-center gap-2">
          {selectedCount > 0 && (
            <>
              <Badge variant="secondary">{selectedCount} selected</Badge>
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  if (detail) {
                    setDeleteTarget({
                      id: detail.record.id,
                      name: detail.record.name,
                      slug: detail.record.slug,
                      version: detail.record.version,
                      installed: detail.record.installed,
                      updatedAt: detail.record.updatedAt,
                    } as Summary);
                  }
                  setShowDeleteDialog(true);
                }}
              >
                <Trash2 className="mr-1 h-4 w-4" />
                Delete
              </Button>
            </>
          )}
          <Button size="sm" onClick={() => setShowAddDialog(true)}>
            <Plus className="mr-1 h-4 w-4" />
            Add
          </Button>
        </div>
      </div>

      {/* Main content: list + detail */}
      <div className="min-h-0 flex-1">
        {skills.length === 0 && !search ? (
          <div className="flex h-full flex-col items-center justify-center gap-4 text-center">
            <div className="rounded-full bg-muted p-4">
              <Package className="h-10 w-10 text-muted-foreground" />
            </div>
            <div>
              <h2 className="text-lg font-medium">{t("skills.empty.title")}</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                {t("skills.empty.desc")}
              </p>
            </div>
            <Button onClick={() => setShowAddDialog(true)}>
              <Plus className="mr-1 h-4 w-4" />
              {t("skills.empty.add")}
            </Button>
          </div>
        ) : (
          <ResizablePanelGroup orientation="horizontal" className="h-full">
            <ResizablePanel defaultSize={30} minSize={20} maxSize={50}>
              <div className="flex h-full min-h-0 flex-col border-r border-border/60">
                <div ref={parentRef} className="min-h-0 flex-1 overflow-auto">
                  <div
                    style={{
                      height: `${virtualizer.getTotalSize()}px`,
                      position: "relative",
                    }}
                  >
                    {virtualizer.getVirtualItems().map((vi) => {
                      const item = flatItems[vi.index];
                      if (!item) return null;

                      if (item.type === "group-header") {
                        const isCollapsed = collapsedGroups.has(item.group);
                        return (
                          <div
                            key={`group-${item.group}`}
                            style={{
                              position: "absolute",
                              top: 0,
                              left: 0,
                              right: 0,
                              transform: `translateY(${vi.start}px)`,
                            }}
                            className="flex h-9 items-center gap-2 border-b border-border/40 bg-muted/30 px-3"
                          >
                            <button
                              type="button"
                              onClick={() => toggleGroup(item.group)}
                              className="rounded p-0.5 text-muted-foreground hover:text-foreground"
                            >
                              {isCollapsed ? (
                                <ChevronRight className="h-4 w-4" />
                              ) : (
                                <ChevronDown className="h-4 w-4" />
                              )}
                            </button>
                            <span className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                              {item.group}
                            </span>
                            <span className="ml-auto text-xs text-muted-foreground">
                              {item.skills.length}
                            </span>
                          </div>
                        );
                      }

                      const skill = item.skill;
                      const isSelected = selectedIds.has(skill.id);
                      return (
                        <div
                          key={`skill-${skill.id}`}
                          style={{
                            position: "absolute",
                            top: 0,
                            left: 0,
                            right: 0,
                            transform: `translateY(${vi.start}px)`,
                          }}
                          className={cn(
                            "flex h-11 items-center gap-2 border-b border-border/40 px-3 text-sm transition-colors cursor-pointer",
                            isSelected
                              ? "bg-primary/10 text-foreground"
                              : "hover:bg-muted/50 text-muted-foreground",
                          )}
                          onClick={(e) => handleSelect(skill, e)}
                          onContextMenu={(e) => {
                            e.preventDefault();
                            setContextMenuSkill(skill);
                          }}
                        >
                          <Layers className="h-4 w-4 shrink-0" />
                          <div className="min-w-0 flex-1">
                            <div className="truncate font-medium text-foreground">
                              {skill.name}
                            </div>
                            <div className="truncate text-xs text-muted-foreground">
                              {skill.slug}
                            </div>
                          </div>
                          <Badge variant="outline" className="shrink-0 text-xs">
                            {skill.version}
                          </Badge>
                          {skill.installed && (
                            <Badge
                              variant="secondary"
                              className="shrink-0 text-xs"
                            >
                              installed
                            </Badge>
                          )}
                          <DropdownMenu open={contextMenuSkill?.id === skill.id} onOpenChange={(open) => { if (!open) setContextMenuSkill(null); }}>
                            <DropdownMenuContent side="bottom" align="start">
                              <DropdownMenuItem
                                onClick={() => {
                                  setSkillsToInstall([skill]);
                                  setShowInstallDialog(true);
                                  setContextMenuSkill(null);
                                }}
                              >
                                <Download className="mr-2 h-4 w-4" />
                                {t("skills.context.install")}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => {
                                  setDeleteTarget(skill);
                                  setShowDeleteDialog(true);
                                  setContextMenuSkill(null);
                                }}
                              >
                                <Trash2 className="mr-2 h-4 w-4 text-destructive" />
                                {t("skills.context.delete")}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => {
                                  navigator.clipboard.writeText(skill.slug);
                                  toast("Copied slug");
                                  setContextMenuSkill(null);
                                }}
                              >
                                <FileText className="mr-2 h-4 w-4" />
                                {t("skills.context.copyPath")}
                              </DropdownMenuItem>
                              <DropdownMenuItem
                                onClick={() => {
                                  navigator.clipboard.writeText(skill.name);
                                  toast("Copied name");
                                  setContextMenuSkill(null);
                                }}
                              >
                                <FileText className="mr-2 h-4 w-4" />
                                {t("skills.context.copyName")}
                              </DropdownMenuItem>
                            </DropdownMenuContent>
                          </DropdownMenu>
                        </div>
                      );
                    })}
                  </div>
                </div>
              </div>
            </ResizablePanel>

            <ResizableHandle withHandle />

            <ResizablePanel defaultSize={70} minSize={40} maxSize={80}>
              <div className="flex h-full min-h-0 flex-col">
                {!detail ? (
                  <div className="flex h-full flex-col items-center justify-center gap-3 text-center">
                    <FileText className="h-12 w-12 text-muted-foreground/50" />
                    <p className="text-sm text-muted-foreground">
                      {t("skills.detail.select")}
                    </p>
                  </div>
                ) : loadingDetail ? (
                  <div className="flex h-full items-center justify-center">
                    <div className="h-6 w-6 animate-spin rounded-full border-2 border-muted border-t-primary" />
                  </div>
                ) : (
                  <>
                    {/* Detail header */}
                    <div className="flex shrink-0 items-center gap-3 border-b border-border/60 px-5 py-3">
                      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                        <Layers className="h-5 w-5 text-primary" />
                      </div>
                      <div className="min-w-0 flex-1">
                        <h2 className="truncate text-base font-semibold">
                          {detail.record.name}
                        </h2>
                        <div className="flex items-center gap-2 text-xs text-muted-foreground">
                          <span>{detail.record.slug}</span>
                          <span>·</span>
                          <span>v{detail.record.version}</span>
                          {detail.record.origin?.type && (
                            <>
                              <span>·</span>
                              <span>{detail.record.origin.type}</span>
                            </>
                          )}
                        </div>
                      </div>
                      {detail.record.installed && (
                        <Badge variant="secondary">installed</Badge>
                      )}
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="sm">
                            <span className="sr-only">Actions</span>
                            <MoreVertical className="h-4 w-4" />
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem
                            onClick={() => {
                              setDeleteTarget({
                                id: detail.record.id,
                                name: detail.record.name,
                                slug: detail.record.slug,
                                version: detail.record.version,
                                installed: detail.record.installed,
                                updatedAt: detail.record.updatedAt,
                              } as Summary);
                              setShowDeleteDialog(true);
                            }}
                          >
                            <Trash2 className="mr-2 h-4 w-4 text-destructive" />
                            Delete
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </div>

                    {/* Tabs */}
                    <div className="flex min-h-0 flex-1 flex-col">
                      <div className="shrink-0 px-5 pt-3">
                        <Tabs value={activeTab} onValueChange={setActiveTab}>
                          <TabsList className="h-8">
                            <TabsTrigger value="description" className="gap-1.5 px-3 text-xs">
                              <FileText className="h-3.5 w-3.5" />
                              Description
                            </TabsTrigger>
                            <TabsTrigger value="install" className="gap-1.5 px-3 text-xs">
                              <Download className="h-3.5 w-3.5" />
                              Install
                            </TabsTrigger>
                            <TabsTrigger value="files" className="gap-1.5 px-3 text-xs">
                              <Folder className="h-3.5 w-3.5" />
                              Files
                            </TabsTrigger>
                          </TabsList>
                        </Tabs>
                      </div>

                      <div className="min-h-0 flex-1 overflow-y-auto px-5 py-3">
                        <Tabs value={activeTab} onValueChange={setActiveTab}>
                          <TabsContent value="description" className="mt-0">
                            {detail.readme ? (
                              <pre className="whitespace-pre-wrap break-words rounded-lg bg-muted/50 p-4 font-mono text-xs leading-relaxed">
                                {detail.readme}
                              </pre>
                            ) : (
                              <p className="text-sm text-muted-foreground">
                                No description available.
                              </p>
                            )}
                          </TabsContent>

                          <TabsContent value="install" className="mt-0">
                            <div className="space-y-4">
                              <div>
                                <h3 className="text-sm font-medium">
                                  Install Targets
                                </h3>
                                <p className="text-xs text-muted-foreground">
                                  Select where to install this skill.
                                </p>
                              </div>
                              <div className="grid gap-2">
                                {[
                                  { label: "Shared (.agents/skills)", value: "shared" },
                                  { label: "Claude (.claude/skills)", value: "claude" },
                                  { label: "Codex (.codex/skills)", value: "codex" },
                                  { label: "Cursor (.cursor/skills)", value: "cursor" },
                                ].map((target) => (
                                  <div
                                    key={target.value}
                                    className="flex items-center gap-3 rounded-lg border border-border/60 p-3"
                                  >
                                    <Checkbox id={target.value} />
                                    <label
                                      htmlFor={target.value}
                                      className="cursor-pointer text-sm"
                                    >
                                      {target.label}
                                    </label>
                                  </div>
                                ))}
                              </div>
                              <Button disabled className="w-full">
                                <Download className="mr-2 h-4 w-4" />
                                Install
                              </Button>
                            </div>
                          </TabsContent>

                          <TabsContent value="files" className="mt-0">
                            {detail.files.length > 0 ? (
                              <div className="rounded-lg border border-border/60">
                                {detail.files.map((file, i) => (
                                  <div
                                    key={i}
                                    className={cn(
                                      "flex items-center gap-2 px-3 py-2 text-sm",
                                      i > 0 && "border-t border-border/40",
                                    )}
                                  >
                                    <FileText className="h-4 w-4 shrink-0 text-muted-foreground" />
                                    <span className="truncate font-mono text-xs">
                                      {file}
                                    </span>
                                  </div>
                                ))}
                              </div>
                            ) : (
                              <p className="text-sm text-muted-foreground">
                                No files found.
                              </p>
                            )}
                          </TabsContent>
                        </Tabs>
                      </div>
                    </div>
                  </>
                )}
              </div>
            </ResizablePanel>
          </ResizablePanelGroup>
        )}
      </div>

      {/* Add skill dialog */}
      <Dialog open={showAddDialog} onOpenChange={setShowAddDialog}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>Add Skill</DialogTitle>
            <DialogDescription>Import a skill from a source.</DialogDescription>
          </DialogHeader>

          <div className="space-y-3 py-2">
            <div>
              <label className="mb-1 block text-xs font-medium text-muted-foreground">
                Source Type
              </label>
              <div className="grid grid-cols-4 gap-2">
                {[
                  { label: "Local", value: "local", icon: Folder },
                  { label: "Npx", value: "npx", icon: Package },
                  { label: "Claude", value: "claude", icon: Layers },
                  { label: "Existing", value: "existing", icon: FileText },
                ].map((kind) => (
                  <button
                    key={kind.value}
                    type="button"
                    onClick={() => setAddKind(kind.value)}
                    className={cn(
                      "flex flex-col items-center gap-1 rounded-lg border p-2 text-xs transition-colors",
                      addKind === kind.value
                        ? "border-primary bg-primary/10 text-primary"
                        : "border-border hover:bg-muted/50 text-muted-foreground",
                    )}
                  >
                    <kind.icon className="h-4 w-4" />
                    {kind.label}
                  </button>
                ))}
              </div>
            </div>

            <div>
              <label className="mb-1 block text-xs font-medium text-muted-foreground">
                {addKind === "local" && "Directory path"}
                {addKind === "npx" && "Package name (e.g. @owner/package)"}
                {addKind === "claude" && "Claude plugin name"}
                {addKind === "existing" && "Comma-separated paths"}
              </label>
              <Input
                value={addInput}
                onChange={(e) => setAddInput(e.target.value)}
                placeholder={
                  addKind === "local"
                    ? "/path/to/skills"
                    : addKind === "npx"
                      ? "@scope/package"
                      : addKind === "claude"
                        ? "plugin-name"
                        : "/path/to/skill1,/path/to/skill2"
                }
              />
            </div>

            <div>
              <label className="mb-1 block text-xs font-medium text-muted-foreground">
                Group (optional)
              </label>
              <Input
                value={addGroup}
                onChange={(e) => setAddGroup(e.target.value)}
                placeholder="My Group"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowAddDialog(false)}>
              Cancel
            </Button>
            <Button
              onClick={handleAdd}
              disabled={adding || !addInput.trim()}
            >
              {adding ? "Adding…" : "Add Skill"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation dialog */}
      <Dialog open={showDeleteDialog} onOpenChange={setShowDeleteDialog}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete Skill</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <strong>{deleteTarget?.name || "this skill"}</strong>? This action
              cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setShowDeleteDialog(false)}
            >
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleting || !deleteTarget?.name}
            >
              {deleting ? "Deleting…" : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

        <SkillInstallDialog
          open={showInstallDialog}
          skills={skillsToInstall}
          onConfirm={() => {
            setShowInstallDialog(false);
            setSkillsToInstall([]);
          }}
          onCancel={() => {
            setShowInstallDialog(false);
            setSkillsToInstall([]);
          }}
        />
    </div>
  );
}
