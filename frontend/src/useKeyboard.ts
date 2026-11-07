import { useEffect } from "react";
import type { Command } from "./state.ts";
import type { Digit } from "./editor/model.ts";

const keys: Record<string, Command> = {
  ".": { type: "dot" },
  "/": { type: "fraction" },
  "^": { type: "power" },
  r: { type: "sqrt" },
  Enter: { type: "evaluate" },
  "=": { type: "evaluate" },
  Escape: { type: "clear" },
  Backspace: { type: "backspace" },
  Delete: { type: "delete" },
  ArrowLeft: { type: "left" },
  ArrowRight: { type: "right" },
  ArrowUp: { type: "up" },
  ArrowDown: { type: "down" },
  Home: { type: "home" },
  End: { type: "end" },
};
export function useKeyboard(
  onCommand: (command: Command) => void,
  historyOpen: boolean,
  onCloseHistory: () => void,
) {
  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (
        event.ctrlKey ||
        event.metaKey ||
        event.altKey ||
        event.isComposing ||
        event.defaultPrevented
      )
        return;
      if (historyOpen) {
        if (event.key === "Escape") {
          event.preventDefault();
          onCloseHistory();
        }
        return;
      }
      if (
        (event.key === "Enter" || event.key === " ") &&
        event.target instanceof HTMLElement &&
        event.target.closest("button")
      )
        return;
      let command: Command | undefined = keys[event.key];
      if (/^[0-9]$/.test(event.key))
        command = { type: "digit", digit: event.key as Digit };
      else if (
        event.key === "+" ||
        event.key === "-" ||
        event.key === "*" ||
        event.key === "%"
      )
        command = { type: "operator", operator: event.key };
      else if (event.key === "(" || event.key === ")")
        command = { type: "paren", paren: event.key };
      if (command) {
        event.preventDefault();
        onCommand(command);
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [onCommand, historyOpen, onCloseHistory]);
}
