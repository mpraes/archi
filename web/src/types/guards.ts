export function asRecord(v: unknown): Record<string, unknown> {
  return v !== null && typeof v === "object" ? (v as Record<string, unknown>) : {};
}

export function asArray(v: unknown): unknown[] {
  return Array.isArray(v) ? v : [];
}

export function asNumber(v: unknown): number {
  return typeof v === "number" && Number.isFinite(v) ? v : 0;
}

export function asBoolean(v: unknown): boolean {
  return v === true;
}

export function asString(v: unknown): string {
  return typeof v === "string" ? v : "";
}

export function asStringArray(v: unknown): string[] {
  if (!Array.isArray(v)) return [];
  return v.filter((item): item is string => typeof item === "string");
}
