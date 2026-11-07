import { expect, it } from "vitest";
import { evaluate } from "../api.ts";
import { chars, rootTwo } from "../test/editor.ts";
import { requests } from "../mocks/handlers.ts";
import type { Row } from "./model.ts";
import {
  clear,
  insertChip,
  insertDigit,
  insertFraction,
  insertOperator,
  moveRight,
} from "./edit.ts";
import { serialize } from "./serialize.ts";

it("edits a fraction, evaluates it and chains its exact result", async () => {
  let state = insertDigit(clear(), "1");
  state = insertFraction(state);
  state = insertDigit(state, "3");
  state = moveRight(state);
  const first = await evaluate(serialize(state.root).text);
  expect(first.type).toBe("result");
  if (first.type !== "result") throw new Error("Expected a result");
  state = insertDigit(
    insertOperator(insertChip(clear(), first.result), "*"),
    "3",
  );
  expect(await evaluate(serialize(state.root).text)).toEqual({
    type: "result",
    result: { value: "1", decimal: "1", kind: "exact" },
  });
  expect(requests.map((request) => request.body)).toEqual([
    '{"expression":"((1)/(3))"}',
    '{"expression":"[1/3]*3"}',
  ]);
});

it.each([
  [
    [...chars("2"), { type: "pow", exp: chars("3") }, ...chars("%")],
    "(2^(3))%",
    "0.08",
  ],
  [
    [
      {
        type: "frac",
        num: chars("1"),
        den: [...chars("2"), { type: "pow", exp: chars("3") }, ...chars("%")],
      },
    ],
    "((1)/(2^(3))%)",
    "12.5",
  ],
  [
    [
      {
        type: "frac",
        num: chars("1"),
        den: [...chars("-2"), { type: "pow", exp: chars("3") }, ...chars("%")],
      },
    ],
    "((1)/-(2^(3))%)",
    "-12.5",
  ],
  [
    [{ type: "frac", num: chars("200"), den: chars("15%") }],
    "((200)/15%)",
    "4000/3",
  ],
  [[{ type: "sqrt", body: chars("2") }], "sqrt(2)", rootTwo.value],
  [
    [
      ...chars("1000*(1+"),
      { type: "frac", num: chars("5"), den: chars("100") },
      ...chars(")"),
      { type: "pow", exp: chars("12") },
    ],
    "1000*(1+((5)/(100)))^(12)",
    "1795.856326022129150390625",
  ],
  [
    [...chars("1500*"), { type: "chip", ...rootTwo }],
    "1500*[~1.4142135623730950488]",
    "~2121.3203435596425732",
  ],
  [
    [
      {
        type: "frac",
        num: [...chars("(2+3)"), { type: "pow", exp: chars("2") }],
        den: chars("3"),
      },
      ...chars("+"),
      { type: "sqrt", body: chars("2.25") },
    ],
    "(((2+3)^(2))/(3))+sqrt(2.25)",
    "59/6",
  ],
] satisfies [Row, string, string][])(
  "sends the approved calculation %j",
  async (row, expression, value) => {
    expect(serialize(row).text).toBe(expression);
    expect(await evaluate(expression)).toMatchObject({
      type: "result",
      result: { value },
    });
  },
);
