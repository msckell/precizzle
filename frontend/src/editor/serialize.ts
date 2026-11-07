import { isChar, isNumberChar, operandStart } from "./model.ts";
import type { Path, Row, Slot } from "./model.ts";

export type Element =
  | {
      readonly type: "node" | "bar";
      readonly path: Path;
      readonly index: number;
    }
  | { readonly type: "slot"; readonly path: Path }
  | { readonly type: "end" };
export interface Span {
  readonly start: number;
  readonly end: number;
  readonly element: Element;
}
export interface Serialized {
  readonly text: string;
  readonly spans: readonly Span[];
}

function barePercentage(row: Row): boolean {
  const end = row.length - 1;
  if (!isChar(row[end], "%") || isChar(row[end - 1], "%")) return false;
  const start = operandStart(row, end);
  return start < end && row.slice(0, start).every((node) => isChar(node, "-"));
}

export function serialize(root: Row): Serialized {
  let text = "";
  const spans: Span[] = [];
  const write = (part: string, element: Element) => {
    const start = text.length;
    text += part;
    spans.push({ start, end: text.length, element });
  };
  const slot = (row: Row, path: Path, grouped = true) => {
    const element: Element = { type: "slot", path };
    if (grouped) write("(", element);
    visit(row, path);
    if (grouped) write(")", element);
  };
  const visit = (row: Row, path: Path) => {
    // A following power or percentage belongs to the whole power, not its exponent.
    const groups = row.flatMap((node, index) => {
      if (
        (node.type !== "pow" && !isChar(node, "%")) ||
        row[index - 1]?.type !== "pow"
      )
        return [];
      const start = operandStart(row, index);
      return start < index ? [{ start, end: index }] : [];
    });
    row.forEach((node, index) => {
      const element: Element = { type: "node", path, index };
      for (const group of groups.filter((group) => group.start === index))
        write("(", { type: "node", path, index: group.end });
      for (const group of groups.filter((group) => group.end === index))
        write(")", { type: "node", path, index: group.end });
      const childPath = (name: Slot): Path => [...path, { index, slot: name }];
      switch (node.type) {
        case "char":
          if (!(node.char === "." && !isNumberChar(row[index + 1])))
            write(node.char, element);
          break;
        case "chip":
          write(`[${node.value}]`, element);
          break;
        case "sqrt":
          write("sqrt", element);
          slot(node.body, childPath("body"));
          break;
        case "pow":
          write("^", element);
          slot(node.exp, childPath("exp"));
          break;
        case "frac":
          write("(", element);
          slot(node.num, childPath("num"));
          write("/", { type: "bar", path, index });
          // Added parentheses would evaluate B% separately and change D55 rounding.
          slot(node.den, childPath("den"), !barePercentage(node.den));
          write(")", element);
          break;
      }
    });
  };
  visit(root, []);
  write("", { type: "end" });
  return { text, spans };
}

export function locate(
  spans: readonly Span[],
  position: number,
): Element | undefined {
  if (!Number.isInteger(position) || position < 0) return undefined;
  return spans.find(
    (span) =>
      position >= span.start &&
      (position < span.end ||
        (span.element.type === "end" && position === span.start)),
  )?.element;
}
