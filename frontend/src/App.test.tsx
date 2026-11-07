import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App.tsx";
import { requests } from "./mocks/handlers.ts";

afterEach(() => {
  localStorage.clear();
  vi.restoreAllMocks();
});
describe("calculator input", () => {
  it("renders the Precizzle heading", () => {
    render(<App />);
    expect(
      screen.getByRole("heading", { name: "Precizzle" }),
    ).toBeInTheDocument();
  });
  it.each(["keyboard", "keypad"])(
    "sends the same decimal calculation with the %s",
    async (method) => {
      const user = userEvent.setup();
      render(<App />);
      if (method === "keyboard") await user.keyboard("0.1+0.2=");
      else
        for (const label of [
          "0",
          "Decimal point",
          "1",
          "Add",
          "0",
          "Decimal point",
          "2",
          "Evaluate",
        ])
          await user.click(screen.getByRole("button", { name: label }));
      await waitFor(() => expect(requests).toHaveLength(1));
      expect(requests[0].body).toBe('{"expression":"0.1+0.2"}');
      await waitFor(() =>
        expect(screen.getByRole("status")).toHaveTextContent(/0.3/),
      );
    },
  );
  it("inserts visible multiplication before a root", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);
    await user.keyboard("2r");
    expect(
      screen.getByRole("group", {
        name: /Expression: 2 times square root of empty/,
      }),
    ).toBeInTheDocument();
    expect(container.querySelector(".operator")).toHaveTextContent("×");
  });
  it("selects then deletes a structure with backspace", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);
    await user.keyboard("1/3{ArrowRight}{Backspace}");
    expect(container.querySelector(".fraction.selected")).toBeInTheDocument();
    await user.keyboard("{Backspace}");
    expect(container.querySelector(".fraction")).not.toBeInTheDocument();
  });
  it("ignores modified keys, composition, and unsupported input", () => {
    render(<App />);
    for (const modifiers of [
      { ctrlKey: true },
      { altKey: true },
      { metaKey: true },
      { isComposing: true },
    ])
      fireEvent.keyDown(window, { key: "2", ...modifiers });
    fireEvent.keyDown(window, { key: "e" });
    expect(
      screen.getByRole("group", { name: "Expression: empty" }),
    ).toBeInTheDocument();
  });
  it("leaves Enter and Space with a focused button", async () => {
    const user = userEvent.setup();
    render(<App />);
    screen.getByRole("button", { name: "2" }).focus();
    await user.keyboard("{Enter} ");
    expect(
      screen.getByRole("group", { name: "Expression: 22" }),
    ).toBeInTheDocument();
    expect(requests).toHaveLength(0);
  });
});
