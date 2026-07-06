import { Summary } from "./types";
import { asArray, asBoolean, asNumber, asRecord, asString, asStringArray } from "./types/guards";

export async function fetchMetrics(): Promise<Summary> {
  const res = await fetch("/api/metrics");
  if (!res.ok) throw new Error("failed to load metrics");
  const raw = await res.json();
  return normalizeSummary(raw);
}

export async function isAIEnabled(): Promise<boolean> {
  try {
    const res = await fetch("/api/ai/enabled");
    const data = asRecord(await res.json());
    return asBoolean(data.enabled);
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
    onError("AI unavailable");
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
        const parsed = asString(JSON.parse(data));
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

/** Normalizes API payload into the frontend Summary contract. */
export function normalizeSummary(raw: unknown): Summary {
  const obj = asRecord(raw);
  const modulesRaw = asArray(obj.modules);
  const modules = modulesRaw.map((mod): Summary["modules"][number] => {
    const m = asRecord(mod);
    return {
      module: String(m.module ?? ""),
      path: String(m.path ?? ""),
      language: String(m.language ?? ""),
      dependencies: asStringArray(m.dependencies),
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

  const connascenceRaw = asArray(obj.connascence);
  const connascence = connascenceRaw.map((rel): Summary["connascence"][number] => {
    const c = asRecord(rel);
    const kind = asConnascenceKind(c.kind);
    return {
      kind,
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

function asConnascenceKind(v: unknown): Summary["connascence"][number]["kind"] {
  return v === "meaning" ? "meaning" : "name";
}