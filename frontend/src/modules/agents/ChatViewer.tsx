import { useMemo, useState, useCallback } from "react";
import {
  User,
  Bot,
  Brain,
  Wrench,
  CheckCircle,
  AlertCircle,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { Badge } from "@/components/ui/badge";
import { parseSessionMessages, type ChatMessage } from "./chatParser";

// ---------------------------------------------------------------------------
// 可折叠行组件（参考 DSH DisclosureRow 设计）
// ---------------------------------------------------------------------------

function DisclosureRow({
  icon,
  title,
  badge,
  summary,
  children,
  defaultOpen = false,
  variant = "default",
}: {
  icon: React.ReactNode;
  title: string;
  badge?: string;
  summary?: string;
  children: React.ReactNode;
  defaultOpen?: boolean;
  variant?: "default" | "thinking" | "tool-call" | "tool-result-error" | "tool-result-ok";
}) {
  const [open, setOpen] = useState(defaultOpen);

  const variantClasses = {
    default: "bg-muted/30 border-muted/50",
    thinking: "bg-amber-50 dark:bg-amber-950/20 border-amber-200/50 dark:border-amber-800/30",
    "tool-call": "bg-blue-50 dark:bg-blue-950/20 border-blue-200/50 dark:border-blue-800/30",
    "tool-result-error": "bg-red-50 dark:bg-red-950/20 border-red-200/50 dark:border-red-800/30",
    "tool-result-ok": "bg-green-50 dark:bg-green-950/20 border-green-200/50 dark:border-green-800/30",
  };

  return (
    <div className={cn("rounded-md border", variantClasses[variant])}>
      <button
        type="button"
        onClick={() => setOpen(!open)}
        className="flex w-full items-center gap-2 px-3 py-2 text-left hover:bg-black/5 dark:hover:bg-white/5"
      >
        <span className="flex-shrink-0">{icon}</span>
        <span className="flex-1 text-xs font-medium truncate">{title}</span>
        {badge && (
          <Badge variant="outline" className="text-[10px] h-4 px-1.5 shrink-0">
            {badge}
          </Badge>
        )}
        {open ? (
          <ChevronUp className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
        ) : (
          <ChevronDown className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
        )}
      </button>

      {/* 折叠时显示摘要 */}
      {!open && summary && (
        <div className="px-3 pb-2">
          <p className="text-[11px] text-muted-foreground truncate">{summary}</p>
        </div>
      )}

      {/* 展开时显示内容 */}
      {open && (
        <div className="border-t px-3 py-2">
          {children}
        </div>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// 时间格式化工具
// ---------------------------------------------------------------------------

function formatTime(ts?: string): string {
  if (!ts) return "";
  try {
    const d = new Date(ts);
    if (isNaN(d.getTime())) return "";
    return d.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
  } catch {
    return "";
  }
}

// ---------------------------------------------------------------------------
// 单条消息渲染
// ---------------------------------------------------------------------------

function MessageBubble({ msg }: { msg: ChatMessage }) {
  const timeStr = formatTime(msg.timestamp);

  // --- 思考内容（可折叠）---
  if (msg.role === "thinking") {
    const firstLine = msg.content.split("\n")[0]?.slice(0, 80) || "Thinking...";
    return (
      <DisclosureRow
        icon={<Brain className="h-4 w-4 text-amber-500" />}
        title="Thinking"
        variant="thinking"
        summary={firstLine}
      >
        <pre className="text-xs font-mono whitespace-pre-wrap break-words text-amber-800 dark:text-amber-300">
          {msg.content}
        </pre>
      </DisclosureRow>
    );
  }

  // --- 工具调用（可折叠）---
  if (msg.role === "tool_call") {
    const argsPreview = msg.toolArgs?.slice(0, 60) || "";
    return (
      <DisclosureRow
        icon={<Wrench className="h-4 w-4 text-blue-600" />}
        title={msg.toolName || "tool"}
        badge="Call"
        variant="tool-call"
        summary={argsPreview}
      >
        {msg.toolArgs && (
          <pre className="text-[11px] font-mono whitespace-pre-wrap break-words text-blue-800 dark:text-blue-300 max-h-48 overflow-y-auto">
            {formatJSON(msg.toolArgs)}
          </pre>
        )}
      </DisclosureRow>
    );
  }

  // --- 工具结果（可折叠）---
  if (msg.role === "tool_result") {
    const isError = msg.content.toLowerCase().includes("error") || msg.content.toLowerCase().includes("exit code: 1");
    const preview = msg.content.slice(0, 80);
    return (
      <DisclosureRow
        icon={isError
          ? <AlertCircle className="h-4 w-4 text-red-500" />
          : <CheckCircle className="h-4 w-4 text-green-600" />}
        title={isError ? "Error" : "Result"}
        badge="Output"
        variant={isError ? "tool-result-error" : "tool-result-ok"}
        summary={preview}
      >
        <pre className="text-[11px] font-mono whitespace-pre-wrap break-words text-muted-foreground max-h-64 overflow-y-auto">
          {truncateContent(msg.content, 3000)}
        </pre>
      </DisclosureRow>
    );
  }

  // --- 系统消息 ---
  if (msg.role === "system") {
    return (
      <div className="flex gap-2 px-3 py-2 rounded-md bg-muted/50 border border-muted">
        <div className="flex-1">
          <pre className="text-[11px] font-mono whitespace-pre-wrap break-words text-muted-foreground">
            {msg.content}
          </pre>
          {timeStr && <span className="text-[10px] text-muted-foreground/60">{timeStr}</span>}
        </div>
      </div>
    );
  }

  // --- 用户消息 ---
  if (msg.role === "user") {
    return (
      <div className="flex gap-2 px-3 py-2 rounded-md bg-primary/5 dark:bg-primary/10 border border-primary/20">
        <div className="flex-shrink-0 mt-0.5">
          <User className="h-4 w-4 text-primary" />
        </div>
        <div className="flex-1 min-w-0">
          <pre className="text-xs whitespace-pre-wrap break-words">{msg.content}</pre>
          {timeStr && <span className="text-[10px] text-muted-foreground/60">{timeStr}</span>}
        </div>
      </div>
    );
  }

  // --- 助手消息 ---
  return (
    <div className="flex gap-2 px-3 py-2 rounded-md bg-muted/30 border border-muted/50">
      <div className="flex-shrink-0 mt-0.5">
        <Bot className="h-4 w-4 text-muted-foreground" />
      </div>
      <div className="flex-1 min-w-0">
        <pre className="text-xs whitespace-pre-wrap break-words">{msg.content}</pre>
        {timeStr && <span className="text-[10px] text-muted-foreground/60">{timeStr}</span>}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// 主组件
// ---------------------------------------------------------------------------

export default function ChatViewer({ raw }: { raw: string }) {
  const messages = useMemo(() => parseSessionMessages(raw), [raw]);

  if (messages.length === 0) {
    return (
      <div className="flex items-center justify-center h-full text-xs text-muted-foreground">
        无法解析会话内容
      </div>
    );
  }

  return (
    <div className="space-y-1">
      {messages.map((msg) => (
        <MessageBubble key={msg.id} msg={msg} />
      ))}
    </div>
  );
}

// ---------------------------------------------------------------------------
// 辅助函数
// ---------------------------------------------------------------------------

function formatJSON(json: string): string {
  try {
    return JSON.stringify(JSON.parse(json), null, 2);
  } catch {
    return json;
  }
}

function truncateContent(content: string, maxLen: number): string {
  if (content.length <= maxLen) return content;
  return content.slice(0, maxLen) + `\n... (${content.length - maxLen} chars truncated)`;
}
