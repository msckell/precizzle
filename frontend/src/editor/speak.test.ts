import { expect, it } from "vitest";
import { chars, rootTwo, third } from "../test/editor.ts";
import type { Row } from "./model.ts";
import { speak } from "./speak.ts";

it.each([
  [[], "empty"],
  [chars("-12.5+3*4%"), "minus 12.5 plus 3 times 4 percent"],
  [chars("(2-3)"), "open parenthesis 2 minus 3 close parenthesis"],
  [
    [{ type: "frac", num: chars("1"), den: chars("3") }],
    "fraction 1 over 3 end fraction",
  ],
  [[{ type: "sqrt", body: [] }], "square root of empty end root"],
  [[...chars("2"), { type: "pow", exp: chars("2") }], "2 squared"],
  [[...chars("2"), { type: "pow", exp: chars("3") }], "2 cubed"],
  [
    [...chars("2"), { type: "pow", exp: chars("-4") }],
    "2 to the power of minus 4 end exponent",
  ],
  [[{ type: "chip", ...third }], "1 over 3"],
  [[{ type: "chip", ...rootTwo }], "approximately 1.4142135623730950488"],
  [
    [
      {
        type: "chip",
        value: "~-1e-25",
        decimal: "-1e-25",
        kind: "approximate",
      },
    ],
    "approximately minus 1 times ten to the power of minus 25",
  ],
] satisfies [Row, string][])("reads %j as %s", (row, expected) =>
  expect(speak(row)).toBe(expected),
);
