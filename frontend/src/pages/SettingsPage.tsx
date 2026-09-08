import { useEffect, useState } from "react";
import {
  Monitor,
  Moon,
  Sun,
  Palette,
  Settings2,
  Info,
  FolderOpen,
  Folder,
  RefreshCw,
  Languages,
  ExternalLink,
  Sparkles,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { usePreferencesStore } from "@/modules/settings/store";
import { useI18n } from "@/modules/i18n";
import { useTheme, listBuiltinThemes, type ThemeColors } from "@/modules/theme";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { toast } from "@/lib/toast";
import { GetInfo } from "@bindings/ai-manager/internal/app/versionservice";
import { CheckForUpdates, CheckAndInstall, Restart, UpdateState } from "@bindings/ai-manager/internal/app/updateservice";
import { IsEnabled as IsAutostartEnabled, SetEnabled as SetAutostartEnabled } from "@bindings/ai-manager/internal/app/autostartservice";
import { GetConfig, UpdateConfig, BrowseLibrary, RescanLibrary } from "@bindings/ai-manager/internal/app/configservice";

// Built from the upstream projects this template abstracts. Opens in the
// browser via the OS default handler — passed to window.open as a plain link.
const SOURCES = [
  {
    name: "CPM OCR Studio",
    repo: "HycJack/models_run_with_go_wails3",
    desc: "模型推理基础设施（OCR / ASR / LLM / YOLO）",
  },
  {
    name: "terax-clone",
    repo: "HycJack/terax-clone",
    desc: "主题与设置系统",
  },
];

/** True when the running binary is a local/dev build (no ldflags injected). */
function isDevBuild(v: string): boolean {
  return v === "" || v === "dev" || v === "unknown";
}

export default function SettingsPage() {
  const { t } = useI18n();

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold">{t("settings.title")}</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          {t("settings.subtitle")}
        </p>
      </div>

      <Tabs defaultValue="general">
        <TabsList className="w-full">
          <TabsTrigger value="general" className="flex-1 gap-1.5">
            <Settings2 className="h-4 w-4" />
            {t("settings.tab.general")}
          </TabsTrigger>
          <TabsTrigger value="themes" className="flex-1 gap-1.5">
            <Palette className="h-4 w-4" />
            {t("settings.tab.themes")}
          </TabsTrigger>
          <TabsTrigger value="library" className="flex-1 gap-1.5">
            <Folder className="h-4 w-4" />
            {t("settings.tab.library")}
          </TabsTrigger>
          <TabsTrigger value="about" className="flex-1 gap-1.5">
            <Info className="h-4 w-4" />
            {t("settings.tab.about")}
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-6">
          <GeneralSection />
        </TabsContent>
        <TabsContent value="themes" className="mt-6">
          <ThemesSection />
        </TabsContent>
        <TabsContent value="library" className="mt-6">
          <LibrarySection />
        </TabsContent>
        <TabsContent value="about" className="mt-6">
          <AboutSection />
        </TabsContent>
      </Tabs>
    </div>
  );
}

// ---------- General ----------

const MODES: { id: "system" | "light" | "dark"; labelKey: string; icon: typeof Sun }[] = [
  { id: "system", labelKey: "settings.mode.system", icon: Monitor },
  { id: "light", labelKey: "settings.mode.light", icon: Sun },
  { id: "dark", labelKey: "settings.mode.dark", icon: Moon },
];

function GeneralSection() {
  const { mode, setMode } = useTheme();
  const { lang, setLang, t } = useI18n();
  const language = usePreferencesStore((s) => s.language);
  const setLanguage = usePreferencesStore((s) => s.setLanguage);
  const zoomLevel = usePreferencesStore((s) => s.zoomLevel);
  const setZoomLevel = usePreferencesStore((s) => s.setZoomLevel);
  const showHidden = usePreferencesStore((s) => s.showHidden);
  const setShowHidden = usePreferencesStore((s) => s.setShowHidden);
  const launchAtLogin = usePreferencesStore((s) => s.launchAtLogin);
  const setLaunchAtLogin = usePreferencesStore((s) => s.setLaunchAtLogin);

  // Sync the i18n language with the settings store on mount.
  useEffect(() => {
    if (lang !== language) {
      setLang(language);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Sync the preferences toggle with the actual platform registration
  // on first mount (the stored pref may be stale if the user removed the
  // registration outside the app).
  useEffect(() => {
    IsAutostartEnabled()
      .then((enabled) => {
        if (enabled !== launchAtLogin) setLaunchAtLogin(enabled);
      })
      .catch(() => {});
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="space-y-6">
      {/* Appearance mode */}
      <section className="space-y-3">
        <Label>{t("settings.appearance")}</Label>
        <div className="grid grid-cols-3 gap-2">
          {MODES.map((o) => (
            <Button
              key={o.id}
              variant={mode === o.id ? "default" : "outline"}
              onClick={() => setMode(o.id)}
              className="h-16 flex-col gap-1.5"
            >
              <o.icon className="h-5 w-5" />
              <span className="text-xs">{t(o.labelKey)}</span>
            </Button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground">
          {t("settings.appearance.hint")}
        </p>
      </section>

      <Separator />

      {/* Language */}
      <section className="space-y-3">
        <Label>{t("settings.language")}</Label>
        <Select
          value={lang}
          onValueChange={(v) => {
            setLang(v as "en" | "zh");
            setLanguage(v as "en" | "zh");
            toast(
              t("toast.languageChanged", { lang: v === "en" ? "English" : "中文" }),
            );
          }}
        >
          <SelectTrigger className="w-full">
            <Languages className="h-4 w-4 mr-2 text-muted-foreground" />
            <SelectValue placeholder={t("settings.language")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="en">English</SelectItem>
            <SelectItem value="zh">中文</SelectItem>
          </SelectContent>
        </Select>
        <p className="text-xs text-muted-foreground">
          {t("settings.language.hint")}
        </p>
      </section>

      <Separator />

      {/* Zoom */}
      <section className="space-y-3">
        <div className="flex items-center justify-between">
          <Label>{t("settings.zoom")}</Label>
          <span className="text-sm tabular-nums text-muted-foreground">
            {Math.round(zoomLevel * 100)}%
          </span>
        </div>
        <Slider
          min={0.5}
          max={2}
          step={0.05}
          value={[zoomLevel]}
          onValueChange={(v) => setZoomLevel(v[0])}
        />
      </section>

      <Separator />

      {/* File browser */}
      <section className="space-y-3">
        <Label>{t("settings.fileBrowser")}</Label>
        <SettingRow
          title={t("settings.showHidden")}
          description={t("settings.showHidden.desc")}
        >
          <Switch checked={showHidden} onCheckedChange={setShowHidden} />
        </SettingRow>
      </section>

      <Separator />

      {/* Startup */}
      <section className="space-y-3">
        <Label>{t("settings.startup")}</Label>
        <SettingRow
          title={t("settings.launchAtLogin")}
          description={t("settings.launchAtLogin.desc")}
        >
          <Switch
            checked={launchAtLogin}
            onCheckedChange={async (v) => {
              setLaunchAtLogin(v);
              try {
                await SetAutostartEnabled(v);
                toast(t(v ? "toast.autostartOn" : "toast.autostartOff"));
              } catch (e) {
                setLaunchAtLogin(!v);
                toast(t("toast.autostartFailed", { error: String(e) }), true);
              }
            }}
          />
        </SettingRow>
      </section>
    </div>
  );
}

// ---------- Themes ----------

function ThemesSection() {
  const { themeId, setThemeId, resolvedMode } = useTheme();
  const { t } = useI18n();
  const themes = listBuiltinThemes();

  return (
    <div className="space-y-3">
      <Label>{t("settings.themes.title")}</Label>
      <p className="text-xs text-muted-foreground">
        {t("settings.themes.hint")}
        {" "}{t("settings.themes.currentMode", { mode: resolvedMode === "dark" ? "深色" : "浅色" })}
      </p>
      <div className="grid grid-cols-2 gap-2">
        {themes.map((th) => {
          const variant =
            th.variants[resolvedMode] ?? th.variants.dark ?? th.variants.light;
          const c: ThemeColors | undefined = variant?.colors;
          const swatchBg = c?.background ?? "var(--background)";
          const swatchFg = c?.foreground ?? "var(--foreground)";
          const swatchAccent = c?.primary ?? c?.accent ?? "var(--primary)";
          const swatchMuted = c?.muted ?? "var(--muted)";
          const selected = themeId === th.id;
          return (
            <button
              key={th.id}
              type="button"
              onClick={() => setThemeId(th.id)}
              className={cn(
                "group flex items-center gap-3 rounded-lg border p-2.5 text-left transition-all",
                selected
                  ? "border-primary ring-1 ring-primary/30"
                  : "border-border hover:border-primary/50",
              )}
            >
              <div
                className="flex h-9 w-12 shrink-0 items-center gap-1 rounded-md border border-border/40"
                style={{ background: swatchBg }}
              >
                <span className="h-6 w-0 flex-1 rounded-sm" style={{ background: swatchAccent }} />
                <span className="h-6 w-0 flex-1 rounded-sm" style={{ background: swatchFg, opacity: 0.7 }} />
                <span className="h-6 w-0 flex-1 rounded-sm" style={{ background: swatchMuted }} />
              </div>
              <div className="flex min-w-0 flex-1 flex-col">
                <span className="truncate text-[12.5px] font-medium">{th.name}</span>
                {th.description ? (
                  <span className="truncate text-[11px] text-muted-foreground">
                    {th.description}
                  </span>
                ) : null}
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

// ---------- Library ----------

function LibrarySection() {
  const { t } = useI18n();
  const [libraryDir, setLibraryDir] = useState("");
  const [loading, setLoading] = useState(true);
  const [scanning, setScanning] = useState(false);
  const [skillCount, setSkillCount] = useState<number | null>(null);

  useEffect(() => {
    GetConfig()
      .then((cfg) => setLibraryDir(cfg.libraryDir || ""))
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const onBrowse = async () => {
    try {
      const path = await BrowseLibrary();
      if (path) {
        setLibraryDir(path);
        await UpdateConfig({ libraryDir: path });
        toast(t("toast.libraryChanged"));
      }
    } catch (e) {
      toast(`Browse failed: ${e}`, true);
    }
  };

  const onOpenFolder = () => {
    if (!libraryDir) return;
    // Use the browser to open the folder — on desktop this works via the OS
    // file protocol. As a fallback, the path is displayed for manual copying.
    window.open(libraryDir, "_blank");
  };

  const onRescan = async () => {
    setScanning(true);
    try {
      const count = await RescanLibrary();
      setSkillCount(count);
      toast(t("toast.libraryScanComplete"));
    } catch (e) {
      toast(t("toast.libraryScanFailed", { error: String(e) }), true);
    } finally {
      setScanning(false);
    }
  };

  return (
    <div className="space-y-6">
      <section className="space-y-3">
        <Label>{t("settings.library.title")}</Label>

        {loading ? (
          <div className="rounded-lg bg-card px-3 py-2.5 text-sm text-muted-foreground">
            加载中…
          </div>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center gap-2 rounded-lg bg-card px-3 py-2.5">
              <Folder className="h-4 w-4 shrink-0 text-muted-foreground" />
              <span className="flex-1 truncate font-mono text-sm">
                {libraryDir || t("settings.library.notConfigured")}
              </span>
            </div>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={onBrowse}>
                <FolderOpen className="h-3.5 w-3.5 mr-1.5" />
                {t("settings.library.browse")}
              </Button>
              {libraryDir && (
                <Button variant="outline" size="sm" onClick={onOpenFolder}>
                  <ExternalLink className="h-3.5 w-3.5 mr-1.5" />
                  {t("settings.library.open")}
                </Button>
              )}
              <Button variant="outline" size="sm" onClick={onRescan} disabled={scanning}>
                <RefreshCw className={cn("h-3.5 w-3.5 mr-1.5", scanning && "animate-spin")} />
                {t("settings.library.rescan")}
              </Button>
            </div>
          </div>
        )}

        <p className="text-xs text-muted-foreground">
          {t("settings.library.path.hint")}
        </p>

        {skillCount !== null && (
          <div className="rounded-lg border border-border/60 bg-card/60 px-3 py-2.5">
            <span className="text-sm text-muted-foreground">
              发现 {skillCount} 个技能
            </span>
          </div>
        )}
      </section>

      <Separator />

      <section className="space-y-3">
        <p className="text-xs text-muted-foreground">
          {t("settings.library.rescan.hint")}
        </p>
      </section>
    </div>
  );
}

// ---------- About ----------

function AboutSection() {
  const { t } = useI18n();
  const [version, setVersion] = useState("");
  const [commit, setCommit] = useState("…");
  const [buildTime, setBuildTime] = useState("…");
  const [checking, setChecking] = useState(false);
  const [updateState, setUpdateState] = useState("idle");
  const [hasUpdate, setHasUpdate] = useState(false);
  const [updateVersion, setUpdateVersion] = useState("");
  const [checked, setChecked] = useState(false);

  const devBuild = isDevBuild(version);

  useEffect(() => {
    GetInfo()
      .then((info) => {
        setVersion(info.version);
        setCommit(info.commit);
        setBuildTime(info.buildTime);
      })
      .catch(() => {});
    UpdateState()
      .then(setUpdateState)
      .catch(() => {});
  }, []);

  const onCheck = async () => {
    setChecking(true);
    try {
      const rel = await CheckForUpdates();
      if (rel) {
        setHasUpdate(true);
        setUpdateVersion(rel.version || "");
        toast(`${t("settings.about.updateAvailable")} ${rel.version}`);
      } else {
        setHasUpdate(false);
        toast(t("settings.about.upToDate"));
      }
    } catch {
      setHasUpdate(false);
      toast("此构建未配置更新源", false);
    } finally {
      setChecking(false);
      setChecked(true);
    }
  };

  const onInstall = async () => {
    try {
      await CheckAndInstall();
    } catch (e) {
      toast(`更新失败: ${e}`, true);
    }
  };

  const onRestart = async () => {
    try {
      await Restart();
    } catch (e) {
      toast(`重启失败: ${e}`, true);
    }
  };

  return (
    <div className="space-y-6">
      {/* Project / brand card */}
      <div className="flex items-center gap-4 rounded-xl border border-border/60 bg-card/60 p-4">
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-accent text-primary-foreground">
          <Sparkles className="h-5 w-5" />
        </div>
        <div className="flex min-w-0 flex-col">
          <span className="text-[15px] font-semibold tracking-tight">
            AI-Manager
          </span>
          <span className="text-[11px] text-muted-foreground">
            Agent 技能管理器
          </span>
          <span className="mt-0.5 font-mono text-[11px] text-muted-foreground">
            v{version || "—"}
          </span>
        </div>
      </div>

      <Separator />

      {/* Build info */}
      <section className="space-y-3">
        <Label>{t("settings.about.title")}</Label>
        <dl className="grid grid-cols-[110px_1fr] gap-y-2.5 text-[12.5px]">
          <dt className="text-muted-foreground">{t("settings.about.version")}</dt>
          <dd className="font-mono">{version}</dd>
          <dt className="text-muted-foreground">{t("settings.about.commit")}</dt>
          <dd className="truncate font-mono text-[11.5px]">{commit}</dd>
          <dt className="text-muted-foreground">{t("settings.about.buildTime")}</dt>
          <dd className="font-mono text-[11.5px]">{buildTime}</dd>
        </dl>
      </section>

      <Separator />

      {/* Sources */}
      <section className="space-y-3">
        <Label>项目来源</Label>
        <div className="space-y-2">
          {SOURCES.map((s) => (
            <div
              key={s.name}
              className="flex items-center justify-between gap-3 rounded-lg bg-card px-3 py-2.5"
            >
              <div className="min-w-0">
                <div className="text-sm font-medium">{s.name}</div>
                <div className="truncate text-xs text-muted-foreground">
                  {s.desc}
                </div>
              </div>
              <span className="shrink-0 font-mono text-[10.5px] text-muted-foreground">
                {s.repo}
              </span>
            </div>
          ))}
        </div>
      </section>

      <Separator />

      {/* Updates */}
      <section className="space-y-3">
        <Label>{t("settings.about.updates")}</Label>
        {devBuild ? (
          <p className="text-xs text-muted-foreground">
            当前为开发构建，自动更新未启用。发布构建（带版本号）后即可检查更新。
          </p>
        ) : checked && !hasUpdate && updateState !== "ready" ? (
          <p className="text-xs text-muted-foreground">{t("settings.about.upToDate")}</p>
        ) : (
          <p className="text-xs text-muted-foreground">
            当前状态：{updateState}
          </p>
        )}
        {!devBuild && (
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={onCheck}
              disabled={checking}
            >
              {checking ? t("settings.about.checking") : t("settings.about.checkForUpdates")}
            </Button>
            {hasUpdate && (
              <Button size="sm" onClick={onInstall}>
                {t("settings.about.installUpdate")} {updateVersion}
              </Button>
            )}
            {updateState === "ready" && (
              <Button size="sm" onClick={onRestart}>
                {t("settings.about.restartToUpdate")}
              </Button>
            )}
          </div>
        )}
      </section>
    </div>
  );
}

// ---------- shared ----------

function SettingRow({
  title,
  description,
  children,
}: {
  title: string;
  description: string;
  children: React.ReactNode;
}) {
  return (
    <div className="flex items-center justify-between gap-4 rounded-lg bg-card px-3 py-2.5">
      <div className="min-w-0">
        <div className="text-sm font-medium">{title}</div>
        <div className="mt-0.5 text-xs text-muted-foreground">{description}</div>
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  );
}
