import type { Result } from "./editor/model.ts";

export interface Problem {
  readonly type: string;
  readonly title: string;
  readonly status: number;
  readonly detail: string;
  readonly code: string;
  readonly position?: number;
  readonly field?: string;
}
export type Evaluation =
  | { readonly type: "result"; readonly result: Result }
  | { readonly type: "problem"; readonly problem: Problem }
  | {
      readonly type: "failure";
      readonly reason: "network_failure" | "unexpected_response";
    };

function object(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

export function isResult(value: unknown): value is Result {
  if (
    !object(value) ||
    typeof value.value !== "string" ||
    typeof value.decimal !== "string" ||
    !/^-?\d+(?:\.\d+)?(?:e[+-]?\d+)?$/.test(value.decimal)
  )
    return false;
  switch (value.kind) {
    case "exact":
      return /^-?\d+(?:\.\d+)?$/.test(value.value);
    case "repeating":
      return /^-?\d+\/[1-9]\d*$/.test(value.value);
    case "approximate":
      return value.value === `~${value.decimal}`;
    default:
      return false;
  }
}

function isProblem(value: unknown, status: number): value is Problem {
  return (
    object(value) &&
    ["type", "title", "detail", "code"].every(
      (key) => typeof value[key] === "string",
    ) &&
    value.status === status &&
    (value.position === undefined ||
      (Number.isSafeInteger(value.position) &&
        (value.position as number) >= 0)) &&
    (value.field === undefined || typeof value.field === "string") &&
    !(value.position !== undefined && value.field !== undefined)
  );
}

export async function evaluate(
  expression: string,
  signal?: AbortSignal,
): Promise<Evaluation> {
  let response: Response;
  try {
    response = await fetch("/api/v1/evaluate", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ expression }),
      signal,
    });
  } catch (error) {
    if (signal?.aborted) throw error;
    return { type: "failure", reason: "network_failure" };
  }
  let body: unknown;
  try {
    body = await response.json();
  } catch (error) {
    if (signal?.aborted) throw error;
    return { type: "failure", reason: "unexpected_response" };
  }
  const contentType = response.headers
    .get("Content-Type")
    ?.split(";")[0]
    .trim()
    .toLowerCase();
  if (
    response.status === 200 &&
    contentType === "application/json" &&
    isResult(body)
  )
    return { type: "result", result: body };
  if (
    response.status >= 400 &&
    response.status <= 599 &&
    contentType === "application/problem+json" &&
    isProblem(body, response.status)
  )
    return { type: "problem", problem: body };
  return { type: "failure", reason: "unexpected_response" };
}
