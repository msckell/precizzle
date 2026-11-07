import { describe, expect, it } from "vitest";
import type { Cursor, EditorState, Row } from "./model.ts";
import {
  clear,
  moveDown,
  moveEnd,
  moveHome,
  moveLeft,
  moveRight,
  moveUp,
  placeCursor,
} from "./edit.ts";

const root: Row = [
  { type: "char", char: "2" },
  {
    type: "frac",
    num: [{ type: "sqrt", body: [{ type: "char", char: "3" }] }],
    den: [{ type: "char", char: "4" }],
  },
  { type: "pow", exp: [{ type: "char", char: "5" }] },
];
const positions: Cursor[] = [
  { path: [], index: 0 },
  { path: [], index: 1 },
  { path: [{ index: 1, slot: "num" }], index: 0 },
  {
    path: [
      { index: 1, slot: "num" },
      { index: 0, slot: "body" },
    ],
    index: 0,
  },
  {
    path: [
      { index: 1, slot: "num" },
      { index: 0, slot: "body" },
    ],
    index: 1,
  },
  { path: [{ index: 1, slot: "num" }], index: 1 },
  { path: [{ index: 1, slot: "den" }], index: 0 },
  { path: [{ index: 1, slot: "den" }], index: 1 },
  { path: [], index: 2 },
  { path: [{ index: 2, slot: "exp" }], index: 0 },
  { path: [{ index: 2, slot: "exp" }], index: 1 },
  { path: [], index: 3 },
];
const state = (cursor: Cursor): EditorState => ({ root, cursor });

describe("cursor movement", () => {
  it.each(
    positions.slice(0, -1).map((from, i) => ({ from, to: positions[i + 1] })),
  )("walks right and left through $from", ({ from, to }) => {
    expect(moveRight(state(from))).toEqual(state(to));
    expect(moveLeft(state(to))).toEqual(state(from));
  });
  it.each([
    [moveLeft, positions[0]],
    [moveRight, positions.at(-1)!],
    [moveDown, positions[0]],
    [moveUp, positions[0]],
  ])("stays at an outer boundary", (move, cursor) => {
    expect(move(state(cursor))).toEqual(state(cursor));
  });
  it.each([
    [moveDown, positions[2], positions[6]],
    [moveUp, positions[7], positions[5]],
    [moveDown, positions[4], positions[7]],
    [moveUp, positions[8], positions[10]],
    [moveUp, positions[11], positions[10]],
    [moveDown, positions[10], positions[11]],
  ])("moves vertically between related slots", (move, from, to) => {
    expect(move(state(from))).toEqual(state(to));
  });
  it("clamps the column in a shorter fraction slot", () => {
    const s: EditorState = {
      root: [{ type: "frac", num: [], den: root }],
      cursor: { path: [{ index: 0, slot: "den" }], index: 3 },
    };
    expect(moveUp(s).cursor).toEqual({
      path: [{ index: 0, slot: "num" }],
      index: 0,
    });
  });
  it("uses Home and End within the active row and clears selection", () => {
    const s = { ...state(positions[7]), selected: positions[1] };
    expect(moveHome(s)).toEqual(state(positions[6]));
    expect(moveEnd(moveHome(s))).toEqual(state(positions[7]));
  });
  it("places a cursor without changing the tree", () => {
    expect(placeCursor(state(positions[0]), positions[4].path, 1)).toEqual(
      state(positions[4]),
    );
  });
  it.each([
    { path: [], index: -1 },
    { path: [], index: 4 },
    { path: [], index: 0.5 },
    { path: [{ index: 0, slot: "num" }], index: 0 },
    { path: [{ index: 99, slot: "exp" }], index: 0 },
  ] satisfies Cursor[])(
    "ignores invalid placement $index $path",
    ({ path, index }) => {
      expect(placeCursor(state(positions[0]), path, index)).toEqual(
        state(positions[0]),
      );
    },
  );
  it("walks through empty slots", () => {
    const s: EditorState = {
      root: [{ type: "frac", num: [], den: [] }],
      cursor: { path: [], index: 0 },
    };
    const num = moveRight(s);
    const den = moveRight(num);
    expect(num.cursor).toEqual({ path: [{ index: 0, slot: "num" }], index: 0 });
    expect(den.cursor).toEqual({ path: [{ index: 0, slot: "den" }], index: 0 });
    expect(moveLeft(den)).toEqual(num);
    expect(moveLeft(num)).toEqual(s);
    expect(moveRight(den).cursor).toEqual({ path: [], index: 1 });
  });
  it("clears the expression and cursor", () => {
    expect(clear()).toEqual({ root: [], cursor: { path: [], index: 0 } });
  });
});
