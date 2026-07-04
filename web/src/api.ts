import { Summary } from "./types";

export async function fetchMetrics(): Promise<Summary> {
  const res = await fetch("/api/metrics");
  if (!res.ok) throw new Error("falha ao carregar métricas");
  return res.json();
}

export async function isAIEnabled(): Promise<boolean> {
  try {
    const res = await fetch("/api/ai/enabled");
    const data = await res.json();
    return data.enabled === true;
  } catch {
    return false;
  }
}

export async function streamAIInsights(
  onChunk: (text: string) => void,
  onDone: () => void,
  onError: (msg: string) => void
): Promise<void> {
  const res = await fetch("/api/ai/insights");
  if (!res.ok || !res.body) {
    onError("IA indisponível");
    return;
  }
  const reader = res.body.getReader();
  const decoder = new TextDecoder();
  let buffer = "";
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    buffer += decoder.decode(value, { stream: true });
    const events = buffer.split("\n\n");
    buffer = events.pop() ?? "";
    for (const block of events) {
      const lines = block.split("\n");
      let event = "chunk";
      let data = "";
      for (const line of lines) {
        if (line.startsWith("event: ")) event = line.slice(7);
        else if (line.startsWith("data: ")) data += line.slice(6);
      }
      try {
        const parsed = JSON.parse(data);
        if (event === "chunk") onChunk(parsed);
        else if (event === "error") onError(parsed);
        else if (event === "done") onDone();
      } catch {
        /* ignore malformed */
      }
    }
  }
  onDone();
}