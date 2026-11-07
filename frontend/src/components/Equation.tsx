import { Fragment, useLayoutEffect, useRef } from "react";
import type { MouseEvent, ReactNode } from "react";
import type { EditorState, Path, Row, Slot } from "../editor/model.ts";
import { isNumberChar } from "../editor/model.ts";
import type { Element } from "../editor/serialize.ts";
import { speak } from "../editor/speak.ts";
import { chipText } from "../format.ts";

export interface EquationError {
  message: string;
  target?: Element;
}
interface Props {
  editor: EditorState;
  onPlace?: (path: Path, index: number) => void;
  error?: EquationError;
  compact?: boolean;
}
const samePath = (a: Path, b: Path) => JSON.stringify(a) === JSON.stringify(b);
const glyph = { "+": "+", "-": "−", "*": "×", "%": "%" };

export function Equation({ editor, onPlace, error, compact = false }: Props) {
  const frame = useRef<HTMLDivElement>(null);
  const scroller = useRef<HTMLDivElement>(null);
  const callout = useRef<HTMLParagraphElement>(null);
  const updateEdges = () => {
    const view = scroller.current!;
    frame.current!.dataset.leftFade = String(view.scrollLeft > 1);
    frame.current!.dataset.rightFade = String(
      view.scrollWidth - view.clientWidth - view.scrollLeft > 1,
    );
  };
  useLayoutEffect(() => {
    const view = scroller.current!;
    const update = () => {
      const target = view.querySelector<HTMLElement>(
        error ? "[data-error]" : "[data-caret]",
      );
      if (target) {
        const box = target.getBoundingClientRect();
        const viewport = view.getBoundingClientRect();
        if (box.left < viewport.left + 12)
          view.scrollLeft += box.left - viewport.left - 12;
        else if (box.right > viewport.right - 12)
          view.scrollLeft += box.right - viewport.right + 12;
        if (callout.current) {
          const after = target.getBoundingClientRect();
          const max = Math.max(
            0,
            view.clientWidth - callout.current.offsetWidth,
          );
          callout.current.style.marginLeft = `${Math.max(0, Math.min(max, after.left - viewport.left - callout.current.offsetWidth / 2 + after.width / 2))}px`;
        }
      }
      updateEdges();
    };
    update();
    const observer =
      typeof ResizeObserver === "undefined"
        ? undefined
        : new ResizeObserver(update);
    observer?.observe(view);
    window.addEventListener("resize", update);
    return () => {
      observer?.disconnect();
      window.removeEventListener("resize", update);
    };
  }, [editor, error]);
  const place = (event: MouseEvent, path: Path, index: number) => {
    event.stopPropagation();
    if (onPlace) {
      event.currentTarget
        .closest<HTMLElement>(".equation-frame")
        ?.focus({ preventScroll: true });
      onPlace(path, index);
    }
  };
  const marked = (target: Element) =>
    error?.target && JSON.stringify(error.target) === JSON.stringify(target);
  const renderRow = (row: Row, path: Path): ReactNode => {
    const current = !!onPlace && samePath(path, editor.cursor.path);
    let caretRendered = false;
    const caret = (index: number) => {
      if (!current || editor.cursor.index !== index || caretRendered)
        return null;
      caretRendered = true;
      return <span key={`caret-${index}`} className="caret" data-caret="" />;
    };
    const pieces: ReactNode[] = [];
    const renderNode = (index: number): ReactNode => {
      const node = row[index];
      const nodeError = marked({ type: "node", path, index });
      const selected =
        editor.selected?.index === index &&
        samePath(editor.selected.path, path);
      const child = (slot: Slot): Path => [...path, { index, slot }];
      const click = (event: MouseEvent<HTMLElement>) => {
        const box = event.currentTarget.getBoundingClientRect();
        place(
          event,
          path,
          index + (event.clientX >= box.left + box.width / 2 ? 1 : 0),
        );
      };
      let content: ReactNode;
      let className = "";
      switch (node.type) {
        case "char":
          className =
            node.char in glyph
              ? "operator"
              : node.char === "(" || node.char === ")"
                ? "parenthesis"
                : "digit";
          content =
            node.char in glyph
              ? glyph[node.char as keyof typeof glyph]
              : node.char;
          break;
        case "chip":
          className = `result-chip ${node.kind}`;
          content = chipText(node);
          break;
        case "frac":
          className = "fraction";
          content = (
            <>
              {renderRow(node.num, child("num"))}
              <span
                className="fraction-bar"
                data-error={
                  marked({ type: "bar", path, index }) ? "" : undefined
                }
              />
              {renderRow(node.den, child("den"))}
            </>
          );
          break;
        case "sqrt":
          className = "square-root";
          content = (
            <>
              <svg viewBox="0 0 24 100" preserveAspectRatio="none">
                <path d="M1 62 L6 56 L12 96 L24 0" />
              </svg>
              <span className="radicand">
                {renderRow(node.body, child("body"))}
              </span>
            </>
          );
          break;
        case "pow":
          className = "power";
          content = renderRow(node.exp, child("exp"));
          break;
      }
      return (
        <span
          key={index}
          className={`${className}${selected ? " selected" : ""}`}
          data-error={nodeError ? "" : undefined}
          onClick={click}
        >
          {content}
        </span>
      );
    };
    for (let i = 0; i < row.length;) {
      if (isNumberChar(row[i])) {
        const start = i;
        const number: ReactNode[] = [];
        while (i < row.length && isNumberChar(row[i])) {
          number.push(caret(i), renderNode(i));
          i++;
        }
        number.push(caret(i));
        pieces.push(
          <span
            key={`number-${start}`}
            className={`number-run${current && path.length === 0 && editor.cursor.index >= start && editor.cursor.index <= i ? " active" : ""}`}
          >
            {number}
          </span>,
        );
      } else {
        pieces.push(caret(i), renderNode(i));
        i++;
      }
    }
    pieces.push(caret(row.length));
    const slotError = marked({ type: "slot", path });
    return (
      <span
        className={`equation-row${row.length === 0 ? " empty-slot" : ""}${current && (path.length > 0 || !row.length) ? " active" : ""}`}
        data-row={JSON.stringify(path)}
        data-error={slotError ? "" : undefined}
        onClick={(event) => place(event, path, row.length)}
      >
        {pieces}
      </span>
    );
  };
  return (
    <div className={`equation-area${compact ? " compact" : ""}`}>
      <div
        ref={frame}
        className="equation-frame"
        role="group"
        aria-label={`Expression: ${speak(editor.root)}`}
        tabIndex={onPlace ? 0 : undefined}
      >
        <div ref={scroller} className="equation-scroll" onScroll={updateEdges}>
          <div className="equation" aria-hidden="true">
            {renderRow(editor.root, [])}
            {error?.target?.type === "end" && (
              <Fragment>
                <span className="end-error" data-error="" />
              </Fragment>
            )}
          </div>
        </div>
      </div>
      {error && (
        <p ref={callout} className="error-callout" role="alert">
          {error.message}
        </p>
      )}
    </div>
  );
}
