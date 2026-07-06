import { fetchMetrics } from "./api";
import { defaultNavItems, renderDashboardShell, setupNavigation } from "./dashboard/layout";
import { buildSnapshot, readPreviousSnapshot, saveSnapshot } from "./dashboard/snapshot";
import { renderActivity } from "./dashboard/sections/activity";
import { renderAI } from "./dashboard/sections/ai";
import { renderArchitectureMap } from "./dashboard/sections/architecture-map";
import { renderHero } from "./dashboard/sections/hero";
import { renderKpiGrid } from "./dashboard/sections/kpi";
import { renderModulePanel } from "./dashboard/sections/module-panel";
import { renderDetailedReport } from "./dashboard/sections/report";
import { renderStatus } from "./dashboard/sections/status";
import { requiredEl } from "./dashboard/ui";
import { Summary } from "./types";
import "./style.css";

let data: Summary | null = null;

async function main() {
  const navItems = defaultNavItems();
  renderDashboardShell(navItems);
  setupNavigation(navItems);

  try {
    data = await fetchMetrics();
  } catch (e) {
    requiredEl<HTMLElement>("status").textContent = "Analysis failed: " + (e as Error).message;
    return;
  }
  const previousSnapshot = readPreviousSnapshot();
  renderStatus(data, previousSnapshot);
  renderHero(data, previousSnapshot);
  renderKpiGrid(data, previousSnapshot);
  renderActivity(data, previousSnapshot);
  renderArchitectureMap(data, (module) => renderModulePanel(data!, module));
  renderDetailedReport(data, previousSnapshot, (module) => renderModulePanel(data!, module));
  renderAI();
  saveSnapshot(buildSnapshot(data));
}

void main();