import { useState, useEffect, useCallback } from "react";
import {
  FolderOpen,
  FolderPlus,
  X,
  AlertTriangle,
  AlertCircle,
  CheckCircle,
  Minus,
  RefreshCw,
  Loader2,
  FolderTree,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { cn } from "@/lib/utils";
import {
  AddProject,
  RemoveProject,
  ListProjects,
  BrowseProject,
  GetEffectiveSkills,
} from "@bindings/ai-manager/internal/app/projectservice";
import type { EffectiveSkill } from "@bindings/ai-manager/internal/effective/models";
import type { Project } from "@bindings/ai-manager/internal/project/models";

import type { AgentKey } from "@/components/agents";
import { AGENT_META, AgentIcon, agentLabel } from "@/components/agents";

type AgentKind = AgentKey;

// All agents to display in the grid (single source of truth: AGENT_META).
const AGENTS: { key: AgentKind; label: string }[] = AGENT_META.map((m) => ({
  key: m.key,
  label: m.label,
}));

interface AgentRow {
  agent: AgentKind;
  label: string;
  skillCount: number;
  totalTokens: number;
  warning: boolean;
  danger: boolean;
}

function StatusBadge({ warning, danger }: { warning: boolean; danger: boolean }) {
  if (danger)
    return (
      <Badge className="bg-red-500/15 text-red-400 hover:bg-red-500/25 gap-1">
        <AlertCircle className="h-3 w-3" /> Danger
      </Badge>
    );
  if (warning)
    return (
      <Badge className="bg-amber-500/15 text-amber-400 hover:bg-amber-500/25 gap-1">
        <AlertTriangle className="h-3 w-3" /> Warn
      </Badge>
    );
  return (
    <Badge variant="outline" className="text-green-500 border-green-500/30 gap-1">
      <CheckCircle className="h-3 w-3" /> OK
    </Badge>
  );
}

export default function ProjectsPage() {
  const [projects, setProjects] = useState<Project[]>([]);
  const [selectedPath, setSelectedPath] = useState<string>("");
  const [agentRows, setAgentRows] = useState<AgentRow[]>([]);
  const [effectiveSkills, setEffectiveSkills] = useState<EffectiveSkill[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<AgentKind | null>(null);
  const [loading, setLoading] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [inputPath, setInputPath] = useState("");

  const loadProjects = useCallback(async () => {
    setLoading(true);
    try {
      const result = await ListProjects();
      setProjects(result);
      if (result.length > 0 && !selectedPath) {
        setSelectedPath(result[0].path);
      }
    } catch (err) {
      console.error("ListProjects failed:", err);
    } finally {
      setLoading(false);
    }
  }, [selectedPath]);

  useEffect(() => {
    loadProjects();
  }, []);

  const scanAgents = useCallback(async (projectPath: string) => {
    if (!projectPath) return;
    setScanning(true);
    try {
      // Fetch all effective skills (deduplicated across all agents)
      const allSkills: EffectiveSkill[] = await GetEffectiveSkills(projectPath);
      const results: AgentRow[] = [];
      for (const { key, label } of AGENTS) {
        // Filter to skills this agent sees
        const skills = allSkills.filter((s) => s.agents?.includes(key));
        const totalTokens = skills.reduce((sum, s) => sum + s.tokens, 0);
        results.push({
          agent: key,
          label,
          skillCount: skills.length,
          totalTokens,
          warning: totalTokens > 2000 || skills.length > 20,
          danger: totalTokens > 5000 || skills.length > 50,
        });
      }
      setAgentRows(results);
    } finally {
      setScanning(false);
    }
  }, []);

  useEffect(() => {
    if (selectedPath) {
      scanAgents(selectedPath);
    }
  }, [selectedPath, scanAgents]);

  const loadEffectiveSkills = useCallback(async (projectPath: string, agent: AgentKind) => {
    try {
      const allSkills: EffectiveSkill[] = await GetEffectiveSkills(projectPath);
      // Filter to skills this agent sees
      const skills = allSkills.filter((s) => s.agents?.includes(agent));
      setEffectiveSkills(skills);
    } catch (err) {
      console.error("GetEffectiveSkills failed:", err);
      setEffectiveSkills([]);
    }
  }, []);

  const handleSelectAgent = (agent: AgentKind) => {
    if (selectedAgent === agent) {
      setSelectedAgent(null);
      setEffectiveSkills([]);
    } else {
      setSelectedAgent(agent);
      loadEffectiveSkills(selectedPath, agent);
    }
  };

  const handleAddProject = async () => {
    if (inputPath.trim()) {
      try {
        await AddProject(inputPath.trim());
        setInputPath("");
        await loadProjects();
      } catch (err) {
        console.error("AddProject failed:", err);
      }
    }
  };

  const handleBrowse = async () => {
    try {
      const path = await BrowseProject();
      if (path) {
        await AddProject(path);
        await loadProjects();
      }
    } catch (err) {
      console.error("BrowseProject failed:", err);
    }
  };

  const handleRemoveProject = async (path: string) => {
    try {
      await RemoveProject(path);
      if (selectedPath === path) {
        setSelectedPath("");
        setAgentRows([]);
        setEffectiveSkills([]);
        setSelectedAgent(null);
      }
      await loadProjects();
    } catch (err) {
      console.error("RemoveProject failed:", err);
    }
  };

  const selectedProject = projects.find((p) => p.path === selectedPath);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Projects</h1>
          <p className="text-sm text-muted-foreground mt-1">
            Manage skill installations across your projects
          </p>
        </div>
        <Button onClick={handleBrowse}>
          <FolderPlus className="h-4 w-4 mr-2" />
          Add Project
        </Button>
      </div>

      {/* Project Selector */}
      <Card>
        <CardHeader className="pb-3">
          <CardTitle className="text-sm font-medium text-muted-foreground">
            <FolderOpen className="h-4 w-4 inline mr-2" />
            Select Project
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3">
          {loading ? (
            <div className="flex items-center gap-2 text-muted-foreground text-sm">
              <Loader2 className="h-4 w-4 animate-spin" />
              Loading projects...
            </div>
          ) : projects.length === 0 ? (
            <div className="text-center py-4">
              <FolderTree className="h-8 w-8 mx-auto text-muted-foreground/50 mb-2" />
              <p className="text-sm text-muted-foreground">No projects yet. Add one to get started.</p>
            </div>
          ) : (
            <>
              <div className="flex flex-wrap gap-2">
                {projects.map((p) => (
                  <button
                    key={p.path}
                    onClick={() => setSelectedPath(p.path)}
                    className={cn(
                      "group flex items-center gap-2 rounded-lg border px-3 py-2 text-sm transition-colors",
                      selectedPath === p.path
                        ? "border-primary bg-primary/10 text-primary"
                        : "border-border hover:bg-secondary"
                    )}
                  >
                    <span className="font-medium truncate max-w-[200px]">
                      {p.name || p.path.split("/").pop() || p.path}
                    </span>
                    <span className="text-xs text-muted-foreground truncate max-w-[150px]">
                      {p.path}
                    </span>
                    <X
                      className="h-3 w-3 text-muted-foreground opacity-0 group-hover:opacity-100 hover:text-destructive transition-opacity"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleRemoveProject(p.path);
                      }}
                    />
                  </button>
                ))}
              </div>
              <div className="flex gap-2">
                <Input
                  placeholder="/path/to/project"
                  value={inputPath}
                  onChange={(e) => setInputPath(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleAddProject()}
                  className="flex-1"
                />
                <Button onClick={handleAddProject} disabled={!inputPath.trim()}>
                  Add
                </Button>
              </div>
            </>
          )}
        </CardContent>
      </Card>

      {/* Agent Grid */}
      {selectedPath && (
        <Card>
          <CardHeader className="pb-3 flex flex-row items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              <FolderTree className="h-4 w-4 inline mr-2" />
              Agent Skills Summary — {selectedProject?.name || selectedPath.split("/").pop()}
            </CardTitle>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => scanAgents(selectedPath)}
              disabled={scanning}
            >
              {scanning ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                <RefreshCw className="h-4 w-4" />
              )}
              <span className="ml-2">Rescan</span>
            </Button>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3 font-medium text-muted-foreground">Agent</th>
                    <th className="text-right py-2 px-3 font-medium text-muted-foreground">Skills</th>
                    <th className="text-right py-2 px-3 font-medium text-muted-foreground">Tokens</th>
                    <th className="text-right py-2 px-3 font-medium text-muted-foreground">Status</th>
                    <th className="w-8" />
                  </tr>
                </thead>
                <tbody>
                  {agentRows.map((row) => (
                    <tr
                      key={row.agent}
                      className={cn(
                        "border-b cursor-pointer transition-colors hover:bg-secondary/50",
                        selectedAgent === row.agent && "bg-secondary/30"
                      )}
                      onClick={() => handleSelectAgent(row.agent)}
                    >
                      <td className="py-2.5 px-3">
                        <div className="flex items-center gap-2.5">
                          <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-md bg-muted/60">
                            <AgentIcon
                              agent={row.agent}
                              size={14}
                              className="text-muted-foreground"
                            />
                          </span>
                          <span className="font-medium">{row.label}</span>
                        </div>
                      </td>
                      <td className="py-2.5 px-3 text-right tabular-nums">{row.skillCount}</td>
                      <td className="py-2.5 px-3 text-right tabular-nums">
                        {row.totalTokens > 0 ? row.totalTokens.toLocaleString() : "—"}
                      </td>
                      <td className="py-2.5 px-3 text-right">
                        {row.skillCount > 0 ? (
                          <StatusBadge warning={row.warning} danger={row.danger} />
                        ) : (
                          <Badge variant="outline" className="text-muted-foreground/50 gap-1">
                            <Minus className="h-3 w-3" /> None
                          </Badge>
                        )}
                      </td>
                      <td className="py-2.5 px-3">
                        {selectedAgent === row.agent && (
                          <span className="text-primary text-xs font-medium">Viewing</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Effective Skills Detail */}
      {selectedAgent && effectiveSkills.length > 0 && (
        <Card>
          <CardHeader className="pb-3 flex flex-row items-center justify-between">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Effective Skills — {AGENTS.find((a) => a.key === selectedAgent)?.label}
            </CardTitle>
            <span className="text-xs text-muted-foreground">
              {effectiveSkills.length} skill{effectiveSkills.length !== 1 ? "s" : ""}
              {effectiveSkills.length > 0 &&
                ` · ${effectiveSkills.reduce((s, sk) => s + sk.tokens, 0).toLocaleString()} tokens`}
            </span>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b">
                    <th className="text-left py-2 px-3 font-medium text-muted-foreground">Skill</th>
                    <th className="text-left py-2 px-3 font-medium text-muted-foreground">Agents</th>
                    <th className="text-center py-2 px-3 font-medium text-muted-foreground">Managed</th>
                    <th className="text-right py-2 px-3 font-medium text-muted-foreground">Tokens</th>
                    <th className="text-right py-2 px-3 font-medium text-muted-foreground">Status</th>
                  </tr>
                </thead>
                <tbody>
                  {effectiveSkills.map((skill, i) => (
                    <tr key={i} className="border-b hover:bg-secondary/30">
                      <td className="py-2.5 px-3 font-medium">{skill.name}</td>
                      <td className="py-2.5 px-3">
                        {skill.agents && skill.agents.length > 0 ? (
                          <div className="flex gap-1 flex-wrap">
                            {skill.agents.map((a) => (
                              <Badge
                                key={a}
                                variant="outline"
                                className="text-xs gap-1 py-0.5 normal-case"
                              >
                                <AgentIcon agent={a} size={11} />
                                {agentLabel(a)}
                              </Badge>
                            ))}
                          </div>
                        ) : (
                          <span className="text-muted-foreground">—</span>
                        )}
                      </td>
                      <td className="py-2.5 px-3 text-center">
                        {skill.builtin ? (
                          <Badge variant="secondary" className="text-xs">Builtin</Badge>
                        ) : skill.managed ? (
                          <CheckCircle className="h-4 w-4 text-green-500 inline" />
                        ) : (
                          <X className="h-4 w-4 text-muted-foreground/50 inline" />
                        )}
                      </td>
                      <td className="py-2.5 px-3 text-right tabular-nums">
                        {skill.tokens > 0 ? skill.tokens.toLocaleString() : "—"}
                      </td>
                      <td className="py-2.5 px-3 text-right">
                        <StatusBadge warning={skill.warning} danger={skill.danger} />
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </CardContent>
        </Card>
      )}

      {selectedAgent && effectiveSkills.length === 0 && (
        <Card>
          <CardHeader className="pb-3">
            <CardTitle className="text-sm font-medium text-muted-foreground">
              Effective Skills — {AGENTS.find((a) => a.key === selectedAgent)?.label}
            </CardTitle>
          </CardHeader>
          <CardContent className="text-center py-8">
            <FolderTree className="h-8 w-8 mx-auto text-muted-foreground/50 mb-2" />
            <p className="text-sm text-muted-foreground">
              No effective skills found for this agent in the selected project.
            </p>
          </CardContent>
        </Card>
      )}
    </div>
  );
}
