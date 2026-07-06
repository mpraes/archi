import { Summary } from "../types";
import { Snapshot } from "./types";

const SNAPSHOT_KEY = "archi:summary:last-v1";

export function buildSnapshot(summary: Summary): Snapshot {
  const modules: Snapshot["modules"] = {};
  for (const module of summary.modules) {
    modules[module.module] = {
      distance: module.distance,
      instability: module.instability,
      maxComplexity: module.maxComplexity,
      totalComplexity: module.totalComplexity,
      orphanCount: module.orphanBlocks.length,
      godCount: module.godBlocks.length,
      efferent: module.efferent,
    };
  }

  const avgDistance = summary.modules.reduce((sum, module) => sum + module.distance, 0) / Math.max(1, summary.modules.length);
  const avgInstability = summary.modules.reduce((sum, module) => sum + module.instability, 0) / Math.max(1, summary.modules.length);
  const avgMaxComplexity = summary.modules.reduce((sum, module) => sum + module.maxComplexity, 0) / Math.max(1, summary.modules.length);

  return {
    capturedAt: new Date().toISOString(),
    projectName: summary.projectName,
    moduleCount: summary.moduleCount,
    hotspotCount: summary.hotspots.length,
    connascenceCount: summary.connascence.length,
    avgDistance,
    avgInstability,
    avgMaxComplexity,
    modules,
  };
}

export function readPreviousSnapshot(): Snapshot | null {
  try {
    const raw = window.localStorage.getItem(SNAPSHOT_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Snapshot;
    if (!parsed || typeof parsed !== "object" || !parsed.modules) return null;
    return parsed;
  } catch {
    return null;
  }
}

export function saveSnapshot(snapshot: Snapshot): void {
  try {
    window.localStorage.setItem(SNAPSHOT_KEY, JSON.stringify(snapshot));
  } catch {
    // Ignore write failures (private mode or blocked storage).
  }
}
