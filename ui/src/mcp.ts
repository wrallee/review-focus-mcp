import { App } from "@modelcontextprotocol/ext-apps";

export const app = new App({ name: "Review Focus", version: "0.1.0" }, {});

type ToolResult = {
  structuredContent?: unknown;
  content?: Array<{ type: string; text?: string }>;
  isError?: boolean;
};

let lastToolResult: ToolResult | undefined;
const listeners = new Set<(result: ToolResult) => void>();

// Register before connect so the initial host tool result is never lost.
app.ontoolresult = (result) => {
  lastToolResult = result as ToolResult;
  for (const listener of listeners) listener(lastToolResult);
};

export async function connectApp() {
  await app.connect();
}

export function subscribeToolResults(listener: (result: ToolResult) => void): () => void {
  listeners.add(listener);
  if (lastToolResult) listener(lastToolResult);
  return () => {
    listeners.delete(listener);
  };
}

export async function callTool<T>(name: string, args: Record<string, unknown> = {}): Promise<T> {
  const result = await app.callServerTool({ name, arguments: args });
  if (result.isError) {
    const text = result.content?.find((item) => item.type === "text")?.text;
    throw new Error(text ?? `Tool ${name} failed`);
  }
  const structured = result.structuredContent;
  if (structured) return structured as T;
  const text = result.content?.find((item) => item.type === "text")?.text;
  if (!text) throw new Error(`Tool ${name} returned no usable payload`);
  return JSON.parse(text) as T;
}
