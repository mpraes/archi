import * as d3 from "d3";
import { Summary, ModuleMetrics, Connascence } from "./types";
import { fetchMetrics, isAIEnabled, streamAIInsights } from "./api";
import "./style.css";

let data: Summary | null = null;
let selected: ModuleMetrics | null = null;
type AxisSelection = d3.Selection<SVGGElement, unknown, HTMLElement, any>;
type LanguageBreakdown = { language: string; modules: number; files: number };
type ConnLineDatum = {
  conn: Connascence;
  from: ModuleMetrics;
  to: ModuleMetrics;
};
type ChartPalette = {
  bg: string;
  axis: string;
  mainSequence: string;
  connName: string;
  connMeaning: string;
  dotStroke: string;
  riskHigh: string;
  riskMid: string;
  riskLow: string;
};
type MetricDelta = { current: number; previous: number | null; delta: number | null };
type Snapshot = {
  capturedAt: string;
  projectName: string;
  moduleCount: number;
  hotspotCount: number;
  connascenceCount: number;
  avgDistance: number;
  avgInstability: number;
  avgMaxComplexity: number;
  modules: Record<string, ModuleSnapshot>;
};
type ModuleSnapshot = {
  distance: number;
  instability: number;
  maxComplexity: number;
  totalComplexity: number;
  orphanCount: number;
  godCount: number;
  efferent: number;
};
type ModuleDelta = {
  module: string;
  currentScore: number;
  previousScore: number | null;
  deltaScore: number | null;
  currentDistance: number;
  previousDistance: number | null;
  currentMaxComplexity: number;
  previousMaxComplexity: number | null;
};

const SNAPSHOT_KEY = "archi:summary:last-v1";
const CHART_MIN_WIDTH = 320;
let chartResizeObserver: ResizeObserver | null = null;
let chartResizeRaf: number | null = null;
let chartResizeHandler: (() => void) | null = null;
let chartOrientationHandler: (() => void) | null = null;
let chartToggleHandler: (() => void) | null = null;
let chartDetailsElement: HTMLDetailsElement | null = null;

async function main() {
  const app = requiredEl<HTMLElement>("app");
  const navItems = ["Resumo", "Módulos", "Conascência", "Risco", "Histórico"];
  const navLinks = navItems
    .map(
      (label, index) =>
        `<a ${index === 0 ? 'class="active" aria-current="page"' : ""} href="#">${label}</a>`,
    )
    .join("");
  app.innerHTML = `
    <div class="dashboard-shell">
      <aside class="sidebar">
        <div class="brand">
          <h1>Archi</h1>
          <p>Painel Analítico</p>
        </div>
        <nav class="side-nav" aria-label="Navegação principal">
          ${navLinks}
        </nav>
      </aside>
      <main class="workspace">
        <header class="topbar">
          <div>
            <h2>Dashboard</h2>
            <p id="status" role="status" aria-live="polite">Carregando diagnóstico…</p>
          </div>
          <div class="top-actions">
            <span id="history-badge" class="badge">Sem baseline</span>
          </div>
        </header>
        <nav class="compact-nav" aria-label="Navegação principal">
          ${navLinks}
        </nav>
        <section id="hero" class="hero"></section>
        <section id="kpi-grid" class="kpi-grid"></section>
        <section class="main">
          <div class="main-left">
            <section id="report" class="report"></section>
            <div class="chart-shell">
              <details>
                <summary>Mapa visual de apoio (secundário)</summary>
                <div id="chart"></div>
              </details>
            </div>
          </div>
          <aside class="main-right">
            <section id="activity" class="panel activity" aria-live="polite"></section>
            <section id="panel" class="panel empty" role="region" aria-live="polite" tabindex="-1">
              <p class="hint">Clique em um módulo na tabela (ou no mapa) para ver o diagnóstico.</p>
            </section>
          </aside>
        </section>
        <section id="ai-panel" class="ai hidden"></section>
      </main>
    </div>
  `;

  try {
    data = await fetchMetrics();
  } catch (e) {
    requiredEl<HTMLElement>("status").textContent = "Falha ao analisar: " + (e as Error).message;
    return;
  }
  const previousSnapshot = readPreviousSnapshot();
  renderStatus(data, previousSnapshot);
  renderHero(data, previousSnapshot);
  renderKpiGrid(data, previousSnapshot);
  renderActivity(data, previousSnapshot);
  renderChart(data);
  setupChartResponsiveness(data);
  renderDetailedReport(data, previousSnapshot);
  renderAI(data);
  saveSnapshot(buildSnapshot(data));
}

function renderStatus(s: Summary, previousSnapshot: Snapshot | null) {
  const hot = s.hotspots.length;
  if (s.moduleCount === 0) {
    requiredEl<HTMLElement>("status").textContent =
      "Nenhum módulo identificado. Execute na raiz do projeto ou use --lang all.";
    return;
  }
  requiredEl<HTMLElement>("status").textContent =
    `${s.moduleCount} módulos monitorados, ${hot} hotspot${hot === 1 ? "" : "s"} ativo${hot === 1 ? "" : "s"}.`;
  const historyBadge = requiredEl<HTMLElement>("history-badge");
  historyBadge.textContent = previousSnapshot
    ? `Comparando com ${formatCapturedAt(previousSnapshot.capturedAt)}`
    : "Sem baseline anterior";
}

function renderHero(s: Summary, previousSnapshot: Snapshot | null): void {
  const hero = requiredEl<HTMLElement>("hero");
  const riskModules = s.modules.filter((m) => m.godBlocks.length > 0 || m.orphanBlocks.length > 0).length;
  const previousRiskModules = previousSnapshot
    ? Object.values(previousSnapshot.modules).filter((m) => m.godCount > 0 || m.orphanCount > 0).length
    : null;
  const riskDelta = previousRiskModules === null ? null : riskModules - previousRiskModules;
  const riskSummary = riskModules === 1 ? "1 módulo em risco" : `${riskModules} módulos em risco`;
  const baselineSummary =
    riskDelta === null
      ? "Sem baseline"
      : riskDelta > 0
        ? `Atenção: +${riskDelta} vs anterior`
        : riskDelta < 0
          ? `Melhorou ${Math.abs(riskDelta)} vs anterior`
          : "Sem mudança vs anterior";
  hero.innerHTML = `
    <div class="hero-intro">
      <h3>Visão geral da arquitetura</h3>
      <p>Foque primeiro no que aumentou risco estrutural e depois avance para o detalhamento por módulo.</p>
    </div>
    <div class="hero-priority">
      <article class="hero-main-metric">
        <span>Prioridade atual</span>
        <strong>${riskSummary}</strong>
        <small class="${deltaClass(riskDelta)}">${baselineSummary}</small>
      </article>
      <ul class="hero-secondary">
        <li>
          <span>Conascência dominante</span>
          <b>${dominantConnascenceLabel(s.connascence)}</b>
        </li>
        <li>
          <span>Projeto analisado</span>
          <b>${s.projectName || "Atual"}</b>
          <small>${new Date().toLocaleDateString("pt-BR")}</small>
        </li>
      </ul>
    </div>
  `;
}

function renderKpiGrid(s: Summary, previousSnapshot: Snapshot | null): void {
  const kpiGrid = requiredEl<HTMLElement>("kpi-grid");
  const currentSnapshot = buildSnapshot(s);
  const averageTotalComplexity = s.modules.reduce((acc, mod) => acc + mod.totalComplexity, 0) / Math.max(1, s.modules.length);
  const cards: Array<{ title: string; value: string; foot: string; state?: string }> = [
    {
      title: "Hotspots",
      value: String(s.hotspots.length),
      foot: formatDelta(metricDelta(s.hotspots.length, previousSnapshot?.hotspotCount ?? null), 0),
      state: deltaClass(metricDelta(s.hotspots.length, previousSnapshot?.hotspotCount ?? null).delta),
    },
    {
      title: "Distância média",
      value: currentSnapshot.avgDistance.toFixed(3),
      foot: formatDelta(metricDelta(currentSnapshot.avgDistance, previousSnapshot?.avgDistance ?? null), 3),
      state: deltaClass(metricDelta(currentSnapshot.avgDistance, previousSnapshot?.avgDistance ?? null).delta),
    },
    {
      title: "Instabilidade média",
      value: currentSnapshot.avgInstability.toFixed(3),
      foot: formatDelta(metricDelta(currentSnapshot.avgInstability, previousSnapshot?.avgInstability ?? null), 3),
      state: deltaClass(metricDelta(currentSnapshot.avgInstability, previousSnapshot?.avgInstability ?? null).delta),
    },
    {
      title: "Cx máx média",
      value: currentSnapshot.avgMaxComplexity.toFixed(1),
      foot: formatDelta(metricDelta(currentSnapshot.avgMaxComplexity, previousSnapshot?.avgMaxComplexity ?? null), 1),
      state: deltaClass(metricDelta(currentSnapshot.avgMaxComplexity, previousSnapshot?.avgMaxComplexity ?? null).delta),
    },
    {
      title: "Cx total média",
      value: averageTotalComplexity.toFixed(1),
      foot: "Carga global por módulo",
      state: "neutral",
    },
    {
      title: "Conascências",
      value: String(s.connascence.length),
      foot: formatDelta(metricDelta(s.connascence.length, previousSnapshot?.connascenceCount ?? null), 0),
      state: deltaClass(metricDelta(s.connascence.length, previousSnapshot?.connascenceCount ?? null).delta),
    },
  ];

  kpiGrid.innerHTML = cards
    .map(
      (card) =>
        `<article class="kpi-card"><span>${card.title}</span><strong>${card.value}</strong><small class="${card.state ?? "neutral"}">${card.foot}</small></article>`,
    )
    .join("");
}

function renderActivity(s: Summary, previousSnapshot: Snapshot | null): void {
  const activity = requiredEl<HTMLElement>("activity");
  const moduleDeltas = compareModules(s.modules, previousSnapshot);
  const events = [
    ...moduleDeltas
      .filter((item) => (item.deltaScore ?? 0) > 0)
      .sort((a, b) => (b.deltaScore ?? 0) - (a.deltaScore ?? 0))
      .slice(0, 4)
      .map((item) => ({
        title: `${item.module} piorou`,
        detail: `Δ risco ${formatSigned(item.deltaScore, 1)} | D ${item.currentDistance.toFixed(3)}`,
        tone: "delta-up",
      })),
    ...moduleDeltas
      .filter((item) => (item.deltaScore ?? 0) < 0)
      .sort((a, b) => (a.deltaScore ?? 0) - (b.deltaScore ?? 0))
      .slice(0, 3)
      .map((item) => ({
        title: `${item.module} melhorou`,
        detail: `Δ risco ${formatSigned(item.deltaScore, 1)} | Cx máx ${item.currentMaxComplexity}`,
        tone: "delta-down",
      })),
  ].slice(0, 7);

  activity.innerHTML = `
    <header><h3>Atividade comparativa</h3><small>${previousSnapshot ? "Baseado no último scan" : "Aguardando baseline para comparar"}</small></header>
    ${events.length === 0 ? `<p class="hint">Sem alterações comparáveis por enquanto.</p>` : `
      <ul class="activity-list">
        ${events.map((event) => `<li><strong class="${event.tone}">${event.title}</strong><small>${event.detail}</small></li>`).join("")}
      </ul>
    `}
  `;
}

function renderDetailedReport(s: Summary, previousSnapshot: Snapshot | null) {
  const report = requiredEl<HTMLElement>("report");
  if (s.modules.length === 0) {
    report.innerHTML = "";
    return;
  }

  const historyLabel = previousSnapshot
    ? `Comparação com scan anterior de ${formatCapturedAt(previousSnapshot.capturedAt)}`
    : "Sem baseline anterior: este scan será usado como referência para a próxima execução.";
  const moduleDeltas = compareModules(s.modules, previousSnapshot);
  const regressions = moduleDeltas.filter((d) => (d.deltaScore ?? 0) > 0).sort((a, b) => (b.deltaScore ?? 0) - (a.deltaScore ?? 0)).slice(0, 8);
  const improvements = moduleDeltas.filter((d) => (d.deltaScore ?? 0) < 0).sort((a, b) => (a.deltaScore ?? 0) - (b.deltaScore ?? 0)).slice(0, 8);
  const riskModules = s.modules.filter((m) => m.godBlocks.length > 0 || m.orphanBlocks.length > 0).length;
  const comparableModules = moduleDeltas.filter((item) => item.deltaScore !== null).length;
  const topActionModules = [...s.modules]
    .sort((a, b) => moduleRiskScore(b) - moduleRiskScore(a))
    .slice(0, 6);
  const byDistance = [...s.modules].sort((a, b) => b.distance - a.distance).slice(0, 5);
  const byComplexity = [...s.modules].sort((a, b) => b.maxComplexity - a.maxComplexity).slice(0, 5);
  const byLanguage = languageBreakdown(s.modules);

  report.innerHTML = `
    <header class="report-header">
      <h3>Painel Analítico</h3>
      <p class="history">${historyLabel}</p>
    </header>
    <div class="report-priority">
      <article><span>Risco imediato</span><strong>${riskModules}</strong><small class="neutral">Com God ou órfão</small></article>
      <article><span>Regressões</span><strong>${regressions.length}</strong><small class="delta-up">${comparableModules > 0 ? "Pioras comparáveis detectadas" : "Sem baseline comparável"}</small></article>
      <article><span>Evoluções</span><strong>${improvements.length}</strong><small class="delta-down">${comparableModules > 0 ? "Melhoras comparáveis detectadas" : "Sem baseline comparável"}</small></article>
    </div>
    <section>
      <h4>Prioridade de ação imediata</h4>
      ${moduleList(topActionModules, (m) => `Risco ${moduleRiskScore(m).toFixed(1)} | D=${m.distance.toFixed(3)} | Cx máx=${m.maxComplexity}`)}
    </section>
    <section>
      <h4>Mudanças desde o último scan</h4>
      <div class="report-flow">
        <section>
          <h5>Regressões</h5>
          ${moduleDeltaList(regressions, "up")}
        </section>
        <section>
          <h5>Evoluções</h5>
          ${moduleDeltaList(improvements, "down")}
        </section>
      </div>
    </section>
    <details class="report-extra">
      <summary>Contexto complementar</summary>
      <div class="report-langs">
        ${byLanguage.map((entry) => `<article><span>${entry.language.toUpperCase()}</span><strong>${entry.modules} módulos</strong><small>${entry.files} arquivos</small></article>`).join("")}
      </div>
      <div class="report-grids">
        <section>
          <h4>Top distância da sequência principal</h4>
          ${moduleList(byDistance, (m) => `D=${m.distance.toFixed(3)} | I=${m.instability.toFixed(3)} | A=${m.abstraction.toFixed(3)}`)}
        </section>
        <section>
          <h4>Top complexidade</h4>
          ${moduleList(byComplexity, (m) => `Máx=${m.maxComplexity} | Total=${m.totalComplexity} | Ce=${m.efferent}`)}
        </section>
      </div>
    </details>
    <section>
      <h4>Todos os módulos</h4>
      <div class="report-table-wrap">
        <table class="report-table">
          <thead>
            <tr><th scope="col">Módulo</th><th scope="col">Lang</th><th scope="col">Risco</th><th scope="col">D</th><th scope="col">Cx Máx</th><th scope="col">Órfãos</th><th scope="col">God</th></tr>
          </thead>
          <tbody>
            ${s.modules.map((m) => {
              const delta = moduleDeltas.find((d) => d.module === m.module) ?? null;
              const deltaCls = deltaClass(delta?.deltaScore ?? null);
              const deltaLabel = deltaTrendLabel(delta?.deltaScore ?? null);
              return `
              <tr>
                <td data-label="Módulo"><button class="module-link" data-module="${escapeAttr(m.module)}" aria-label="Abrir diagnóstico do módulo ${escapeAttr(m.module)}"><code>${m.module}</code></button></td>
                <td data-label="Lang">${(m.language || "-").toUpperCase()}</td>
                <td data-label="Risco" class="${deltaCls}">${deltaLabel} ${moduleRiskScore(m).toFixed(1)}</td>
                <td data-label="D">${m.distance.toFixed(3)}</td>
                <td data-label="Cx Máx">${m.maxComplexity}</td>
                <td data-label="Órfãos">${m.orphanBlocks.length}</td>
                <td data-label="God">${m.godBlocks.length}</td>
              </tr>`;
            }).join("")}
          </tbody>
        </table>
      </div>
    </section>
  `;
  report.querySelectorAll<HTMLButtonElement>("button.module-link").forEach((button) => {
    button.addEventListener("click", () => {
      const moduleName = button.dataset.module;
      const module = s.modules.find((m) => m.module === moduleName);
      if (module) selectModule(module);
    });
  });
}

function moduleList(items: ModuleMetrics[], line: (m: ModuleMetrics) => string): string {
  if (items.length === 0) return `<p class="hint">Sem dados suficientes.</p>`;
  return `<ul class="report-list">${items.map((m) => `<li><code>${m.module}</code><small>${line(m)}</small></li>`).join("")}</ul>`;
}

function languageBreakdown(modules: ModuleMetrics[]): LanguageBreakdown[] {
  const map = new Map<string, { modules: number; files: number }>();
  for (const m of modules) {
    const lang = (m.language || "unknown").toLowerCase();
    const current = map.get(lang) ?? { modules: 0, files: 0 };
    current.modules += 1;
    current.files += m.files;
    map.set(lang, current);
  }
  return [...map.entries()]
    .map(([language, counts]) => ({ language, modules: counts.modules, files: counts.files }))
    .sort((a, b) => b.modules - a.modules || b.files - a.files);
}

function moduleDeltaList(items: ModuleDelta[], direction: "up" | "down"): string {
  if (items.length === 0) {
    return `<p class="hint">${direction === "up" ? "Nenhuma regressão registrada no baseline atual." : "Nenhuma evolução registrada no baseline atual."}</p>`;
  }
  return `<ul class="report-list">${items
    .map(
      (m) =>
        `<li><code>${m.module}</code><small>Δ risco ${formatSigned(m.deltaScore, 1)} | D ${m.currentDistance.toFixed(3)}${m.previousDistance === null ? "" : ` (antes ${m.previousDistance.toFixed(3)})`} | Cx máx ${m.currentMaxComplexity}${m.previousMaxComplexity === null ? "" : ` (antes ${m.previousMaxComplexity})`}</small></li>`,
    )
    .join("")}</ul>`;
}

function renderChart(s: Summary) {
  const chart = d3.select<HTMLDivElement, unknown>("#chart");
  chart.selectAll("*").remove();
  const palette = readChartPalette();
  if (s.modules.length === 0) {
    chart.append("div")
      .attr("class", "hint")
      .text("Nenhum módulo detectado para desenhar o mapa. Tente um projeto Go ou force --lang go.");
    return;
  }
  const chartNode = chart.node();
  if (!chartNode) return;
  const width = Math.max(CHART_MIN_WIDTH, Math.round(chartNode.getBoundingClientRect().width || CHART_MIN_WIDTH));
  const height = 460;
  const margin = { top: 20, right: 20, bottom: 40, left: 50 };
  const svg = chart.append<SVGSVGElement>("svg").attr("viewBox", `0 0 ${width} ${height}`);
  const g = svg.append<SVGGElement>("g").attr("transform", `translate(${margin.left},${margin.top})`);
  const iw = width - margin.left - margin.right;
  const ih = height - margin.top - margin.bottom;

  const x = d3.scaleLinear().domain([0, 1]).range([0, iw]);
  const y = d3.scaleLinear().domain([0, 1]).range([ih, 0]);

  g.append("rect")
    .attr("x", 0).attr("y", 0).attr("width", iw).attr("height", ih)
    .attr("fill", palette.bg);

  const color = (m: ModuleMetrics) =>
    m.distance > 0.5 ? palette.riskHigh : m.distance > 0.3 ? palette.riskMid : palette.riskLow;

  // Main sequence line.
  g.append("line")
    .attr("x1", x(0)).attr("y1", y(1))
    .attr("x2", x(1)).attr("y2", y(0))
    .attr("stroke", palette.mainSequence).attr("stroke-dasharray", "4 4");

  g.append("g")
    .attr("transform", `translate(0,${ih})`)
    .call(d3.axisBottom(x).ticks(5))
    .call(formatAxisTickLabels);
  g.append("g")
    .call(d3.axisLeft(y).ticks(5))
    .call(formatAxisTickLabels);

  g.append("text").attr("x", iw / 2).attr("y", ih + 36)
    .attr("text-anchor", "middle").attr("fill", palette.axis).text("Instabilidade (I)");
  g.append("text").attr("transform", "rotate(-90)").attr("x", -ih / 2).attr("y", -38)
    .attr("text-anchor", "middle").attr("fill", palette.axis).text("Abstração (A)");

  const r = (m: ModuleMetrics) => 4 + Math.min(18, Math.sqrt(m.maxComplexity) * 3);

  const byModuleName = new Map<string, ModuleMetrics>(s.modules.map((m) => [m.module, m]));
  const connascence: ConnLineDatum[] = s.connascence.flatMap((conn) => {
    const from = byModuleName.get(conn.from);
    const to = byModuleName.get(conn.to);
    if (!from || !to) return [];
    return [{ conn, from, to }];
  });
  g.selectAll<SVGLineElement, Connascence>(".conn")
    .data(connascence)
    .enter()
    .append("line")
    .attr("class", "conn")
    .attr("x1", (d: ConnLineDatum) => x(d.from.instability))
    .attr("y1", (d: ConnLineDatum) => y(d.from.abstraction))
    .attr("x2", (d: ConnLineDatum) => x(d.to.instability))
    .attr("y2", (d: ConnLineDatum) => y(d.to.abstraction))
    .attr("stroke", (d: ConnLineDatum) => (d.conn.kind === "name" ? palette.connName : palette.connMeaning))
    .attr("opacity", 0.5);

  const dots = g.selectAll<SVGCircleElement, ModuleMetrics>(".dot")
    .data(s.modules)
    .enter()
    .append("circle")
    .attr("class", "dot")
    .attr("cx", (m: ModuleMetrics) => x(m.instability))
    .attr("cy", (m: ModuleMetrics) => y(m.abstraction))
    .attr("r", r)
    .attr("fill", color)
    .attr("stroke", palette.dotStroke)
    .style("cursor", "pointer")
    .on("click", (_e: MouseEvent, m: ModuleMetrics) => selectModule(m));

  dots.append("title").text((m: ModuleMetrics) => `${m.module}\nD=${m.distance.toFixed(2)}`);
}

function formatAxisTickLabels(sel: AxisSelection): void {
  const axisColor = readCssVar("--chart-axis", "rgb(95 115 145)");
  sel
    .selectAll<SVGTextElement, number>("text")
    .attr("fill", axisColor)
    .text((d: number) => d.toFixed(1));
}

function setupChartResponsiveness(summary: Summary): void {
  teardownChartResponsiveness();
  const chartElement = document.getElementById("chart");
  const detailsElement = chartElement?.closest("details");
  if (!chartElement) return;

  const scheduleRender = () => {
    if (chartResizeRaf !== null) {
      window.cancelAnimationFrame(chartResizeRaf);
    }
    chartResizeRaf = window.requestAnimationFrame(() => {
      chartResizeRaf = null;
      renderChart(summary);
    });
  };

  chartResizeHandler = scheduleRender;
  chartOrientationHandler = scheduleRender;
  window.addEventListener("resize", chartResizeHandler);
  window.addEventListener("orientationchange", chartOrientationHandler);

  if ("ResizeObserver" in window) {
    chartResizeObserver = new ResizeObserver(() => scheduleRender());
    chartResizeObserver.observe(chartElement);
  }

  if (detailsElement instanceof HTMLDetailsElement) {
    chartDetailsElement = detailsElement;
    chartToggleHandler = () => scheduleRender();
    chartDetailsElement.addEventListener("toggle", chartToggleHandler);
  }
}

function teardownChartResponsiveness(): void {
  if (chartResizeObserver) {
    chartResizeObserver.disconnect();
    chartResizeObserver = null;
  }
  if (chartResizeHandler) {
    window.removeEventListener("resize", chartResizeHandler);
    chartResizeHandler = null;
  }
  if (chartOrientationHandler) {
    window.removeEventListener("orientationchange", chartOrientationHandler);
    chartOrientationHandler = null;
  }
  if (chartDetailsElement && chartToggleHandler) {
    chartDetailsElement.removeEventListener("toggle", chartToggleHandler);
  }
  chartDetailsElement = null;
  chartToggleHandler = null;
  if (chartResizeRaf !== null) {
    window.cancelAnimationFrame(chartResizeRaf);
    chartResizeRaf = null;
  }
}

function readCssVar(name: string, fallback: string): string {
  const value = window.getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return value || fallback;
}

function readChartPalette(): ChartPalette {
  return {
    bg: readCssVar("--chart-bg", "rgb(248 251 255)"),
    axis: readCssVar("--chart-axis", "rgb(95 115 145)"),
    mainSequence: readCssVar("--chart-main-sequence", "rgb(183 199 223)"),
    connName: readCssVar("--chart-conn-name", "rgb(160 92 255)"),
    connMeaning: readCssVar("--chart-conn-meaning", "rgb(255 138 61)"),
    dotStroke: readCssVar("--chart-dot-stroke", "rgb(216 226 241)"),
    riskHigh: readCssVar("--risk-high", "rgb(196 58 58)"),
    riskMid: readCssVar("--risk-mid", "rgb(238 159 47)"),
    riskLow: readCssVar("--risk-low", "rgb(31 125 80)"),
  };
}

function selectModule(m: ModuleMetrics) {
  selected = m;
  const panel = requiredEl<HTMLElement>("panel");
  panel.classList.remove("empty");
  panel.setAttribute("aria-label", `Diagnóstico detalhado do módulo ${m.module}`);
  const last = m.module.split("/").pop();
  const conns = data!.connascence.filter((c) => c.from === m.module || c.to === m.module);
  panel.innerHTML = `
    <header><h2>${last}</h2><code>${m.module}</code></header>
    <section>
      <h3>O que está rígido</h3>
      <ul>
        <li><span class="m">Quem ele puxa</span> <b>${m.efferent}</b> módulos</li>
        <li><span class="m">Quem puxa ele</span> <b>${m.afferent}</b> módulos dependem dele</li>
        <li><span class="m">Abstração</span> ${pct(m.abstracts, m.abstracts + m.concretes)}</li>
        <li><span class="m">Complexidade máx</span> <b>${m.maxComplexity}</b> (total ${m.totalComplexity})</li>
      </ul>
    </section>
    ${m.godBlocks.length ? section("Funções God", m.godBlocks) : ""}
    ${m.orphanBlocks.length ? section("Código órfão", m.orphanBlocks) : ""}
    ${conns.length ? connSection(conns, m.module) : ""}
  `;
  panel.focus();
}

function pct(a: number, b: number): string {
  if (b === 0) return "—";
  return `<b>${Math.round((a / b) * 100)}%</b> (${a}/${b})`;
}

function section(title: string, items: string[]): string {
  return `<section><h3>${title}</h3><ul>${items.map((i) => `<li><code>${i}</code></li>`).join("")}</ul></section>`;
}

function connSection(conns: Connascence[], self: string): string {
  const rows = conns.map((c) => {
    const other = c.from === self ? c.to : c.from;
    const dir = c.from === self ? "→" : "←";
    const label = c.kind === "name" ? "Conascência de Nome" : "Conascência de Significado";
    return `<li><span class="m">${label}</span> ${dir} <code>${other}</code><br><small>${c.detail}</small></li>`;
  }).join("");
  return `<section><h3>Conascência Detectada</h3><ul>${rows}</ul></section>`;
}

function renderAI(s: Summary) {
  const aiPanel = requiredEl<HTMLElement>("ai-panel");
  isAIEnabled().then((enabled) => {
    if (!enabled) return;
    aiPanel.classList.remove("hidden");
    aiPanel.innerHTML = `<h3>✨ Insight do Consultor Virtual</h3><div id="ai-stream" class="skeleton">Carregando insights…</div>`;
    let text = "";
    streamAIInsights(
      (chunk) => {
        text += chunk;
        const el = requiredEl<HTMLElement>("ai-stream");
        el.classList.remove("skeleton");
        el.textContent = text;
      },
      () => {},
      (msg) => {
        const el = document.getElementById("ai-stream") as HTMLElement | null;
        if (el) { el.classList.remove("skeleton"); el.textContent = "IA indisponível: " + msg; }
      }
    );
  });
}

function dominantConnascenceLabel(conns: Connascence[]): string {
  if (conns.length === 0) return "Sem dependências críticas";
  const nameCount = conns.filter((conn) => conn.kind === "name").length;
  const meaningCount = conns.length - nameCount;
  if (nameCount === meaningCount) return "Nome e significado equilibrados";
  return nameCount > meaningCount ? "Predomínio de conascência de nome" : "Predomínio de conascência de significado";
}

function requiredEl<T extends HTMLElement>(id: string): T {
  const el = document.getElementById(id);
  if (!el) {
    throw new Error(`elemento #${id} não encontrado`);
  }
  return el as T;
}

function buildSnapshot(summary: Summary): Snapshot {
  const modules: Record<string, ModuleSnapshot> = {};
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

function readPreviousSnapshot(): Snapshot | null {
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

function saveSnapshot(snapshot: Snapshot): void {
  try {
    window.localStorage.setItem(SNAPSHOT_KEY, JSON.stringify(snapshot));
  } catch {
    // Ignore write failures (private mode or blocked storage).
  }
}

function compareModules(modules: ModuleMetrics[], previousSnapshot: Snapshot | null): ModuleDelta[] {
  return modules.map((m): ModuleDelta => {
    const previous = previousSnapshot?.modules[m.module] ?? null;
    const currentScore = moduleRiskScore(m);
    const previousScore = previous ? moduleRiskScoreFromSnapshot(previous) : null;
    return {
      module: m.module,
      currentScore,
      previousScore,
      deltaScore: previousScore === null ? null : currentScore - previousScore,
      currentDistance: m.distance,
      previousDistance: previous?.distance ?? null,
      currentMaxComplexity: m.maxComplexity,
      previousMaxComplexity: previous?.maxComplexity ?? null,
    };
  });
}

function moduleRiskScore(module: ModuleMetrics): number {
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

function moduleRiskScoreFromSnapshot(module: ModuleSnapshot): number {
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

function metricDelta(current: number, previous: number | null): MetricDelta {
  return {
    current,
    previous,
    delta: previous === null ? null : current - previous,
  };
}

function formatSigned(value: number | null, digits: number): string {
  if (value === null) return "n/a";
  const rounded = value.toFixed(digits);
  return `${value > 0 ? "+" : ""}${rounded}`;
}

function formatDelta(delta: MetricDelta, digits: number): string {
  if (delta.previous === null || delta.delta === null) return "Sem baseline";
  if (delta.delta > 0) return `Subiu ${formatSigned(delta.delta, digits)} vs anterior`;
  if (delta.delta < 0) return `Caiu ${formatSigned(delta.delta, digits)} vs anterior`;
  return "Sem mudança vs anterior";
}

function deltaClass(delta: number | null): string {
  if (delta === null) return "neutral";
  if (delta > 0) return "delta-up";
  if (delta < 0) return "delta-down";
  return "neutral";
}

function deltaTrendLabel(delta: number | null): string {
  if (delta === null) return "•";
  if (delta > 0) return "↑";
  if (delta < 0) return "↓";
  return "=";
}

function formatCapturedAt(isoDate: string): string {
  const parsed = new Date(isoDate);
  if (Number.isNaN(parsed.getTime())) return "data desconhecida";
  return new Intl.DateTimeFormat("pt-BR", { dateStyle: "short", timeStyle: "short" }).format(parsed);
}

function escapeAttr(value: string): string {
  return value.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;");
}

void main();