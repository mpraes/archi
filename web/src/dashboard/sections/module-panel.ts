import { Connascence, ModuleMetrics, Summary } from "../../types";
import { requestNavigation } from "../layout";
import { requiredEl } from "../ui";

export function renderModulePanel(summary: Summary, module: ModuleMetrics): void {
  const panel = requiredEl<HTMLElement>("panel");
  panel.classList.remove("empty");
  panel.setAttribute("aria-label", `Detailed diagnosis for module ${module.module}`);
  const last = module.module.split("/").pop();
  const conns = summary.connascence.filter((conn) => conn.from === module.module || conn.to === module.module);

  panel.innerHTML = `
    <header><h2>${last}</h2><code>${module.module}</code></header>
    <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 16px; margin-top: 16px;">
      <section>
        <h3>What is rigid here</h3>
        <ul>
          <li><span class="m">Outgoing dependencies</span> <b>${module.efferent}</b> modules</li>
          <li><span class="m">Incoming dependencies</span> <b>${module.afferent}</b> modules depend on it</li>
          <li><span class="m">Abstraction ratio</span> ${pct(module.abstracts, module.abstracts + module.concretes)}</li>
          <li><span class="m">Max complexity</span> <b>${module.maxComplexity}</b> (total ${module.totalComplexity})</li>
        </ul>
      </section>
      ${module.godBlocks.length ? section("God functions", module.godBlocks) : ""}
      ${module.orphanBlocks.length ? section("Orphan code", module.orphanBlocks) : ""}
      ${conns.length ? connSection(conns, module.module) : ""}
    </div>
  `;
  requestNavigation("report");
  panel.focus();
}

function pct(a: number, b: number): string {
  if (b === 0) return "—";
  return `<b>${Math.round((a / b) * 100)}%</b> (${a}/${b})`;
}

function section(title: string, items: string[]): string {
  return `<section><h3>${title}</h3><ul>${items.map((item) => `<li><code>${item}</code></li>`).join("")}</ul></section>`;
}

function connSection(conns: Connascence[], self: string): string {
  const maxRows = 32;
  const visibleConns = conns.slice(0, maxRows);
  const rows = visibleConns
    .map((conn) => {
      const other = conn.from === self ? conn.to : conn.from;
      const direction = conn.from === self ? "→" : "←";
      const label = conn.kind === "name" ? "Name connascence" : "Meaning connascence";
      return `<li><span class="m">${label}</span> ${direction} <code>${other}</code><br><small>${conn.detail}</small></li>`;
    })
    .join("");
  const overflowNote =
    conns.length > maxRows
      ? `<p class="hint">Showing ${maxRows} of ${conns.length} relationships. Refine the module scope to inspect fewer links at once.</p>`
      : "";

  return `<section><h3>Shared meaning links</h3>${overflowNote}<ul class="conn-list">${rows}</ul></section>`;
}
