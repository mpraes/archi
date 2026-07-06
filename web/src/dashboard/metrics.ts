import { ModuleMetrics } from "../types";
import { MetricDelta, ModuleDelta, ModuleSnapshot, Snapshot } from "./types";

export function moduleRiskScore(module: ModuleMetrics): number {
  return (
    module.distance * 35 +
    module.instability * 12 +
    module.maxComplexity * 1.2 +
    module.totalComplexity * 0.08 +
    module.efferent * 1.5 +
    module.orphanBlocks.length * 8 +
    module.godBlocks.length * 10
  );
}

export function moduleRiskScoreFromSnapshot(module: ModuleSnapshot): number {
  return (
    module.distance * 35 +
    module.instability * 12 +
    module.maxComplexity * 1.2 +
    module.totalComplexity * 0.08 +
    module.efferent * 1.5 +
    module.orphanCount * 8 +
    module.godCount * 10
  );
}

export function compareModules(modules: ModuleMetrics[], previousSnapshot: Snapshot | null): ModuleDelta[] {
  return modules.map((module): ModuleDelta => {
    const previous = previousSnapshot?.modules[module.module] ?? null;
    const currentScore = moduleRiskScore(module);
    const previousScore = previous ? moduleRiskScoreFromSnapshot(previous) : null;

    return {
      module: module.module,
      currentScore,
      previousScore,
      deltaScore: previousScore === null ? null : currentScore - previousScore,
      currentDistance: module.distance,
      previousDistance: previous?.distance ?? null,
      currentMaxComplexity: module.maxComplexity,
      previousMaxComplexity: previous?.maxComplexity ?? null,
    };
  });
}

export function metricDelta(current: number, previous: number | null): MetricDelta {
  return {
    current,
    previous,
    delta: previous === null ? null : current - previous,
  };
}

export function formatSigned(value: number | null, digits: number): string {
  if (value === null) return "n/a";
  const rounded = value.toFixed(digits);
  return `${value > 0 ? "+" : ""}${rounded}`;
}

export function formatDelta(delta: MetricDelta, digits: number): string {
  if (delta.previous === null || delta.delta === null) return "No baseline";
  if (delta.delta > 0) return `Increased ${formatSigned(delta.delta, digits)} vs previous run`;
  if (delta.delta < 0) return `Dropped ${formatSigned(delta.delta, digits)} vs previous run`;
  return "No change vs previous run";
}

export function deltaClass(delta: number | null): string {
  if (delta === null) return "neutral";
  if (delta > 0) return "delta-up";
  if (delta < 0) return "delta-down";
  return "neutral";
}

export function deltaTrendLabel(delta: number | null): string {
  if (delta === null) return "•";
  if (delta > 0) return "↑";
  if (delta < 0) return "↓";
  return "=";
}
