import { Summary } from "../../types";
import { compareModules, formatSigned } from "../metrics";
import { Snapshot } from "../types";
import { requiredEl } from "../ui";

export function renderActivity(summary: Summary, previousSnapshot: Snapshot | null): void {
  const activity = requiredEl<HTMLElement>("activity");
  const moduleDeltas = compareModules(summary.modules, previousSnapshot);
  const events = [
    ...moduleDeltas
      .filter((item) => (item.deltaScore ?? 0) > 0)
      .sort((a, b) => (b.deltaScore ?? 0) - (a.deltaScore ?? 0))
      .slice(0, 4)
      .map((item) => ({
        title: `${item.module} regressed`,
        detail: `Δ risk ${formatSigned(item.deltaScore, 1)} | D ${item.currentDistance.toFixed(3)}`,
        tone: "delta-up",
      })),
    ...moduleDeltas
      .filter((item) => (item.deltaScore ?? 0) < 0)
      .sort((a, b) => (a.deltaScore ?? 0) - (b.deltaScore ?? 0))
      .slice(0, 3)
      .map((item) => ({
        title: `${item.module} improved`,
        detail: `Δ risk ${formatSigned(item.deltaScore, 1)} | Max complexity ${item.currentMaxComplexity}`,
        tone: "delta-down",
      })),
  ].slice(0, 7);

  activity.innerHTML = `
    <header><h3>Comparative activity</h3><small>${previousSnapshot ? "Based on the latest scan" : "Waiting for a baseline to compare"}</small></header>
    ${events.length === 0 ? `<p class="hint">No comparable changes yet.</p>` : `
      <ul class="activity-list">
        ${events.map((event) => `<li><strong class="${event.tone}">${event.title}</strong><small>${event.detail}</small></li>`).join("")}
      </ul>
    `}
  `;
}
