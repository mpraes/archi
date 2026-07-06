import { describe, expect, it } from "vitest";
import { asArray, asBoolean, asNumber, asRecord, asString, asStringArray } from "./types/guards";

describe("runtime guards", () => {
  it("coerces primitives safely", () => {
    expect(asNumber(3)).toBe(3);
    expect(asNumber("x")).toBe(0);
    expect(asBoolean(true)).toBe(true);
    expect(asBoolean("true")).toBe(false);
    expect(asString("ok")).toBe("ok");
    expect(asString(1)).toBe("");
  });

  it("normalizes records and arrays", () => {
    expect(asRecord({ a: 1 }).a).toBe(1);
    expect(asRecord(null)).toEqual({});
    expect(asArray([1, 2])).toHaveLength(2);
    expect(asStringArray(["a", 1])).toEqual(["a"]);
  });
});
