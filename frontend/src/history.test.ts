import { afterEach, describe, expect, it, vi } from "vitest";
import { chars, rootTwo, third } from "./test/editor.ts";
import { load, save, HISTORY_KEY } from "./history.ts";
import type { HistoryEntry, HistoryStorage } from "./history.ts";

const entry: HistoryEntry = {
  expression: [
    {
      type: "frac",
      num: [{ type: "chip", ...third }],
      den: [{ type: "sqrt", body: chars("2") }],
    },
    { type: "pow", exp: chars("-2") },
  ],
  result: rootTwo,
};
afterEach(() => {
  vi.restoreAllMocks();
  localStorage.clear();
});
describe("history", () => {
  it("round trips the editable tree and exact values", () => {
    expect(save([entry])).toBe(true);
    expect(load()).toEqual([entry]);
    expect(load()[0]).not.toBe(entry);
    expect(localStorage.getItem(HISTORY_KEY)).toBe(
      JSON.stringify({ version: 1, entries: [entry] }),
    );
  });
  it("keeps 100 entries newest first and drops the oldest", () => {
    const entries = Array.from({ length: 101 }, (_, i) => ({
      ...entry,
      expression: chars(String(i)),
    }));
    save(entries);
    expect(load()).toEqual(entries.slice(0, 100));
    localStorage.setItem(HISTORY_KEY, JSON.stringify({ version: 1, entries }));
    expect(load()).toEqual(entries.slice(0, 100));
  });
  it("clears stored history", () => {
    save([entry]);
    save([]);
    expect(load()).toEqual([]);
  });
  it("treats missing storage as empty", () => expect(load()).toEqual([]));
  it.each([
    "{",
    "null",
    "[]",
    "{}",
    '{"version":2,"entries":[]}',
    '{"version":1,"entries":{}}',
  ])("treats corrupt or wrong-version data as empty: %s", (raw) => {
    localStorage.setItem(HISTORY_KEY, raw);
    expect(load()).toEqual([]);
  });
  it.each([
    null,
    {},
    { expression: [], result: { value: 2, decimal: "2", kind: "exact" } },
    { ...entry, expression: null },
    { ...entry, expression: [null] },
    { ...entry, expression: [{ type: "unknown" }] },
    { ...entry, expression: [{ type: "char", char: "e" }] },
    { ...entry, expression: [{ type: "char", char: "12" }] },
    { ...entry, expression: [{ type: "char", char: 2 }] },
    {
      ...entry,
      expression: [
        { type: "chip", value: "[1]+2", decimal: "3", kind: "exact" },
      ],
    },
    { ...entry, expression: [{ type: "frac", num: [], den: {} }] },
    { ...entry, expression: [{ type: "sqrt" }] },
    { ...entry, expression: [{ type: "pow", exp: false }] },
    { ...entry, result: { value: "1", decimal: "1", kind: "other" } },
  ])("rejects the whole history when an entry is invalid: %j", (invalid) => {
    localStorage.setItem(
      HISTORY_KEY,
      JSON.stringify({ version: 1, entries: [entry, invalid] }),
    );
    expect(load()).toEqual([]);
  });
  it("bounds nesting from untrusted storage", () => {
    let expression: unknown = [];
    for (let i = 0; i < 102; i++)
      expression = [{ type: "sqrt", body: expression }];
    localStorage.setItem(
      HISTORY_KEY,
      JSON.stringify({ version: 1, entries: [{ ...entry, expression }] }),
    );
    expect(load()).toEqual([]);
  });
  it("uses the versioned key with an injected storage", () => {
    const storage: HistoryStorage = {
      getItem: vi.fn(() => null),
      setItem: vi.fn(),
    };
    expect(load(storage)).toEqual([]);
    expect(storage.getItem).toHaveBeenCalledWith("precizzle.history.v1");
    expect(save([entry], storage)).toBe(true);
    expect(storage.setItem).toHaveBeenCalledWith(
      HISTORY_KEY,
      JSON.stringify({ version: 1, entries: [entry] }),
    );
  });
  it("handles storage methods that throw", () => {
    const fail = () => {
      throw new Error("Storage unavailable");
    };
    const storage = { getItem: fail, setItem: fail };
    expect(load(storage)).toEqual([]);
    expect(save([entry], storage)).toBe(false);
  });
  it("handles the browser denying access to localStorage itself", () => {
    vi.spyOn(window, "localStorage", "get").mockImplementation(() => {
      throw new DOMException("Denied", "SecurityError");
    });
    expect(load()).toEqual([]);
    expect(save([entry])).toBe(false);
  });
});
