import { useEffect, useState, useMemo, useCallback } from "react";
import {
  Search,
  Trash2,
  FolderOpen,
  Package,
  X,
  Loader2,
  RefreshCw,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "@/lib/toast";
import { useI18n } from "@/modules/i18n";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  GetSkillAgentStatus,
  ToggleSkillAgent,
  RemoveSkill,
  OpenSkillDirectory,
  GetAgentConfig,
  GetExternalSkills,
  RegisterExternalSkill,
} from "@bindings/ai-manager/internal/app/skillservice";
import { DiscoverSkillsDialog } from "@/components/DiscoverSkillsDialog";
import type { SkillAgentStatus, ExternalSkill } from "@bindings/ai-manager/internal/app/models";

// ---------------------------------------------------------------------------
// Agent icon config
// ---------------------------------------------------------------------------

interface AgentIconConfig {
  key: string;
  label: string;
  icon: React.ReactNode;
  colorClass: string;
}

// Maps iconType string from backend config to a React component.
function renderAgentIcon(iconType: string): React.ReactNode {
  switch (iconType) {
    case "claude":
      return (
        <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor">
          <path d="M4.709 15.955l4.72-2.647.08-.23-.08-.128H9.2l-.79-.048-2.698-.073-2.339-.097-2.266-.122-.571-.121L0 11.784l.055-.352.48-.321.686.06 1.52.103 2.278.158 1.652.097 2.449.255h.389l.055-.157-.134-.098-.103-.097-2.358-1.596-2.552-1.688-1.336-.972-.724-.491-.364-.462-.158-1.008.656-.722.881.06.225.061.893.686 1.908 1.476 2.491 1.833.365.304.145-.103.019-.073-.164-.274-1.355-2.446-1.446-2.49-.644-1.032-.17-.619a2.97 2.97 0 01-.104-.729L6.283.134 6.696 0l.996.134.42.364.62 1.414 1.002 2.229 1.555 3.03.456.898.243.832.091.255h.158V9.01l.128-1.706.237-2.095.23-2.695.08-.76.376-.91.747-.492.584.28.48.685-.067.444-.286 1.851-.559 2.903-.364 1.942h.212l.243-.242.985-1.306 1.652-2.064.73-.82.85-.904.547-.431h1.033l.76 1.129-.34 1.166-1.064 1.347-.881 1.142-1.264 1.7-.79 1.36.073.11.188-.02 2.856-.606 1.543-.28 1.841-.315.833.388.091.395-.328.807-1.969.486-2.309.462-3.439.813-.042.03.049.061 1.549.146.662.036h1.622l3.02.225.79 .522.474.638-.079.485-1.215.62-1.64-.389-3.829-.91-1.312-.329h-.182v.11l1.093 1.068 2.006 1.81 2.509 2.33.127.578-.322.455-.34-.049-2.205-1.657-.851-.747-1.926-1.62h-.128v.17l.444.649 2.345 3.521.122 1.08-.17.353-.608.213-.668-.122-1.374-1.925-1.415-2.167-1.143-1.943-.14.08-.674 7.254-.316.37-.729.28-.607-.461-.322-.747.322-1.476.389-1.924.315-1.53.286-1.9.17-.632-.012-.042-.14.018-1.434 1.967-2.18 2.945-1.726 1.845-.414.164-.717-.37.067-.662.401-.589 2.388-3.036 1.44-1.882.93-1.086-.006-.158h-.055L4.132 18.56l-1.13.146-.487-.456.061-.746.231-.243 1.908-1.312-.006.006z" fill="#D97757" fill-rule="nonzero"/>
        </svg>
      );
    case "codex":
      return (
        <svg className="h-3.5 w-3.5" viewBox="0 0 24 24" fill="currentColor">
          <path fillRule="evenodd" d="M21.55 10.004a5.416 5.416 0 00-.478-4.501c-1.217-2.09-3.662-3.166-6.05-2.66A5.59 5.59 0 0010.831 1C8.39.995 6.224 2.546 5.473 4.838A5.553 5.553 0 001.76 7.496a5.487 5.487 0 00.691 6.5 5.416 5.416 0 00.477 4.502c1.217 2.09 3.662 3.165 6.05 2.66A5.586 5.586 0 0013.168 23c2.443.006 4.61-1.546 5.361-3.84a5.553 5.553 0 003.715-2.66 5.488 5.488 0 00-.693-6.497v.001zm-8.381 11.558a4.199 4.199 0 01-2.675-.954c.034-.018.093-.05.132-.074l4.44-2.53a.71.71 0 00.364-.623v-6.176l1.877 1.069c.02.01.033.029.036.05v5.115c-.003 2.274-1.87 4.118-4.174 4.123zM4.192 17.78a4.059 4.059 0 01-.498-2.763c.032.02.09.055.131.078l4.44 2.53c.225.13.504.13.73 0l5.42-3.088v2.138a.068.068 0 01-.027.057L9.9 19.288c-1.999 1.136-4.552.46-5.707-1.51h-.001zM3.023 8.216A4.15 4.15 0 015.198 6.41l-.002.151v5.06a.711.711 0 00.364.624l5.42 3.087-1.876 1.07a.067.067 0 01-.063.005l-4.489-2.559c-1.995-1.14-2.679-3.658-1.53-5.63h.001zm15.417 3.54l-5.42-3.088L14.896 7.6a.067.067 0 01.063-.006l4.489 2.557c1.998 1.14 2.683 3.662 1.529 5.633a4.163 4.163 0 01-2.174 1.807V12.38a.71 .71 0 00-.363-.623zm1.867-2.773a6.04 6.04 0 00-.132-.078l-4.44-2.53a.731.731 0 00-.729 0l-5.42 3.088V7.325a.068.068 0 01.027-.057L14.1 4.713c2-1.137 4.555-.46 5.707 1.513.487.833.664 1.809.499 2.757h.001zm-11.741 3.81l-1.877-1.068a.065.065 0 01-.036-.051V6.559c.001-2.277 1.873-4.122 4.181-4.12.976 0 1.92.338 2.671.954-.034.018-.092.05-.131.073l-4.44 2.53a.71.71 0 00-.365.623l-.003 6.173v.002zm1.02-2.168L12 9.25l2.414 1.375v2.75L12 14.75l-2.415-1.375v-2.75z"/>
        </svg>
      );
    case "pi":
      return <span className="text-[10px] font-bold">π</span>;
    case "opencode":
      return (
        <svg className="h-3.5 w-3.5" viewBox="0 0 240 300" fill="currentColor">
          <path d="M180 240H60V120H180V240Z" fill="#CFCECD"/>
          <path d="M180 60H60V240H180V60ZM240 300H0V0H240V300Z" fill="#211E1E"/>
        </svg>
      );
    case "hermes":
      return <span className="text-[9px] font-bold tracking-tighter">H</span>;
    default:
      return <Package className="h-3.5 w-3.5" />;
  }
}

// ---------------------------------------------------------------------------
// Main component
// ---------------------------------------------------------------------------

export default function SkillsPage() {
  const { t } = useI18n();
  const [search, setSearch] = useState("");
  const [skills, setSkills] = useState<SkillAgentStatus[]>([]);
  const [loading, setLoading] = useState(true);
  const [togglingKey, setTogglingKey] = useState<string | null>(null);
  const [showDiscoverDialog, setShowDiscoverDialog] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<SkillAgentStatus | null>(null);
  const [deleting, setDeleting] = useState(false);
  const [agentConfigs, setAgentConfigs] = useState<AgentIconConfig[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<string | null>(null);
  const [externalSkills, setExternalSkills] = useState<ExternalSkill[]>([]);
  const [loadingExternal, setLoadingExternal] = useState(false);
  const [registeringExternal, setRegisteringExternal] = useState<string | null>(null);

  // Fetch agent config on mount
  useEffect(() => {
    GetAgentConfig().then((cfg) => {
      if (cfg && cfg.agents) {
        setAgentConfigs(
          cfg.agents.map((a) => ({
            key: a.key,
            label: a.label,
            icon: renderAgentIcon(a.iconType),
            colorClass: a.colorClass,
          })),
        );
      }
    });
  }, []);

  // Fetch skills on mount
  const loadSkills = useCallback(async () => {
    setLoading(true);
    try {
      const data = await GetSkillAgentStatus();
      setSkills(data);
      // Also load external skills
      setLoadingExternal(true);
      const extData = await GetExternalSkills();
      setExternalSkills(extData);
    } catch (e) {
      toast(String(e), true);
    } finally {
      setLoading(false);
      setLoadingExternal(false);
    }
  }, []);

  useEffect(() => {
    loadSkills();
  }, [loadSkills]);

  // Filter by search and selected agent
  const filtered = useMemo(() => {
    let result = skills;
    if (search.trim()) {
      const q = search.toLowerCase();
      result = result.filter(
        (s) =>
          s.name.toLowerCase().includes(q) ||
          s.slug.toLowerCase().includes(q) ||
          s.version.toLowerCase().includes(q),
      );
    }
    if (selectedAgent) {
      result = result.filter((s) => s.agents?.[selectedAgent]);
    }
    return result;
  }, [skills, search, selectedAgent]);

  // Count per agent
  const agentCounts = useMemo(() => {
    const counts: Record<string, number> = {};
    for (const cfg of agentConfigs) {
      counts[cfg.key] = skills.filter((s) => s.agents?.[cfg.key]).length;
    }
    return counts;
  }, [skills, agentConfigs]);

  // Toggle a skill for an agent — optimistically update local state
  const handleToggle = async (skillName: string, agentKey: string) => {
    const toggleKey = `${skillName}:${agentKey}`;
    setTogglingKey(toggleKey);
    try {
      setSkills((prev) =>
        prev.map((s) => {
          if (s.name !== skillName) return s;
          const agents = { ...s.agents };
          agents[agentKey] = !agents[agentKey];
          return { ...s, agents };
        }),
      );
      await ToggleSkillAgent(skillName, agentKey);
    } catch (e) {
      setSkills((prev) =>
        prev.map((s) => {
          if (s.name !== skillName) return s;
          const agents = { ...s.agents };
          agents[agentKey] = !agents[agentKey];
          return { ...s, agents };
        }),
      );
      toast(String(e), true);
    } finally {
      setTogglingKey(null);
    }
  };

  // Delete skill
  const handleDelete = async () => {
    if (!deleteTarget) return;
    setDeleting(true);
    try {
      await RemoveSkill(deleteTarget.name);
      toast(`Deleted ${deleteTarget.name}`);
      setDeleteTarget(null);
      await loadSkills();
    } catch (e) {
      toast(String(e), true);
    } finally {
      setDeleting(false);
    }
  };

  // Open in folder
  const handleOpen = (skillName: string) => {
    OpenSkillDirectory(skillName).catch((e) => toast(String(e), true));
  };

  // Register an external skill into the library
  const handleRegisterExternal = async (skill: ExternalSkill) => {
    setRegisteringExternal(skill.canonical);
    try {
      await RegisterExternalSkill(skill.canonical);
      toast(`Registered ${skill.name}`);
      await loadSkills();
    } catch (e) {
      toast(String(e), true);
    } finally {
      setRegisteringExternal(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">{t("skills.title")}</h1>
          <p className="text-sm text-muted-foreground mt-1">
            {t("skills.empty.desc")}
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button variant="outline" size="sm" onClick={loadSkills} disabled={loading}>
            {loading ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : (
              <RefreshCw className="h-4 w-4" />
            )}
            <span className="ml-2">{loading ? "Refreshing..." : "Refresh"}</span>
          </Button>
          <Button onClick={() => setShowDiscoverDialog(true)}>
            <Search className="h-4 w-4 mr-2" />
            {t("skills.discover")}
          </Button>
        </div>
      </div>

      {/* Search + Agent filters */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium text-muted-foreground">
            <Search className="h-4 w-4 inline mr-2" />
            {t("skills.search")}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          <div className="flex items-center gap-2">
            <div className="relative flex-1">
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder={t("skills.search")}
                className="pl-3"
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
            {skills.length > 0 && (
              <Badge variant="secondary" className="shrink-0">
                {filtered.length} / {skills.length}
              </Badge>
            )}
          </div>

          {/* Agent filter buttons */}
          {agentConfigs.length > 0 && (
            <div className="flex flex-wrap gap-2">
              {agentConfigs.map((cfg) => {
                const isActive = selectedAgent === cfg.key;
                return (
                  <button
                    key={cfg.key}
                    type="button"
                    onClick={() => setSelectedAgent(isActive ? null : cfg.key)}
                    className={cn(
                      "group flex items-center gap-2 rounded-lg border px-3 py-2 text-sm transition-colors",
                      isActive
                        ? "border-primary bg-primary/10 text-primary"
                        : "border-border hover:bg-secondary"
                    )}
                  >
                    <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md bg-muted/60">
                      {cfg.icon}
                    </span>
                    <span className="font-medium">{cfg.label}</span>
                    <span className="text-xs text-muted-foreground">{agentCounts[cfg.key]}</span>
                  </button>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Skill list */}
      {loading ? (
        <Card>
          <CardContent className="flex h-64 items-center justify-center">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </CardContent>
        </Card>
      ) : skills.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <div className="rounded-full bg-muted p-4 mb-4">
              <Package className="h-10 w-10 text-muted-foreground" />
            </div>
            <h2 className="text-lg font-medium">{t("skills.empty.title")}</h2>
            <p className="mt-1 text-sm text-muted-foreground">{t("skills.empty.desc")}</p>
            <Button className="mt-4" onClick={() => setShowDiscoverDialog(true)}>
              <Search className="mr-1 h-4 w-4" />
              {t("skills.empty.discover")}
            </Button>
          </CardContent>
        </Card>
      ) : filtered.length === 0 ? (
        <Card>
          <CardContent className="flex flex-col items-center justify-center py-12">
            <Search className="h-10 w-10 text-muted-foreground/40 mb-3" />
            <p className="text-sm text-muted-foreground">
              {search.trim()
                ? `No skills matching "${search}"`
                : selectedAgent
                  ? `No skills installed for ${agentConfigs.find((a) => a.key === selectedAgent)?.label}`
                  : "No matching skills"}
            </p>
            {(search || selectedAgent) && (
              <Button
                variant="outline"
                size="sm"
                className="mt-3"
                onClick={() => {
                  setSearch("");
                  setSelectedAgent(null);
                }}
              >
                Clear filters
              </Button>
            )}
          </CardContent>
        </Card>
      ) : (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              <Package className="h-4 w-4 inline mr-2" />
              Skills ({filtered.length})
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3 font-medium text-muted-foreground">Skill</th>
                    <th className="text-left py-2 px-3 font-medium text-muted-foreground">Agents</th>
                    <th className="w-20 py-2 px-3" />
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((skill) => (
                    <tr
                      key={skill.slug}
                      className="group border-b transition-colors hover:bg-secondary/50"
                    >
                      <td className="py-2.5 px-3">
                        <div className="flex items-center gap-2">
                          <span className="font-medium">{skill.name}</span>
                          {skill.hasUpdate && (
                            <Badge variant="outline" className="text-xs gap-1 py-0.5 text-amber-500 border-amber-500/30">
                              <span className="h-1.5 w-1.5 rounded-full bg-amber-500" />
                              Update
                            </Badge>
                          )}
                        </div>
                        {skill.description && (
                          <div className="mt-0.5 text-xs text-muted-foreground/70 truncate max-w-[300px]">
                            {skill.description}
                          </div>
                        )}
                      </td>
                      <td className="py-2.5 px-3">
                        <div className="flex items-center gap-1">
                          {agentConfigs.map((cfg) => {
                            const enabled = skill.agents?.[cfg.key] ?? false;
                            const isPending = togglingKey === `${skill.name}:${cfg.key}`;
                            return (
                              <button
                                key={cfg.key}
                                type="button"
                                onClick={() => handleToggle(skill.name, cfg.key)}
                                disabled={isPending}
                                aria-label={cfg.label}
                                aria-pressed={enabled}
                                title={cfg.label + (enabled ? " ✓" : "")}
                                className={cn(
                                  "flex h-6 w-6 items-center justify-center rounded-md transition-colors",
                                  enabled
                                    ? "bg-primary/10 text-primary ring-1 ring-primary/20"
                                    : "text-muted-foreground/40 hover:bg-muted hover:text-muted-foreground",
                                  "disabled:cursor-not-allowed"
                                )}
                              >
                                {isPending ? (
                                  <Loader2 className="h-3 w-3 animate-spin" />
                                ) : (
                                  cfg.icon
                                )}
                              </button>
                            );
                          })}
                        </div>
                      </td>
                      <td className="py-2.5 px-3">
                        <div className="flex items-center justify-end gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7"
                            onClick={() => handleOpen(skill.name)}
                            title="Open in Folder"
                          >
                            <FolderOpen className="h-3.5 w-3.5" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-7 w-7 text-destructive hover:text-destructive"
                            onClick={() => setDeleteTarget(skill)}
                            title="Delete"
                          >
                            <Trash2 className="h-3.5 w-3.5" />
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* External skills (not in library) */}
      {!loading && externalSkills.length > 0 && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              <Package className="h-4 w-4 inline mr-2" />
              External Skills ({externalSkills.length})
              <span className="ml-2 text-xs font-normal text-muted-foreground/70">
                — installed in agent directories but not managed by AI-Manager
              </span>
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3 font-medium text-muted-foreground">Skill</th>
                    <th className="text-left py-2 px-3 font-medium text-muted-foreground">Agents</th>
                    <th className="w-24 py-2 px-3" />
                  </tr>
                </thead>
                <tbody>
                  {externalSkills.map((skill) => (
                    <tr
                      key={skill.canonical}
                      className="group border-b transition-colors hover:bg-secondary/50"
                    >
                      <td className="py-2.5 px-3">
                        <div className="font-medium">{skill.name}</div>
                        {skill.description && (
                          <div className="mt-0.5 text-xs text-muted-foreground/70 truncate max-w-[300px]">
                            {skill.description}
                          </div>
                        )}
                      </td>
                      <td className="py-2.5 px-3">
                        <div className="flex items-center gap-1">
                          {agentConfigs.map((cfg) => {
                            const enabled = skill.agents?.[cfg.key] ?? false;
                            return (
                              <span
                                key={cfg.key}
                                title={cfg.label + (enabled ? " ✓" : "")}
                                className={cn(
                                  "flex h-6 w-6 items-center justify-center rounded-md",
                                  enabled
                                    ? "bg-primary/10 text-primary ring-1 ring-primary/20"
                                    : "text-muted-foreground/30"
                                )}
                              >
                                {cfg.icon}
                              </span>
                            );
                          })}
                        </div>
                      </td>
                      <td className="py-2.5 px-3">
                        <Button
                          variant="outline"
                          size="sm"
                          className="h-7 text-xs"
                          onClick={() => handleRegisterExternal(skill)}
                          disabled={registeringExternal === skill.canonical}
                        >
                          {registeringExternal === skill.canonical ? (
                            <Loader2 className="h-3.5 w-3.5 animate-spin" />
                          ) : (
                            <span>Register</span>
                          )}
                        </Button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Delete confirmation */}
      <Dialog open={!!deleteTarget} onOpenChange={(open) => !open && setDeleteTarget(null)}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>Delete Skill</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete{" "}
              <strong>{deleteTarget?.name}</strong>? This will remove it from
              the Library and all agent installations.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={deleting}
            >
              {deleting ? "Deleting…" : "Delete"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Discover dialog */}
      <DiscoverSkillsDialog
        open={showDiscoverDialog}
        onClose={() => setShowDiscoverDialog(false)}
        onInstalled={() => {
          setShowDiscoverDialog(false);
          loadSkills();
        }}
      />
    </div>
  );
}
