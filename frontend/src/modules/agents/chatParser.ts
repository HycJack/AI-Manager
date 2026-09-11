/**
 * chatParser.ts — 将 Claude Code / Codex JSONL session 解析为结构化聊天消息
 */

// ---------------------------------------------------------------------------
// 类型定义
// ---------------------------------------------------------------------------

export interface ChatMessage {
  id: string;
  role: "user" | "assistant" | "thinking" | "tool_call" | "tool_result" | "system";
  content: string;
  /** 工具调用名称（role === "tool_call" 时） */
  toolName?: string;
  /** 工具调用参数 JSON 字符串（role === "tool_call" 时） */
  toolArgs?: string;
  /** 时间戳 */
  timestamp?: string;
}

// ---------------------------------------------------------------------------
// 解析入口
// ---------------------------------------------------------------------------

/**
 * 自动检测格式并解析 JSONL 文本为聊天消息列表。
 * 支持 Claude Code / Pi JSONL 和 Codex JSONL 两种格式。
 */
export function parseSessionMessages(raw: string): ChatMessage[] {
  const lines = raw.split("\n").filter((l) => l.trim());
  if (lines.length === 0) return [];

  // 检测格式：第一行是否包含 "session_meta"（Codex）或 "response_item"（Codex）
  // 或 "session"（Pi）
  const firstLine = lines[0];
  const isCodex = firstLine.includes('"session_meta"') || firstLine.includes('"response_item"');

  if (isCodex) {
    return parseCodexFormat(lines);
  }
  // Pi 和 Claude Code 使用类似的格式（message type + message.role）
  return parseClaudeFormat(lines);
}

// ---------------------------------------------------------------------------
// Claude Code JSONL 解析
// ---------------------------------------------------------------------------

function parseClaudeFormat(lines: string[]): ChatMessage[] {
  const messages: ChatMessage[] = [];
  let idCounter = 0;

  for (const line of lines) {
    try {
      const obj = JSON.parse(line);
      const type = obj.type;

      // 跳过非消息类型（Claude Code + Pi）
      if (["queue-operation", "attachment", "skill_listing", "agent_listing_delta",
        "last-prompt", "session", "model_change", "thinking_level_change",
        "system", "error"].includes(type)) continue;

      const ts = obj.timestamp;

      // user / assistant / message 消息（Claude Code + Pi）
      if (type === "user" || type === "assistant" || type === "message") {
        const msg = obj.message;
        if (!msg) continue;

        const role = msg.role === "user" ? "user" : "assistant";
        const contentItems = msg.content;

        if (typeof contentItems === "string") {
          messages.push({ id: `m${idCounter++}`, role, content: contentItems, timestamp: ts });
        } else if (Array.isArray(contentItems)) {
          for (const item of contentItems) {
            if (item.type === "text") {
              messages.push({ id: `m${idCounter++}`, role, content: item.text ?? "", timestamp: ts });
            } else if (item.type === "thinking") {
              messages.push({ id: `m${idCounter++}`, role: "thinking", content: item.thinking ?? "", timestamp: ts });
            } else if (item.type === "tool_use") {
              messages.push({
                id: `m${idCounter++}`, role: "tool_call",
                content: item.name ?? "tool",
                toolName: item.name,
                toolArgs: JSON.stringify(item.input ?? {}),
                timestamp: ts,
              });
            } else if (item.type === "tool_result") {
              messages.push({
                id: `m${idCounter++}`, role: "tool_result",
                content: typeof item.content === "string" ? item.content : JSON.stringify(item.content ?? ""),
                timestamp: ts,
              });
            }
          }
        }
      }

      // text 类型（独立消息）
      if (type === "text") {
        messages.push({ id: `m${idCounter++}`, role: "assistant", content: obj.text ?? obj.content ?? "", timestamp: ts });
      }
    } catch {
      // 跳过无法解析的行
    }
  }
  return messages;
}

// ---------------------------------------------------------------------------
// Codex JSONL 解析
// ---------------------------------------------------------------------------

function parseCodexFormat(lines: string[]): ChatMessage[] {
  const messages: ChatMessage[] = [];
  let idCounter = 0;

  for (const line of lines) {
    try {
      const obj = JSON.parse(line);
      const outerType = obj.type;
      const payload = obj.payload;
      const ts = obj.timestamp;

      // 跳过元数据和非消息类型
      if (outerType === "session_meta" || outerType === "event_msg" || outerType === "turn_context" ||
        outerType === "token_count") continue;

      if (outerType !== "response_item" || !payload) continue;

      const ptype = payload.type;

      // 消息类型
      if (ptype === "message") {
        const role = payload.role === "user" ? "user" : "assistant";
        const contentItems = payload.content;

        if (Array.isArray(contentItems)) {
          for (const item of contentItems) {
            if (item.type === "input_text" || item.type === "output_text") {
              messages.push({ id: `m${idCounter++}`, role, content: item.text ?? "", timestamp: ts });
            }
          }
        } else if (typeof contentItems === "string") {
          messages.push({ id: `m${idCounter++}`, role, content: contentItems, timestamp: ts });
        }
      }

      // 思考/推理
      if (ptype === "reasoning") {
        const contentItems = payload.content;
        if (Array.isArray(contentItems)) {
          for (const item of contentItems) {
            if (item.type === "output_text" || item.type === "input_text") {
              messages.push({ id: `m${idCounter++}`, role: "thinking", content: item.text ?? "", timestamp: ts });
            }
          }
        }
      }

      // 工具调用
      if (ptype === "function_call") {
        messages.push({
          id: `m${idCounter++}`, role: "tool_call",
          content: payload.name ?? "tool",
          toolName: payload.name,
          toolArgs: payload.arguments ?? "",
          timestamp: ts,
        });
      }

      // 工具结果
      if (ptype === "function_call_output" || ptype === "custom_tool_call_output") {
        const output = payload.output ?? payload.content ?? "";
        messages.push({
          id: `m${idCounter++}`, role: "tool_result",
          content: typeof output === "string" ? output : JSON.stringify(output),
          timestamp: ts,
        });
      }

      // 自定义工具调用
      if (ptype === "custom_tool_call") {
        messages.push({
          id: `m${idCounter++}`, role: "tool_call",
          content: payload.name ?? "custom_tool",
          toolName: payload.name,
          toolArgs: payload.arguments ?? "",
          timestamp: ts,
        });
      }

    } catch {
      // 跳过无法解析的行
    }
  }
  return messages;
}
