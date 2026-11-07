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
