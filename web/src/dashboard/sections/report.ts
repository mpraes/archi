import { ModuleMetrics, Summary } from "../../types";
import { formatCapturedAt, languageBreakdown } from "../format";
import { compareModules, deltaClass, deltaTrendLabel, formatSigned, moduleRiskScore } from "../metrics";
import { ModuleDelta, Snapshot } from "../types";
import { escapeAttr, infoButton, requiredEl } from "../ui";

export function renderDetailedReport(
  summary: Summary,
  previousSnapshot: Snapshot | null,
  onModuleSelect: (module: ModuleMetrics) => void,
): void {
  const report = requiredEl<HTMLElement>("report");
  if (summary.modules.length === 0) {
    report.innerHTML = "";
    return;
  }

  const historyLabel = previousSnapshot
    ? `Comparison against previous scan from ${formatCapturedAt(previousSnapshot.capturedAt)}`
    : "No previous baseline: this scan will be used as reference for the next run.";
  const moduleDeltas = compareModules(summary.modules, previousSnapshot);
  const regressions = moduleDeltas
    .filter((d) => (d.deltaScore ?? 0) > 0)
    .sort((a, b) => (b.deltaScore ?? 0) - (a.deltaScore ?? 0))
    .slice(0, 8);
  const improvements = moduleDeltas
    .filter((d) => (d.deltaScore ?? 0) < 0)
    .sort((a, b) => (a.deltaScore ?? 0) - (b.deltaScore ?? 0))
    .slice(0, 8);
  const riskModules = summary.modules.filter((m) => m.godBlocks.length > 0 || m.orphanBlocks.length > 0).length;
  const comparableModules = moduleDeltas.filter((item) => item.deltaScore !== null).length;
  const topActionModules = [...summary.modules].sort((a, b) => moduleRiskScore(b) - moduleRiskScore(a)).slice(0, 6);
  const byDistance = [...summary.modules].sort((a, b) => b.distance - a.distance).slice(0, 5);
  const byComplexity = [...summary.modules].sort((a, b) => b.maxComplexity - a.maxComplexity).slice(0, 5);
  const byLanguage = languageBreakdown(summary.modules);

  report.innerHTML = `
    <header class="report-header">
      <h3>Analytical panel</h3>
      <p class="history">${historyLabel}</p>
    </header>
    <div class="report-priority">
      <article><div class="card-head-inline"><span>Immediate risk</span>${infoButton("Modules that currently contain at least one god block or orphan block. This is your short-term architecture debt queue.", "Immediate risk info")}</div><strong>${riskModules}</strong><small class="neutral">With god block or orphan code</small></article>
      <article><div class="card-head-inline"><span>Regressions</span>${infoButton("Modules whose composite risk score got worse compared to the stored baseline.", "Regressions info")}</div><strong>${regressions.length}</strong><small class="delta-up">${comparableModules > 0 ? "Comparable regressions detected" : "No comparable baseline"}</small></article>
      <article><div class="card-head-inline"><span>Improvements</span>${infoButton("Modules whose composite risk score improved compared to the stored baseline.", "Improvements info")}</div><strong>${improvements.length}</strong><small class="delta-down">${comparableModules > 0 ? "Comparable improvements detected" : "No comparable baseline"}</small></article>
    </div>
    <section>
      <h4>Immediate action priority</h4>
      ${moduleList(topActionModules, (m) => `Risk ${moduleRiskScore(m).toFixed(1)} | D=${m.distance.toFixed(3)} | Max complexity=${m.maxComplexity}`)}
    </section>
    <section>
      <h4>Changes since the latest scan</h4>
      <div class="report-flow">
        <section>
          <h5>Regressions</h5>
          ${moduleDeltaList(regressions, "up")}
        </section>
        <section>
          <h5>Improvements</h5>
          ${moduleDeltaList(improvements, "down")}
        </section>
      </div>
    </section>
    <details class="report-extra">
      <summary>Additional context</summary>
      <div class="report-langs">
        ${byLanguage.map((entry) => `<article><span>${entry.language.toUpperCase()}</span><strong>${entry.modules} modules</strong><small>${entry.files} files</small></article>`).join("")}
      </div>
      <div class="report-grids">
        <section>
          <h4>Top distance from main sequence</h4>
          ${moduleList(byDistance, (m) => `D=${m.distance.toFixed(3)} | I=${m.instability.toFixed(3)} | A=${m.abstraction.toFixed(3)}`)}
        </section>
        <section>
          <h4>Top complexity</h4>
          ${moduleList(byComplexity, (m) => `Max=${m.maxComplexity} | Total=${m.totalComplexity} | Ce=${m.efferent}`)}
        </section>
      </div>
    </details>
    <section>
      <h4>All modules</h4>
      <div class="report-table-wrap">
        <table class="report-table">
          <thead>
            <tr><th scope="col">Module</th><th scope="col">Lang</th><th scope="col">Risk</th><th scope="col">D</th><th scope="col">Max Complexity</th><th scope="col">Orphans</th><th scope="col">God</th></tr>
          </thead>
          <tbody>
            ${summary.modules
              .map((module) => {
                const delta = moduleDeltas.find((d) => d.module === module.module) ?? null;
                const deltaCls = deltaClass(delta?.deltaScore ?? null);
                const deltaLabel = deltaTrendLabel(delta?.deltaScore ?? null);
                return `
              <tr>
                <td data-label="Module"><button class="module-link" data-module="${escapeAttr(module.module)}" aria-label="Open module diagnosis for ${escapeAttr(module.module)}"><code>${module.module}</code></button></td>
                <td data-label="Lang">${(module.language || "-").toUpperCase()}</td>
                <td data-label="Risk" class="${deltaCls}">${deltaLabel} ${moduleRiskScore(module).toFixed(1)}</td>
                <td data-label="D">${module.distance.toFixed(3)}</td>
                <td data-label="Max Complexity">${module.maxComplexity}</td>
                <td data-label="Orphans">${module.orphanBlocks.length}</td>
                <td data-label="God">${module.godBlocks.length}</td>
              </tr>`;
              })
              .join("")}
          </tbody>
        </table>
      </div>
    </section>
  `;

  report.querySelectorAll<HTMLButtonElement>("button.module-link").forEach((button) => {
    button.addEventListener("click", () => {
      const moduleName = button.dataset.module;
      const module = summary.modules.find((candidate) => candidate.module === moduleName);
      if (module) onModuleSelect(module);
    });
  });
}

function moduleList(items: ModuleMetrics[], line: (module: ModuleMetrics) => string): string {
  if (items.length === 0) return `<p class="hint">Not enough data.</p>`;
  return `<ul class="report-list">${items.map((module) => `<li><code>${module.module}</code><small>${line(module)}</small></li>`).join("")}</ul>`;
}

function moduleDeltaList(items: ModuleDelta[], direction: "up" | "down"): string {
  if (items.length === 0) {
    return `<p class="hint">${direction === "up" ? "No regressions recorded in the current baseline." : "No improvements recorded in the current baseline."}</p>`;
  }

  return `<ul class="report-list">${items
    .map(
      (module) =>
        `<li><code>${module.module}</code><small>Δ risk ${formatSigned(module.deltaScore, 1)} | D ${module.currentDistance.toFixed(3)}${module.previousDistance === null ? "" : ` (was ${module.previousDistance.toFixed(3)})`} | Max complexity ${module.currentMaxComplexity}${module.previousMaxComplexity === null ? "" : ` (was ${module.previousMaxComplexity})`}</small></li>`,
    )
    .join("")}</ul>`;
}
