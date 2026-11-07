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
});
