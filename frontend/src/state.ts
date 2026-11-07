import * as edit from "./editor/edit.ts";
import type { Digit, EditorState, Path, Result, Row } from "./editor/model.ts";
import { locate } from "./editor/serialize.ts";
import type { Span } from "./editor/serialize.ts";
import type { Evaluation } from "./api.ts";
import { load } from "./history.ts";
import type { HistoryEntry } from "./history.ts";
import { formatResult } from "./format.ts";
import type { Expansion } from "./format.ts";
import { messageFor } from "./messages.ts";
import type { EquationError } from "./components/Equation.tsx";

export type EditorCommand =
  | { type: "digit"; digit: Digit }
  | { type: "operator"; operator: "+" | "-" | "*" | "%" }
  | { type: "paren"; paren: "(" | ")" }
  | { type: "place"; path: Path; index: number }
  | { type: "chip"; result: Result }
  | {
      type:
        | "dot"
        | "fraction"
        | "power"
        | "sqrt"
        | "sign"
        | "backspace"
        | "delete"
        | "left"
        | "right"
        | "up"
        | "down"
        | "home"
        | "end"
        | "clear";
    };
export type Command = EditorCommand | { type: "evaluate" };
export interface AppState {
  editor: EditorState;
  phase: "editing" | "pending" | "result" | "error";
  result?: Result;
  error?: EquationError;
  history: HistoryEntry[];
  historyOpen: boolean;
  confirmClear: boolean;
  expansion: Expansion;
  copyStatus: "" | "copied" | "failed";
  resultVersion: number;
  requestId: number;
}
export type Action =
  | { type: "edit"; command: EditorCommand }
  | { type: "pending"; id: number }
  | {
      type: "evaluated";
      id: number;
      outcome: Evaluation;
      expression: Row;
      spans: readonly Span[];
    }
  | { type: "history"; open: boolean }
  | { type: "load"; entry: HistoryEntry }
  | { type: "use"; result: Result }
  | { type: "confirm-clear"; confirm: boolean }
  | { type: "clear-history" | "expand" }
  | { type: "copied"; version: number; success: boolean };

export function initialState(): AppState {
  return {
    editor: edit.clear(),
    phase: "editing",
    history: load(),
    historyOpen: false,
    confirmClear: false,
    expansion: 0,
    copyStatus: "",
    resultVersion: 0,
    requestId: 0,
  };
}
function apply(editor: EditorState, command: EditorCommand): EditorState {
  switch (command.type) {
    case "digit":
      return edit.insertDigit(editor, command.digit);
    case "operator":
      return edit.insertOperator(editor, command.operator);
    case "paren":
      return edit.insertParen(editor, command.paren);
    case "place":
      return edit.placeCursor(editor, command.path, command.index);
    case "chip":
      return edit.insertChip(editor, command.result);
    case "dot":
      return edit.insertDot(editor);
    case "fraction":
      return edit.insertFraction(editor);
    case "power":
      return edit.insertPower(editor);
    case "sqrt":
      return edit.insertSqrt(editor);
    case "sign":
      return edit.toggleSign(editor);
    case "backspace":
      return edit.backspace(editor);
    case "delete":
      return edit.deleteForward(editor);
    case "left":
      return edit.moveLeft(editor);
    case "right":
      return edit.moveRight(editor);
    case "up":
      return edit.moveUp(editor);
    case "down":
      return edit.moveDown(editor);
    case "home":
      return edit.moveHome(editor);
    case "end":
      return edit.moveEnd(editor);
    case "clear":
      return edit.clear();
  }
}
function input(state: AppState, command: EditorCommand): AppState {
  let editor = state.editor;
  let startsNew = false;
  if (state.phase === "result" && state.result) {
    startsNew = [
      "operator",
      "fraction",
      "power",
      "digit",
      "dot",
      "paren",
      "sqrt",
      "chip",
    ].includes(command.type);
    if (["operator", "fraction", "power"].includes(command.type))
      editor = edit.insertChip(edit.clear(), state.result);
    else if (["digit", "dot", "paren", "sqrt", "chip"].includes(command.type))
      editor = edit.clear();
  }
  return {
    ...state,
    editor: apply(editor, command),
    phase: "editing",
    error: undefined,
    copyStatus: "",
    ...(command.type === "clear" || startsNew
      ? { result: undefined, expansion: 0 }
      : {}),
  };
}
export function reducer(state: AppState, action: Action): AppState {
  switch (action.type) {
    case "edit":
      return input(state, action.command);
    case "pending":
      return {
        ...state,
        phase: "pending",
        requestId: action.id,
        error: undefined,
        copyStatus: "",
      };
    case "evaluated": {
      if (state.phase !== "pending" || state.requestId !== action.id)
        return state;
      const { outcome } = action;
      if (outcome.type === "result")
        return {
          ...state,
          phase: "result",
          result: outcome.result,
          error: undefined,
          expansion: 0,
          copyStatus: "",
          resultVersion: state.resultVersion + 1,
          history: [
            { expression: action.expression, result: outcome.result },
            ...state.history,
          ].slice(0, 100),
        };
      const code =
        outcome.type === "problem" ? outcome.problem.code : outcome.reason;
      const position =
        outcome.type === "problem" ? outcome.problem.position : undefined;
      return {
        ...state,
        phase: "error",
        error: {
          message: messageFor(code),
          target:
            position === undefined ? undefined : locate(action.spans, position),
        },
      };
    }
    case "history":
      return { ...state, historyOpen: action.open, confirmClear: false };
    case "load":
      return {
        ...state,
        editor: {
          root: action.entry.expression,
          cursor: { path: [], index: action.entry.expression.length },
        },
        result: action.entry.result,
        phase: "result",
        error: undefined,
        historyOpen: false,
        confirmClear: false,
        expansion: 0,
        copyStatus: "",
        resultVersion: state.resultVersion + 1,
      };
    case "use":
      return {
        ...input(state, { type: "chip", result: action.result }),
        historyOpen: false,
        confirmClear: false,
      };
    case "confirm-clear":
      return { ...state, confirmClear: action.confirm };
    case "clear-history":
      return { ...state, history: [], confirmClear: false };
    case "expand":
      return state.result
        ? {
            ...state,
            expansion:
              formatResult(state.result, state.expansion).nextExpansion ??
              state.expansion,
            copyStatus: "",
          }
        : state;
    case "copied":
      return action.version === state.resultVersion && state.result
        ? { ...state, copyStatus: action.success ? "copied" : "failed" }
        : state;
  }
}
