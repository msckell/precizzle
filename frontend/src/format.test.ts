import { describe, expect, it } from "vitest";
import type { Kind, Result } from "./editor/model.ts";
import { chipText, formatResult, splitDecimal } from "./format.ts";
import type { Expansion } from "./format.ts";

const result = (decimal: string, kind: Kind = "exact"): Result => ({
  value: decimal,
  decimal,
  kind,
});
describe("display formatting", () => {
  it.each([
    ["0", "exact", "0", false],
    ["-12", "exact", "-12", false],
    ["0.1234", "repeating", "0.1234", false],
    ["0.12345", "repeating", "0.12345", false],
    ["0.123456", "repeating", "0.1234…", true],
    ["0.12345", "exact", "0.12345", false],
    ["-0.123459", "exact", "-0.12345…", true],
    ["1.4142135623730950488", "approximate", "1.41421…", true],
    [
      "9999999999999999999999999999999999999999",
      "exact",
      "9999999999999999999999999999999999999999",
      false,
    ],
    ["-1.23456789e-900", "approximate", "-1.23456…e-900", true],
    ["1e-25", "exact", "1e-25", false],
  ] satisfies [string, Kind, string, boolean][])(
    "formats %s (%s) without rounding",
    (decimal, kind, text, hasMore) => {
      const input = result(decimal, kind);
      expect(formatResult(input)).toMatchObject({ text, hasMore });
      expect(input.decimal).toBe(decimal);
    },
  );
  it.each([1, 2, 3, 4, 5] as const)(
    "expands to step %i and caps at 100",
    (expansion) => {
      const display = formatResult(
        result("0." + "1234567890".repeat(100)),
        expansion,
      );
      expect(display.shownDecimals).toBe(expansion * 20);
      expect(display.mantissa).toBe("0." + "1234567890".repeat(expansion * 2));
      expect(display.hasMore).toBe(expansion < 5);
      expect(display.nextExpansion).toBe(expansion < 5 ? expansion + 1 : null);
    },
  );
  it.each([4, 5, 19, 20, 21, 39, 40, 41, 59, 60, 61, 79, 80, 81, 99, 100, 101])(
    "stops early for %i fractional digits",
    (length) => {
      const input = result("0." + "1".repeat(length));
      for (const step of [1, 2, 3, 4, 5] satisfies Expansion[]) {
        const display = formatResult(input, step);
        expect(display.shownDecimals).toBe(Math.min(length, step * 20));
        expect(display.hasMore).toBe(Math.min(length, 100) > step * 20);
      }
    },
  );
  it.each([
    ["-1.234e-25", "-1.234", "-25"],
    ["12", "12", undefined],
    ["2E+3", "2", "+3"],
  ])("splits scientific %s as strings", (decimal, mantissa, exponent) => {
    expect(splitDecimal(decimal)).toEqual({ mantissa, exponent });
  });
  it("keeps the exponent when expanding", () => {
    expect(
      formatResult(result("3." + "3".repeat(100) + "e-31", "repeating"), 1),
    ).toMatchObject({
      text: "3." + "3".repeat(20) + "…e-31",
      exponent: "-31",
      shownDecimals: 20,
    });
  });
  it.each([
    [result("0.33333333", "repeating"), "0.3333…"],
    [result("1.4142135", "approximate"), "≈1.41421…"],
    [result("-2"), "-2"],
  ])("builds a short chip label", (input, expected) =>
    expect(chipText(input)).toBe(expected),
  );
});
