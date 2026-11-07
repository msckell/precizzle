import { isResult } from "./api.ts";
import type { Result, Row } from "./editor/model.ts";

export const HISTORY_KEY = "precizzle.history.v1";
export interface HistoryEntry {
  readonly expression: Row;
  readonly result: Result;
}
export type HistoryStorage = Pick<Storage, "getItem" | "setItem">;

function isRow(value: unknown, depth = 0): value is Row {
  if (!Array.isArray(value) || depth > 100) return false;
  return value.every((node: unknown) => {
    if (typeof node !== "object" || node === null) return false;
    const item = node as Record<string, unknown>;
    switch (item.type) {
      case "char":
        return (
          typeof item.char === "string" && /^[0-9.+*%()-]$/.test(item.char)
        );
      case "chip":
        return isResult(item);
      case "frac":
        return isRow(item.num, depth + 1) && isRow(item.den, depth + 1);
      case "sqrt":
        return isRow(item.body, depth + 1);
      case "pow":
        return isRow(item.exp, depth + 1);
      default:
        return false;
    }
  });
}

function isEntry(value: unknown): value is HistoryEntry {
  if (typeof value !== "object" || value === null) return false;
  const entry = value as Record<string, unknown>;
  return isRow(entry.expression) && isResult(entry.result);
}

export function load(storage?: HistoryStorage): HistoryEntry[] {
  try {
    const raw = (storage ?? window.localStorage).getItem(HISTORY_KEY);
    if (raw === null) return [];
    const data: unknown = JSON.parse(raw);
    if (typeof data !== "object" || data === null) return [];
    const record = data as Record<string, unknown>;
    return record.version === 1 &&
      Array.isArray(record.entries) &&
      record.entries.every(isEntry)
      ? record.entries.slice(0, 100)
      : [];
  } catch {
    return [];
  }
}

export function save(
  entries: readonly HistoryEntry[],
  storage?: HistoryStorage,
): boolean {
  try {
    (storage ?? window.localStorage).setItem(
      HISTORY_KEY,
      JSON.stringify({ version: 1, entries: entries.slice(0, 100) }),
    );
    return true;
  } catch {
    return false;
  }
}
