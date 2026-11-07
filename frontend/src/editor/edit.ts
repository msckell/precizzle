import { rowAt, slots } from "./model.ts";
import type { Cursor, EditorState, Path } from "./model.ts";

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
