import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterAll, afterEach, beforeAll } from "vitest";
import { requests } from "../mocks/handlers.ts";
import { server } from "../mocks/server.ts";

beforeAll(() => {
  server.listen({ onUnhandledFrame: "error" });
});

// Vitest runs without globals, so Testing Library cannot register its own cleanup.
afterEach(() => {
  cleanup();
  server.resetHandlers();
  requests.length = 0;
});

afterAll(() => {
  server.close();
});
