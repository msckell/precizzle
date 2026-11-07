import { describe, expect, it } from "vitest";
import { chars, rootTwo, third } from "../test/editor.ts";
import type { Row } from "./model.ts";
import { locate, serialize } from "./serialize.ts";

const fraction = (num: Row, den: Row): Row => [{ type: "frac", num, den }];
const chip: Row = [{ type: "chip", ...third }];
const approx: Row = [{ type: "chip", ...rootTwo }];

describe("serialization", () => {
  it.each([
    [[...chip, ...chars("%")], "100", "100", "exact"],
    [[...chars("-"), ...chip, ...chars("%")], "-100", "-100", "exact"],
    [
      [...approx, ...chars("%")],
      "~23.57022603955158414672",
      "23.57022603955158414672",
      "approximate",
    ],
    [
      [...chars("-"), ...approx, ...chars("%")],
      "~-23.57022603955158414672",
      "-23.57022603955158414672",
      "approximate",
    ],
    [
      [...chars("--"), ...approx, ...chars("%")],
      "~23.57022603955158414672",
      "23.57022603955158414672",
      "approximate",
    ],
    [
      [...chars("("), ...approx, ...chars("%)")],
      "~23.57022603955158414339",
      "23.57022603955158414339",
      "approximate",
    ],
    [
      [...chars("-("), ...approx, ...chars("%)")],
      "~-23.57022603955158414339",
      "-23.57022603955158414339",
      "approximate",
    ],
    [
      [...chars("("), ...approx, ...chars(")%")],
      "~23.57022603955158414672",
      "23.57022603955158414672",
      "approximate",
    ],
  ] satisfies [Row, string, string, string][])(
    "sends percentage denominators with the verified API semantics: %j",
    async (den, value, decimal, kind) => {
      const expression = serialize(fraction(chip, den)).text;
      const response = await fetch("/api/v1/evaluate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ expression }),
      });
      expect(response.status).toBe(200);
      expect(await response.json()).toEqual({ value, decimal, kind });
    },
  );
  it.each([
    [chars("2.+3.5*4"), "2+3.5*4"],
    [fraction(chars("1"), chars("2")), "((1)/(2))"],
    [
      [...fraction(chars("1"), chars("2")), { type: "pow", exp: chars("2") }],
      "((1)/(2))^(2)",
    ],
    [[...chars("2+3"), { type: "pow", exp: chars("2") }], "2+3^(2)"],
    [
      [
        ...chars("2"),
        { type: "pow", exp: chars("3") },
        { type: "pow", exp: chars("2") },
      ],
      "(2^(3))^(2)",
    ],
    [
      [
        ...chars("2"),
        { type: "pow", exp: [...chars("3"), { type: "pow", exp: chars("2") }] },
      ],
      "2^(3^(2))",
    ],
    [
      [{ type: "sqrt", body: fraction(chars("9"), chars("4")) }],
      "sqrt(((9)/(4)))",
    ],
    [
      [
        ...fraction(
          [...chars("(2+3)"), { type: "pow", exp: chars("2") }],
          chars("3"),
        ),
        ...chars("+"),
        { type: "sqrt", body: chars("2.25") },
      ],
      "(((2+3)^(2))/(3))+sqrt(2.25)",
    ],
    [[...chip, ...chars("*7")], "[1/3]*7"],
    [approx, "[~1.4142135623730950488]"],
    [fraction([], []), "(()/())"],
    [[{ type: "sqrt", body: [] }], "sqrt()"],
    [[{ type: "pow", exp: [] }], "^()"],
    [chars(""), ""],
  ] satisfies [Row, string][])("serializes %j as %s", (row, expected) => {
    const { text, spans } = serialize(row);
    expect(text).toBe(expected);
    for (let position = 0; position < text.length; position++)
      expect(locate(spans, position)).toBeDefined();
    expect(locate(spans, text.length)).toEqual({ type: "end" });
  });

  it.each([
    [chars("15%"), "15%"],
    [chars("-15%"), "-15%"],
    [chars("--15%"), "--15%"],
    [[...approx, ...chars("%")], "[~1.4142135623730950488]%"],
    [[...chars("-"), ...chip, ...chars("%")], "-[1/3]%"],
    [chars("(15)%"), "(15)%"],
    [chars("-(15)%"), "-(15)%"],
    [chars("(15%)"), "((15%))"],
    [chars("-(15%)"), "(-(15%))"],
    [chars("15%+2"), "(15%+2)"],
    [chars("2*15%"), "(2*15%)"],
    [chars("15%%"), "(15%%)"],
    [chars("%"), "(%)"],
    [
      [...chars("2"), { type: "pow", exp: chars("3") }, ...chars("%")],
      "(2^(3))%",
    ],
    [[{ type: "sqrt", body: chars("2") }, ...chars("%")], "sqrt(2)%"],
    [[...fraction(chars("1"), chars("3")), ...chars("%")], "((1)/(3))%"],
  ] satisfies [Row, string][])(
    "limits the D81 exception to a bare percentage: %j",
    (den, expected) => {
      expect(serialize(fraction(chip, den)).text).toBe(`(([1/3])/${expected})`);
    },
  );

  it.each([
    [fraction(chars("1"), chars("0")), 4, { type: "bar", path: [], index: 0 }],
    [
      [...chars("0"), { type: "pow", exp: chars("-1") }],
      1,
      { type: "node", path: [], index: 1 },
    ],
    [
      [{ type: "sqrt", body: chars("-4") }],
      0,
      { type: "node", path: [], index: 0 },
    ],
    [
      [...chars("2"), { type: "pow", exp: chars("0.5") }],
      1,
      { type: "node", path: [], index: 1 },
    ],
    [
      [...chars("2"), { type: "pow", exp: chars("1001") }],
      1,
      { type: "node", path: [], index: 1 },
    ],
    [
      [...chars("10"), { type: "pow", exp: chars("50") }],
      2,
      { type: "node", path: [], index: 2 },
    ],
    [chars("2+"), 2, { type: "end" }],
    [chars("(2+3"), 0, { type: "node", path: [], index: 0 }],
    [chars("(2+"), 3, { type: "end" }],
    [chars("2(3)"), 1, { type: "node", path: [], index: 1 }],
    [chars("15%%"), 3, { type: "node", path: [], index: 3 }],
    [chars("1.2.3"), 0, { type: "node", path: [], index: 0 }],
    [chip, 0, { type: "node", path: [], index: 0 }],
    [fraction([], []), 2, { type: "slot", path: [{ index: 0, slot: "num" }] }],
    [fraction([], []), 5, { type: "slot", path: [{ index: 0, slot: "den" }] }],
    [
      [{ type: "sqrt", body: [] }],
      5,
      { type: "slot", path: [{ index: 0, slot: "body" }] },
    ],
    [
      [{ type: "pow", exp: [] }],
      2,
      { type: "slot", path: [{ index: 0, slot: "exp" }] },
    ],
  ] as const)("maps an error at %j / %i", (row, position, target) => {
    expect(locate(serialize(row).spans, position)).toEqual(target);
  });
  it("keeps positions correct after dropping a dot and applying the percentage exception", () => {
    const row = fraction(chars("1."), [
      ...chars("--"),
      ...approx,
      ...chars("%"),
    ]);
    const { text, spans } = serialize(row);
    expect(locate(spans, text.indexOf("/"))).toEqual({
      type: "bar",
      path: [],
      index: 0,
    });
    expect(locate(spans, text.indexOf("%"))).toEqual({
      type: "node",
      path: [{ index: 0, slot: "den" }],
      index: 3,
    });
  });
  it.each([
    [
      [...chars("2"), { type: "pow", exp: chars("3") }, ...chars("%")],
      "(2^(3))%",
    ],
    [[...chars("2"), { type: "pow", exp: chars("3%") }], "2^(3%)"],
    [
      [
        ...chars("2"),
        { type: "pow", exp: chars("3") },
        ...chars("%"),
        { type: "pow", exp: chars("2") },
      ],
      "(2^(3))%^(2)",
    ],
    [
      [...chars("-2"), { type: "pow", exp: chars("3") }, ...chars("%")],
      "-(2^(3))%",
    ],
  ] satisfies [Row, string][])(
    "preserves which operand owns the percentage: %j",
    (row, expected) => {
      expect(serialize(row).text).toBe(expected);
    },
  );
  it.each([-1, 0.5, 10, NaN])(
    "rejects an out-of-range position %s",
    (position) => {
      expect(locate(serialize(chars("2")).spans, position)).toBeUndefined();
    },
  );
});
