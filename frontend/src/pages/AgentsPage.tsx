import { useEffect, useState, useCallback } from "react";
import {
  RefreshCw,
  ExternalLink,
  Settings2,
  MessageSquare,
  Brain,
  Loader2,
  AlertTriangle,
  ChevronRight,
  FileText,
  Pencil,
  X,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { toast } from "@/lib/toast";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  ListAgents,
  GetProviderConfig,
  UpdateProviderConfig,
  ListSessions,
  GetSessionTranscript,
  OpenSessionInEditor,
  ListMemory,
  GetMemoryEntry,
  OpenMemoryInEditor,
  SetEnableUnverifiedPaths,
  GetCustomPaths,
  SaveAgentPaths,
} from "@bindings/ai-manager/internal/app/agentservice";
import type { AgentInfo } from "@bindings/ai-manager/internal/app/models";
import type { ProviderConfig } from "@bindings/ai-manager/internal/providers/models";
import type { Session } from "@bindings/ai-manager/internal/sessions/models";
import type { MemoryEntry } from "@bindings/ai-manager/internal/memory/models";
import { Switch } from "@/components/ui/switch";
import ChatViewer from "@/modules/agents/ChatViewer";

function AgentIcon({ iconType, className }: { iconType: string; className?: string }) {
  const cls = className ?? "h-4 w-4";
  switch (iconType) {
    case "claude":
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="currentColor">
          <path fill="#D97757" d="M4.709 15.955l4.72-2.647.08-.23-.08-.128H9.2l-.79-.048-2.698-.073-2.339-.097-2.266-.122-.571-.121L0 11.784l.055-.352.48-.321.686.06 1.52.103 2.278.158 1.652.097 2.449.255h.389l.055-.157-.134-.098-.103-.097-2.358-1.596-2.552-1.688-1.336-.972-.724-.491-.364-.462-.158-1.008.656-.722.881.06.225.061.893.686 1.908 1.476 2.491 1.833.365.304.145-.103.019-.073-.164-.274-1.355-2.446-1.446-2.49-.644-1.032-.17-.619a2.97 2.97 0 01-.104-.729L6.283.134 6.696 0l.996.134.42.364.62 1.414 1.002 2.229 1.555 3.03.456.898.243.832.091.255h.158V9.01l.128-1.706.237-2.095.23-2.695.08-.76.376-.91.747-.492.584.28.48.685-.067.444-.286 1.851-.559 2.903-.364 1.942h.212l.243-.242.985-1.306 1.652-2.064.73-.82.85-.904.547-.431h1.033l.76 1.129-.34 1.166-1.064 1.347-.881 1.142-1.264 1.7-.79 1.36.073.11.188-.02 2.856-.606 1.543-.28 1.841-.315.833.388.091.395-.328.807-1.969.486-2.309.462-3.439.813-.042.03.049.061 1.549.146.662.036h1.622l3.02.225.79.522.474.638-.079.485-1.215.62-1.64-.389-3.829-.91-1.312-.329h-.182v.11l1.093 1.068 2.006 1.81 2.509 2.33.127.578-.322.455-.34-.049-2.205-1.657-.851-.747-1.926-1.62h-.128v.17l.444.649 2.345 3.521.122 1.08-.17.353-.608.213-.668-.122-1.374-1.925-1.415-2.167-1.143-1.943-.14.08-.674 7.254-.316.37-.729.28-.607-.461-.322-.747.322-1.476.389-1.924.315-1.53.286-1.9.17-.632-.012-.042-.14.018-1.434 1.967-2.18 2.945-1.726 1.845-.414.164-.717-.37.067-.662.401-.589 2.388-3.036 1.44-1.882.93-1.086-.006-.158h-.055L4.132 18.56l-1.13.146-.487-.456.061-.746.231-.243 1.908-1.312-.006.006z"/>
        </svg>
      );
    case "codex":
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="currentColor">
          <path fillRule="evenodd" d="M21.55 10.004a5.416 5.416 0 00-.478-4.501c-1.217-2.09-3.662-3.166-6.05-2.66A5.59 5.59 0 0010.831 1C8.39.995 6.224 2.546 5.473 4.838A5.553 5.553 0 001.76 7.496a5.487 5.487 0 00.691 6.5 5.416 5.416 0 00.477 4.502c1.217 2.09 3.662 3.165 6.05 2.66A5.586 5.586 0 0013.168 23c2.443.006 4.61-1.546 5.361-3.84a5.553 5.553 0 003.715-2.66 5.488 5.488 0 00-.693-6.497v.001zm-8.381 11.558a4.199 4.199 0 01-2.675-.954c.034-.018.093-.05.132-.074l4.44-2.53a.71.71 0 00.364-.623v-6.176l1.877 1.069c.02.01.033.029.036.05v5.115c-.003 2.274-1.87 4.118-4.174 4.123zM4.192 17.78a4.059 4.059 0 01-.498-2.763c.032.02.09.055.131.078l4.44 2.53c.225.13.504.13.73 0l5.42-3.088v2.138a.068.068 0 01-.027.057L9.9 19.288c-1.999 1.136-4.552.46-5.707-1.51h-.001zM3.023 8.216A4.15 4.15 0 015.198 6.41l-.002.151v5.06a.711.711 0 00.364.624l5.42 3.087-1.876 1.07a.067.067 0 01-.063.005l-4.489-2.559c-1.995-1.14-2.679-3.658-1.53-5.63h.001zm15.417 3.54l-5.42-3.088L14.896 7.6a.067.067 0 01.063-.006l4.489 2.557c1.998 1.14 2.683 3.662 1.529 5.633a4.163 4.163 0 01-2.174 1.807V12.38a.71 .71 0 00-.363-.623zm1.867-2.773a6.04 6.04 0 00-.132-.078l-4.44-2.53a.731.731 0 00-.729 0l-5.42 3.088V7.325a.068.068 0 01.027.057L14.1 4.713c2-1.137 4.555-.46 5.707 1.513.487.833.664 1.809.499 2.757h.001zm-11.741 3.81l-1.877-1.068a.065.065 0 01-.036-.051V6.559c.001-2.277 1.873-4.122 4.181-4.12.976 0 1.92.338 2.671.954-.034.018-.092.05-.131.073l-4.44 2.53a.71.71 0 00-.365.623l-.003 6.173v.002zm1.02-2.168L12 9.25l2.414 1.375v2.75L12 14.75l-2.415-1.375v-2.75z"/>
        </svg>
      );
    default:
      return (
        <svg className={cls} viewBox="0 0 24 24" fill="currentColor">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-1 17.93c-3.95-.49-7-3.85-7-7.93 0-.62.08-1.21.21-1.79L9 15v1c0 1.1.9 2 2 2v1.93zm6.9-2.54c-.26-.81-1-1.39-1.9-1.39h-1v-3c0-.55-.45-1-1-1H8v-2h2c.55 0 1-.45 1-1V7h2c1.1 0 2-.9 2-2v-.41c2.93 1.19 5 4.06 5 7.41 0 2.08-.8 3.97-2.1 5.39z"/>
        </svg>
      );
  }
}

type TabKey = "provider" | "sessions" | "memory";

export default function AgentsPage() {
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<AgentInfo | null>(null);
  const [tab, setTab] = useState<TabKey>("provider");
  const [loading, setLoading] = useState(true);
  const [enableUnverified, setEnableUnverified] = useState(false);

  // Path config dialog
  const [showPathDialog, setShowPathDialog] = useState(false);
  const [pathForm, setPathForm] = useState({
    provider_config_path: "",
    session_root_path: "",
    session_format: "",
    memory_root_path: "",
    memory_format: "",
  });

  // Provider state
  const [providerConfig, setProviderConfig] = useState<ProviderConfig | null>(null);
  const [editModel, setEditModel] = useState("");
  const [editBaseUrl, setEditBaseUrl] = useState("");
  const [editApiKey, setEditApiKey] = useState("");
  const [savingProvider, setSavingProvider] = useState(false);

  // Sessions state
  const [sessions, setSessions] = useState<Session[]>([]);
  const [sessionFilter, setSessionFilter] = useState("");
  const [selectedSession, setSelectedSession] = useState<Session | null>(null);
  const [transcript, setTranscript] = useState("");

  // Memory state
  const [memoryEntries, setMemoryEntries] = useState<MemoryEntry[]>([]);
  const [selectedMemory, setSelectedMemory] = useState<MemoryEntry | null>(null);
  const [memoryContent, setMemoryContent] = useState("");

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const list = await ListAgents();
      setAgents(list ?? []);
    } catch (err) {
      toast(`加载 Agent 列表失败: ${err}`, true);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { refresh(); }, [refresh, enableUnverified]);

  const selectAgent = useCallback(async (agent: AgentInfo) => {
    setSelectedAgent(agent);
    setProviderConfig(null);
    setSessions([]);
    setMemoryEntries([]);
    setSelectedSession(null);
    setSelectedMemory(null);
    setTranscript("");
    setMemoryContent("");

    if (agent.supports_provider) {
      try {
        const cfg = await GetProviderConfig(agent.kind);
        setProviderConfig(cfg);
        if (cfg) {
          setEditModel(cfg.model ?? "");
          setEditBaseUrl(cfg.base_url ?? "");
          setEditApiKey(cfg.api_key ?? "");
        }
      } catch { }
    }
    if (agent.supports_session) {
      try {
        const list = await ListSessions(agent.kind, { project: "", since: "", limit: 50 });
        setSessions(list ?? []);
      } catch { }
    }
    if (agent.supports_memory) {
      try {
        const list = await ListMemory(agent.kind);
        setMemoryEntries(list ?? []);
      } catch { }
    }
  }, []);

  const saveProvider = useCallback(async () => {
    if (!selectedAgent || !providerConfig) return;
    setSavingProvider(true);
    try {
      const result = await UpdateProviderConfig(selectedAgent.kind, {
        agent: selectedAgent.kind,
        path: providerConfig.path,
        format: providerConfig.format,
        exists: providerConfig.exists,
        provider: providerConfig.provider,
        model: editModel,
        base_url: editBaseUrl,
        api_key: editApiKey,
        extra: providerConfig.extra,
        errors: providerConfig.errors,
        raw: providerConfig.raw,
      });
      if (result) { setProviderConfig(result); toast("Provider 配置已保存"); }
    } catch (err) { toast(`保存失败: ${err}`, true); }
    finally { setSavingProvider(false); }
  }, [selectedAgent, providerConfig, editModel, editBaseUrl, editApiKey]);

  const viewSession = useCallback(async (session: Session) => {
    if (!selectedAgent) return;
    setSelectedSession(session);
    setTranscript("");
    try { setTranscript((await GetSessionTranscript(selectedAgent.kind, session.id)) ?? ""); }
    catch (err) { toast(`加载会话失败: ${err}`, true); }
  }, [selectedAgent]);

  const viewMemory = useCallback(async (entry: MemoryEntry) => {
    if (!selectedAgent) return;
    setSelectedMemory(entry);
    setMemoryContent("");
    try { setMemoryContent((await GetMemoryEntry(selectedAgent.kind, entry.path))?.content ?? ""); }
    catch (err) { toast(`加载 Memory 失败: ${err}`, true); }
  }, [selectedAgent]);

  const toggleUnverified = useCallback(async (checked: boolean) => {
    setEnableUnverified(checked);
    try { await SetEnableUnverifiedPaths(checked); } catch { }
  }, []);

  // --- Path config ---
  const openPathConfig = useCallback(async () => {
    if (!selectedAgent) return;
    try {
      const all = await GetCustomPaths();
      const custom = all?.[selectedAgent.kind];
      const p = selectedAgent.paths;
      setPathForm({
        provider_config_path: custom?.provider_config_path ?? p.provider_config_path ?? "",
        session_root_path: custom?.session_root_path ?? p.session_root_path ?? "",
        session_format: custom?.session_format ?? p.session_format ?? "",
        memory_root_path: custom?.memory_root_path ?? p.memory_root_path ?? "",
        memory_format: custom?.memory_format ?? p.memory_format ?? "",
      });
      setShowPathDialog(true);
    } catch (err) {
      toast(`加载路径配置失败: ${err}`, true);
    }
  }, [selectedAgent]);

  const savePathConfig = useCallback(async () => {
    if (!selectedAgent) return;
    try {
      await SaveAgentPaths(selectedAgent.kind, pathForm);
      toast("路径配置已保存");
      setShowPathDialog(false);
      // Refresh agent list to reflect new paths
      await refresh();
    } catch (err) {
      toast(`保存路径失败: ${err}`, true);
    }
  }, [selectedAgent, pathForm, refresh]);

  const filteredSessions = sessions.filter(
    (s) => !sessionFilter || s.id.toLowerCase().includes(sessionFilter.toLowerCase()) ||
      (s.project && s.project.toLowerCase().includes(sessionFilter.toLowerCase()))
  );

  return (
    <div className="flex h-full gap-4 min-h-0">
      {/* Left: Agent list */}
      <div className="flex w-52 shrink-0 flex-col gap-2 min-h-0">
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-semibold">Agents</h2>
          <button type="button" onClick={refresh} className="rounded-md p-1 text-muted-foreground hover:bg-secondary hover:text-foreground">
            <RefreshCw className="h-3.5 w-3.5" />
          </button>
        </div>

        <div className="flex items-center gap-2 rounded-md bg-muted/50 px-2 py-1.5">
          <Switch checked={enableUnverified} onCheckedChange={toggleUnverified} className="scale-90" />
          <span className="text-xs text-muted-foreground">未验证路径</span>
        </div>

        <div className="flex-1 space-y-0.5 overflow-y-auto min-h-0">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />
            </div>
          ) : (
            agents.map((agent) => (
              <button
                key={agent.kind}
                type="button"
                onClick={() => selectAgent(agent)}
                className={cn(
                  "flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-sm text-left transition-colors hover:bg-secondary",
                  selectedAgent?.kind === agent.kind && "bg-secondary font-medium"
                )}
              >
                <AgentIcon iconType={agent.icon_type} className="h-4 w-4 shrink-0" />
                <span className="flex-1 truncate">{agent.label}</span>
                {agent.has_provider_config && <Settings2 className="h-3 w-3 text-primary" />}
                {agent.has_sessions && <MessageSquare className="h-3 w-3 text-primary" />}
                {agent.has_memory && <Brain className="h-3 w-3 text-primary" />}
              </button>
            ))
          )}
        </div>
      </div>

      {/* Right: Detail panel */}
      <div className="flex min-w-0 flex-1 flex-col min-h-0">
        {!selectedAgent ? (
          <div className="flex h-full items-center justify-center text-sm text-muted-foreground">
            选择一个 Agent 查看详情
          </div>
        ) : (
          <>
            {/* Agent header */}
            <div className="flex items-center gap-2 border-b pb-3">
              <AgentIcon iconType={selectedAgent.icon_type} className="h-5 w-5" />
              <h3 className="text-base font-semibold">{selectedAgent.label}</h3>
              <Badge variant="secondary" className="text-xs">{selectedAgent.status}</Badge>
              <div className="ml-auto flex gap-1">
                <Button variant="ghost" size="sm" onClick={openPathConfig} className="h-7 gap-1 text-xs">
                  <Pencil className="h-3 w-3" /> 路径配置
                </Button>
              </div>
            </div>

            <Tabs value={tab} onValueChange={(v) => setTab(v as TabKey)} className="flex-1 min-h-0 flex flex-col">
              <TabsList className="mt-3">
                {selectedAgent.supports_provider && (
                  <TabsTrigger value="provider" className="gap-1.5">
                    <Settings2 className="h-3.5 w-3.5" /> Provider
                  </TabsTrigger>
                )}
                {selectedAgent.supports_session && (
                  <TabsTrigger value="sessions" className="gap-1.5">
                    <MessageSquare className="h-3.5 w-3.5" /> Sessions ({sessions.length})
                  </TabsTrigger>
                )}
                {selectedAgent.supports_memory && (
                  <TabsTrigger value="memory" className="gap-1.5">
                    <Brain className="h-3.5 w-3.5" /> Memory ({memoryEntries.length})
                  </TabsTrigger>
                )}
              </TabsList>

              {/* Provider tab */}
              <TabsContent value="provider" className="mt-3 overflow-y-auto">
                {providerConfig ? (
                  <Card>
                    <CardHeader className="pb-2">
                      <CardTitle className="text-sm">Provider Config</CardTitle>
                      <p className="text-xs text-muted-foreground">{providerConfig.path} · {providerConfig.format}</p>
                    </CardHeader>
                    <CardContent className="space-y-3">
                      {providerConfig.errors?.length > 0 && (
                        <div className="flex items-start gap-2 rounded-md bg-destructive/10 p-2 text-xs text-destructive">
                          <AlertTriangle className="h-3.5 w-3.5 shrink-0" />
                          <div>{providerConfig.errors.join("; ")}</div>
                        </div>
                      )}
                      <div className="grid grid-cols-2 gap-3">
                        <div className="space-y-1">
                          <label className="text-xs font-medium text-muted-foreground">Provider</label>
                          <Input value={providerConfig.provider} disabled className="h-7 text-xs" />
                        </div>
                        <div className="space-y-1">
                          <label className="text-xs font-medium text-muted-foreground">Model</label>
                          <Input value={editModel} onChange={(e) => setEditModel(e.target.value)} className="h-7 text-xs" />
                        </div>
                        <div className="space-y-1">
                          <label className="text-xs font-medium text-muted-foreground">Base URL</label>
                          <Input value={editBaseUrl} onChange={(e) => setEditBaseUrl(e.target.value)} className="h-7 text-xs" />
                        </div>
                        <div className="space-y-1">
                          <label className="text-xs font-medium text-muted-foreground">API Key</label>
                          <Input value={editApiKey} onChange={(e) => setEditApiKey(e.target.value)} type="password" className="h-7 text-xs" />
                        </div>
                      </div>
                      {providerConfig.raw && (
                        <div className="space-y-1">
                          <label className="text-xs font-medium text-muted-foreground">Raw Content</label>
                          <textarea value={providerConfig.raw} readOnly className="h-32 w-full rounded-md border bg-muted/30 p-2 text-xs font-mono resize-none" />
                        </div>
                      )}
                      <div className="flex justify-end">
                        <Button size="sm" onClick={saveProvider} disabled={savingProvider} className="h-7 text-xs">
                          {savingProvider && <Loader2 className="mr-1 h-3 w-3 animate-spin" />}
                          保存
                        </Button>
                      </div>
                    </CardContent>
                  </Card>
                ) : (
                  <p className="text-sm text-muted-foreground">该 Agent 无本地 Provider 配置</p>
                )}
              </TabsContent>

              {/* Sessions tab */}
              <TabsContent value="sessions" className="mt-3 flex-1 min-h-0 overflow-y-auto">
                <div className="flex items-center gap-2 mb-2">
                  <Input
                    value={sessionFilter}
                    onChange={(e) => setSessionFilter(e.target.value)}
                    placeholder="搜索 session ID 或项目路径..."
                    className="h-7 text-xs flex-1"
                  />
                </div>
                <div className="grid grid-cols-2 gap-2 h-[calc(100%-3rem)]">
                  <div className="space-y-1 overflow-y-auto min-h-0">
                    {filteredSessions.length === 0 ? (
                      <p className="text-xs text-muted-foreground py-4 text-center">无会话</p>
                    ) : (
                      filteredSessions.map((session) => (
                        <button
                          key={session.id}
                          type="button"
                          onClick={() => viewSession(session)}
                          className={cn(
                            "w-full rounded-md border p-2 text-left text-xs transition-colors hover:bg-secondary",
                            selectedSession?.id === session.id && "bg-secondary border-primary"
                          )}
                        >
                          <div className="flex items-center gap-1.5">
                            <ChevronRight className="h-3 w-3 shrink-0" />
                            <span className="flex-1 truncate font-medium">{session.id}</span>
                            <Badge variant="outline" className="text-[10px]">{session.message_count} msg</Badge>
                          </div>
                          {session.project && <div className="truncate text-[10px] text-muted-foreground pl-4">{session.project}</div>}
                          <div className="text-[10px] text-muted-foreground pl-4">{new Date(session.started_at).toLocaleString()}</div>
                        </button>
                      ))
                    )}
                  </div>
                  <div className="rounded-md border bg-muted/20 overflow-y-auto min-h-0">
                    {selectedSession ? (
                      <>
                        <div className="flex items-center justify-between sticky top-0 bg-muted/50 z-10 px-2 py-1.5 border-b">
                          <span className="text-xs font-medium truncate">{selectedSession.id}</span>
                          <button type="button" onClick={() => OpenSessionInEditor(selectedAgent.kind, selectedSession.id)} className="rounded p-1 text-muted-foreground hover:bg-secondary hover:text-foreground" title="在外部编辑器打开">
                            <ExternalLink className="h-3 w-3" />
                          </button>
                        </div>
                        {transcript ? (
                          <div className="p-2">
                            <ChatViewer raw={transcript} />
                          </div>
                        ) : (
                          <div className="flex items-center justify-center py-8"><Loader2 className="h-4 w-4 animate-spin text-muted-foreground" /></div>
                        )}
                      </>
                    ) : (
                      <p className="text-xs text-muted-foreground py-8 text-center">选择一个会话查看内容</p>
                    )}
                  </div>
                </div>
              </TabsContent>

              {/* Memory tab */}
              <TabsContent value="memory" className="mt-3 flex-1 min-h-0 overflow-y-auto">
                <div className="grid grid-cols-2 gap-2 h-[calc(100%-1rem)]">
                  <div className="space-y-1 overflow-y-auto min-h-0">
                    {memoryEntries.length === 0 ? (
                      <p className="text-xs text-muted-foreground py-4 text-center">无 Memory 条目</p>
                    ) : (
                      memoryEntries.map((entry) => (
                        <button
                          key={entry.path}
                          type="button"
                          onClick={() => viewMemory(entry)}
                          className={cn(
                            "w-full rounded-md border p-2 text-left text-xs transition-colors hover:bg-secondary",
                            selectedMemory?.path === entry.path && "bg-secondary border-primary"
                          )}
                        >
                          <div className="flex items-center gap-1.5">
                            <ChevronRight className="h-3 w-3 shrink-0" />
                            <span className="flex-1 truncate font-medium">{entry.title}</span>
                            <Badge variant="outline" className="text-[10px]">{entry.kind}</Badge>
                          </div>
                          <div className="truncate text-[10px] text-muted-foreground pl-4">{entry.size_bytes} bytes</div>
                        </button>
                      ))
                    )}
                  </div>
                  <div className="rounded-md border bg-muted/20 overflow-y-auto min-h-0">
                    {selectedMemory ? (
                      <>
                        <div className="flex items-center justify-between sticky top-0 bg-muted/50 z-10 px-2 py-1.5 border-b">
                          <span className="text-xs font-medium truncate">{selectedMemory.title}</span>
                          <button type="button" onClick={() => OpenMemoryInEditor(selectedAgent.kind, selectedMemory.path)} className="rounded p-1 text-muted-foreground hover:bg-secondary hover:text-foreground" title="在外部编辑器打开">
                            <ExternalLink className="h-3 w-3" />
                          </button>
                        </div>
                        {memoryContent ? (
                          <div className="p-2">
                            <pre className="text-[11px] font-mono whitespace-pre-wrap break-all">{memoryContent}</pre>
                          </div>
                        ) : (
                          <div className="flex items-center justify-center py-8"><Loader2 className="h-4 w-4 animate-spin text-muted-foreground" /></div>
                        )}
                      </>
                    ) : (
                      <p className="text-xs text-muted-foreground py-8 text-center">选择一个条目查看内容</p>
                    )}
                  </div>
                </div>
              </TabsContent>
            </Tabs>
          </>
        )}
      </div>

      {/* Path config dialog */}
      <Dialog open={showPathDialog} onOpenChange={setShowPathDialog}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>路径配置 — {selectedAgent?.label}</DialogTitle>
          </DialogHeader>
          <div className="space-y-3 py-2">
            <p className="text-xs text-muted-foreground">
              自定义 Provider 配置文件、Session 目录、Memory 目录的位置。留空使用默认路径。
            </p>
            <div className="space-y-1">
              <label className="text-xs font-medium text-muted-foreground">Provider 配置文件</label>
              <Input value={pathForm.provider_config_path} onChange={(e) => setPathForm((f) => ({ ...f, provider_config_path: e.target.value }))} placeholder="留空使用默认路径" className="h-8 text-xs font-mono" />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <label className="text-xs font-medium text-muted-foreground">Session 目录</label>
                <Input value={pathForm.session_root_path} onChange={(e) => setPathForm((f) => ({ ...f, session_root_path: e.target.value }))} placeholder="留空使用默认路径" className="h-8 text-xs font-mono" />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-medium text-muted-foreground">Session 格式</label>
                <select value={pathForm.session_format} onChange={(e) => setPathForm((f) => ({ ...f, session_format: e.target.value }))} className="h-8 w-full rounded-md border bg-muted/30 px-2 text-xs">
                  <option value="">使用默认</option>
                  <option value="jsonl_per_project">JSONL (per project)</option>
                  <option value="json_flat">JSON (flat)</option>
                  <option value="sqlite_global">SQLite (global)</option>
                  <option value="vscode_storage">VS Code Storage</option>
                </select>
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-1">
                <label className="text-xs font-medium text-muted-foreground">Memory 目录</label>
                <Input value={pathForm.memory_root_path} onChange={(e) => setPathForm((f) => ({ ...f, memory_root_path: e.target.value }))} placeholder="留空使用默认路径" className="h-8 text-xs font-mono" />
              </div>
              <div className="space-y-1">
                <label className="text-xs font-medium text-muted-foreground">Memory 格式</label>
                <select value={pathForm.memory_format} onChange={(e) => setPathForm((f) => ({ ...f, memory_format: e.target.value }))} className="h-8 w-full rounded-md border bg-muted/30 px-2 text-xs">
                  <option value="">使用默认</option>
                  <option value="markdown_index">Markdown Index</option>
                  <option value="markdown_flat">Markdown Flat</option>
                  <option value="sqlite">SQLite</option>
                  <option value="none">None</option>
                  <option value="rules_only">Rules Only</option>
                  <option value="server_side">Server Side</option>
                </select>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="ghost" size="sm" onClick={() => setShowPathDialog(false)}>取消</Button>
            <Button size="sm" onClick={savePathConfig}>保存路径</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
