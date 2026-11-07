import { useId } from "react";
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
  const id = useId().replace(/[^\w-]/g, "");
  const url = (name: string) => `url(#${id}-${name})`;
  return (
    <header className="header">
      <h1 className="brand">
        <svg viewBox="2 10 96 100" role="img" aria-label="P">
          <defs>
            <linearGradient
              id={`${id}-s`}
              gradientUnits="userSpaceOnUse"
              x1="0"
              y1="10"
              x2="0"
              y2="110"
            >
              <stop offset="0" stopColor="#3ED9AB" />
              <stop offset="1" stopColor="#12C488" />
            </linearGradient>
            <linearGradient
              id={`${id}-b`}
              gradientUnits="userSpaceOnUse"
              x1="0"
              y1="10"
              x2="0"
              y2="82"
            >
              <stop offset="0" stopColor="#FC7A78" />
              <stop offset=".5" stopColor="#FF7925" />
              <stop offset="1" stopColor="#FE9B09" />
            </linearGradient>
            <linearGradient
              id={`${id}-f`}
              gradientUnits="userSpaceOnUse"
              x1="13"
              y1="10"
              x2="30"
              y2="27"
            >
              <stop offset="0" stopColor="#FC7A78" />
              <stop offset="1" stopColor="#EE6072" />
            </linearGradient>
            <linearGradient
              id={`${id}-sh`}
              gradientUnits="userSpaceOnUse"
              x1="35"
              y1="0"
              x2="45"
              y2="0"
            >
              <stop offset="0" stopColor="#7a2a00" stopOpacity=".38" />
              <stop offset="1" stopColor="#7a2a00" stopOpacity="0" />
            </linearGradient>
            <linearGradient
              id={`${id}-cr`}
              gradientUnits="userSpaceOnUse"
              x1="24"
              y1="21"
              x2="32"
              y2="29"
            >
              <stop offset="0" stopColor="#0b5e45" stopOpacity=".32" />
              <stop offset="1" stopColor="#0b5e45" stopOpacity="0" />
            </linearGradient>
          </defs>
          <path
            d="M30 21 H52 C72 21 85 31 85 46 C85 61 72 71 52 71 H30"
            fill="none"
            stroke={url("b")}
            strokeWidth="22"
            strokeLinejoin="round"
          />
          <rect x="35" y="60" width="10" height="22" fill={url("sh")} />
          <path d="M13 34 L13 21 Q13 10 24 10 L37 10 Z" fill={url("f")} />
          <path d="M13 34 L37 10 L35 10 L35 103 L13 110 Z" fill={url("s")} />
          <path d="M13 34 L35 12 L35 32 L13 54 Z" fill={url("cr")} />
        </svg>
        recizzle
      </h1>
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
