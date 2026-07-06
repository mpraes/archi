import * as d3 from "d3";
import { ModuleMetrics, Summary } from "../../types";
import { dominantConnascenceLabel } from "../format";
import { moduleRiskScore } from "../metrics";
import { readCssVar, requiredEl } from "../ui";

type GraphNodeDatum = d3.SimulationNodeDatum & {
  id: string;
  module: ModuleMetrics;
  risk: number;
};

type GraphEdgeDatum = d3.SimulationLinkDatum<GraphNodeDatum> & {
  kind: "dependency" | "conn-name" | "conn-meaning";
  weight: number;
};

const CHART_MIN_WIDTH = 320;
const ARCH_GRAPH_MAX_NODES = 120;
const ARCH_GRAPH_HEIGHT = 560;

export function renderArchitectureMap(summary: Summary, onModuleSelect: (module: ModuleMetrics) => void): void {
  const meta = requiredEl<HTMLElement>("architecture-map-meta");
  const content = requiredEl<HTMLElement>("architecture-map-content");
  const graph = buildArchitectureGraph(summary);
  const connascenceView = document.getElementById("connascence-view");

  meta.innerHTML = `
    <div class="architecture-kpis">
      <article><span>Modules rendered</span><strong>${graph.nodes.length}</strong><small>${graph.wasTrimmed ? `Auto-limited from ${summary.modules.length}` : "Full project graph"}</small></article>
      <article><span>Edges rendered</span><strong>${graph.edges.length}</strong><small>${graph.dependencyEdges} dependencies + ${graph.connEdges} connascence</small></article>
      <article><span>Dominant connascence</span><strong>${dominantConnascenceLabel(summary.connascence)}</strong><small>Cross-module coupling signal</small></article>
    </div>
    ${graph.wasTrimmed ? `<p class="hint">Large repository detected. Rendering top ${graph.nodes.length} highest-risk modules to keep interaction smooth.</p>` : ""}
    ${graph.dependencyEdges === 0 ? `<p class="hint">No explicit dependency edges were provided by the metrics payload. Showing connascence links and module risk map.</p>` : ""}
  `;
  content.innerHTML = `
    <div class="architecture-tools">
      <section class="architecture-legend">
        <h4>Legend</h4>
        <ul>
          <li><span class="legend-dot risk-high"></span> High risk node</li>
          <li><span class="legend-dot risk-mid"></span> Medium risk node</li>
          <li><span class="legend-dot risk-low"></span> Low risk node</li>
          <li><span class="legend-line dependency"></span> Dependency edge</li>
          <li><span class="legend-line conn-name"></span> Name connascence edge</li>
          <li><span class="legend-line conn-meaning"></span> Meaning connascence edge</li>
        </ul>
      </section>
      <section class="architecture-note">
        <h4>How to use</h4>
        <p>Node size and color are driven by module risk score. Click a node to open module diagnostics in the side panel.</p>
      </section>
    </div>
    <div id="arch-graph" class="architecture-graph"></div>
  `;
  if (connascenceView) {
    connascenceView.innerHTML = renderConnascenceSection(summary);
  }

  const graphEl = requiredEl<HTMLElement>("arch-graph");
  if (graph.nodes.length === 0) {
    graphEl.innerHTML = `<p class="hint">No modules available for architecture graph rendering.</p>`;
    return;
  }

  const width = Math.max(CHART_MIN_WIDTH, Math.round(graphEl.getBoundingClientRect().width || CHART_MIN_WIDTH));
  const height = ARCH_GRAPH_HEIGHT;
  const riskExtent = d3.extent(graph.nodes, (node) => node.risk);
  const maxRisk = riskExtent[1] ?? 1;
  const minRisk = riskExtent[0] ?? 0;
  const riskRange = Math.max(0.001, maxRisk - minRisk);
  const nodeRadius = (node: GraphNodeDatum) => 7 + ((node.risk - minRisk) / riskRange) * 14;
  const nodeColor = (node: GraphNodeDatum) => {
    if (node.risk > 80) return readCssVar("--risk-high", "#c43a3a");
    if (node.risk > 45) return readCssVar("--risk-mid", "#ee9f2f");
    return readCssVar("--risk-low", "#1f7d50");
  };
  const edgeColor = (edge: GraphEdgeDatum): string => {
    if (edge.kind === "dependency") return readCssVar("--chart-main-sequence", "#b7c7df");
    if (edge.kind === "conn-name") return readCssVar("--chart-conn-name", "#a05cff");
    return readCssVar("--chart-conn-meaning", "#ff8a3d");
  };

  const svg = d3
    .select(graphEl)
    .append<SVGSVGElement>("svg")
    .attr("viewBox", `0 0 ${width} ${height}`)
    .attr("role", "img")
    .attr("aria-label", "Architecture graph map");
  const scene = svg.append("g");
  svg.call(
    d3.zoom<SVGSVGElement, unknown>().scaleExtent([0.35, 2.2]).on("zoom", (event) => {
      scene.attr("transform", event.transform.toString());
    }),
  );

  const edges = scene
    .append("g")
    .attr("class", "edge-layer")
    .selectAll("line")
    .data(graph.edges)
    .enter()
    .append("line")
    .attr("class", (edge) => `edge ${edge.kind}`)
    .attr("stroke", edgeColor)
    .attr("stroke-opacity", 0.7)
    .attr("stroke-width", (edge) => (edge.kind === "dependency" ? 1.2 : 2));

  let simulation: d3.Simulation<GraphNodeDatum, GraphEdgeDatum> | null = null;
  const nodes = scene
    .append("g")
    .attr("class", "node-layer")
    .selectAll("circle")
    .data(graph.nodes)
    .enter()
    .append("circle")
    .attr("class", "graph-node")
    .attr("r", nodeRadius)
    .attr("fill", nodeColor)
    .attr("stroke", readCssVar("--chart-dot-stroke", "#d8e2f1"))
    .attr("stroke-width", 1.5)
    .style("cursor", "pointer")
    .on("click", (_event: MouseEvent, node: GraphNodeDatum) => onModuleSelect(node.module))
    .call(
      d3
        .drag<SVGCircleElement, GraphNodeDatum>()
        .on("start", (event, node) => {
          if (!event.active) simulation?.alphaTarget(0.25).restart();
          node.fx = node.x;
          node.fy = node.y;
        })
        .on("drag", (event, node) => {
          node.fx = event.x;
          node.fy = event.y;
        })
        .on("end", (event, node) => {
          if (!event.active) simulation?.alphaTarget(0);
          node.fx = null;
          node.fy = null;
        }),
    );
  nodes
    .append("title")
    .text((node) => `${node.id}\nRisk=${node.risk.toFixed(1)}\nD=${node.module.distance.toFixed(3)}\nMaxCx=${node.module.maxComplexity}`);

  const labels = scene
    .append("g")
    .attr("class", "label-layer")
    .selectAll("text")
    .data(graph.nodes)
    .enter()
    .append("text")
    .attr("class", "graph-label")
    .attr("font-size", 11)
    .attr("fill", readCssVar("--text", "#273752"))
    .attr("text-anchor", "middle")
    .text((node) => node.id.split("/").pop() || node.id);

  simulation = d3
    .forceSimulation(graph.nodes)
    .force(
      "link",
      d3
        .forceLink(graph.edges)
        .id((node) => (node as GraphNodeDatum).id)
        .distance((edge) => (edge.kind === "dependency" ? 85 : 115)),
    )
    .force("charge", d3.forceManyBody().strength(-240))
    .force("center", d3.forceCenter(width / 2, height / 2))
    .force("collide", d3.forceCollide<GraphNodeDatum>().radius((node) => nodeRadius(node) + 6))
    .on("tick", () => {
      edges
        .attr("x1", (edge) => (edge.source as GraphNodeDatum).x ?? 0)
        .attr("y1", (edge) => (edge.source as GraphNodeDatum).y ?? 0)
        .attr("x2", (edge) => (edge.target as GraphNodeDatum).x ?? 0)
        .attr("y2", (edge) => (edge.target as GraphNodeDatum).y ?? 0);
      nodes.attr("cx", (node) => node.x ?? 0).attr("cy", (node) => node.y ?? 0);
      labels.attr("x", (node) => node.x ?? 0).attr("y", (node) => (node.y ?? 0) + nodeRadius(node) + 11);
    });
}

function renderConnascenceSection(summary: Summary): string {
  const grouped = new Map<string, { kind: "name" | "meaning"; from: string; to: string; detail: string; count: number }>();
  for (const conn of summary.connascence) {
    const key = `${conn.kind}|${conn.from}|${conn.to}|${conn.detail}`;
    const current = grouped.get(key);
    if (current) current.count += 1;
    else grouped.set(key, { ...conn, count: 1 });
  }

  const rows = [...grouped.values()]
    .sort((a, b) => b.count - a.count)
    .slice(0, 120)
    .map((item) => {
      const kindLabel = item.kind === "name" ? "Name connascence" : "Meaning connascence";
      const repeated = item.count > 1 ? `<small>Repeated ${item.count} times</small>` : "";
      return `<li><span class="m">${kindLabel}</span> <code>${item.from}</code> → <code>${item.to}</code><br><small>${item.detail}</small>${repeated}</li>`;
    })
    .join("");

  if (rows.length === 0) {
    return `
      <header>
        <h3>Shared meaning hotspots</h3>
        <p class="hint">No cross-module meaning or naming links were detected in this scan.</p>
      </header>
    `;
  }

  const overflowHint =
    grouped.size > 120
      ? `<p class="hint">Showing the top 120 relationships (out of ${grouped.size}) sorted by repetition count.</p>`
      : "";

  return `
    <header class="connascence-header">
      <h3>Shared meaning hotspots</h3>
      <p class="hint">Review the strongest cross-module naming and meaning couplings.</p>
    </header>
    ${overflowHint}
    <ul class="connascence-list">${rows}</ul>
  `;
}

function buildArchitectureGraph(summary: Summary): {
  nodes: GraphNodeDatum[];
  edges: GraphEdgeDatum[];
  dependencyEdges: number;
  connEdges: number;
  wasTrimmed: boolean;
} {
  const orderedNodes = summary.modules
    .map((module) => ({ id: module.module, module, risk: moduleRiskScore(module) }))
    .sort((a, b) => b.risk - a.risk);
  const wasTrimmed = orderedNodes.length > ARCH_GRAPH_MAX_NODES;
  const nodes = wasTrimmed ? orderedNodes.slice(0, ARCH_GRAPH_MAX_NODES) : orderedNodes;
  const moduleSet = new Set(nodes.map((node) => node.id));
  const edges: GraphEdgeDatum[] = [];
  const seen = new Set<string>();
  let dependencyEdges = 0;
  let connEdges = 0;

  for (const node of nodes) {
    const deps = node.module.dependencies ?? [];
    for (const dep of deps) {
      if (!moduleSet.has(dep) || dep === node.id) continue;
      const key = `${node.id}->${dep}:dependency`;
      if (seen.has(key)) continue;
      seen.add(key);
      edges.push({ source: node.id, target: dep, kind: "dependency", weight: 1 });
      dependencyEdges += 1;
    }
  }

  for (const conn of summary.connascence) {
    if (!moduleSet.has(conn.from) || !moduleSet.has(conn.to) || conn.from === conn.to) continue;
    const kind: GraphEdgeDatum["kind"] = conn.kind === "name" ? "conn-name" : "conn-meaning";
    const key = `${conn.from}->${conn.to}:${kind}`;
    if (seen.has(key)) continue;
    seen.add(key);
    edges.push({ source: conn.from, target: conn.to, kind, weight: 1.4 });
    connEdges += 1;
  }

  return { nodes, edges, dependencyEdges, connEdges, wasTrimmed };
}
