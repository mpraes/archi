export function requiredEl<T extends HTMLElement>(id: string): T {
  const el = document.getElementById(id);
  if (!el) {
    throw new Error(`element #${id} not found`);
  }
  return el as T;
}

export function readCssVar(name: string, fallback: string): string {
  const value = window.getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return value || fallback;
}

export function escapeAttr(value: string): string {
  return value.replace(/&/g, "&amp;").replace(/"/g, "&quot;").replace(/</g, "&lt;");
}

export function infoButton(text: string, label: string): string {
  return `<button class="info-button" type="button" aria-label="${escapeAttr(label)}" data-info="${escapeAttr(text)}">i</button>`;
}
