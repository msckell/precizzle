import type { Command } from "../state.ts";
import type { Digit } from "../editor/model.ts";

const operators: { label: string; glyph: string; command: Command }[] = [
  { label: "Add", glyph: "+", command: { type: "operator", operator: "+" } },
  {
    label: "Subtract",
    glyph: "−",
    command: { type: "operator", operator: "-" },
  },
  {
    label: "Multiply",
    glyph: "×",
    command: { type: "operator", operator: "*" },
  },
  { label: "Fraction", glyph: "÷", command: { type: "fraction" } },
  {
    label: "Open parenthesis",
    glyph: "(",
    command: { type: "paren", paren: "(" },
  },
  {
    label: "Close parenthesis",
    glyph: ")",
    command: { type: "paren", paren: ")" },
  },
  { label: "Power", glyph: "xʸ", command: { type: "power" } },
  { label: "Square root (R)", glyph: "√", command: { type: "sqrt" } },
  {
    label: "Percent",
    glyph: "%",
    command: { type: "operator", operator: "%" },
  },
];
export function Keypad({
  onCommand,
}: {
  onCommand: (command: Command) => void;
}) {
  const key = (
    label: string,
    content: React.ReactNode,
    command: Command,
    className = "key",
  ) => (
    <button
      key={label}
      type="button"
      className={className}
      aria-label={label}
      onClick={() => onCommand(command)}
    >
      {content}
    </button>
  );
  return (
    <div className="keypad" aria-label="Calculator keypad">
      <p className="key-label">Operation</p>
      <div className="operator-keys">
        {operators.map((op) =>
          key(
            op.label,
            op.command.type === "fraction" ? (
              <span className="fraction-key" aria-hidden="true">
                <i />
                <b />
                <i />
              </span>
            ) : (
              op.glyph
            ),
            op.command,
            "operator-key",
          ),
        )}
      </div>
      <p className="key-label">Number</p>
      <div className="number-keys">
        {["7", "8", "9", "4", "5", "6", "1", "2", "3"].map((digit) =>
          key(digit, digit, { type: "digit", digit: digit as Digit }),
        )}
        {key("Toggle sign", "±", { type: "sign" })}
        {key("0", "0", { type: "digit", digit: "0" })}
        {key("Decimal point", ".", { type: "dot" })}
      </div>
      <div className="action-keys">
        {key("Clear expression", "C", { type: "clear" }, "key clear-key")}
        {key(
          "Backspace",
          <svg
            className="icon"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="1.8"
            aria-hidden="true"
          >
            <path d="M9 5h12v14H9l-7-7Z" />
            <path d="m12 9 6 6m0-6-6 6" />
          </svg>,
          { type: "backspace" },
        )}
        {key("Evaluate", "=", { type: "evaluate" }, "equals-key")}
      </div>
    </div>
  );
}
