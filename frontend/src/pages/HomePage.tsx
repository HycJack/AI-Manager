import { useEffect, useState } from "react";
import { Greet, Ping } from "@bindings/ai-manager/internal/app/greeterservice";
import { Search } from "@bindings/ai-manager/internal/app/searchservice";
import type { SearchResult } from "@bindings/ai-manager/internal/app/models";
import { toast } from "@/lib/toast";

// HomePage demonstrates the core Wails3 frontend↔backend contract:
//   - calling a bound Go method (request/response)
//   - emitting + receiving a backend event
//
// Replace this page (and GreeterService) with your own UI and services.
export default function HomePage() {
  const [name, setName] = useState("");
  const [greeting, setGreeting] = useState("");
  const [pong, setPong] = useState("");

  const [root, setRoot] = useState("");
  const [fileTypes, setFileTypes] = useState("");
  const [skipDirs, setSkipDirs] = useState("");
  const [results, setResults] = useState<SearchResult[] | null>(null);
  const [searching, setSearching] = useState(false);

  useEffect(() => {
    Ping().then(setPong).catch((e) => toast(String(e), true));
  }, []);

  // Run the backend file search. FileTypes / SkipDirs are comma-separated
  // text (e.g. "go, ts" and "node_modules, .git").
  const runSearch = async (e: React.FormEvent) => {
    e.preventDefault();
    setSearching(true);
    try {
      const opts = {
        root: root.trim(),
        fileTypes: splitList(fileTypes),
        skipDirs: splitList(skipDirs),
        includeHidden: false,
      };
      const found = await Search(opts);
      setResults(found);
    } catch (err) {
      setResults(null);
      toast(String(err), true);
    } finally {
      setSearching(false);
    }
  };

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const msg = await Greet(name);
      setGreeting(msg);
    } catch (err) {
      toast(String(err), true);
    }
  };

  return (
    <div className="mx-auto max-w-xl space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Home</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Wails3 模板骨架。左侧导航 + 右侧内容区；此页演示前后端桥接。
        </p>
      </div>

      <div className="rounded-xl bg-card p-5">
        <div className="mb-4 flex items-center gap-2 text-xs">
          <span
            className={`rounded-full px-2 py-0.5 ${
              pong === "pong"
                ? "bg-emerald-500/15 text-emerald-400"
                : "bg-muted text-muted-foreground"
            }`}
          >
            bridge: {pong || "…"}
          </span>
        </div>

        <form onSubmit={onSubmit} className="flex gap-2">
          <input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="输入你的名字"
            className="flex-1 rounded-lg border border-input bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
          />
          <button
            type="submit"
            className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90"
          >
            Greet
          </button>
        </form>

        {greeting && (
          <p className="mt-4 rounded-lg bg-secondary/50 px-4 py-3 text-sm">
            {greeting}
          </p>
        )}
      </div>

      {/* File search — scoped by directory + file type, with skip dirs. */}
      <div className="rounded-xl bg-card p-5">
        <div className="mb-4">
          <h2 className="text-base font-semibold">文件搜索</h2>
          <p className="mt-0.5 text-xs text-muted-foreground">
            指定目录或文件类型，自动跳过依赖目录（如 node_modules）。
          </p>
        </div>

        <form onSubmit={runSearch} className="space-y-3">
          <div className="grid gap-2 sm:grid-cols-3">
            <label className="flex flex-col gap-1 text-xs text-muted-foreground">
              目录
              <input
                value={root}
                onChange={(e) => setRoot(e.target.value)}
                placeholder="留空使用应用数据目录"
                className="rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
              />
            </label>
            <label className="flex flex-col gap-1 text-xs text-muted-foreground">
              文件类型（逗号分隔）
              <input
                value={fileTypes}
                onChange={(e) => setFileTypes(e.target.value)}
                placeholder="如：go, ts"
                className="rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
              />
            </label>
            <label className="flex flex-col gap-1 text-xs text-muted-foreground">
              跳过的目录（逗号分隔）
              <input
                value={skipDirs}
                onChange={(e) => setSkipDirs(e.target.value)}
                placeholder="默认跳过 node_modules 等"
                className="rounded-lg border border-input bg-background px-3 py-2 text-sm text-foreground outline-none focus:ring-2 focus:ring-ring"
              />
            </label>
          </div>

          <div className="flex items-center gap-3">
            <button
              type="submit"
              disabled={searching}
              className="rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition-opacity hover:opacity-90 disabled:opacity-60"
            >
              {searching ? "搜索中…" : "搜索"}
            </button>
            {results && (
              <span className="text-xs text-muted-foreground">
                找到 {results.length} 个文件
              </span>
            )}
          </div>
        </form>

        {results && (
          <ul className="mt-4 max-h-64 divide-y divide-border/60 overflow-y-auto rounded-lg border border-border/60 bg-background">
            {results.length === 0 ? (
              <li className="px-4 py-3 text-sm text-muted-foreground">
                没有匹配的文件。
              </li>
            ) : (
              results.map((r) => (
                <li
                  key={r.path}
                  className="flex items-center justify-between gap-2 px-4 py-2 text-sm"
                >
                  <span className="truncate font-mono text-[13px]">
                    {r.path}
                  </span>
                  {r.size > 0 && (
                    <span className="shrink-0 text-xs text-muted-foreground">
                      {formatSize(r.size)}
                    </span>
                  )}
                </li>
              ))
            )}
          </ul>
        )}
      </div>
    </div>
  );
}

// Splits a comma-separated string into trimmed, non-empty parts.
function splitList(value: string): string[] {
  return value
    .split(",")
    .map((p) => p.trim())
    .filter(Boolean);
}

// Returns a human-readable file size, e.g. "1.2 MB".
function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let v = bytes / 1024;
  let i = 0;
  while (v >= 1024 && i < units.length - 1) {
    v /= 1024;
    i++;
  }
  return `${v.toFixed(1)} ${units[i]}`;
}
