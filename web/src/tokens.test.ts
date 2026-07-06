import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

describe("design tokens consistency", () => {
  it("defines core risk and layout CSS variables", () => {
    const tokens = readFileSync(resolve(__dirname, "styles/tokens.css"), "utf8");
    for (const token of ["--bg", "--text", "--accent", "--danger", "--warn", "--ok", "--risk-high"]) {
      expect(tokens).toContain(token);
    }
  });

  it("defines hero section classes", () => {
    const sections = readFileSync(resolve(__dirname, "styles/sections.css"), "utf8");
    for (const cls of [".hero", ".hero-tips", ".hero-priority", ".hero-main-metric"]) {
      expect(sections).toContain(cls);
    }
  });
});
