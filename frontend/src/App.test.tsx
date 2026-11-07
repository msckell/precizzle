import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";
import App from "./App.tsx";
import { requests } from "./mocks/handlers.ts";
import { http, HttpResponse } from "msw/http";
import { server } from "./mocks/server.ts";

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

describe("results", () => {
  it("expands repeating decimals through every step and copies the full decimal", async () => {
    const user = userEvent.setup();
    render(<App />);
    await user.keyboard("1/3=");
    await waitFor(() =>
      expect(screen.getByTestId("result-value")).toHaveTextContent("0.3333…"),
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
