import type { Result } from "./editor/model.ts";

export type Expansion = 0 | 1 | 2 | 3 | 4 | 5;

export function splitDecimal(decimal: string): {
  mantissa: string;
  exponent: string | undefined;
} {
  const [mantissa, exponent] = decimal.split(/e/i);
  return { mantissa, exponent };
}

export function formatResult(result: Result, expansion: Expansion = 0) {
  const { mantissa, exponent } = splitDecimal(result.decimal);
  const [integer, fraction = ""] = mantissa.split(".");
  // Below 1 the short view counts digits from the first non-zero one, so a tiny value never reads as 0.
  const zeros = /^-?0$/.test(integer)
    ? Math.max(fraction.search(/[1-9]/), 0)
    : 0;
  const first = (result.kind === "repeating" ? 4 : 5) + zeros;
  const limits = [first, ...[20, 40, 60, 80, 100].filter((n) => n > first)];
  const limit =
    expansion === 0 && fraction.length <= 5 + zeros
      ? fraction.length
      : limits[expansion];
  const shownDecimals = Math.min(fraction.length, limit);
  const visible =
    integer + (shownDecimals ? "." + fraction.slice(0, shownDecimals) : "");
  const hasMore = shownDecimals < Math.min(fraction.length, 100);
  return {
    text:
      visible +
      (hasMore ? "…" : "") +
      (exponent === undefined ? "" : `e${exponent}`),
    mantissa: visible,
    exponent,
    shownDecimals,
    hasMore,
    nextExpansion: hasMore ? ((expansion + 1) as Expansion) : null,
  };
}

export function expansionCaption(result: Result, expansion: Expansion): string {
  const { mantissa } = splitDecimal(result.decimal);
  const total = Math.min(100, mantissa.split(".")[1]?.length ?? 0);
  return `Showing ${formatResult(result, expansion).shownDecimals} of ${total} decimals`;
}

export function chipText(result: Result): string {
  return (result.kind === "approximate" ? "≈" : "") + formatResult(result).text;
}
