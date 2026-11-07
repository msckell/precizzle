import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App.tsx";
import { requests } from "./mocks/handlers.ts";
import { http, HttpResponse } from "msw/http";
import { server } from "./mocks/server.ts";
import { HISTORY_KEY, load, save } from "./history.ts";
import { chars } from "./test/editor.ts";

afterEach(() => {
  localStorage.clear();
  vi.restoreAllMocks();
});
describe("calculator input", () => {
  it("renders the ribbon P brand and hides the decorations", () => {
    const { container } = render(<App />);
    const heading = screen.getByRole("heading", { name: "Precizzle" });
    expect(within(heading).getByRole("img", { name: "P" })).toBeInTheDocument();
    expect(container.querySelector(".decorations")).toHaveAttribute(
      "aria-hidden",
      "true",
    );
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
        expect(screen.getByRole("status")).toHaveTextContent(/^exact: 0\.3$/),
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
  it("moves through structures with the arrow keys, Home and End", async () => {
    const user = userEvent.setup();
    render(<App />);
    const expression = (name: string) =>
      expect(
        screen.getByRole("group", { name: `Expression: ${name}` }),
      ).toBeInTheDocument();
    await user.keyboard("1/2{ArrowUp}5");
    expression("fraction 15 over 2 end fraction");
    await user.keyboard("{ArrowDown}{End}{ArrowRight}^2{ArrowDown}+3");
    expression("fraction 15 over 2 end fraction squared plus 3");
    await user.keyboard("{ArrowLeft}{ArrowLeft}{ArrowUp}0");
    expression(
      "fraction 15 over 2 end fraction to the power of 20 end exponent plus 3",
    );
    await user.keyboard("{ArrowDown}{Home}{Delete}");
    expect(document.querySelector(".fraction.selected")).toBeInTheDocument();
    await user.keyboard("{Escape}");
    expression("empty");
  });
  it("edits with the sign, power and clear keys", async () => {
    const user = userEvent.setup();
    render(<App />);
    for (const label of ["7", "Toggle sign", "Power", "2"])
      await user.click(screen.getByRole("button", { name: label }));
    expect(
      screen.getByRole("group", { name: "Expression: minus 7 squared" }),
    ).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Clear expression" }));
    expect(
      screen.getByRole("group", { name: "Expression: empty" }),
    ).toBeInTheDocument();
  });
  it("dims the result when the cursor is placed in the evaluated expression", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);
    await user.keyboard("0.1+0.2=");
    await screen.findByTestId("result-value");
    fireEvent.click(container.querySelector(".equation .digit")!, {
      clientX: 0,
    });
    expect(screen.getByRole("region", { name: "Result" })).toHaveClass(
      "dimmed",
    );
    await user.keyboard("3");
    expect(
      screen.getByRole("group", { name: "Expression: 03.1 plus 0.2" }),
    ).toBeInTheDocument();
  });
});

describe("results", () => {
  it("expands repeating decimals through every step and copies the full decimal", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("1/3=");
    await waitFor(() =>
      expect(screen.getByTestId("result-value")).toHaveTextContent("0.3333…"),
    );
    expect(screen.getByRole("status")).toHaveTextContent(
      /^repeating: 0\.3333…$/,
    );
    expect(screen.queryByText(/Showing/)).not.toBeInTheDocument();
    const write = vi.spyOn(navigator.clipboard, "writeText");
    await user.click(screen.getByRole("button", { name: "Copy full decimal" }));
    expect(write).toHaveBeenCalledWith("0." + "3".repeat(100));
    expect(screen.getByRole("status")).toHaveTextContent("Full decimal copied");
    for (const count of [20, 40, 60, 80, 100]) {
      await user.click(
        screen.getByRole("button", { name: "Show more decimals" }),
      );
      expect(
        screen.getByText(`Showing ${count} of 100 decimals`),
      ).toBeInTheDocument();
      expect(screen.getByTestId("result-value").textContent).toBe(
        "0." + "3".repeat(count) + (count < 100 ? "…" : ""),
      );
    }
    expect(
      screen.queryByRole("button", { name: "Show more decimals" }),
    ).not.toBeInTheDocument();
  });
  it("chains the exact value, not the displayed decimal", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("1/3=");
    await screen.findByTestId("result-value");
    await user.keyboard("*3=");
    await waitFor(() =>
      expect(screen.getByTestId("result-value")).toHaveTextContent(/^1$/),
    );
    expect(requests.at(-1)?.body).toBe('{"expression":"[1/3]*3"}');
  });
  it("dims the result when moving the cursor and starts a fresh expression with a digit", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("0.1+0.2=");
    await screen.findByTestId("result-value");
    await user.keyboard("{ArrowLeft}");
    expect(screen.getByRole("region", { name: "Result" })).toHaveClass(
      "dimmed",
    );
    await user.keyboard("=");
    await waitFor(() =>
      expect(screen.getByRole("region", { name: "Result" })).not.toHaveClass(
        "dimmed",
      ),
    );
    await user.keyboard("7");
    expect(
      screen.getByRole("group", { name: "Expression: 7" }),
    ).toBeInTheDocument();
    expect(screen.queryByTestId("result-value")).not.toBeInTheDocument();
  });
  it("marks approximate results and stops expansion at their last digit", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("r2=");
    await waitFor(() =>
      expect(screen.getByTestId("result-value")).toHaveTextContent("≈1.41421…"),
    );
    expect(screen.getByText("approximate")).toBeInTheDocument();
    expect(screen.getByRole("status")).toHaveTextContent(
      /^approximate: 1\.41421…$/,
    );
    await user.click(
      screen.getByRole("button", { name: "Show more decimals" }),
    );
    expect(screen.getByText("Showing 19 of 19 decimals")).toBeInTheDocument();
    expect(
      screen.queryByRole("button", { name: "Show more decimals" }),
    ).not.toBeInTheDocument();
  });
  it("reports a clipboard failure without changing the calculation", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("0.1+0.2=");
    await screen.findByTestId("result-value");
    vi.spyOn(navigator.clipboard, "writeText").mockRejectedValue(
      new Error("Denied"),
    );
    await user.click(screen.getByRole("button", { name: "Copy full decimal" }));
    expect(screen.getByRole("status")).toHaveTextContent("Could not copy");
    expect(screen.getByTestId("result-value")).toHaveTextContent("0.3");
  });
});

describe("errors", () => {
  it("marks the failing fraction bar and keeps the expression for correction", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);
    await user.keyboard("250+1/(3-3)=");
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "divides by zero",
    );
    expect(
      container.querySelector(".fraction-bar[data-error]"),
    ).toBeInTheDocument();
    expect(
      screen.getByRole("group", {
        name: /Expression: 250 plus fraction 1 over/,
      }),
    ).toBeInTheDocument();
    await user.keyboard("{Backspace}");
    expect(screen.queryByRole("alert")).not.toBeInTheDocument();
  });
  it("marks the expression end when an operand is missing", async () => {
    const user = userEvent.setup();
    const { container } = render(<App />);
    await user.keyboard("(2+=");
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "expression is incomplete",
    );
    expect(container.querySelector(".end-error")).toBeInTheDocument();
  });
  it.each(["network", "malformed"])(
    "shows a useful %s failure without marking an unrelated element",
    async (kind) => {
      server.use(
        http.post("/api/v1/evaluate", () =>
          kind === "network"
            ? HttpResponse.error()
            : HttpResponse.json({ value: 3 }),
        ),
      );
      const user = userEvent.setup();
      const { container } = render(<App />);
      await user.keyboard("2=");
      expect(await screen.findByRole("alert")).toHaveTextContent(
        kind === "network" ? "Check your connection" : "unexpected response",
      );
      expect(container.querySelector("[data-error]")).not.toBeInTheDocument();
    },
  );
});

const historyButton = () => screen.getByRole("button", { name: /^History/ });
const entries = () =>
  within(screen.getByRole("region", { name: "History" })).getAllByRole(
    "listitem",
  );

describe("history", () => {
  it("loads an entry with its structures and result, then chains it", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("1/3=");
    await screen.findByTestId("result-value");
    await user.keyboard("0.1+0.2=");
    await waitFor(() =>
      expect(historyButton()).toHaveAccessibleName("History, 2 calculations"),
    );
    await user.click(historyButton());
    expect(historyButton()).toHaveAttribute("aria-expanded", "true");
    expect(
      screen.queryByRole("button", { name: "Evaluate" }),
    ).not.toBeInTheDocument();
    expect(entries()).toHaveLength(2);
    expect(entries()[0]).toHaveTextContent("= 0.3");
    expect(
      within(entries()[1]).getByRole("group", {
        name: "Expression: fraction 1 over 3 end fraction",
      }),
    ).toBeInTheDocument();
    await user.click(
      within(entries()[1]).getByRole("button", { name: "Load" }),
    );
    expect(
      screen.queryByRole("region", { name: "History" }),
    ).not.toBeInTheDocument();
    expect(historyButton()).toHaveFocus();
    expect(
      screen.getByRole("group", {
        name: "Expression: fraction 1 over 3 end fraction",
      }),
    ).toBeInTheDocument();
    expect(screen.getByTestId("result-value")).toHaveTextContent("0.3333…");
    await user.keyboard("*3=");
    await waitFor(() =>
      expect(screen.getByTestId("result-value")).toHaveTextContent(/^1$/),
    );
    expect(requests.at(-1)?.body).toBe('{"expression":"[1/3]*3"}');
  });
  it.each(["1500*", "1500"])(
    "uses an approximate result as a chip at the cursor after %s",
    async (typed) => {
      const user = userEvent.setup();
      const { container } = render(<App />);
      await user.keyboard("r2=");
      await screen.findByTestId("result-value");
      await user.keyboard(typed);
      await user.click(historyButton());
      expect(entries()[0]).toHaveTextContent("Approx.≈1.41421…");
      await user.click(
        within(entries()[0]).getByRole("button", { name: "Use" }),
      );
      expect(historyButton()).toHaveFocus();
      expect(screen.getByText("≈1.41421…")).toHaveClass("result-chip");
      expect(container.querySelectorAll(".equation .operator")).toHaveLength(1);
      await user.keyboard("=");
      await waitFor(() =>
        expect(screen.getByTestId("result-value")).toHaveTextContent(
          "≈2121.32034…",
        ),
      );
      expect(requests.at(-1)?.body).toBe(
        '{"expression":"1500*[~1.4142135623730950488]"}',
      );
    },
  );
  it("asks before clearing and empties both the list and the storage", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("0.1+0.2=");
    await screen.findByTestId("result-value");
    await user.click(historyButton());
    await user.click(screen.getByRole("button", { name: "Clear history" }));
    expect(
      screen.getByRole("group", { name: "Confirm clearing the history" }),
    ).toHaveTextContent("Delete 1 calculation?");
    expect(screen.getByRole("button", { name: "Cancel" })).toHaveFocus();
    await user.click(screen.getByRole("button", { name: "Cancel" }));
    expect(entries()).toHaveLength(1);
    expect(screen.getByRole("button", { name: "Clear history" })).toHaveFocus();
    await user.click(screen.getByRole("button", { name: "Clear history" }));
    await user.click(screen.getByRole("button", { name: "Delete all" }));
    expect(screen.queryByRole("listitem")).not.toBeInTheDocument();
    expect(screen.getByText(/No calculations yet/)).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Close history" })).toHaveFocus();
    expect(historyButton()).toHaveAccessibleName("History, 0 calculations");
    expect(load()).toEqual([]);
    expect(JSON.parse(localStorage.getItem(HISTORY_KEY)!)).toEqual({
      version: 1,
      entries: [],
    });
  });
  it("ignores calculator keys while open and closes with Escape", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("12");
    await user.click(historyButton());
    expect(screen.getByRole("button", { name: "Close history" })).toHaveFocus();
    await user.keyboard("5+r={Backspace}");
    expect(requests).toHaveLength(0);
    await user.keyboard("{Escape}");
    expect(
      screen.queryByRole("region", { name: "History" }),
    ).not.toBeInTheDocument();
    expect(historyButton()).toHaveFocus();
    expect(
      screen.getByRole("group", { name: "Expression: 12" }),
    ).toBeInTheDocument();
    await user.click(historyButton());
    await user.click(screen.getByRole("button", { name: "Close history" }));
    expect(historyButton()).toHaveFocus();
    await user.click(historyButton());
    await user.click(historyButton());
    expect(
      screen.queryByRole("region", { name: "History" }),
    ).not.toBeInTheDocument();
  });
  it("keeps the history across reloads", async () => {
    const user = userEvent.setup();
    const { unmount } = render(<App />);
    await user.keyboard("0.1+0.2=");
    await screen.findByTestId("result-value");
    unmount();
    render(<App />);
    expect(historyButton()).toHaveAccessibleName("History, 1 calculation");
    await user.click(historyButton());
    expect(entries()[0]).toHaveTextContent("= 0.3");
  });
});

describe("superseded requests", () => {
  const slow = () => {
    let respond = () => {};
    const signals: AbortSignal[] = [];
    server.use(
      http.post("/api/v1/evaluate", async ({ request }) => {
        signals.push(request.signal);
        await new Promise<void>((resolve) => (respond = resolve));
        return HttpResponse.json({ value: "14", decimal: "14", kind: "exact" });
      }),
    );
    return { signals, respond: () => respond() };
  };
  const settle = () => new Promise((resolve) => setTimeout(resolve, 20));
  it.each(["Load", "Use"])(
    "cancels a pending evaluation on %s",
    async (action) => {
      save([
        {
          expression: chars("0.1+0.2"),
          result: { value: "0.3", decimal: "0.3", kind: "exact" },
        },
      ]);
      const pending = slow();
      const user = userEvent.setup();
      render(<App />);
      await user.keyboard("2*=");
      await waitFor(() => expect(pending.signals).toHaveLength(1));
      expect(screen.getByRole("status")).toHaveTextContent("Calculating…");
      await user.click(historyButton());
      await user.click(
        within(entries()[0]).getByRole("button", { name: action }),
      );
      expect(pending.signals[0].aborted).toBe(true);
      pending.respond();
      await settle();
      expect(screen.queryByText("14")).not.toBeInTheDocument();
      if (action === "Load")
        expect(screen.getByTestId("result-value")).toHaveTextContent(/^0.3$/);
      else
        expect(
          screen.getByRole("group", {
            name: "Expression: 2 times 0.3",
          }),
        ).toBeInTheDocument();
      expect(load()).toHaveLength(1);
    },
  );
  it("drops the pending result when the expression is edited", async () => {
    const pending = slow();
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("2+3*4=");
    await waitFor(() => expect(pending.signals).toHaveLength(1));
    await user.keyboard("5");
    expect(pending.signals[0].aborted).toBe(true);
    pending.respond();
    await settle();
    expect(screen.queryByTestId("result-value")).not.toBeInTheDocument();
    expect(
      screen.getByRole("group", { name: "Expression: 2 plus 3 times 45" }),
    ).toBeInTheDocument();
  });
});
