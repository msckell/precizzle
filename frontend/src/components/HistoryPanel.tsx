import { useEffect, useRef } from "react";
import type { Result } from "../editor/model.ts";
import { formatResult } from "../format.ts";
import type { HistoryEntry } from "../history.ts";
import { Equation } from "./Equation.tsx";

interface Props {
  entries: readonly HistoryEntry[];
  confirmClear: boolean;
  onClose: () => void;
  onLoad: (entry: HistoryEntry) => void;
  onUse: (result: Result) => void;
  onConfirmClear: (confirm: boolean) => void;
  onClear: () => void;
}
const plural = (count: number) =>
  `${count} calculation${count === 1 ? "" : "s"}`;

export function HistoryPanel({
  entries,
  confirmClear,
  onClose,
  onLoad,
  onUse,
  onConfirmClear,
  onClear,
}: Props) {
  const close = useRef<HTMLButtonElement>(null);
  const cancel = useRef<HTMLButtonElement>(null);
  const clear = useRef<HTMLButtonElement>(null);
  const asked = useRef(false);
  useEffect(() => {
    close.current?.focus();
  }, []);
  useEffect(() => {
    if (confirmClear) cancel.current?.focus();
    else if (asked.current) (clear.current ?? close.current)?.focus();
    asked.current = confirmClear;
  }, [confirmClear]);
  return (
    <section className="history" aria-labelledby="history-title">
      <div className="history-heading">
        <h2 id="history-title">History</h2>
        <button
          ref={close}
          type="button"
          className="icon-button"
          aria-label="Close history"
          onClick={onClose}
        >
          <svg
            className="icon"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            strokeLinecap="round"
            aria-hidden="true"
          >
            <path d="M18 6 6 18M6 6l12 12" />
          </svg>
        </button>
      </div>
      <p className="history-note">Saved on this device, up to 100.</p>
      {entries.length === 0 ? (
        <p className="history-empty">
          No calculations yet. Results appear here after you press =.
        </p>
      ) : (
        <ol className="history-list">
          {entries.map((entry, index) => {
            const display = formatResult(entry.result);
            return (
              <li key={index} className="history-entry">
                <div className="history-calculation">
                  <Equation
                    editor={{
                      root: entry.expression,
                      cursor: { path: [], index: entry.expression.length },
                    }}
                    compact
                  />
                  <p className="history-value">
                    {entry.result.kind === "approximate" ? (
                      <>
                        <span className="approximate-tag">Approx.</span>
                        <span className="approximate-mark">≈</span>
                      </>
                    ) : (
                      "= "
                    )}
                    {display.text}
                  </p>
                </div>
                <button
                  type="button"
                  className="history-action"
                  onClick={() => onLoad(entry)}
                >
                  Load
                </button>
                <button
                  type="button"
                  className="history-action use"
                  onClick={() => onUse(entry.result)}
                >
                  Use
                </button>
              </li>
            );
          })}
        </ol>
      )}
      {confirmClear ? (
        <div
          className="clear-confirmation"
          role="group"
          aria-label="Confirm clearing the history"
        >
          <p>
            Delete all {plural(entries.length)}?
            <span>
              They are removed from this device and can&apos;t be recovered.
            </span>
          </p>
          <div className="confirmation-buttons">
            <button
              ref={cancel}
              type="button"
              onClick={() => onConfirmClear(false)}
            >
              Cancel
            </button>
            <button type="button" className="delete-all" onClick={onClear}>
              Delete all
            </button>
          </div>
        </div>
      ) : (
        entries.length > 0 && (
          <div className="history-footer">
            {plural(entries.length)}
            <button
              ref={clear}
              type="button"
              className="clear-history"
              onClick={() => onConfirmClear(true)}
            >
              <svg
                className="icon"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M3 6h18M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
              </svg>
              Clear history
            </button>
          </div>
        )
      )}
    </section>
  );
}
