import { NavItem } from "./types";
import { requiredEl } from "./ui";

export function defaultNavItems(): NavItem[] {
  return [
    { label: "Overview", targetId: "hero" },
    { label: "Modules", targetId: "report" },
    { label: "Architecture Map", targetId: "architecture-map" },
    { label: "Connascence", targetId: "connascence-view" },
    { label: "Risk", targetId: "kpi-grid" },
    { label: "History", targetId: "activity" },
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
            <h2>Dashboard</h2>
            <p id="status" role="status" aria-live="polite">Loading architecture diagnosis…</p>
          </div>
          <div class="top-actions">
            <span id="history-badge" class="badge">No baseline yet</span>
          </div>
        </header>
        <nav class="compact-nav" aria-label="Primary navigation">
          ${navLinks}
        </nav>
        <section id="hero" class="hero"></section>
        <section class="main">
          <div class="main-left">
            <section id="report" class="report"></section>
          </div>
          <aside class="main-right">
            <section id="panel" class="panel empty" role="region" aria-live="polite" tabindex="-1">
              <p class="hint">Click a module in the table (or map) to see its diagnosis.</p>
            </section>
          </aside>
        </section>
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
        <section id="kpi-grid" class="kpi-grid"></section>
        <section id="activity" class="panel activity" aria-live="polite"></section>
        <section id="ai-panel" class="ai hidden"></section>
      </main>
    </div>
  `;
}

export function setupNavigation(navItems: NavItem[]): void {
  const links = [...document.querySelectorAll<HTMLAnchorElement>(".side-nav a[data-target], .compact-nav a[data-target]")];
  const linksByTarget = new Map<string, HTMLAnchorElement[]>();

  for (const link of links) {
    const targetId = link.dataset.target;
    if (!targetId) continue;
    const bucket = linksByTarget.get(targetId) ?? [];
    bucket.push(link);
    linksByTarget.set(targetId, bucket);
    link.addEventListener("click", (event) => {
      event.preventDefault();
      const target = document.getElementById(targetId);
      if (!target) return;
      target.scrollIntoView({ behavior: "smooth", block: "start" });
      setActiveNavTarget(linksByTarget, targetId);
    });
  }

  const observer = new IntersectionObserver(
    (entries) => {
      const visible = entries
        .filter((entry) => entry.isIntersecting)
        .sort((a, b) => b.intersectionRatio - a.intersectionRatio);
      const top = visible[0];
      if (top?.target?.id) {
        setActiveNavTarget(linksByTarget, top.target.id);
      }
    },
    {
      root: null,
      threshold: [0.25, 0.4, 0.6],
      rootMargin: "-15% 0px -55% 0px",
    },
  );

  for (const item of navItems) {
    const section = document.getElementById(item.targetId);
    if (section) observer.observe(section);
  }
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
