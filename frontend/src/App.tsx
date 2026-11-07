import { useEffect, useReducer, useRef } from "react";
import { evaluate } from "./api.ts";
import { serialize } from "./editor/serialize.ts";
import { save } from "./history.ts";
import { initialState, reducer } from "./state.ts";
import type { Command } from "./state.ts";
import { useKeyboard } from "./useKeyboard.ts";
import { Equation } from "./components/Equation.tsx";
import { Keypad } from "./components/Keypad.tsx";
import { ResultCard } from "./components/ResultCard.tsx";
import { Header } from "./components/Header.tsx";
import { HistoryPanel } from "./components/HistoryPanel.tsx";

function App() {
  const [state, dispatch] = useReducer(reducer, undefined, initialState);
  const pending = useRef<AbortController | undefined>(undefined);
  const nextRequest = useRef(0);
  useEffect(() => () => pending.current?.abort(), []);
  const historyButton = useRef<HTMLButtonElement>(null);
  const historyWasOpen = useRef(false);
  useEffect(() => {
    save(state.history);
  }, [state.history]);
  useEffect(() => {
    if (historyWasOpen.current && !state.historyOpen)
      historyButton.current?.focus();
    historyWasOpen.current = state.historyOpen;
  }, [state.historyOpen]);
  const onCommand = (command: Command) => {
    pending.current?.abort();
    if (command.type !== "evaluate") {
      dispatch({ type: "edit", command });
      return;
    }
    const controller = new AbortController();
    pending.current = controller;
    const id = ++nextRequest.current;
    const expression = state.editor.root;
    const { text, spans } = serialize(expression);
    dispatch({ type: "pending", id });
    void evaluate(text, controller.signal)
      .then((outcome) => {
        if (!controller.signal.aborted)
          dispatch({ type: "evaluated", id, outcome, expression, spans });
      })
      .catch(() => {
        if (!controller.signal.aborted)
          dispatch({
            type: "evaluated",
            id,
            outcome: { type: "failure", reason: "network_failure" },
            expression,
            spans,
          });
      });
  };
  const closeHistory = () => dispatch({ type: "history", open: false });
  useKeyboard(onCommand, state.historyOpen, closeHistory);
  const copy = async () => {
    if (!state.result) return;
    const version = state.resultVersion;
    try {
      await navigator.clipboard.writeText(state.result.decimal);
      dispatch({ type: "copied", version, success: true });
    } catch {
      dispatch({ type: "copied", version, success: false });
    }
  };
  return (
    <main className="calculator">
      <section className="sheet">
        <Header
          count={state.history.length}
          historyOpen={state.historyOpen}
          onToggleHistory={() =>
            dispatch({ type: "history", open: !state.historyOpen })
          }
          historyButton={historyButton}
        />
        <div className="workspace">
          <Equation
            editor={state.editor}
            onPlace={(path, index) => onCommand({ type: "place", path, index })}
            error={state.error}
          />
          <ResultCard
            result={state.result}
            dimmed={state.phase !== "result"}
            pending={state.phase === "pending"}
            expansion={state.expansion}
            copyStatus={state.copyStatus}
            onExpand={() => dispatch({ type: "expand" })}
            onCopy={() => void copy()}
          />
        </div>
      </section>
      <aside className="controls">
        {state.historyOpen ? (
          <HistoryPanel
            entries={state.history}
            confirmClear={state.confirmClear}
            onClose={closeHistory}
            onLoad={(entry) => {
              pending.current?.abort();
              dispatch({ type: "load", entry });
            }}
            onUse={(result) => {
              pending.current?.abort();
              dispatch({ type: "use", result });
            }}
            onConfirmClear={(confirm) =>
              dispatch({ type: "confirm-clear", confirm })
            }
            onClear={() => dispatch({ type: "clear-history" })}
          />
        ) : (
          <Keypad onCommand={onCommand} />
        )}
      </aside>
      <p role="status" className="sr-only">
        {state.copyStatus === "copied"
          ? "Full decimal copied."
          : state.copyStatus === "failed"
            ? "Could not copy. Try again."
            : state.phase === "pending"
              ? "Calculating…"
              : state.phase === "result" && state.result
                ? `${state.result.kind}: ${state.result.decimal}`
                : ""}
      </p>
    </main>
  );
}
export default App;
