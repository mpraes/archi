import { describe, expect, it } from "vitest";
import { normalizeSummary } from "./api";
import { Summary } from "./types";

describe("normalizeSummary API contract", () => {
  it("maps Go model.Summary JSON fields", () => {
    const raw = {
      projectName: "demo",
      moduleCount: 1,
      modules: [
        {
          module: "alpha",
          path: "/alpha",
          language: "go",
          files: 2,
          afferent: 1,
          efferent: 2,
          instability: 0.66,
          abstraction: 0.33,
          distance: 0.1,
          maxComplexity: 5,
          totalComplexity: 8,
          abstracts: 1,
          concretes: 2,
          orphanBlocks: ["orphan"],
          godBlocks: ["god"],
        },
      ],
      connascence: [{ kind: "meaning", from: "a", to: "b", detail: "shared" }],
      hotspots: ["alpha"],
    };

    const summary: Summary = normalizeSummary(raw);
    expect(summary.projectName).toBe("demo");
    expect(summary.moduleCount).toBe(1);
    expect(summary.modules[0].module).toBe("alpha");
    expect(summary.modules[0].godBlocks).toEqual(["god"]);
    expect(summary.connascence[0].kind).toBe("meaning");
    expect(summary.hotspots).toEqual(["alpha"]);
  });

  it("coerces invalid values safely", () => {
    const summary = normalizeSummary({ modules: "bad", connascence: null });
    expect(summary.modules).toEqual([]);
    expect(summary.connascence).toEqual([]);
    expect(summary.moduleCount).toBe(0);
  });
});
