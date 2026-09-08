import { useEffect, useState } from "react";
import { Monitor, Moon, Sun, Palette, Settings2, Info } from "lucide-react";
import { cn } from "@/lib/utils";
import { usePreferencesStore } from "@/modules/settings/store";
import { useTheme, listBuiltinThemes, type ThemeColors } from "@/modules/theme";
import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { Slider } from "@/components/ui/slider";
import { Switch } from "@/components/ui/switch";
import { Separator } from "@/components/ui/separator";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { toast } from "@/lib/toast";
import { GetInfo } from "@bindings/ai-manager/internal/app/versionservice";
import { CheckForUpdates, CheckAndInstall, Restart, UpdateState } from "@bindings/ai-manager/internal/app/updateservice";
import { IsEnabled as IsAutostartEnabled, SetEnabled as SetAutostartEnabled } from "@bindings/ai-manager/internal/app/autostartservice";

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

const MODES: { id: "system" | "light" | "dark"; label: string; icon: typeof Sun }[] = [
  { id: "system", label: "System", icon: Monitor },
  { id: "light", label: "Light", icon: Sun },
  { id: "dark", label: "Dark", icon: Moon },
];

export default function SettingsPage() {
  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold">设置</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          外观、主题与应用偏好（参考 terax-clone 的 General + Themes 配置）。
        </p>
      </div>

      <Tabs defaultValue="general">
        <TabsList className="w-full">
          <TabsTrigger value="general" className="flex-1 gap-1.5">
            <Settings2 className="h-4 w-4" />
            通用
          </TabsTrigger>
          <TabsTrigger value="themes" className="flex-1 gap-1.5">
            <Palette className="h-4 w-4" />
            主题
          </TabsTrigger>
          <TabsTrigger value="about" className="flex-1 gap-1.5">
            <Info className="h-4 w-4" />
            关于
          </TabsTrigger>
        </TabsList>

        <TabsContent value="general" className="mt-6">
          <GeneralSection />
        </TabsContent>
        <TabsContent value="themes" className="mt-6">
          <ThemesSection />
        </TabsContent>
        <TabsContent value="about" className="mt-6">
          <AboutSection />
        </TabsContent>
      </Tabs>
    </div>
  );
}

// ---------- General ----------

function GeneralSection() {
  const { mode, setMode } = useTheme();
  const zoomLevel = usePreferencesStore((s) => s.zoomLevel);
  const setZoomLevel = usePreferencesStore((s) => s.setZoomLevel);
  const showHidden = usePreferencesStore((s) => s.showHidden);
  const setShowHidden = usePreferencesStore((s) => s.setShowHidden);
  const launchAtLogin = usePreferencesStore((s) => s.launchAtLogin);
  const setLaunchAtLogin = usePreferencesStore((s) => s.setLaunchAtLogin);

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
        <Label>外观模式</Label>
        <div className="grid grid-cols-3 gap-2">
          {MODES.map((o) => (
            <Button
              key={o.id}
              variant={mode === o.id ? "default" : "outline"}
              onClick={() => setMode(o.id)}
              className="h-16 flex-col gap-1.5"
            >
              <o.icon className="h-5 w-5" />
              <span className="text-xs">{o.label}</span>
            </Button>
          ))}
        </div>
        <p className="text-xs text-muted-foreground">
          主题颜色与背景在「主题」页切换。
        </p>
      </section>

      <Separator />

      {/* Zoom */}
      <section className="space-y-3">
        <div className="flex items-center justify-between">
          <Label>界面缩放</Label>
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
        <Label>文件浏览器</Label>
        <SettingRow
          title="显示隐藏文件"
          description="在文件列表与搜索中显示点开头的文件（.env、.gitignore 等）。"
        >
          <Switch
            checked={showHidden}
            onCheckedChange={setShowHidden}
          />
        </SettingRow>
      </section>

      <Separator />

      {/* Startup */}
      <section className="space-y-3">
        <Label>启动</Label>
        <SettingRow
          title="开机自启"
          description="登录系统时自动打开应用。"
        >
          <Switch
            checked={launchAtLogin}
            onCheckedChange={async (v) => {
              setLaunchAtLogin(v);
              try {
                await SetAutostartEnabled(v);
                toast(v ? "已设置开机自启" : "已关闭开机自启");
              } catch (e) {
                setLaunchAtLogin(!v); // revert on failure
                toast(`设置失败: ${e}`, true);
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
  const themes = listBuiltinThemes();

  return (
    <div className="space-y-3">
      <Label>主题</Label>
      <p className="text-xs text-muted-foreground">
        点击选择主题并立即应用。当前模式：{resolvedMode === "dark" ? "深色" : "浅色"}。
      </p>
      <div className="grid grid-cols-2 gap-2">
        {themes.map((t) => {
          const variant =
            t.variants[resolvedMode] ?? t.variants.dark ?? t.variants.light;
          const c: ThemeColors | undefined = variant?.colors;
          const swatchBg = c?.background ?? "var(--background)";
          const swatchFg = c?.foreground ?? "var(--foreground)";
          const swatchAccent = c?.primary ?? c?.accent ?? "var(--primary)";
          const swatchMuted = c?.muted ?? "var(--muted)";
          const selected = themeId === t.id;
          return (
            <button
              key={t.id}
              type="button"
              onClick={() => setThemeId(t.id)}
              className={cn(
                "group flex items-center gap-3 rounded-lg border p-2.5 text-left transition-all",
                selected
                  ? "border-primary ring-1 ring-primary/30"
                  : "border-border hover:border-primary/50"
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
                <span className="truncate text-[12.5px] font-medium">{t.name}</span>
                {t.description ? (
                  <span className="truncate text-[11px] text-muted-foreground">
                    {t.description}
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

// ---------- About ----------

function AboutSection() {
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
        toast(`发现新版本 ${rel.version}`);
      } else {
        setHasUpdate(false);
        toast("已是最新版本");
      }
    } catch {
      // Dev builds (no update repo configured) reject the check; surface that
      // as a hint rather than a scary error toast.
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
        <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br from-primary to-accent text-sm font-extrabold text-primary-foreground">
          W
        </div>
        <div className="flex min-w-0 flex-col">
          <span className="text-[15px] font-semibold tracking-tight">
            Wails3 Skeleton
          </span>
          <span className="text-[11px] text-muted-foreground">
            从开源项目抽象出的桌面应用模板
          </span>
          <span className="mt-0.5 font-mono text-[11px] text-muted-foreground">
            v{version || "—"}
          </span>
        </div>
      </div>

      <Separator />

      {/* Build info */}
      <section className="space-y-3">
        <Label>构建信息</Label>
        <dl className="grid grid-cols-[110px_1fr] gap-y-2.5 text-[12.5px]">
          <dt className="text-muted-foreground">版本</dt>
          <dd className="font-mono">{version}</dd>
          <dt className="text-muted-foreground">提交</dt>
          <dd className="truncate font-mono text-[11.5px]">{commit}</dd>
          <dt className="text-muted-foreground">构建时间</dt>
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
        <Label>更新</Label>
        {devBuild ? (
          <p className="text-xs text-muted-foreground">
            当前为开发构建，自动更新未启用。发布构建（带版本号）后即可检查更新。
          </p>
        ) : checked && !hasUpdate && updateState !== "ready" ? (
          <p className="text-xs text-muted-foreground">已是最新版本。</p>
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
              {checking ? "检查中…" : "检查更新"}
            </Button>
            {hasUpdate && (
              <Button size="sm" onClick={onInstall}>
                下载并安装 {updateVersion}
              </Button>
            )}
            {updateState === "ready" && (
              <Button size="sm" onClick={onRestart}>
                重启以完成更新
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
