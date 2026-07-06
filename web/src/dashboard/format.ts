import { Connascence, ModuleMetrics } from "../types";
import { LanguageBreakdown } from "./types";

export function dominantConnascenceLabel(conns: Connascence[]): string {
  if (conns.length === 0) return "No critical dependencies";
  const nameCount = conns.filter((conn) => conn.kind === "name").length;
  const meaningCount = conns.length - nameCount;
  if (nameCount === meaningCount) return "Balanced between name and meaning";
  return nameCount > meaningCount ? "Name connascence dominates" : "Meaning connascence dominates";
}

export function languageBreakdown(modules: ModuleMetrics[]): LanguageBreakdown[] {
  const map = new Map<string, { modules: number; files: number }>();
  for (const module of modules) {
    const lang = (module.language || "unknown").toLowerCase();
    const current = map.get(lang) ?? { modules: 0, files: 0 };
    current.modules += 1;
    current.files += module.files;
    map.set(lang, current);
  }

  return [...map.entries()]
    .map(([language, counts]) => ({ language, modules: counts.modules, files: counts.files }))
    .sort((a, b) => b.modules - a.modules || b.files - a.files);
}

export function formatCapturedAt(isoDate: string): string {
  const parsed = new Date(isoDate);
  if (Number.isNaN(parsed.getTime())) return "unknown date";
  return new Intl.DateTimeFormat("en-US", { dateStyle: "short", timeStyle: "short" }).format(parsed);
}
