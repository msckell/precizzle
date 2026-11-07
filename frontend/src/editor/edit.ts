import { isChar, isNumberChar, operandStart, rowAt, slots } from "./model.ts";
import type {
  Character,
  Cursor,
  Digit,
  EditorState,
  Node,
  Path,
  Result,
  Row,
  Slot,
} from "./model.ts";

function at(state: EditorState, cursor: Cursor): EditorState {
  return { root: state.root, cursor };
}

export function clear(): EditorState {
  return { root: [], cursor: { path: [], index: 0 } };
}

export function placeCursor(
  state: EditorState,
  path: Path,
  index: number,
): EditorState {
  const row = rowAt(state.root, path);
  return row && Number.isInteger(index) && index >= 0 && index <= row.length
    ? at(state, { path, index })
    : state;
}

function horizontal(state: EditorState, right: boolean): EditorState {
  const { path, index } = state.cursor;
  const row = rowAt(state.root, path)!;
  const adjacent = row[right ? index : index - 1];
  if (adjacent) {
    const children = slots(adjacent);
    const child = right ? children[0] : children.at(-1);
    return at(
      state,
      child
        ? {
            path: [
              ...path,
              { index: right ? index : index - 1, slot: child[0] },
            ],
            index: right ? 0 : child[1].length,
          }
        : { path, index: index + (right ? 1 : -1) },
    );
  }
  const step = path.at(-1);
  if (!step) return at(state, state.cursor);
  const parentPath = path.slice(0, -1);
  const children = slots(rowAt(state.root, parentPath)![step.index]);
  const sibling =
    children[
      children.findIndex(([slot]) => slot === step.slot) + (right ? 1 : -1)
    ];
  return at(
    state,
    sibling
      ? {
          path: [...parentPath, { index: step.index, slot: sibling[0] }],
          index: right ? 0 : sibling[1].length,
        }
      : { path: parentPath, index: step.index + (right ? 1 : 0) },
  );
}

export const moveLeft = (state: EditorState): EditorState =>
  horizontal(state, false);
export const moveRight = (state: EditorState): EditorState =>
  horizontal(state, true);
export const moveHome = (state: EditorState): EditorState =>
  at(state, { ...state.cursor, index: 0 });
export const moveEnd = (state: EditorState): EditorState =>
  at(state, {
    ...state.cursor,
    index: rowAt(state.root, state.cursor.path)!.length,
  });

function vertical(state: EditorState, up: boolean): EditorState {
  const { path, index } = state.cursor;
  for (let depth = path.length - 1; depth >= 0; depth--) {
    const step = path[depth];
    const parentPath = path.slice(0, depth);
    if (step.slot === (up ? "den" : "num")) {
      const targetPath: Path = [
        ...parentPath,
        { index: step.index, slot: up ? "num" : "den" },
      ];
      return at(state, {
        path: targetPath,
        index: Math.min(index, rowAt(state.root, targetPath)!.length),
      });
    }
    if (!up && step.slot === "exp")
      return at(state, { path: parentPath, index: step.index + 1 });
  }
  if (up) {
    const row = rowAt(state.root, path)!;
    const powerIndex = row[index]?.type === "pow" ? index : index - 1;
    const power = row[powerIndex];
    if (power?.type === "pow") {
      return at(state, {
        path: [...path, { index: powerIndex, slot: "exp" }],
        index: power.exp.length,
      });
    }
  }
  return at(state, state.cursor);
}
export const moveUp = (state: EditorState): EditorState =>
  vertical(state, true);
export const moveDown = (state: EditorState): EditorState =>
  vertical(state, false);

const charNode = (char: Character): Node => ({ type: "char", char });

function replaceRow(root: Row, path: Path, replacement: Row): Row {
  const [step, ...rest] = path;
  if (!step) return replacement;
  return root.map((node, index) =>
    index === step.index
      ? {
          ...node,
          [step.slot]: replaceRow(rowAt(root, [step])!, rest, replacement),
        }
      : node,
  );
}

function splice(
  state: EditorState,
  start: number,
  count: number,
  nodes: Row,
  cursorIndex = start + nodes.length,
): EditorState {
  const { path } = state.cursor;
  const row = rowAt(state.root, path)!;
  const next = [...row.slice(0, start), ...nodes, ...row.slice(start + count)];
  const normalized: Node[] = [];
  let inNumber = false;
  let dot = false;
  for (let i = 0; i < next.length; i++) {
    const node = next[i];
    if (isNumberChar(node)) {
      if (isChar(node, ".")) {
        // Joining two decimals must not silently discard either operand's digits.
        if (dot) return state;
        if (!inNumber) {
          normalized.push(charNode("0"));
          if (i < cursorIndex) cursorIndex++;
        }
        dot = true;
      }
      inNumber = true;
    } else {
      inNumber = false;
      dot = false;
    }
    normalized.push(node);
  }
  return {
    root: replaceRow(state.root, path, normalized),
    cursor: { path, index: cursorIndex },
  };
}

export function insertDigit(state: EditorState, digit: Digit): EditorState {
  return splice(state, state.cursor.index, 0, [charNode(digit)]);
}

export function insertDot(state: EditorState): EditorState {
  const row = rowAt(state.root, state.cursor.path)!;
  const { index } = state.cursor;
  let start = index;
  let end = index;
  while (isNumberChar(row[start - 1])) start--;
  while (isNumberChar(row[end])) end++;
  if (row.slice(start, end).some((node) => isChar(node, "."))) return state;
  return splice(
    state,
    index,
    0,
    index === start ? [charNode("0"), charNode(".")] : [charNode(".")],
  );
}

export function insertOperator(
  state: EditorState,
  operator: "+" | "-" | "*" | "%",
): EditorState {
  return splice(state, state.cursor.index, 0, [charNode(operator)]);
}

function needsMultiply(state: EditorState): boolean {
  const node = rowAt(state.root, state.cursor.path)![state.cursor.index - 1];
  return (
    !!node &&
    (node.type !== "char" ||
      isNumberChar(node) ||
      node.char === ")" ||
      node.char === "%")
  );
}

export function insertParen(state: EditorState, paren: "(" | ")"): EditorState {
  const nodes =
    paren === "(" && needsMultiply(state)
      ? [charNode("*"), charNode(paren)]
      : [charNode(paren)];
  return splice(state, state.cursor.index, 0, nodes);
}

function enter(state: EditorState, index: number, slot: Slot): EditorState {
  return at(state, { path: [...state.cursor.path, { index, slot }], index: 0 });
}

export function insertFraction(state: EditorState): EditorState {
  const row = rowAt(state.root, state.cursor.path)!;
  const end = state.cursor.index;
  const start = operandStart(row, end);
  const num = row.slice(start, end);
  const next = splice(state, start, end - start, [
    { type: "frac", num, den: [] },
  ]);
  return enter(next, start, num.length ? "den" : "num");
}

export function insertPower(state: EditorState): EditorState {
  const index = state.cursor.index;
  return enter(
    splice(state, index, 0, [{ type: "pow", exp: [] }]),
    index,
    "exp",
  );
}

export function insertSqrt(state: EditorState): EditorState {
  const multiply = needsMultiply(state);
  const index = state.cursor.index;
  const node: Node = { type: "sqrt", body: [] };
  return enter(
    splice(state, index, 0, multiply ? [charNode("*"), node] : [node]),
    index + (multiply ? 1 : 0),
    "body",
  );
}

export function insertChip(state: EditorState, result: Result): EditorState {
  return splice(state, state.cursor.index, 0, [{ type: "chip", ...result }]);
}

function unaryMinus(row: Row, index: number): boolean {
  return (
    isChar(row[index], "-") &&
    (index === 0 || operandStart(row, index) === index)
  );
}

export function toggleSign(state: EditorState): EditorState {
  const row = rowAt(state.root, state.cursor.path)!;
  let start = state.cursor.index;
  if (isNumberChar(row[start])) {
    while (isNumberChar(row[start - 1])) start--;
  } else start = operandStart(row, start);
  if (unaryMinus(row, start - 1))
    return splice(
      state,
      start - 1,
      1,
      [],
      Math.max(start - 1, state.cursor.index - 1),
    );
  return splice(state, start, 0, [charNode("-")], state.cursor.index + 1);
}

function remove(state: EditorState, forward: boolean): EditorState {
  if (state.selected) {
    const selected = state.selected;
    return splice(at(state, selected), selected.index, 1, []);
  }
  const { path, index } = state.cursor;
  const row = rowAt(state.root, path)!;
  const target = forward ? index : index - 1;
  const node = row[target];
  if (node) {
    const children = slots(node);
    if (children.some(([, child]) => child.length))
      return { ...state, selected: { path, index: target } };
    return splice(state, target, 1, []);
  }
  const parent = path.at(-1);
  if (parent) {
    const parentPath = path.slice(0, -1);
    const structure = rowAt(state.root, parentPath)![parent.index];
    if (slots(structure).every(([, child]) => child.length === 0)) {
      return splice(
        at(state, { path: parentPath, index: parent.index }),
        parent.index,
        1,
        [],
      );
    }
  }
  return horizontal(state, forward);
}

export const backspace = (state: EditorState): EditorState =>
  remove(state, false);
export const deleteForward = (state: EditorState): EditorState =>
  remove(state, true);
