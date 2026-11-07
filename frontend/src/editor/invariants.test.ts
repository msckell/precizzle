import { expect, it } from "vitest";
import {
  backspace,
  clear,
  deleteForward,
  insertDigit,
  insertDot,
  insertFraction,
  insertOperator,
  insertParen,
  insertPower,
  insertSqrt,
  moveDown,
  moveEnd,
  moveHome,
  moveLeft,
  moveRight,
  moveUp,
  toggleSign,
} from "./edit.ts";
import { rowAt, slots } from "./model.ts";
import type { EditorState, Row } from "./model.ts";
import { locate, serialize } from "./serialize.ts";

function freeze(value: object) {
  Object.freeze(value);
  for (const child of Object.values(value))
    if (typeof child === "object" && child !== null) freeze(child);
}

function checkNumbers(row: Row) {
  let number = "";
  const check = () => {
    if (number) expect(number).toMatch(/^\d+(?:\.\d*)?$/);
    number = "";
  };
  for (const node of row) {
    if (node.type === "char" && /^[0-9.]$/.test(node.char)) number += node.char;
    else check();
    for (const [, child] of slots(node)) checkNumbers(child);
  }
  check();
}

it.each([7, 31, 101])(
  "preserves cursor, numeric input and immutability through mixed editing (seed %i)",
  (seed) => {
    const actions: ((state: EditorState) => EditorState)[] = [
      (s) => insertDigit(s, "0"),
      (s) => insertDigit(s, "7"),
      insertDot,
      (s) => insertOperator(s, "+"),
      (s) => insertOperator(s, "%"),
      (s) => insertParen(s, "("),
      (s) => insertParen(s, ")"),
      insertFraction,
      insertPower,
      insertSqrt,
      toggleSign,
      backspace,
      deleteForward,
      moveLeft,
      moveRight,
      moveUp,
      moveDown,
      moveHome,
      moveEnd,
    ];
    let state = clear();
    for (let i = 0; i < 300; i++) {
      seed = (seed * 48271) % 2147483647;
      freeze(state);
      state = actions[seed % actions.length](state);
      const row = rowAt(state.root, state.cursor.path);
      expect(row).toBeDefined();
      expect(state.cursor.index).toBeGreaterThanOrEqual(0);
      expect(state.cursor.index).toBeLessThanOrEqual(row!.length);
      checkNumbers(state.root);
      const { text, spans } = serialize(state.root);
      expect(locate(spans, text.length)).toEqual({ type: "end" });
      expect(
        spans.map((span) => text.slice(span.start, span.end)).join(""),
      ).toBe(text);
    }
  },
);
