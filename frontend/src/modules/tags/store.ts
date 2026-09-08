import { create } from "zustand";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

/** A single tag node in the two-level tree. */
export interface TagNode {
  id: string;
  name: string;
  /** Parent tag ID, or null for root-level (parent) tags. */
  parentId: string | null;
  /** IDs of skills assigned to this tag. */
  skillIds: string[];
}

/** A named group of skills for batch installation. */
export interface GroupNode {
  id: string;
  name: string;
  /** IDs of skills in this group. */
  skillIds: string[];
}

// ---------------------------------------------------------------------------
// Persistence
// ---------------------------------------------------------------------------

const STORAGE_KEY = "ai-manager-tags-v1";

interface PersistedState {
  tags: TagNode[];
  groups: GroupNode[];
}

function readStore(): PersistedState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) return JSON.parse(raw) as PersistedState;
  } catch {
    /* ignore */
  }
  return { tags: [], groups: [] };
}

function writeStore(state: PersistedState): void {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
  } catch {
    /* ignore quota / privacy errors */
  }
}

// ---------------------------------------------------------------------------
// ID generation
// ---------------------------------------------------------------------------

function slugify(name: string): string {
  return name
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "") || "tag";
}

function uid(prefix: string): string {
  return `${prefix}-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`;
}

// ---------------------------------------------------------------------------
// Store
// ---------------------------------------------------------------------------

interface TagsStore {
  tags: TagNode[];
  groups: GroupNode[];
  hydrated: boolean;

  // Tag operations
  init: () => void;
  addTag: (name: string, parentId?: string | null) => TagNode;
  addParentTag: (name: string) => TagNode;
  addChildTag: (parentId: string, name: string) => TagNode;
  renameTag: (id: string, newName: string) => void;
  deleteTag: (id: string) => void;
  assignTag: (skillId: string, tagId: string) => void;
  unassignTag: (skillId: string, tagId: string) => void;
  getTagsBySkill: (skillId: string) => TagNode[];
  getSkillsByTag: (tagId: string) => string[];
  getChildren: (parentId: string) => TagNode[];
  getParents: () => TagNode[];

  // Group operations
  addGroup: (name: string) => GroupNode;
  renameGroup: (id: string, newName: string) => void;
  deleteGroup: (id: string) => void;
  moveSkillToGroup: (skillId: string, groupId: string) => void;
  moveSkillOutOfGroup: (skillId: string, groupId: string) => void;

  // Internal
  _persist: () => void;
}

let flushTimer: ReturnType<typeof setTimeout> | null = null;
const FLUSH_DELAY_MS = 300;

export const useTagsStore = create<TagsStore>((set, get) => ({
  tags: readStore().tags,
  groups: readStore().groups,
  hydrated: false,

  init: () => {
    // Already loaded from localStorage during construction — just mark hydrated.
    set({ hydrated: true });
  },

  addTag: (name, parentId = null) => {
    const trimmed = name.trim();
    if (!trimmed) return get().addTag("untagged");
    const existing = get().tags.find(
      (t) => t.name.toLowerCase() === trimmed.toLowerCase() && t.parentId === parentId,
    );
    if (existing) return existing;

    const tag: TagNode = {
      id: uid(slugify(trimmed)),
      name: trimmed,
      parentId,
      skillIds: [],
    };
    set((s) => ({ tags: [...s.tags, tag] }));
    get()._persist();
    return tag;
  },

  addParentTag: (name) => get().addTag(name, null),

  addChildTag: (parentId, name) => {
    const parent = get().tags.find((t) => t.id === parentId);
    if (!parent) return { id: "", name: name.trim(), parentId, skillIds: [] } as TagNode;
    return get().addTag(name, parentId);
  },

  renameTag: (id, newName) => {
    const trimmed = newName.trim();
    if (!trimmed) return;
    set((s) => ({
      tags: s.tags.map((t) => (t.id === id ? { ...t, name: trimmed } : t)),
    }));
    get()._persist();
  },

  deleteTag: (id) => {
    // Remove the tag AND all its children (if it's a parent tag).
    const toRemove = new Set([id]);
    // Find children
    for (const t of get().tags) {
      if (t.parentId === id) toRemove.add(t.id);
    }
    // Also unassign from skills
    set((s) => ({
      tags: s.tags.filter((t) => !toRemove.has(t.id)),
    }));
    get()._persist();
  },

  assignTag: (skillId, tagId) => {
    set((s) => ({
      tags: s.tags.map((t) => {
        if (t.id !== tagId) return t;
        if (t.skillIds.includes(skillId)) return t;
        return { ...t, skillIds: [...t.skillIds, skillId] };
      }),
    }));
    get()._persist();
  },

  unassignTag: (skillId, tagId) => {
    set((s) => ({
      tags: s.tags.map((t) => {
        if (t.id !== tagId) return t;
        return { ...t, skillIds: t.skillIds.filter((id) => id !== skillId) };
      }),
    }));
    get()._persist();
  },

  getTagsBySkill: (skillId) =>
    get().tags.filter((t) => t.skillIds.includes(skillId)),

  getSkillsByTag: (tagId) => {
    const tag = get().tags.find((t) => t.id === tagId);
    return tag ? tag.skillIds : [];
  },

  getChildren: (parentId) => get().tags.filter((t) => t.parentId === parentId),

  getParents: () => get().tags.filter((t) => t.parentId === null),

  addGroup: (name) => {
    const trimmed = name.trim();
    if (!trimmed) return { id: "", name, skillIds: [] } as GroupNode;
    const existing = get().groups.find((g) => g.name.toLowerCase() === trimmed.toLowerCase());
    if (existing) return existing;

    const group: GroupNode = {
      id: uid("group-" + slugify(trimmed)),
      name: trimmed,
      skillIds: [],
    };
    set((s) => ({ groups: [...s.groups, group] }));
    get()._persist();
    return group;
  },

  renameGroup: (id, newName) => {
    const trimmed = newName.trim();
    if (!trimmed) return;
    set((s) => ({
      groups: s.groups.map((g) => (g.id === id ? { ...g, name: trimmed } : g)),
    }));
    get()._persist();
  },

  deleteGroup: (id) => {
    set((s) => ({ groups: s.groups.filter((g) => g.id !== id) }));
    get()._persist();
  },

  moveSkillToGroup: (skillId, groupId) => {
    set((s) => ({
      groups: s.groups.map((g) => {
        if (g.id !== groupId) {
          // Remove from other groups
          return { ...g, skillIds: g.skillIds.filter((id) => id !== skillId) };
        }
        if (g.skillIds.includes(skillId)) return g;
        return { ...g, skillIds: [...g.skillIds, skillId] };
      }),
    }));
    get()._persist();
  },

  moveSkillOutOfGroup: (skillId, groupId) => {
    set((s) => ({
      groups: s.groups.map((g) => {
        if (g.id !== groupId) return g;
        return { ...g, skillIds: g.skillIds.filter((id) => id !== skillId) };
      }),
    }));
    get()._persist();
  },

  _persist: () => {
    if (flushTimer) clearTimeout(flushTimer);
    flushTimer = setTimeout(() => {
      flushTimer = null;
      writeStore({ tags: get().tags, groups: get().groups });
    }, FLUSH_DELAY_MS);
  },
}));
