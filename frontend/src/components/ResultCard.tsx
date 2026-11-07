import type { Result } from "../editor/model.ts";
import { expansionCaption, formatResult } from "../format.ts";
import type { Expansion } from "../format.ts";

interface Props {
  result?: Result;
  dimmed: boolean;
  pending: boolean;
  expansion: Expansion;
  copyStatus: "" | "copied" | "failed";
  onExpand: () => void;
  onCopy: () => void;
}
export function ResultCard({
  result,
  dimmed,
  pending,
  expansion,
  copyStatus,
  onExpand,
  onCopy,
}: Props) {
  if (!result)
    return (
      <section className="result-card empty-result" aria-label="Result">
        <b aria-hidden="true">=</b>
        {pending ? "Calculating…" : "Press = to calculate"}
      </section>
    );
  const display = formatResult(result, expansion);
  return (
    <section
      className={`result-card${dimmed ? " dimmed" : ""}`}
      aria-label="Result"
      aria-busy={pending}
    >
      <div className="result-heading">
        <span>Result</span>
        <span className={`kind-badge ${result.kind}`}>{result.kind}</span>
        <button
          type="button"
          className="copy-button"
          onClick={onCopy}
          aria-label="Copy full decimal"
        >
          <svg
            className="icon"
            aria-hidden="true"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
          >
            <rect x="8" y="8" width="13" height="13" rx="2" />
            <path d="M4 16H3V3h13v1" />
          </svg>
          {copyStatus === "copied" ? "Copied" : "Copy"}
        </button>
      </div>
      <div className="result-value" data-testid="result-value">
        {result.kind === "approximate" && (
          <span className="approximate-mark" aria-label="Approximately">
            ≈
          </span>
        )}
        <span>{display.mantissa}</span>
        {display.hasMore && (
          <button
            type="button"
            className="more-decimals"
            aria-label="Show more decimals"
            onClick={onExpand}
          >
            …
          </button>
        )}
        {display.exponent !== undefined && (
          <span className="scientific-exponent">
            {" "}
            × 10<sup>{display.exponent}</sup>
          </span>
        )}
      </div>
      {expansion > 0 && (
        <p className="decimal-caption">{expansionCaption(result, expansion)}</p>
      )}
      {copyStatus === "failed" && (
        <p className="copy-error">Could not copy. Try again.</p>
      )}
    </section>
  );
}
