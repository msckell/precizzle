import { fireEvent, render, screen } from "@testing-library/react";
import { expect, it, vi } from "vitest";
import { Equation } from "./Equation.tsx";
import { chars, rootTwo, state, third } from "../test/editor.ts";
import type { EditorState } from "../editor/model.ts";

it("reads the expression as a whole and renders nested editable slots", () => {
  const editor: EditorState = {
    root: [
      { type: "frac", num: chars("2"), den: [{ type: "sqrt", body: [] }] },
      { type: "pow", exp: chars("3") },
      { type: "chip", ...third },
      { type: "chip", ...rootTwo },
    ],
    cursor: {
      path: [
        { index: 0, slot: "den" },
        { index: 0, slot: "body" },
      ],
      index: 0,
    },
  };
  const { container } = render(<Equation editor={editor} onPlace={vi.fn()} />);
  expect(
    screen.getByRole("group", {
      name: /Expression: fraction 2 over square root of empty/,
    }),
  ).toBeInTheDocument();
  expect(container.querySelectorAll("[data-caret]")).toHaveLength(1);
  expect(container.querySelector(".empty-slot.active")).toBeInTheDocument();
  expect(screen.getByText("0.3333…")).toBeInTheDocument();
  expect(screen.getByText("≈1.41421…")).toBeInTheDocument();
});
it.each([
  [11, 0],
  [29, 1],
])(
  "places a cursor on the nearer side of a character (%i)",
  (clientX, index) => {
    const onPlace = vi.fn();
    render(<Equation editor={state("2")} onPlace={onPlace} />);
    const digit = screen.getByText("2");
    vi.spyOn(digit, "getBoundingClientRect").mockReturnValue({
      left: 10,
      width: 20,
    } as DOMRect);
    fireEvent.click(digit, { clientX });
    expect(onPlace).toHaveBeenCalledWith([], index);
  },
);
it("places the cursor at the end of a clicked slot", () => {
  const onPlace = vi.fn();
  const { container } = render(
    <Equation
      editor={state([{ type: "frac", num: chars("12"), den: [] }])}
      onPlace={onPlace}
    />,
  );
  fireEvent.click(container.querySelectorAll("[data-row]")[1]);
  expect(onPlace).toHaveBeenCalledWith([{ index: 0, slot: "num" }], 2);
});
it("marks a selected structure and the fraction bar at fault", () => {
  const editor = {
    ...state([{ type: "frac", num: chars("1"), den: chars("0") }]),
    selected: { path: [], index: 0 },
  };
  const { container } = render(
    <Equation
      editor={editor}
      onPlace={vi.fn()}
      error={{
        message: "Cannot divide by zero. Change the denominator.",
        target: { type: "bar", path: [], index: 0 },
      }}
    />,
  );
  expect(container.querySelector(".fraction.selected")).toBeInTheDocument();
  expect(
    container.querySelector(".fraction-bar[data-error]"),
  ).toBeInTheDocument();
  expect(screen.getByRole("alert")).toHaveTextContent("Cannot divide by zero");
});
it("renders a history equation without a cursor or editing focus", () => {
  const { container } = render(<Equation editor={state("(2+3)*4%")} compact />);
  expect(container.querySelector("[data-caret]")).not.toBeInTheDocument();
  expect(screen.getByRole("group")).not.toHaveAttribute("tabindex");
});
