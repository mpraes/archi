import { Summary } from "./types";

export async function fetchMetrics(): Promise<Summary> {
  const res = await fetch("/api/metrics");
  if (!res.ok) throw new Error("falha ao carregar métricas");
  const raw = await res.json();
  return normalizeSummary(raw);
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

function normalizeSummary(raw: unknown): Summary {
  const obj = (raw ?? {}) as Record<string, unknown>;
  const modulesRaw = Array.isArray(obj.modules) ? obj.modules : [];
  const modules = modulesRaw.map((mod): Summary["modules"][number] => {
    const m = (mod ?? {}) as Record<string, unknown>;
    return {
      module: String(m.module ?? ""),
      path: String(m.path ?? ""),
      files: asNumber(m.files),
      afferent: asNumber(m.afferent),
      efferent: asNumber(m.efferent),
      instability: asNumber(m.instability),
      abstraction: asNumber(m.abstraction),
      distance: asNumber(m.distance),
      maxComplexity: asNumber(m.maxComplexity),
      totalComplexity: asNumber(m.totalComplexity),
      abstracts: asNumber(m.abstracts),
      concretes: asNumber(m.concretes),
      orphanBlocks: asStringArray(m.orphanBlocks),
      godBlocks: asStringArray(m.godBlocks),
    };
  });

  const connascenceRaw = Array.isArray(obj.connascence) ? obj.connascence : [];
  const connascence = connascenceRaw.map((rel) => {
    const c = (rel ?? {}) as Record<string, unknown>;
    return {
      kind: c.kind === "meaning" ? "meaning" : "name",
      from: String(c.from ?? ""),
      to: String(c.to ?? ""),
      detail: String(c.detail ?? ""),
    };
  });

  return {
    projectName: String(obj.projectName ?? ""),
    moduleCount: asNumber(obj.moduleCount),
    modules,
    connascence,
    hotspots: asStringArray(obj.hotspots),
  };
}

function asNumber(v: unknown): number {
  return typeof v === "number" && Number.isFinite(v) ? v : 0;
}

function asStringArray(v: unknown): string[] {
  if (!Array.isArray(v)) return [];
  return v.filter((item): item is string => typeof item === "string");
}