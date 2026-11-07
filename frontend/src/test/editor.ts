import type { Character, EditorState, Row, Result } from "../editor/model.ts";

export const chars = (text: string): Row =>
  [...text].map((char) => ({ type: "char", char: char as Character }));
export const state = (row: Row | string, index?: number): EditorState => {
  const root = typeof row === "string" ? chars(row) : row;
  return { root, cursor: { path: [], index: index ?? root.length } };
};
export const third: Result = {
  value: "1/3",
  decimal: "0." + "3".repeat(100),
  kind: "repeating",
};
export const rootTwo: Result = {
  value: "~1.4142135623730950488",
  decimal: "1.4142135623730950488",
  kind: "approximate",
};
