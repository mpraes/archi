import { NavItem } from "./types";
import { requiredEl } from "./ui";

export function defaultNavItems(): NavItem[] {
  return [
    { label: "Architecture overview", targetId: "hero" },
    { label: "Module diagnosis", targetId: "report" },
    { label: "Dependency map", targetId: "architecture-map" },
    { label: "How it works", targetId: "how-it-works" },
    { label: "Shared meaning hotspots", targetId: "connascence-view" },
    { label: "Risk indicators", targetId: "kpi-grid" },
    { label: "Scan history", targetId: "activity" },
  ];
}

export function renderDashboardShell(navItems: NavItem[]): void {
  const app = requiredEl<HTMLElement>("app");
  const navLinks = navItems
    .map(
      (item, index) =>
        `<a ${index === 0 ? 'class="active" aria-current="page"' : ""} href="#${item.targetId}" data-target="${item.targetId}">${item.label}</a>`,
    )
    .join("");

  app.innerHTML = `
    <div class="dashboard-shell">
      <aside class="sidebar">
        <div class="brand">
          <h1>Archi</h1>
          <p>Architecture Dashboard</p>
        </div>
        <nav class="side-nav" aria-label="Primary navigation">
          ${navLinks}
        </nav>
      </aside>
      <main class="workspace">
        <header class="topbar">
          <div>
            <h2 id="workspace-title">Architecture overview</h2>
            <p id="status" role="status" aria-live="polite">Loading architecture diagnosis…</p>
          </div>
          <div class="top-actions">
            <span id="history-badge" class="badge">No baseline yet</span>
            <a class="doc-link" href="#how-it-works">How it works</a>
          </div>
        </header>
        <nav class="compact-nav" aria-label="Primary navigation">
          ${navLinks}
        </nav>
        <section class="dashboard-page is-active" data-page="overview">
          <section id="hero" class="hero"></section>
        </section>
        <section class="dashboard-page" data-page="modules" hidden>
          <section class="main stacked">
            <div class="main-left">
              <section id="report" class="report"></section>
            </div>
          </section>
          <section id="panel" class="panel empty module-diagnosis" role="region" aria-live="polite" tabindex="-1">
            <p class="hint">Click a module in the table (or map) to see its diagnosis.</p>
            <p class="hint">Need context first? Read <a href="#how-it-works">How it works</a>.</p>
          </section>
        </section>
        <section class="dashboard-page" data-page="architecture" hidden>
          <section id="architecture-map" class="panel architecture-map" aria-live="polite">
            <header class="architecture-map-header">
              <h3>Architecture Map</h3>
              <p>Interactive block-code architecture graph with module-level dependencies and connascence links.</p>
            </header>
            <section class="architecture-map-workspace">
              <div id="architecture-map-meta" class="architecture-map-meta"></div>
              <div id="architecture-map-content" class="architecture-map-content"></div>
            </section>
          </section>
        </section>
        <section class="dashboard-page" data-page="connascence" hidden>
          <section id="connascence-view" class="panel connascence-view" aria-live="polite"></section>
        </section>
        <section class="dashboard-page" data-page="docs" hidden>
          <section id="how-it-works" class="panel docs-panel" aria-live="polite">
            <header>
              <h3>How it works</h3>
              <p class="hint">A quick guide to read this dashboard and take action.</p>
            </header>
            <section>
              <h4>1) Run an analysis</h4>
              <p>The CLI scans modules, computes architecture and complexity metrics, and opens this dashboard with the latest snapshot.</p>
            </section>
            <section>
              <h4>2) Prioritize what to fix first</h4>
              <p>Start in <strong>Module diagnosis</strong>. Use Immediate risk, regressions, and the module list to focus on the highest-impact modules.</p>
            </section>
            <section>
              <h4>3) Inspect coupling in the map</h4>
              <p>Use <strong>Dependency map</strong> to understand cross-module links. Click a node to load detailed diagnosis in the module section.</p>
            </section>
            <section>
              <h4>4) Track progress over time</h4>
              <p>When baseline history exists, regressions and improvements compare the current run against the previous snapshot.</p>
            </section>
            <section>
              <h4>Metric hints</h4>
              <ul>
                <li><strong>D</strong> (distance): how far a module is from the ideal balance between abstraction and stability.</li>
                <li><strong>Instability</strong>: tendency to change due to outgoing dependencies.</li>
                <li><strong>Complexity</strong>: max and total cyclomatic load per module.</li>
                <li><strong>Connascence</strong>: shared names/meanings that couple modules.</li>
              </ul>
            </section>
          </section>
        </section>
        <section class="dashboard-page" data-page="risk" hidden>
          <section id="kpi-grid" class="kpi-grid"></section>
        </section>
        <section class="dashboard-page" data-page="history" hidden>
          <section id="activity" class="panel activity" aria-live="polite"></section>
          <section id="ai-panel" class="ai hidden"></section>
        </section>
      </main>
    </div>
  `;
}

export function setupNavigation(navItems: NavItem[]): void {
  const links = [...document.querySelectorAll<HTMLAnchorElement>(".side-nav a[data-target], .compact-nav a[data-target]")];
  const linksByTarget = new Map<string, HTMLAnchorElement[]>();
  const pages = [...document.querySelectorAll<HTMLElement>(".dashboard-page")];

  for (const link of links) {
    const targetId = link.dataset.target;
    if (!targetId) continue;
    const bucket = linksByTarget.get(targetId) ?? [];
    bucket.push(link);
    linksByTarget.set(targetId, bucket);
    link.addEventListener("click", (event) => {
      event.preventDefault();
      activateTarget(targetId, linksByTarget, pages, true);
    });
  }

  window.addEventListener("hashchange", () => {
    const targetId = pickHashTarget(window.location.hash, navItems);
    if (targetId) activateTarget(targetId, linksByTarget, pages, false);
  });

  window.addEventListener("archi:navigate", (event) => {
    const custom = event as CustomEvent<{ targetId?: string }>;
    const targetId = custom.detail?.targetId;
    if (!targetId) return;
    activateTarget(targetId, linksByTarget, pages, false);
  });

  const initialTarget = pickHashTarget(window.location.hash, navItems) ?? navItems[0]?.targetId;
  if (initialTarget) activateTarget(initialTarget, linksByTarget, pages, false);
}

export function requestNavigation(targetId: string): void {
  window.dispatchEvent(new CustomEvent("archi:navigate", { detail: { targetId } }));
}

function activateTarget(
  targetId: string,
  linksByTarget: Map<string, HTMLAnchorElement[]>,
  pages: HTMLElement[],
  updateHash: boolean,
): void {
  const target = document.getElementById(targetId);
  if (!target) return;

  const targetPage = target.closest<HTMLElement>(".dashboard-page");
  if (targetPage) {
    for (const page of pages) {
      const isActive = page === targetPage;
      page.classList.toggle("is-active", isActive);
      page.hidden = !isActive;
    }
  }

  setActiveNavTarget(linksByTarget, targetId);
  updateWorkspaceTitle(targetId, linksByTarget);
  if (updateHash) {
    window.location.hash = targetId;
  }
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function pickHashTarget(hash: string, navItems: NavItem[]): string | null {
  if (!hash || hash === "#") return null;
  const target = hash.slice(1);
  return navItems.some((item) => item.targetId === target) ? target : null;
}

function updateWorkspaceTitle(targetId: string, linksByTarget: Map<string, HTMLAnchorElement[]>): void {
  const title = document.getElementById("workspace-title");
  if (!title) return;
  const links = linksByTarget.get(targetId);
  const label = links?.[0]?.textContent?.trim();
  if (!label) return;
  title.textContent = label;
  document.title = `${label} | Archi`;
}

function setActiveNavTarget(linksByTarget: Map<string, HTMLAnchorElement[]>, targetId: string): void {
  linksByTarget.forEach((group, id) => {
    for (const link of group) {
      const isActive = id === targetId;
      link.classList.toggle("active", isActive);
      if (isActive) link.setAttribute("aria-current", "page");
      else link.removeAttribute("aria-current");
    }
  });
}
