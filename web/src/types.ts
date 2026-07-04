export interface ModuleMetrics {
  module: string;
  path: string;
  files: number;
  afferent: number;
  efferent: number;
  instability: number;
  abstraction: number;
  distance: number;
  maxComplexity: number;
  totalComplexity: number;
  abstracts: number;
  concretes: number;
  orphanBlocks: string[] | null;
  godBlocks: string[] | null;
}

export interface Connascence {
  kind: "name" | "meaning";
  from: string;
  to: string;
  detail: string;
}

export interface Summary {
  projectName: string;
  moduleCount: number;
  modules: ModuleMetrics[];
  connascence: Connascence[];
  hotspots: string[];
}