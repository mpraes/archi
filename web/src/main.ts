import * as d3 from "d3";
import { Summary, ModuleMetrics, Connascence } from "./types";
import { fetchMetrics, isAIEnabled, streamAIInsights } from "./api";
import "./style.css";

let data: Summary | null = null;
let selected: ModuleMetrics | null = null;

async function main() {
  const app = document.getElementById("app")!;
  app.innerHTML = `
    <header class="status"><p id="status">Carregando diagnóstico…</p></header>
    <section class="main">
      <div id="chart"></div>
      <aside id="panel" class="panel empty">
        <p class="hint">Clique em um módulo para ver o diagnóstico.</p>
      </aside>
    </section>
    <section id="ai-panel" class="ai hidden"></section>
  `;

  try {
    data = await fetchMetrics();
  } catch (e) {
    document.getElementById("status")!.textContent = "Falha ao analisar: " + (e as Error).message;
    return;
  }
  renderStatus(data);
  renderChart(data);
  renderAI(data);
}

function renderStatus(s: Summary) {
  const hot = s.hotspots.length;
  if (s.moduleCount === 0) {
    document.getElementById("status")!.innerHTML =
      `📊 <strong>Status da Arquitetura:</strong> nenhum módulo foi identificado neste scan. ` +
      `Se o projeto for JS/TS, o parser ainda não está implementado nesta versão.`;
    return;
  }
  document.getElementById("status")!.innerHTML =
    `📊 <strong>Status da Arquitetura:</strong> Seu projeto possui ` +
    `<strong>${s.moduleCount} módulos</strong>. O design está predominantemente ` +
    `<strong>${hot === 0 ? "Saudável" : "Atenção"}</strong>, mas existem ` +
    `<strong>${hot} hotspot${hot > 1 ? "s" : ""}</strong> que ${hot === 1 ? "está travando" : "estão travando"} sua evolução.`;
}

function renderChart(s: Summary) {
  const chart = d3.select("#chart");
  if (s.modules.length === 0) {
    chart.append("div")
      .attr("class", "hint")
      .text("Nenhum módulo detectado para desenhar o mapa. Tente um projeto Go ou force --lang go.");
    return;
  }
  const width = chart.node()!.getBoundingClientRect().width;
  const height = 460;
  const margin = { top: 20, right: 20, bottom: 40, left: 50 };
  const svg = chart.append("svg").attr("viewBox", `0 0 ${width} ${height}`);
  const g = svg.append("g").attr("transform", `translate(${margin.left},${margin.top})`);
  const iw = width - margin.left - margin.right;
  const ih = height - margin.top - margin.bottom;

  const x = d3.scaleLinear().domain([0, 1]).range([0, iw]);
  const y = d3.scaleLinear().domain([0, 1]).range([ih, 0]);

  g.append("rect")
    .attr("x", 0).attr("y", 0).attr("width", iw).attr("height", ih)
    .attr("fill", "#11151c");

  const color = (m: ModuleMetrics) =>
    m.distance > 0.5 ? "#ff5d5d" : m.distance > 0.3 ? "#ffd166" : "#8ddc8d";

  // Main sequence line.
  g.append("line")
    .attr("x1", x(0)).attr("y1", y(1))
    .attr("x2", x(1)).attr("y2", y(0))
    .attr("stroke", "#3a4050").attr("stroke-dasharray", "4 4");

  g.append("g")
    .attr("transform", `translate(0,${ih})`)
    .call(d3.axisBottom(x).ticks(5))
    .call((sel) => sel.selectAll("text").attr("fill", "#8b93a7").text((d) => (d as number).toFixed(1)));
  g.append("g")
    .call(d3.axisLeft(y).ticks(5))
    .call((sel) => sel.selectAll("text").attr("fill", "#8b93a7").text((d) => (d as number).toFixed(1)));

  g.append("text").attr("x", iw / 2).attr("y", ih + 36)
    .attr("text-anchor", "middle").attr("fill", "#8b93a7").text("Instabilidade (I)");
  g.append("text").attr("transform", "rotate(-90)").attr("x", -ih / 2).attr("y", -38)
    .attr("text-anchor", "middle").attr("fill", "#8b93a7").text("Abstração (A)");

  const r = (m: ModuleMetrics) => 4 + Math.min(18, Math.sqrt(m.maxComplexity) * 3);

  const connascence = s.connascence;
  g.selectAll(".conn")
    .data(connascence)
    .enter()
    .append("line")
    .attr("class", "conn")
    .attr("x1", (c) => x(find(s, c.from)!.instability))
    .attr("y1", (c) => y(find(s, c.from)!.abstraction))
    .attr("x2", (c) => x(find(s, c.to)!.instability))
    .attr("y2", (c) => y(find(s, c.to)!.abstraction))
    .attr("stroke", (c) => (c.kind === "name" ? "#a05cff" : "#ff8a3d"))
    .attr("opacity", 0.5);

  const dots = g.selectAll(".dot")
    .data(s.modules)
    .enter()
    .append("circle")
    .attr("class", "dot")
    .attr("cx", (m) => x(m.instability))
    .attr("cy", (m) => y(m.abstraction))
    .attr("r", r)
    .attr("fill", color)
    .attr("stroke", "#1b1f29")
    .style("cursor", "pointer")
    .on("click", (_e, m) => selectModule(m));

  dots.append("title").text((m) => `${m.module}\nD=${m.distance.toFixed(2)}`);
}

function find(s: Summary, name: string): ModuleMetrics | undefined {
  return s.modules.find((m) => m.module === name);
}

function selectModule(m: ModuleMetrics) {
  selected = m;
  const panel = document.getElementById("panel")!;
  panel.classList.remove("empty");
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
  const aiPanel = document.getElementById("ai-panel")!;
  isAIEnabled().then((enabled) => {
    if (!enabled) return;
    aiPanel.classList.remove("hidden");
    aiPanel.innerHTML = `<h3>✨ Insight do Consultor Virtual</h3><div id="ai-stream" class="skeleton">Carregando insights…</div>`;
    let text = "";
    streamAIInsights(
      (chunk) => {
        text += chunk;
        const el = document.getElementById("ai-stream")!;
        el.classList.remove("skeleton");
        el.textContent = text;
      },
      () => {},
      (msg) => {
        const el = document.getElementById("ai-stream");
        if (el) { el.classList.remove("skeleton"); el.textContent = "IA indisponível: " + msg; }
      }
    );
  });
}

void main();