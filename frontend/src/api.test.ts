import { describe, expect, it } from "vitest";
import { http, HttpResponse } from "msw/http";
import { server } from "./mocks/server.ts";
import { fixtures, requests, requestFixtures } from "./mocks/handlers.ts";
import { evaluate, isResult } from "./api.ts";
import { messageFor } from "./messages.ts";

describe("evaluate client", () => {
  it.each(["0.1+0.2", "200/15%", "[1/3]/[~1.4142135623730950488]%"])(
    "returns %s with string precision intact",
    async (expression) => {
      expect(await evaluate(expression)).toEqual({
        type: "result",
        result: fixtures[expression],
      });
      expect(requests).toEqual([
        {
          url: "/api/v1/evaluate",
          method: "POST",
          contentType: "application/json",
          body: JSON.stringify({ expression }),
        },
      ]);
    },
  );
  it.each(["(2+", "(2+3", "1/0", "unknown"])(
    "returns the Problem Details for %s",
    async (expression) => {
      const answer = await evaluate(expression);
      expect(answer.type).toBe("problem");
      if (answer.type !== "problem") throw new Error("Expected a problem");
      expect(answer.problem).toMatchObject(
        expression === "unknown"
          ? { code: "internal_error", status: 500 }
          : fixtures[expression],
      );
      if (expression === "unknown")
        expect(answer.problem).not.toHaveProperty("position");
    },
  );
  it("accepts a request-level error with field instead of position", async () => {
    const problem = requestFixtures['{"expression":2}'];
    server.use(
      http.post("/api/v1/evaluate", () =>
        HttpResponse.json(problem, {
          status: 400,
          headers: { "Content-Type": "application/problem+json" },
        }),
      ),
    );
    expect(await evaluate("2")).toEqual({ type: "problem", problem });
  });
  it.each([
    null,
    [],
    {},
    { value: 0.3, decimal: "0.3", kind: "exact" },
    { value: "0.3", decimal: 0.3, kind: "exact" },
    { value: "0.3", decimal: "0.3", kind: "rounded" },
  ])("rejects a malformed result %j", async (body) => {
    server.use(http.post("/api/v1/evaluate", () => HttpResponse.json(body)));
    expect(await evaluate("0.1+0.2")).toEqual({
      type: "failure",
      reason: "unexpected_response",
    });
    expect(isResult(body)).toBe(false);
  });
  it.each([
    { value: "", decimal: "", kind: "exact" },
    { value: "1+2", decimal: "3", kind: "exact" },
    { value: "1/0", decimal: "0", kind: "repeating" },
    { value: "1", decimal: "NaN", kind: "exact" },
    { value: "1.4", decimal: "1.4", kind: "approximate" },
    { value: "~1.4", decimal: "1.5", kind: "approximate" },
  ])("rejects invalid result text %j", (body) =>
    expect(isResult(body)).toBe(false),
  );
  it.each([
    { type: 1 },
    { title: null },
    { detail: false },
    { code: 123 },
    { status: 422 },
    { position: -1 },
    { position: 0.5 },
    { position: "0" },
    { position: null },
    { field: 123 },
    { field: "expression" },
  ])("rejects malformed Problem Details %j", async (overrides) => {
    server.use(
      http.post("/api/v1/evaluate", () =>
        HttpResponse.json(
          { ...fixtures["(2+"], ...overrides },
          {
            status: 400,
            headers: { "Content-Type": "application/problem+json" },
          },
        ),
      ),
    );
    expect(await evaluate("(2+")).toEqual({
      type: "failure",
      reason: "unexpected_response",
    });
  });
  it.each([
    [200, "text/html", "<html>oops</html>"],
    [200, "application/json", "{"],
    [400, "application/problem+json", "null"],
    [200, "application/problem+json", JSON.stringify(fixtures["(2+"])],
    [400, "application/json", JSON.stringify(fixtures["(2+"])],
    [201, "application/json", JSON.stringify(fixtures["0.1+0.2"])],
    [400, "application/problem+json", "{}"],
    [400, "application/problem+json", "[]"],
  ] as const)(
    "rejects unexpected HTTP response %i %s",
    async (status, contentType, body) => {
      server.use(
        http.post(
          "/api/v1/evaluate",
          () =>
            new HttpResponse(body, {
              status,
              headers: { "Content-Type": contentType },
            }),
        ),
      );
      expect(await evaluate("2")).toEqual({
        type: "failure",
        reason: "unexpected_response",
      });
    },
  );
  it("accepts a content type parameter", async () => {
    server.use(
      http.post("/api/v1/evaluate", () =>
        HttpResponse.json(fixtures["0.1+0.2"], {
          headers: { "Content-Type": "application/json; charset=utf-8" },
        }),
      ),
    );
    expect((await evaluate("0.1+0.2")).type).toBe("result");
  });
  it("returns network failure", async () => {
    server.use(http.post("/api/v1/evaluate", () => HttpResponse.error()));
    expect(await evaluate("2")).toEqual({
      type: "failure",
      reason: "network_failure",
    });
  });
  it("passes cancellation through without reporting a network failure", async () => {
    const controller = new AbortController();
    server.use(
      http.post("/api/v1/evaluate", () => {
        controller.abort();
        return HttpResponse.json(fixtures["0.1+0.2"]);
      }),
    );
    await expect(evaluate("0.1+0.2", controller.signal)).rejects.toMatchObject({
      name: "AbortError",
    });
  });
});

it.each([
  ["invalid_json", "request"],
  ["missing_field", "missing"],
  ["unknown_field", "field"],
  ["invalid_number", "number"],
  ["operand_out_of_range", "number"],
  ["exponent_not_integer", "integer"],
  ["exponent_out_of_range", "1000"],
  ["invalid_operator", "operator"],
  ["invalid_expression", "expression"],
  ["expression_too_complex", "100"],
  ["division_by_zero", "zero"],
  ["negative_square_root", "negative"],
  ["result_too_large", "result"],
  ["result_too_precise", "digits"],
  ["not_found", "service"],
  ["method_not_allowed", "request"],
  ["payload_too_large", "expression"],
  ["unsupported_media_type", "request"],
  ["internal_error", "again"],
  ["network_failure", "connection"],
  ["unexpected_response", "response"],
])("explains %s and offers a correction", (code, word) => {
  expect(messageFor(code).toLowerCase()).toContain(word);
  expect(messageFor(code).split(". ").length).toBeGreaterThanOrEqual(2);
});
it("falls back for unknown error codes", () =>
  expect(messageFor("new_code")).toBe(messageFor("unexpected_response")));
