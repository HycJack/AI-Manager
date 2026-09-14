import { useState, useEffect, useCallback } from "react";
import {
  Search,
  X,
  ExternalLink,
  Download,
  Loader2,
  Github,
  Globe,
  FolderOpen,
  Plus,
  Flame,
  RefreshCw,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { toast } from "@/lib/toast";
import { useI18n } from "@/modules/i18n";
import { RepoInstallPicker } from "@/components/RepoInstallPicker";
import {
  SearchSkillsSh,
  SearchGitHubSkills,
  AddSkill,
  PopularSkillsSh,
} from "@bindings/ai-manager/internal/app/skillservice";
import { BrowseLibrary } from "@bindings/ai-manager/internal/app/configservice";
import type { SkillsShSearchResult, GitHubSearchResult, PopularSkill } from "@bindings/ai-manager/internal/app/models";

/** formatCount renders an install count the way skills.sh does: 3.39M / 209.7K. */
function formatCount(n: number): string {
  if (n >= 1_000_000) return (n / 1_000_000).toFixed(2) + "M";
  if (n >= 1_000) return (n / 1_000).toFixed(1) + "K";
  return String(n);
}

interface DiscoverSkillsDialogProps {
  open: boolean;
  onClose: () => void;
  onInstalled: () => void;
}

type Source = "skills-sh" | "github" | "popular";

export function DiscoverSkillsDialog({ open, onClose, onInstalled }: DiscoverSkillsDialogProps) {
  const { t } = useI18n();
  const [source, setSource] = useState<Source>("skills-sh");
  const [query, setQuery] = useState("");
  const [searching, setSearching] = useState(false);
  const [skillsShResults, setSkillsShResults] = useState<SkillsShSearchResult | null>(null);
  const [githubResults, setGitHubResults] = useState<GitHubSearchResult | null>(null);
  const [selectedFolder, setSelectedFolder] = useState("");
  const [adding, setAdding] = useState(false);
  const [popular, setPopular] = useState<PopularSkill[] | null>(null);
  const [popularLoading, setPopularLoading] = useState(false);
  const [addingKey, setAddingKey] = useState<string | null>(null);
  // A row that opened the "which skills does this repo ship" picker.
  const [installSource, setInstallSource] = useState<{ source: string; skillKey: string } | null>(null);

  const doSearch = useCallback(async () => {
    if (!query.trim()) return;
    setSearching(true);
    try {
      if (source === "skills-sh") {
        const result = await SearchSkillsSh(query, 20, 0);
        setSkillsShResults(result);
      } else {
        const result = await SearchGitHubSkills(query);
        setGitHubResults(result);
      }
    } catch (e) {
      toast(String(e), true);
    } finally {
      setSearching(false);
    }
  }, [query, source]);

  // Leaderboard fetch. Runs once per open so the Popular tab is already warm
  // when the user switches to it.
  const loadPopular = useCallback(async () => {
    setPopularLoading(true);
    try {
      setPopular(await PopularSkillsSh(20));
    } catch (e) {
      toast(String(e), true);
    } finally {
      setPopularLoading(false);
    }
  }, []);

  useEffect(() => {
    if (open) {
      setQuery("");
      setSkillsShResults(null);
      setGitHubResults(null);
      setSelectedFolder("");
      loadPopular();
    }
  }, [open, loadPopular]);

  const handleAddLocal = async () => {
    if (!selectedFolder) return;
    setAdding(true);
    try {
      await AddSkill("local", selectedFolder, "", []);
      toast(t("skills.added"));
      onInstalled();
    } catch (e) {
      toast(String(e), true);
    } finally {
      setAdding(false);
    }
  };

  // Leaderboard rows install the same way the npx flow does: the source is
  // "owner/repo", which the library clones and scans.
  const handleAddRemote = async (p: PopularSkill) => {
    setAddingKey(p.key);
    try {
      await AddSkill("npx", p.source, "", []);
      toast(t("skills.added"));
      onInstalled();
      onClose();
    } catch (e) {
      toast(String(e), true);
    } finally {
      setAddingKey(null);
    }
  };

  if (!open) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div
        className="w-full max-w-2xl max-h-[80vh] rounded-xl border bg-background shadow-lg flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center gap-3 border-b px-5 py-3">
          <h2 className="text-lg font-semibold shrink-0">{t("skills.discover")}</h2>
          <div className="ml-auto flex items-center gap-1 shrink-0">
            <Button
              variant="ghost"
              size="sm"
              className={cn(source === "popular" && "bg-muted")}
              onClick={() => {
                setSource("popular");
                if (!popular) loadPopular();
              }}
            >
              <Flame className="mr-1 h-4 w-4" />
              {t("skills.popular")}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className={cn(source === "skills-sh" && "bg-muted")}
              onClick={() => setSource("skills-sh")}
            >
              <Globe className="mr-1 h-4 w-4" />
              skills.sh
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className={cn(source === "github" && "bg-muted")}
              onClick={() => setSource("github")}
            >
              <Github className="mr-1 h-4 w-4" />
              GitHub
            </Button>
            <div className="w-px h-6 bg-border mx-1" />
            <button
              onClick={onClose}
              className="flex h-8 w-8 items-center justify-center rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Add from local section */}
        <div className="border-b px-5 py-3">
          <div className="flex items-center gap-2 mb-2">
            <FolderOpen className="h-4 w-4 text-muted-foreground" />
            <span className="text-sm font-medium">{t("skills.addLocal")}</span>
          </div>
          <div className="flex gap-2">
            <div className="flex-1">
              <div className="flex items-center gap-2 rounded-lg bg-muted/50 px-3 py-2">
                <span className="flex-1 truncate text-sm text-muted-foreground">
                  {selectedFolder || t("skills.noFolder")}
                </span>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              onClick={async () => {
                const path = await BrowseLibrary();
                if (path) setSelectedFolder(path);
              }}
            >
              {t("skills.browse")}
            </Button>
            <Button size="sm" onClick={handleAddLocal} disabled={!selectedFolder || adding}>
              {adding ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="mr-1 h-4 w-4" />}
              {adding ? t("skills.adding") : t("skills.add")}
            </Button>
          </div>
        </div>

        {/* Popular tab has nothing to search, so hide the search bar */}
        {source !== "popular" ? (
          <div className="flex items-center gap-2 border-b px-5 py-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && doSearch()}
                placeholder={source === "skills-sh" ? t("skills.searchSkillsSh") : t("skills.searchGithub")}
                className="pl-9"
                autoFocus
              />
            </div>
            <Button onClick={doSearch} disabled={searching || !query.trim()}>
              {searching ? <Loader2 className="h-4 w-4 animate-spin" /> : t("skills.search")}
            </Button>
          </div>
        ) : null}

        {source === "popular" && (
          <div className="flex items-center gap-2 border-b px-5 py-2.5">
            <Badge variant="secondary">{popular?.length ?? 0}</Badge>
            <span className="text-xs text-muted-foreground">{t("skills.popular.hint")}</span>
            <Button
              variant="ghost"
              size="icon"
              className="ml-auto h-7 w-7"
              disabled={popularLoading}
              onClick={loadPopular}
              title={t("skills.popular.refresh")}
            >
              <RefreshCw className={cn("h-3.5 w-3.5", popularLoading && "animate-spin")} />
            </Button>
          </div>
        )}

        {/* Results */}
        <div className="flex-1 overflow-auto p-3">
          {source !== "popular" && !skillsShResults && !githubResults && !searching ? (
            <div className="flex h-full flex-col items-center justify-center text-muted-foreground">
              <Globe className="h-12 w-12 mb-3 opacity-40" />
              <p>{t("skills.searchHint")}</p>
            </div>
          ) : (
            <>
              {source === "skills-sh" && skillsShResults && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2 px-1">
                    <Badge variant="secondary">{skillsShResults.totalCount}</Badge>
                    <span className="text-xs text-muted-foreground">{t("skills.results")} "{skillsShResults.query}"</span>
                  </div>
                  {skillsShResults.skills.length === 0 ? (
                    <p className="text-center text-sm text-muted-foreground py-8">{t("skills.noResults")}</p>
                  ) : (
                    skillsShResults.skills.map((s) => (
                      <div key={s.key} className="flex items-center gap-3 rounded-lg border p-3 hover:bg-muted/30">
                        <div className="min-w-0 flex-1">
                          <div className="font-medium truncate">{s.name}</div>
                          <div className="text-xs text-muted-foreground truncate">
                            {s.repoOwner}/{s.repoName}
                          </div>
                        </div>
                        <span className="text-xs text-muted-foreground shrink-0">
                          {s.installs} {t("skills.installs")}
                        </span>
                        <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
                          <a href={s.readmeUrl} target="_blank" rel="noreferrer">
                            <ExternalLink className="h-4 w-4" />
                          </a>
                        </Button>
                        <Button
                          size="sm"
                          className={cn(installSource?.source === s.source && "border-primary bg-primary/10")}
                          onClick={() => setInstallSource({ source: s.source, skillKey: s.key })}
                        >
                          <Download className="mr-1 h-4 w-4" />
                          {t("skills.add")}
                        </Button>
                      </div>
                    ))
                  )}
                </div>
              )}

              {source === "github" && githubResults && (
                <div className="space-y-2">
                  <div className="flex items-center gap-2 px-1">
                    <Badge variant="secondary">{githubResults.total_count}</Badge>
                    <span className="text-xs text-muted-foreground">{t("skills.githubRepos")}</span>
                  </div>
                  {githubResults.items.length === 0 ? (
                    <p className="text-center text-sm text-muted-foreground py-8">{t("skills.noResults")}</p>
                  ) : (
                    githubResults.items.map((r) => (
                      <div key={r.full_name} className="flex items-center gap-3 rounded-lg border p-3 hover:bg-muted/30">
                        <div className="min-w-0 flex-1">
                          <div className="font-medium truncate">{r.full_name}</div>
                          <div className="text-xs text-muted-foreground truncate">
                            {r.description || t("skills.noDesc")}
                          </div>
                        </div>
                        <span className="text-xs text-muted-foreground shrink-0">
                          ★ {r.stargazers_count}
                        </span>
                        <Button variant="ghost" size="icon" className="h-8 w-8" asChild>
                          <a href={r.html_url} target="_blank" rel="noreferrer">
                            <ExternalLink className="h-4 w-4" />
                          </a>
                        </Button>
                        <Button
                          size="sm"
                          className={cn(installSource?.source === r.full_name && "border-primary bg-primary/10")}
                          onClick={() => setInstallSource({ source: r.full_name, skillKey: "" })}
                        >
                          <Download className="mr-1 h-4 w-4" />
                          {t("skills.add")}
                        </Button>
                      </div>
                    ))
                  )}
                </div>
              )}

              {source === "popular" &&
                (popularLoading && !popular ? (
                  <div className="flex h-40 items-center justify-center">
                    <Loader2 className="h-6 w-6 animate-spin text-muted-foreground" />
                  </div>
                ) : popular && popular.length > 0 ? (
                  <div className="space-y-2">
                    {popular.map((p) => (
                      <div key={p.key} className="flex items-center gap-3 rounded-lg border p-3 hover:bg-muted/30">
                        <span
                          className={cn(
                            "w-8 shrink-0 text-center font-mono text-sm font-semibold",
                            p.rank <= 3 ? "text-amber-500" : "text-muted-foreground",
                          )}
                        >
                          {p.rank}
                        </span>
                        <div className="min-w-0 flex-1">
                          <div className="flex items-center gap-1.5">
                            <span className="truncate font-medium">{p.name}</span>
                            {p.isOfficial && (
                              <Badge variant="outline" className="shrink-0 px-1 py-0 text-[10px]">
                                {t("skills.popular.official")}
                              </Badge>
                            )}
                          </div>
                          <div className="truncate text-xs text-muted-foreground">
                            {p.repoOwner}/{p.repoName}
                          </div>
                        </div>
                        <div className="shrink-0 text-right">
                          <div className="text-sm font-medium">{formatCount(p.installs)}</div>
                          <div className="text-[10px] text-muted-foreground">
                            {formatCount(p.weeklyActivity)} · {t("skills.popular.weekly")}
                          </div>
                        </div>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 shrink-0"
                          title={t("skills.popular.github")}
                          asChild
                        >
                          <a href={p.githubUrl} target="_blank" rel="noreferrer">
                            <Github className="h-4 w-4" />
                          </a>
                        </Button>
                        <Button
                          size="sm"
                          className="shrink-0"
                          disabled={addingKey !== null}
                          onClick={() => handleAddRemote(p)}
                        >
                          {addingKey === p.key ? (
                            <Loader2 className="mr-1 h-4 w-4 animate-spin" />
                          ) : (
                            <Plus className="mr-1 h-4 w-4" />
                          )}
                          {addingKey === p.key ? t("skills.adding") : t("skills.add")}
                        </Button>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="py-8 text-center text-sm text-muted-foreground">{t("skills.noResults")}</p>
                ))}
            </>
          )}
        </div>
      </div>

      <RepoInstallPicker
        open={installSource !== null}
        source={installSource?.source ?? ""}
        skillKey={installSource?.skillKey ?? ""}
        onClose={() => setInstallSource(null)}
        onInstalled={onInstalled}
      />
    </div>
  );
}
