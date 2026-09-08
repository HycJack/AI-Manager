import { useState, useRef, useCallback } from "react";
import {
  DndContext,
  DragOverlay,
  useDraggable,
  useDroppable,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from "@dnd-kit/core";
import {
  FolderTree,
  Folder,
  FolderOpen,
  Hash,
  Plus,
  Trash2,
  GripVertical,
  Package,
  PackageOpen,
  ChevronRight,
  ChevronDown,
  X,
  Check,
  Tag,
  Layers,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
  TooltipProvider,
} from "@/components/ui/tooltip";
import {
  useTagsStore,
  type TagNode,
  type GroupNode,
} from "@/modules/tags/store";
import { cn } from "@/lib/utils";
import { toast } from "@/lib/toast";

// ---------------------------------------------------------------------------
// Inline editable label
// ---------------------------------------------------------------------------

function InlineEdit({
  value,
  onSave,
  size = "md",
  className,
}: {
  value: string;
  onSave: (v: string) => void;
  size?: "sm" | "md";
  className?: string;
}) {
  const [editing, setEditing] = useState(false);
  const [draft, setDraft] = useState(value);
  const inputRef = useRef<HTMLInputElement>(null);

  const start = () => {
    setDraft(value);
    setEditing(true);
  };

  const commit = () => {
    const v = draft.trim();
    setEditing(false);
    if (v && v !== value) onSave(v);
  };

  if (!editing) {
    return (
      <span
        className={cn(
          "inline-flex cursor-text select-none truncate rounded px-1 hover:bg-accent/40",
          size === "sm" ? "text-xs" : "text-sm",
          className,
        )}
        onDoubleClick={start}
      >
        {value}
      </span>
    );
  }

  return (
    <Input
      ref={inputRef}
      value={draft}
      onChange={(e) => setDraft(e.target.value)}
      onBlur={commit}
      onKeyDown={(e) => {
        if (e.key === "Enter") {
          e.preventDefault();
          commit();
        } else if (e.key === "Escape") {
          setEditing(false);
        }
      }}
      className={cn("h-6 px-1.5 text-xs", className)}
      autoFocus
    />
  );
}

// ---------------------------------------------------------------------------
// Confirm delete (two-click)
// ---------------------------------------------------------------------------

function ConfirmDelete({
  label,
  onConfirm,
}: {
  label: string;
  onConfirm: () => void;
}) {
  const [confirming, setConfirming] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>(0);

  const handleClick = () => {
    if (!confirming) {
      setConfirming(true);
      timerRef.current = setTimeout(() => setConfirming(false), 2500);
    } else {
      clearTimeout(timerRef.current);
      setConfirming(false);
      onConfirm();
    }
  };

  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <Button
          variant="ghost"
          size="icon"
          className={cn(
            "h-6 w-6 rounded",
            confirming && "bg-destructive/20 text-destructive animate-pulse",
          )}
          onClick={handleClick}
        >
          <Trash2 className="h-3 w-3" />
          <span className="sr-only">Delete {label}</span>
        </Button>
      </TooltipTrigger>
      <TooltipContent side="top">
        {confirming ? "Click again to confirm" : `Delete ${label}`}
      </TooltipContent>
    </Tooltip>
  );
}

// ---------------------------------------------------------------------------
// Draggable skill chip (uses dnd-kit useDraggable)
// ---------------------------------------------------------------------------

function SkillChip({
  skillId,
  skillName,
}: {
  skillId: string;
  skillName: string;
}) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `skill:${skillId}`,
    data: { type: "skill", skillId, skillName },
  });

  return (
    <div
      ref={setNodeRef}
      {...listeners}
      {...attributes}
      className={cn(
        "flex cursor-grab items-center gap-1 rounded-md bg-secondary/60 px-2 py-1 text-xs active:cursor-grabbing",
        isDragging && "opacity-40",
      )}
    >
      <GripVertical className="h-3 w-3 text-muted-foreground/50" />
      <span className="truncate">{skillName}</span>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Tag row (droppable via useDroppable)
// ---------------------------------------------------------------------------

function TagRow({
  tag,
  depth = 0,
  children,
}: {
  tag: TagNode;
  depth?: number;
  children?: React.ReactNode;
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: `tag:${tag.id}`,
    data: { type: "tag", tagId: tag.id },
  });

  const { assignTag, deleteTag, renameTag, addChildTag } = useTagsStore();

  const [expanded, setExpanded] = useState(true);
  const [addingChild, setAddingChild] = useState(false);
  const [childName, setChildName] = useState("");

  const isParent = tag.parentId === null;

  return (
    <div className="select-none">
      <div
        ref={setNodeRef}
        className={cn(
          "group flex items-center gap-1.5 rounded-md px-1.5 py-1 transition-colors",
          isOver
            ? "bg-primary/15 ring-1 ring-primary/40"
            : "hover:bg-accent/30",
        )}
        style={{ paddingLeft: `${depth * 16 + 6}px` }}
      >
        {isParent ? (
          <button
            className="flex h-4 w-4 items-center justify-center rounded text-muted-foreground hover:text-foreground"
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronRight className="h-3 w-3" />
            )}
          </button>
        ) : (
          <span className="w-4" />
        )}

        {isParent ? (
          <FolderOpen className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        ) : (
          <Folder className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        )}

        <InlineEdit
          value={tag.name}
          onSave={(v) => renameTag(tag.id, v)}
          className="min-w-0 flex-1"
          size="sm"
        />

        {tag.skillIds.length > 0 && (
          <Badge variant="secondary" className="h-4 px-1 text-[10px] leading-none">
            {tag.skillIds.length}
          </Badge>
        )}

        <div className="flex items-center gap-0.5 opacity-0 transition-opacity group-hover:opacity-100">
          {isParent && (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 rounded"
              onClick={() => setAddingChild(true)}
            >
              <Plus className="h-3 w-3" />
              <span className="sr-only">Add child</span>
            </Button>
          )}
          <ConfirmDelete
            label={tag.name}
            onConfirm={() => {
              deleteTag(tag.id);
              toast(`Deleted "${tag.name}"`);
            }}
          />
        </div>
      </div>

      {addingChild && (
        <div
          className="flex items-center gap-1"
          style={{ paddingLeft: `${depth * 16 + 36}px` }}
        >
          <Input
            value={childName}
            onChange={(e) => setChildName(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter" && childName.trim()) {
                addChildTag(tag.id, childName);
                setChildName("");
                setAddingChild(false);
              } else if (e.key === "Escape") {
                setAddingChild(false);
                setChildName("");
              }
            }}
            placeholder="Child tag name…"
            className="h-6 flex-1 px-1.5 text-xs"
            autoFocus
          />
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={() => {
              if (childName.trim()) {
                addChildTag(tag.id, childName);
                setChildName("");
                setAddingChild(false);
              }
            }}
          >
            <Check className="h-3 w-3" />
          </Button>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={() => {
              setAddingChild(false);
              setChildName("");
            }}
          >
            <X className="h-3 w-3" />
          </Button>
        </div>
      )}

      {isParent && expanded && children}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Group row (droppable via useDroppable)
// ---------------------------------------------------------------------------

function GroupRow({
  group,
  skills,
  onDelete,
  onRename,
}: {
  group: GroupNode;
  skills: { id: string; name: string }[];
  onDelete: () => void;
  onRename: (v: string) => void;
}) {
  const { setNodeRef, isOver } = useDroppable({
    id: `group:${group.id}`,
    data: { type: "group", groupId: group.id },
  });

  const { moveSkillToGroup, moveSkillOutOfGroup } = useTagsStore();

  return (
    <div
      ref={setNodeRef}
      className={cn(
        "rounded-lg border border-border p-2 transition-colors",
        isOver && "border-primary/40 bg-primary/5",
      )}
    >
      <div className="flex items-center gap-1.5">
        <PackageOpen className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
        <InlineEdit
          value={group.name}
          onSave={onRename}
          className="min-w-0 flex-1"
          size="sm"
        />
        <Badge variant="secondary" className="h-4 px-1 text-[10px] leading-none">
          {group.skillIds.length}
        </Badge>
        <ConfirmDelete label={group.name} onConfirm={onDelete} />
      </div>

      {skills.length > 0 && (
        <div className="mt-1.5 flex flex-wrap gap-1">
          {skills.map((skill) => (
            <span
              key={skill.id}
              className="flex items-center gap-0.5 rounded bg-secondary/50 px-1.5 py-0.5 text-[10px]"
            >
              {skill.name}
              <button
                className="ml-0.5 text-muted-foreground hover:text-destructive"
                onClick={() => {
                  moveSkillOutOfGroup(skill.id, group.id);
                }}
              >
                <X className="h-2.5 w-2.5" />
              </button>
            </span>
          ))}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export function TagsPanel() {
  const {
    tags,
    groups,
    addParentTag,
    addGroup,
    renameGroup,
    deleteGroup,
    moveSkillToGroup,
    assignTag,
    getParents,
    getChildren,
  } = useTagsStore();

  // Inline add-parent form
  const [addingParent, setAddingParent] = useState(false);
  const [parentName, setParentName] = useState("");

  // Inline add-group form
  const [addingGroup, setAddingGroup] = useState(false);
  const [groupName, setGroupName] = useState("");

  // DnD state
  const [activeSkill, setActiveSkill] = useState<{
    id: string;
    name: string;
  } | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 4 } }),
  );

  const handleDragStart = (event: DragStartEvent) => {
    const data = event.active.data.current;
    if (data?.type === "skill") {
      setActiveSkill({ id: data.skillId, name: data.skillName });
    }
  };

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveSkill(null);

    const { active, over } = event;
    if (!over) return;

    const dragData = active.data.current;
    const dropData = over.data.current;

    if (dragData?.type === "skill" && dropData?.type === "tag") {
      assignTag(dragData.skillId, dropData.tagId);
      const tagName = tags.find((t) => t.id === dropData.tagId)?.name ?? "tag";
      toast(`Assigned "${dragData.skillName}" to "${tagName}"`);
    } else if (
      dragData?.type === "skill" &&
      dropData?.type === "group"
    ) {
      moveSkillToGroup(dragData.skillId, dropData.groupId);
      const groupName = groups.find((g) => g.id === dropData.groupId)?.name ?? "group";
      toast(`Added "${dragData.skillName}" to group "${groupName}"`);
    }
  };

  // Build tag tree
  const parents = getParents();
  const parentsWithChildren = parents.map((p) => ({
    tag: p,
    children: getChildren(p.id),
  }));

  // Demo skills (placeholder — would come from the skill library)
  const demoSkills = [
    { id: "skill-react", name: "React Patterns" },
    { id: "skill-vue", name: "Vue Composition" },
    { id: "skill-go", name: "Go Concurrency" },
    { id: "skill-python", name: "Python Async" },
    { id: "skill-svelte", name: "SvelteKit Guide" },
    { id: "skill-tailwind", name: "Tailwind v4" },
    { id: "skill-astro", name: "Astro Components" },
  ];

  return (
    <TooltipProvider>
      <DndContext
        sensors={sensors}
        onDragStart={handleDragStart}
        onDragEnd={handleDragEnd}
      >
        <div className="flex h-full gap-4">
          {/* ── Tags Tree ────────────────────────────────────────── */}
          <div className="flex w-72 flex-shrink-0 flex-col rounded-xl border border-border bg-card">
            <div className="flex items-center justify-between border-b border-border px-3 py-2.5">
              <div className="flex items-center gap-2">
                <Tag className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-medium">Tags</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6"
                onClick={() => setAddingParent(true)}
              >
                <Plus className="h-3.5 w-3.5" />
                <span className="sr-only">Add parent tag</span>
              </Button>
            </div>

            <div className="flex-1 overflow-y-auto p-2">
              {addingParent && (
                <div className="mb-2 flex items-center gap-1">
                  <Input
                    value={parentName}
                    onChange={(e) => setParentName(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && parentName.trim()) {
                        addParentTag(parentName);
                        setParentName("");
                        setAddingParent(false);
                      } else if (e.key === "Escape") {
                        setAddingParent(false);
                        setParentName("");
                      }
                    }}
                    placeholder="Parent tag name…"
                    className="h-7 flex-1 px-2 text-xs"
                    autoFocus
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => {
                      if (parentName.trim()) {
                        addParentTag(parentName);
                        setParentName("");
                        setAddingParent(false);
                      }
                    }}
                  >
                    <Check className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => {
                      setAddingParent(false);
                      setParentName("");
                    }}
                  >
                    <X className="h-3.5 w-3.5" />
                  </Button>
                </div>
              )}

              {parentsWithChildren.length === 0 && !addingParent ? (
                <div className="flex flex-col items-center justify-center py-8 text-center">
                  <FolderTree className="mb-2 h-8 w-8 text-muted-foreground/40" />
                  <p className="text-xs text-muted-foreground">
                    No tags yet. Add a parent tag to get started.
                  </p>
                </div>
              ) : (
                <div className="space-y-0.5">
                  {parentsWithChildren.map(({ tag, children }) => (
                    <TagRow
                      key={tag.id}
                      tag={tag}
                      depth={0}
                      children={
                        children.length > 0 ? (
                          <div className="space-y-0.5">
                            {children.map((child) => (
                              <TagRow key={child.id} tag={child} depth={1} />
                            ))}
                          </div>
                        ) : undefined
                      }
                    />
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* ── Groups Panel ─────────────────────────────────────── */}
          <div className="flex w-72 flex-shrink-0 flex-col rounded-xl border border-border bg-card">
            <div className="flex items-center justify-between border-b border-border px-3 py-2.5">
              <div className="flex items-center gap-2">
                <Layers className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-medium">Groups</span>
              </div>
              <Button
                variant="ghost"
                size="icon"
                className="h-6 w-6"
                onClick={() => setAddingGroup(true)}
              >
                <Plus className="h-3.5 w-3.5" />
                <span className="sr-only">Add group</span>
              </Button>
            </div>

            <div className="flex-1 overflow-y-auto p-2">
              {addingGroup && (
                <div className="mb-2 flex items-center gap-1">
                  <Input
                    value={groupName}
                    onChange={(e) => setGroupName(e.target.value)}
                    onKeyDown={(e) => {
                      if (e.key === "Enter" && groupName.trim()) {
                        addGroup(groupName);
                        setGroupName("");
                        setAddingGroup(false);
                      } else if (e.key === "Escape") {
                        setAddingGroup(false);
                        setGroupName("");
                      }
                    }}
                    placeholder="Group name…"
                    className="h-7 flex-1 px-2 text-xs"
                    autoFocus
                  />
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => {
                      if (groupName.trim()) {
                        addGroup(groupName);
                        setGroupName("");
                        setAddingGroup(false);
                      }
                    }}
                  >
                    <Check className="h-3.5 w-3.5" />
                  </Button>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-7 w-7"
                    onClick={() => {
                      setAddingGroup(false);
                      setGroupName("");
                    }}
                  >
                    <X className="h-3.5 w-3.5" />
                  </Button>
                </div>
              )}

              {groups.length === 0 && !addingGroup ? (
                <div className="flex flex-col items-center justify-center py-8 text-center">
                  <Package className="mb-2 h-8 w-8 text-muted-foreground/40" />
                  <p className="text-xs text-muted-foreground">
                    No groups yet. Create a group to organize skills.
                  </p>
                </div>
              ) : (
                <div className="space-y-1.5">
                  {groups.map((group) => (
                    <GroupRow
                      key={group.id}
                      group={group}
                      skills={demoSkills.filter((s) => group.skillIds.includes(s.id))}
                      onDelete={() => {
                        deleteGroup(group.id);
                        toast(`Deleted group "${group.name}"`);
                      }}
                      onRename={(v) => renameGroup(group.id, v)}
                    />
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* ── Skills Pool ──────────────────────────────────────── */}
          <div className="flex flex-1 flex-col rounded-xl border border-border bg-card">
            <div className="flex items-center justify-between border-b border-border px-3 py-2.5">
              <div className="flex items-center gap-2">
                <Hash className="h-4 w-4 text-muted-foreground" />
                <span className="text-sm font-medium">Skills</span>
              </div>
              <span className="text-xs text-muted-foreground">
                Drag skills to tags or groups
              </span>
            </div>

            <div className="flex-1 overflow-y-auto p-3">
              <div className="flex flex-wrap gap-2">
                {demoSkills.map((skill) => (
                  <SkillChip
                    key={skill.id}
                    skillId={skill.id}
                    skillName={skill.name}
                  />
                ))}
              </div>

              <div className="mt-6 rounded-lg bg-muted/40 p-3">
                <p className="text-xs leading-relaxed text-muted-foreground">
                  <span className="font-medium text-foreground">Tip:</span>{" "}
                  Drag skills onto tags or groups to organize them. Double-click a
                  tag or group name to rename. Click the trash icon twice to
                  delete.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Drag overlay */}
        <DragOverlay>
          {activeSkill ? (
            <div className="rounded-md bg-primary px-2 py-1 text-xs text-primary-foreground shadow-lg">
              {activeSkill.name}
            </div>
          ) : null}
        </DragOverlay>
      </DndContext>
    </TooltipProvider>
  );
}
