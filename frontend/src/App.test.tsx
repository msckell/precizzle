import { renderToString } from "react-dom/server";
import { describe, expect, it } from "vitest";
import App from "./App.tsx";

describe("App", () => {
  it("renders the Precizzle heading", () => {
    expect(renderToString(<App />)).toContain("<h1>Precizzle</h1>");
  });
});
