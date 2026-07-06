import { describe, expect, it, beforeEach } from "vitest";
import { defaultNavItems, renderDashboardShell } from "./dashboard/layout";
import { renderHero } from "./dashboard/sections/hero";
import { Summary } from "./types";

function mountApp(): void {
  document.body.innerHTML = `<div id="app"></div>`;
}

describe("dashboard landing consistency", () => {
  beforeEach(() => {
    mountApp();
  });

  it("renders shell with hero entry point and primary navigation", () => {
    const nav = defaultNavItems();
    renderDashboardShell(nav);

    expect(document.querySelector("#hero")).not.toBeNull();
    expect(document.querySelector(".dashboard-shell")).not.toBeNull();
    expect(document.querySelector(".side-nav a[href='#hero']")).not.toBeNull();
    expect(document.querySelector("#status")?.textContent).toContain("Loading");
    expect(nav[0].targetId).toBe("hero");
  });

  it("renders hero priority metrics and connascence summary", () => {
    renderDashboardShell(defaultNavItems());
    const summary: Summary = {
      projectName: "mini-go",
      moduleCount: 2,
      modules: [
        { module: "a", path: "a", language: "go", files: 1, afferent: 0, efferent: 1, instability: 1, abstraction: 0, distance: 0, maxComplexity: 1, totalComplexity: 1, abstracts: 0, concretes: 0, orphanBlocks: ["x"], godBlocks: [] },
        { module: "b", path: "b", language: "go", files: 1, afferent: 1, efferent: 0, instability: 0, abstraction: 0, distance: 1, maxComplexity: 1, totalComplexity: 1, abstracts: 0, concretes: 0, orphanBlocks: [], godBlocks: [] },
      ],
      connascence: [{ kind: "name", from: "a", to: "b", detail: "calls" }],
      hotspots: [],
    };

    renderHero(summary, null);
    const hero = document.querySelector("#hero");
    expect(hero?.querySelector(".hero-main-metric strong")?.textContent).toContain("module");
    expect(hero?.textContent).toContain("Architecture overview");
    expect(hero?.textContent).toContain("mini-go");
  });
});
