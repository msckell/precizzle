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

function App() {
  const [state, dispatch] = useReducer(reducer, undefined, initialState);
  const pending = useRef<AbortController | undefined>(undefined);
  const nextRequest = useRef(0);
  useEffect(() => () => pending.current?.abort(), []);
  useEffect(() => {
    save(state.history);
  }, [state.history]);
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
  useKeyboard(onCommand, state.historyOpen, () =>
    dispatch({ type: "history", open: false }),
  );
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
        <h1>Precizzle</h1>
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
        <Keypad onCommand={onCommand} />
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
