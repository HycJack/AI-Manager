import { Layers } from "lucide-react";
import type { ComponentType } from "react";
import {
  ProviderAntigravityIcon,
  ProviderClaudeIcon,
  ProviderCodexIcon,
  ProviderCopilotIcon,
  ProviderCursorIcon,
  ProviderDroidIcon,
  ProviderGrokIcon,
  ProviderOpencodeIcon,
  ProviderPiIcon,
} from "./agents_icons";

/**
 * Agent metadata, aligned with kitter's AGENT_ICON_ORDER.
 *
 * The set mirrors the backend `internal/agents` kinds (universal + 9
 * providers). `universal` has no provider mark of its own in kitter — it is
 * the shared `.agents/skills` directory consumed by many agents — so we use a
 * neutral layered icon as its mark.
 */
export type AgentKey =
  | "universal"
  | "claude-code"
  | "codex"
  | "cursor"
  | "opencode"
  | "pi"
  | "grok"
  | "antigravity"
  | "droid"
  | "copilot"
  | "hermes"
  | "cc-switch";

export interface AgentIconProps {
  size?: number;
  className?: string;
  "aria-hidden"?: string;
}

export type AgentIconType = ComponentType<AgentIconProps>;

export interface AgentMeta {
  key: AgentKey;
  label: string;
  icon: AgentIconType;
}

export const AGENT_META: AgentMeta[] = [
  { key: "universal", label: "Universal", icon: Layers as unknown as AgentIconType },
  { key: "claude-code", label: "Claude Code", icon: ProviderClaudeIcon },
  { key: "codex", label: "Codex", icon: ProviderCodexIcon },
  { key: "cursor", label: "Cursor", icon: ProviderCursorIcon },
  { key: "opencode", label: "OpenCode", icon: ProviderOpencodeIcon },
  { key: "pi", label: "Pi", icon: ProviderPiIcon },
  { key: "grok", label: "Grok", icon: ProviderGrokIcon },
  { key: "antigravity", label: "Antigravity", icon: ProviderAntigravityIcon },
  { key: "droid", label: "Droid", icon: ProviderDroidIcon },
  { key: "copilot", label: "GitHub Copilot", icon: ProviderCopilotIcon },
  { key: "hermes", label: "Hermes", icon: Layers as unknown as AgentIconType },
  { key: "cc-switch", label: "CC-Switch", icon: Layers as unknown as AgentIconType },
];

const META_BY_KEY = new Map(AGENT_META.map((meta) => [meta.key, meta]));

/** Resolve a key (possibly unknown/case-variant from backend data) to metadata. */
export function agentMeta(key: string): AgentMeta | undefined {
  const normalized = key.toLowerCase() as AgentKey;
  return META_BY_KEY.get(normalized);
}

export function agentLabel(key: string): string {
  return agentMeta(key)?.label ?? key;
}

/**
 * Agent mark used across the UI (project rows, install targets, badges).
 * Brand marks keep their original colors; monochrome marks inherit the
 * current text color via `currentColor`, so they follow the theme.
 */
export function AgentIcon({
  agent,
  size = 16,
  className,
}: {
  agent: string;
  size?: number;
  className?: string;
}) {
  const meta = agentMeta(agent);
  if (!meta) {
    return null;
  }
  const Icon = meta.icon;
  return <Icon size={size} className={className} aria-hidden="true" />;
}
