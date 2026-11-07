import type { Ref } from "react";

interface Props {
  count: number;
  historyOpen: boolean;
  onToggleHistory: () => void;
  historyButton: Ref<HTMLButtonElement>;
}
export function Header({
  count,
  historyOpen,
  onToggleHistory,
  historyButton,
}: Props) {
  return (
    <header className="header">
      <h1 className="brand">Precizzle</h1>
      <button
        ref={historyButton}
        type="button"
        className={`pill history-button${historyOpen ? " on" : ""}`}
        aria-label={`History, ${count} calculation${count === 1 ? "" : "s"}`}
        aria-expanded={historyOpen}
        onClick={onToggleHistory}
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
          <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
          <path d="M3 3v5h5M12 7v5l4 2" />
        </svg>
        <span className="history-label">
          History{count > 0 && <b>{count}</b>}
        </span>
      </button>
    </header>
  );
}
