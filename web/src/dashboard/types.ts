import { ModuleMetrics } from "../types";

export type LanguageBreakdown = { language: string; modules: number; files: number };
export type MetricDelta = { current: number; previous: number | null; delta: number | null };

export type Snapshot = {
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

export type ModuleSnapshot = {
  distance: number;
  instability: number;
  maxComplexity: number;
  totalComplexity: number;
  orphanCount: number;
  godCount: number;
  efferent: number;
};

export type ModuleDelta = {
  module: string;
  currentScore: number;
  previousScore: number | null;
  deltaScore: number | null;
  currentDistance: number;
  previousDistance: number | null;
  currentMaxComplexity: number;
  previousMaxComplexity: number | null;
};

export type NavItem = { label: string; targetId: string };

export type ModuleSelectHandler = (module: ModuleMetrics) => void;
