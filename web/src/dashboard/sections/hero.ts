import { Summary } from "../../types";
import { dominantConnascenceLabel } from "../format";
import { deltaClass } from "../metrics";
import { Snapshot } from "../types";
import { infoButton, requiredEl } from "../ui";

export function renderHero(summary: Summary, previousSnapshot: Snapshot | null): void {
  const hero = requiredEl<HTMLElement>("hero");
  const riskModules = summary.modules.filter((m) => m.godBlocks.length > 0 || m.orphanBlocks.length > 0).length;
  const previousRiskModules = previousSnapshot
    ? Object.values(previousSnapshot.modules).filter((m) => m.godCount > 0 || m.orphanCount > 0).length
    : null;
  const riskDelta = previousRiskModules === null ? null : riskModules - previousRiskModules;
  const riskSummary = riskModules === 1 ? "1 module at risk" : `${riskModules} modules at risk`;
  const baselineSummary =
    riskDelta === null
      ? "No baseline yet"
      : riskDelta > 0
        ? `Attention: +${riskDelta} vs previous run`
        : riskDelta < 0
          ? `Improved by ${Math.abs(riskDelta)} vs previous run`
          : "No change vs previous run";

  hero.innerHTML = `
    <div class="hero-intro">
      <h3>Architecture overview</h3>
      <p>This panel shows where structural risk is concentrating now and where to focus your next refactoring cycle.</p>
      <ul class="hero-tips">
        <li>Start with modules marked as high risk and validate if their dependencies can be reduced.</li>
        <li>Use regressions from the latest scan to prioritize fixes with immediate impact.</li>
        <li>Open module diagnostics to move from high-level risk signals to concrete code actions.</li>
      </ul>
    </div>
    <div class="hero-priority">
      <article class="hero-main-metric">
        <div class="card-head-inline"><span>Current priority</span>${infoButton("How many modules currently show god blocks or orphan code. Use this as your primary triage signal for architecture risk.", "Current priority info")}</div>
        <strong>${riskSummary}</strong>
        <small class="${deltaClass(riskDelta)}">${baselineSummary}</small>
      </article>
      <ul class="hero-secondary">
        <li>
          <span>Dominant connascence</span>
          <b>${dominantConnascenceLabel(summary.connascence)}</b>
        </li>
        <li>
          <span>Analyzed project</span>
          <b>${summary.projectName || "Current project"}</b>
          <small>${new Date().toLocaleDateString("en-US")}</small>
        </li>
      </ul>
    </div>
  `;
}
