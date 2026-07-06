import { Summary } from "../../types";
import { metricDelta, formatDelta, deltaClass } from "../metrics";
import { buildSnapshot } from "../snapshot";
import { Snapshot } from "../types";
import { infoButton, requiredEl } from "../ui";

export function renderKpiGrid(summary: Summary, previousSnapshot: Snapshot | null): void {
  const kpiGrid = requiredEl<HTMLElement>("kpi-grid");
  const currentSnapshot = buildSnapshot(summary);
  const averageTotalComplexity =
    summary.modules.reduce((acc, mod) => acc + mod.totalComplexity, 0) / Math.max(1, summary.modules.length);

  const cards: Array<{ title: string; value: string; foot: string; info: string; state?: string }> = [
    {
      title: "Hotspots",
      value: String(summary.hotspots.length),
      foot: formatDelta(metricDelta(summary.hotspots.length, previousSnapshot?.hotspotCount ?? null), 0),
      info: "Count of modules that deserve immediate attention due to structural imbalance or complexity pressure.",
      state: deltaClass(metricDelta(summary.hotspots.length, previousSnapshot?.hotspotCount ?? null).delta),
    },
    {
      title: "Average distance",
      value: currentSnapshot.avgDistance.toFixed(3),
      foot: formatDelta(metricDelta(currentSnapshot.avgDistance, previousSnapshot?.avgDistance ?? null), 3),
      info: "Average distance from the main sequence. Higher values usually mean architecture is less balanced.",
      state: deltaClass(metricDelta(currentSnapshot.avgDistance, previousSnapshot?.avgDistance ?? null).delta),
    },
    {
      title: "Average instability",
      value: currentSnapshot.avgInstability.toFixed(3),
      foot: formatDelta(metricDelta(currentSnapshot.avgInstability, previousSnapshot?.avgInstability ?? null), 3),
      info: "Average dependency volatility. Closer to 1 means modules depend heavily on external modules.",
      state: deltaClass(metricDelta(currentSnapshot.avgInstability, previousSnapshot?.avgInstability ?? null).delta),
    },
    {
      title: "Average max complexity",
      value: currentSnapshot.avgMaxComplexity.toFixed(1),
      foot: formatDelta(metricDelta(currentSnapshot.avgMaxComplexity, previousSnapshot?.avgMaxComplexity ?? null), 1),
      info: "Average of each module's peak cyclomatic complexity, useful for spotting difficult-to-maintain code paths.",
      state: deltaClass(metricDelta(currentSnapshot.avgMaxComplexity, previousSnapshot?.avgMaxComplexity ?? null).delta),
    },
    {
      title: "Average total complexity",
      value: averageTotalComplexity.toFixed(1),
      foot: "Global complexity load per module",
      info: "Total cyclomatic complexity aggregated by module and averaged across the project.",
      state: "neutral",
    },
    {
      title: "Connascence links",
      value: String(summary.connascence.length),
      foot: formatDelta(metricDelta(summary.connascence.length, previousSnapshot?.connascenceCount ?? null), 0),
      info: "Detected cross-module connascence relationships that can increase change coordination cost.",
      state: deltaClass(metricDelta(summary.connascence.length, previousSnapshot?.connascenceCount ?? null).delta),
    },
  ];

  kpiGrid.innerHTML = cards
    .map(
      (card) =>
        `<article class="kpi-card"><header class="card-head"><span>${card.title}</span>${infoButton(card.info, `${card.title} info`)}</header><strong>${card.value}</strong><small class="${card.state ?? "neutral"}">${card.foot}</small></article>`,
    )
    .join("");
}
