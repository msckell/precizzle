import { describe, expect, it } from "vitest";
import { requests } from "./handlers.ts";

async function evaluate(expression: string) {
  const response = await fetch("/api/v1/evaluate", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ expression }),
  });
  return {
    status: response.status,
    type: response.headers.get("Content-Type"),
    body: await response.json(),
  };
}

describe("evaluate mock", () => {
  it("answers a SPEC success with its fixture", async () => {
    expect(await evaluate("200/15%")).toEqual({
      status: 200,
      type: "application/json",
      body: {
        value: "4000/3",
        decimal: "1333." + "3".repeat(100),
        kind: "repeating",
      },
    });
  });

  it("answers a SPEC error with Problem Details and its position", async () => {
    const { status, type, body } = await evaluate("1/0");
    expect(status).toBe(422);
    expect(type).toBe("application/problem+json");
    expect(body).toMatchObject({
      code: "division_by_zero",
      status: 422,
      position: 1,
    });
  });

  it("answers an expression without a fixture with internal_error", async () => {
    const { status, body } = await evaluate("1+1");
    expect(status).toBe(500);
    expect(body).toMatchObject({ code: "internal_error" });
    expect(body).not.toHaveProperty("position");
  });

  it("records each request", async () => {
    await evaluate("15%");
    expect(requests).toEqual([
      {
        url: "/api/v1/evaluate",
        method: "POST",
        contentType: "application/json",
        body: '{"expression":"15%"}',
      },
    ]);
  });

  it.each([
    ["(2+", "invalid_expression", 3],
    ["(2+3", "invalid_expression", 0],
    ["2 1.2.3", "invalid_expression", 2],
    ["sqrx(4)", "invalid_expression", 3],
    ["sqrt (4)", "invalid_expression", 4],
    ["((2+3", "invalid_expression", 1],
    ["1" + "+1".repeat(91) + "^1".repeat(11), "expression_too_complex", 201],
    ["1" + "^1".repeat(11) + "+1".repeat(90), "expression_too_complex", 21],
    ["200/0%", "division_by_zero", 3],
    ["1e39*1000%", "result_too_large", 9],
    ["1e39/0.1%", "result_too_large", 8],
    ["1e-1000*1%", "result_too_precise", 9],
  ] as const)(
    "locates %s according to D54 and D55",
    async (expression, code, position) => {
      const { body } = await evaluate(expression);
      expect(body).toMatchObject({ code, position });
      expect(body).not.toHaveProperty("field");
    },
  );

  it.each([
    ["200*15%", "30", "30", "exact"],
    ["200/15%", "4000/3", "1333." + "3".repeat(100), "repeating"],
    [
      "[1/3]/[~1.4142135623730950488]%",
      "~23.57022603955158414672",
      "23.57022603955158414672",
      "approximate",
    ],
    ["100*1e-999%", "0." + "0".repeat(998) + "1", "1e-999", "exact"],
  ])(
    "preserves D55's single operation for %s",
    async (expression, value, decimal, kind) => {
      const { status, body } = await evaluate(expression);
      expect(status).toBe(200);
      expect(body).toEqual({ value, decimal, kind });
    },
  );

  it.each([
    ["{}", "missing_field"],
    ['{"expression":null}', "missing_field"],
    ['{"expression":2}', "invalid_expression"],
  ])(
    "keeps request errors separate from expression positions: %s",
    async (body, code) => {
      const response = await fetch("/api/v1/evaluate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body,
      });
      expect(response.status).toBe(400);
      expect(response.headers.get("Content-Type")).toBe(
        "application/problem+json",
      );
      const answer: unknown = await response.json();
      expect(answer).toMatchObject({ code, field: "expression", status: 400 });
      expect(answer).not.toHaveProperty("position");
    },
  );
});
