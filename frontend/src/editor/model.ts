export type Kind = "exact" | "repeating" | "approximate";
export interface Result {
  readonly value: string;
  readonly decimal: string;
  readonly kind: Kind;
}
export type Digit = "0" | "1" | "2" | "3" | "4" | "5" | "6" | "7" | "8" | "9";
export type Character = Digit | "." | "+" | "-" | "*" | "%" | "(" | ")";
export type Node =
  | { readonly type: "char"; readonly char: Character }
  | ({ readonly type: "chip" } & Result)
  | { readonly type: "frac"; readonly num: Row; readonly den: Row }
  | { readonly type: "sqrt"; readonly body: Row }
  | { readonly type: "pow"; readonly exp: Row };
export type Row = readonly Node[];
export type Slot = "num" | "den" | "body" | "exp";
export interface Step {
  readonly index: number;
  readonly slot: Slot;
}
export type Path = readonly Step[];
export interface Cursor {
  readonly path: Path;
  readonly index: number;
}
export interface EditorState {
  readonly root: Row;
  readonly cursor: Cursor;
  readonly selected?: Cursor;
}

export function slots(node: Node): readonly (readonly [Slot, Row])[] {
  switch (node.type) {
    case "frac":
      return [
        ["num", node.num],
        ["den", node.den],
      ];
    case "sqrt":
      return [["body", node.body]];
    case "pow":
      return [["exp", node.exp]];
    default:
      return [];
  }
}

export function rowAt(root: Row, path: Path): Row | undefined {
  let row = root;
  for (const step of path) {
    const node = row[step.index];
    const child = node && slots(node).find(([slot]) => slot === step.slot);
    if (!child) return undefined;
    row = child[1];
  }
  return row;
}

export function isNumberChar(node: Node | undefined): boolean {
  return node?.type === "char" && /^[0-9.]$/.test(node.char);
}

export function isChar(node: Node | undefined, char: Character): boolean {
  return node?.type === "char" && node.char === char;
}

export function operandStart(row: Row, end: number): number {
  let index = end;
  while (
    index > 0 &&
    (row[index - 1].type === "pow" || isChar(row[index - 1], "%"))
  )
    index--;
  if (isNumberChar(row[index - 1])) {
    while (isNumberChar(row[index - 1])) index--;
    return index;
  }
  const node = row[index - 1];
  if (node?.type === "chip" || node?.type === "frac" || node?.type === "sqrt")
    return index - 1;
  if (isChar(node, ")")) {
    let depth = 0;
    for (let i = index - 1; i >= 0; i--) {
      if (isChar(row[i], ")")) depth++;
      if (isChar(row[i], "(") && --depth === 0) return i;
    }
  }
  return end;
}
