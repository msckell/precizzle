import { describe, expect, it } from "vitest";
import { chars, state, third } from "../test/editor.ts";
import type { EditorState, Node, Row } from "./model.ts";
import {
  backspace,
  deleteForward,
  insertChip,
  insertDigit,
  insertDot,
  insertFraction,
  insertOperator,
  insertParen,
  insertPower,
  insertSqrt,
  moveRight,
  placeCursor,
  toggleSign,
} from "./edit.ts";

describe("insertion", () => {
  it("keeps the cursor before the repaired right-hand number when splitting a decimal", () => {
    expect(insertOperator(state("1.2", 1), "+")).toEqual(state("1+0.2", 2));
  });
  it.each([
    ["", 0, "0."],
    ["0", 1, "0."],
    ["12", 1, "1.2"],
    ["12", 0, "0.12"],
    ["1.2", 0, "1.2"],
    ["1.2", 3, "1.2"],
  ])("inserts a decimal in %s at %i", (input, index, expected) => {
    expect(insertDot(state(input, index)).root).toEqual(chars(expected));
  });
  it.each([
    ["", 0, "7"],
    ["12", 1, "172"],
    ["00", 2, "007"],
    ["0.", 2, "0.7"],
  ])("inserts a digit into %s", (input, index, expected) => {
    expect(insertDigit(state(input, index), "7")).toEqual(
      state(expected, index + 1),
    );
  });
  it.each(["+", "-", "*", "%"] as const)("inserts operator %s", (op) => {
    expect(insertOperator(state("12", 1), op)).toEqual(state(`1${op}2`, 2));
  });
  it.each(["2", "2.", ")", "15%"])(
    "inserts visible multiplication after %s",
    (input) => {
      expect(insertParen(state(input), "(").root).toEqual(chars(input + "*("));
      expect(insertSqrt(state(input)).root).toEqual([
        ...chars(input + "*"),
        { type: "sqrt", body: [] },
      ]);
    },
  );
  const chip: Node = { type: "chip", ...third };
  const times: Node = { type: "char", char: "*" };
  const half: Node = { type: "frac", num: chars("1"), den: chars("2") };
  const root: Node = { type: "sqrt", body: chars("2") };
  const square: Node = { type: "pow", exp: chars("2") };
  it.each<[string, Row, number, Row, number]>([
    ["after a number", chars("2"), 1, [...chars("2*"), chip], 3],
    ["after a trailing dot", chars("2."), 2, [...chars("2.*"), chip], 4],
    ["after a group", chars("(2)"), 3, [...chars("(2)*"), chip], 5],
    ["after a percentage", chars("15%"), 3, [...chars("15%*"), chip], 5],
    ["after a chip", [chip], 1, [chip, times, chip], 3],
    ["after a fraction", [half], 1, [half, times, chip], 3],
    ["after a root", [root], 1, [root, times, chip], 3],
    [
      "after a power",
      [...chars("2"), square],
      2,
      [...chars("2"), square, times, chip],
      4,
    ],
    ["before a number", chars("5"), 0, [chip, ...chars("*5")], 1],
    ["before a group", chars("(5)"), 0, [chip, ...chars("*(5)")], 1],
    ["before a chip", [chip], 0, [chip, times, chip], 1],
    ["before a fraction", [half], 0, [chip, times, half], 1],
    ["before a root", [root], 0, [chip, times, root], 1],
    [
      "between two digits",
      chars("23"),
      1,
      [...chars("2*"), chip, ...chars("*3")],
      3,
    ],
    ["in an empty row", [], 0, [chip], 1],
    ["after an operator", chars("2+"), 2, [...chars("2+"), chip], 3],
    ["after a unary minus", chars("-"), 1, [...chars("-"), chip], 2],
    ["after an opening parenthesis", chars("("), 1, [...chars("("), chip], 2],
    ["before an operator", chars("+2"), 0, [chip, ...chars("+2")], 1],
    ["before a closing parenthesis", chars(")"), 0, [chip, ...chars(")")], 1],
    ["before a percent sign", chars("%"), 0, [chip, ...chars("%")], 1],
    ["before a power it becomes the base of", [square], 0, [chip, square], 1],
  ])(
    "inserts a chip %s with a visible multiplication only where needed",
    (_, row, index, expected, cursor) => {
      expect(insertChip(state(row, index), third)).toEqual(
        state(expected, cursor),
      );
    },
  );
  it("inserts a chip with multiplication inside a slot", () => {
    const editor: EditorState = {
      root: [{ type: "sqrt", body: chars("4") }],
      cursor: { path: [{ index: 0, slot: "body" }], index: 1 },
    };
    expect(insertChip(editor, third)).toEqual({
      root: [{ type: "sqrt", body: [...chars("4*"), chip] }],
      cursor: { path: [{ index: 0, slot: "body" }], index: 3 },
    });
  });
  it.each(["", "2+", "("])("does not multiply after %s", (input) => {
    expect(insertParen(state(input), "(").root).toEqual(chars(input + "("));
    expect(insertSqrt(state(input)).root).toEqual([
      ...chars(input),
      { type: "sqrt", body: [] },
    ]);
  });
  it("inserts a closing parenthesis literally", () => {
    expect(insertParen(state("2"), ")")).toEqual(state("2)"));
  });
  it.each([
    { input: chars("123"), prefix: [], operand: chars("123") },
    { input: chars("2+3"), prefix: chars("2+"), operand: chars("3") },
    { input: chars("-23"), prefix: chars("-"), operand: chars("23") },
    { input: chars("2+(3+4)"), prefix: chars("2+"), operand: chars("(3+4)") },
    { input: chars("((2+3)*4)%"), prefix: [], operand: chars("((2+3)*4)%") },
    {
      input: [{ type: "chip", ...third }],
      prefix: [],
      operand: [{ type: "chip", ...third }],
    },
    {
      input: [{ type: "sqrt", body: chars("4") }],
      prefix: [],
      operand: [{ type: "sqrt", body: chars("4") }],
    },
    {
      input: [{ type: "frac", num: chars("1"), den: chars("2") }],
      prefix: [],
      operand: [{ type: "frac", num: chars("1"), den: chars("2") }],
    },
    {
      input: [...chars("2%"), { type: "pow", exp: chars("3") }],
      prefix: [],
      operand: [...chars("2%"), { type: "pow", exp: chars("3") }],
    },
    { input: [], prefix: [], operand: [] },
    { input: chars("2+"), prefix: chars("2+"), operand: [] },
    { input: chars("2)"), prefix: chars("2)"), operand: [] },
  ] satisfies { input: Row; prefix: Row; operand: Row }[])(
    "captures only the preceding operand $input",
    ({ input, prefix, operand }) => {
      const fraction = insertFraction(state(input));
      expect(fraction.root).toEqual([
        ...prefix,
        { type: "frac", num: operand, den: [] },
      ]);
      expect(fraction.cursor).toEqual({
        path: [{ index: prefix.length, slot: operand.length ? "den" : "num" }],
        index: 0,
      });
      const power = insertPower(state(input));
      expect(power.root).toEqual([...input, { type: "pow", exp: [] }]);
      expect(power.cursor).toEqual({
        path: [{ index: input.length, slot: "exp" }],
        index: 0,
      });
    },
  );
  it("captures only digits before a cursor inside a number", () => {
    expect(insertFraction(state("1234", 2)).root).toEqual([
      { type: "frac", num: chars("12"), den: [] },
      ...chars("34"),
    ]);
  });
  it("keeps exact chip values and edits nested slots immutably", () => {
    const initial = insertSqrt(state(""));
    const snapshot = JSON.stringify(initial);
    const result = insertChip(initial, third);
    expect(result.root).toEqual([
      { type: "sqrt", body: [{ type: "chip", ...third }] },
    ]);
    expect(JSON.stringify(initial)).toBe(snapshot);
    expect(result.cursor.index).toBe(1);
  });
  it.each([
    { type: "chip", ...third },
    { type: "sqrt", body: chars("2") },
    { type: "pow", exp: chars("2") },
  ] satisfies Node[])("multiplies before a new root after $type", (node) => {
    expect(insertSqrt(state([node])).root[1]).toEqual({
      type: "char",
      char: "*",
    });
  });
});

describe("deletion and signs", () => {
  it("toggles a unary sign when the cursor is immediately before it", () => {
    expect(toggleSign(state("-12", 0))).toEqual(state("12", 0));
  });
  it("does not remove a selected structure when typing at its cursor", () => {
    const original = state([{ type: "sqrt", body: chars("2") }]);
    const changed = insertDigit(backspace(original), "3");
    expect(changed.root).toEqual([...original.root, ...chars("3")]);
    expect(changed.selected).toBeUndefined();
  });
  it.each([
    { type: "frac", num: chars("1"), den: chars("2") },
    { type: "sqrt", body: chars("2") },
    { type: "pow", exp: chars("2") },
  ] satisfies Node[])(
    "selects and then deletes $type in either direction",
    (node) => {
      for (const [remove, index] of [
        [backspace, 2],
        [deleteForward, 1],
      ] as const) {
        const before = state([...chars("3"), node], index);
        const selected = remove(before);
        expect(selected.root).toEqual(before.root);
        expect(selected.selected).toEqual({ path: [], index: 1 });
        expect(remove(selected)).toEqual(state("3"));
        expect(moveRight(selected).selected).toBeUndefined();
      }
    },
  );
  it.each([
    { type: "frac", num: [], den: [] },
    { type: "sqrt", body: [] },
    { type: "pow", exp: [] },
  ] satisfies Node[])("deletes empty $type immediately", (node) => {
    expect(backspace(state([node]))).toEqual(state(""));
    expect(deleteForward(state([node], 0))).toEqual(state(""));
  });
  it.each([
    ["123", 2, "13"],
    ["123", 0, "123"],
    ["0.5", 1, "0.5"],
    ["1.2+3.4", 4, "1.2+3.4"],
  ])("backspaces %s at %i", (input, index, expected) => {
    expect(backspace(state(input, index)).root).toEqual(chars(expected));
  });
  it.each([
    ["123", 1, "13"],
    ["123", 3, "123"],
    ["0.5", 0, "0.5"],
    ["1.2+3.4", 3, "1.2+3.4"],
  ])("deletes %s at %i", (input, index, expected) => {
    expect(deleteForward(state(input, index)).root).toEqual(chars(expected));
  });
  it("deletes a chip atomically", () => {
    expect(backspace(state([{ type: "chip", ...third }]))).toEqual(state(""));
  });
  it("moves out of a populated slot at its edge without destroying it", () => {
    const s = insertDigit(insertSqrt(state("")), "2");
    expect(deleteForward(s).root).toEqual(s.root);
    expect(deleteForward(s).cursor).toEqual({ path: [], index: 1 });
    expect(backspace(placeCursor(s, s.cursor.path, 0)).cursor).toEqual({
      path: [],
      index: 0,
    });
  });
  it("deletes an empty structure from inside its empty slot", () => {
    expect(backspace(insertSqrt(state("")))).toEqual(state(""));
    expect(deleteForward(insertPower(state("2")))).toEqual(state("2"));
  });
  it.each([
    ["", 0, "-"],
    ["-", 1, ""],
    ["12", 2, "-12"],
    ["12", 1, "-12"],
    ["-12", 3, "12"],
    ["2+", 2, "2+-"],
    ["2-", 2, "2--"],
    ["2--", 3, "2-"],
    ["2-3", 3, "2--3"],
    ["2+-3", 4, "2+3"],
    ["(2+3)", 5, "-(2+3)"],
    ["15%", 3, "-15%"],
  ])("toggles sign in %s at %i", (input, index, expected) => {
    expect(toggleSign(state(input, index)).root).toEqual(chars(expected));
  });
  it("edits a deep numerator without changing other rows", () => {
    const initial: EditorState = {
      root: [
        {
          type: "sqrt",
          body: [{ type: "frac", num: chars("12"), den: chars("3") }],
        },
      ],
      cursor: {
        path: [
          { index: 0, slot: "body" },
          { index: 0, slot: "num" },
        ],
        index: 1,
      },
    };
    const changed = insertDigit(initial, "7");
    expect(changed.root).toEqual([
      {
        type: "sqrt",
        body: [{ type: "frac", num: chars("172"), den: chars("3") }],
      },
    ]);
    expect(initial.root).toEqual([
      {
        type: "sqrt",
        body: [{ type: "frac", num: chars("12"), den: chars("3") }],
      },
    ]);
  });
});
