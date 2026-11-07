import { http, HttpResponse } from "msw/http";

type Kind = "exact" | "repeating" | "approximate";

interface Success {
  value: string;
  decimal: string;
  kind: Kind;
}

interface Problem {
  type: string;
  title: string;
  status: number;
  detail: string;
  code: string;
  position?: number;
  field?: string;
}

export interface RecordedRequest {
  url: string;
  method: string;
  contentType: string | null;
  body: string;
}

const titles: Record<string, [string, number]> = {
  missing_field: ["Missing field", 400],
  invalid_expression: ["Invalid expression", 400],
  invalid_number: ["Invalid number", 400],
  expression_too_complex: ["Expression too complex", 400],
  exponent_not_integer: ["Exponent not integer", 400],
  exponent_out_of_range: ["Exponent out of range", 400],
  division_by_zero: ["Division by zero", 422],
  negative_square_root: ["Negative square root", 422],
  result_too_large: ["Result too large", 422],
  result_too_precise: ["Result too precise", 422],
  internal_error: ["Internal error", 500],
};

const details: Record<string, (position: number) => string> = {
  invalid_expression: (position) =>
    `The expression cannot continue at position ${position}.`,
  invalid_number: (position) =>
    `The number at position ${position} does not match the operand syntax.`,
  expression_too_complex: (position) =>
    `The expression is too complex at position ${position}: it allows at most 100 operations, 10 powers and 100 nested levels.`,
  division_by_zero: (position) =>
    `The expression divides by zero at position ${position}.`,
  negative_square_root: (position) =>
    `The square root at position ${position} receives a negative value.`,
  exponent_not_integer: (position) =>
    `The exponent of the power at position ${position} is not an exact integer.`,
  exponent_out_of_range: (position) =>
    `The exponent of the power at position ${position} is not between -1000 and 1000.`,
  result_too_large: (position) =>
    `The absolute value of the result at position ${position} is 10^40 or more.`,
  result_too_precise: (position) =>
    `The result at position ${position} needs more digits than the limits allow.`,
};

function problem(
  code: string,
  position?: number,
  detail = position === undefined
    ? "The server could not complete the calculation."
    : details[code](position),
  field?: string,
): Problem {
  const [title, status] = titles[code];
  return {
    type: `https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#${code}`,
    title,
    status,
    detail,
    code,
    ...(position === undefined ? {} : { position }),
    ...(field === undefined ? {} : { field }),
  };
}

function ok(value: string, decimal: string, kind: Kind): Success {
  return { value, decimal, kind };
}

// Every row of SPEC §5 "Expression examples", answered as the real API answers it.
// The long decimals are built from their digits instead of being written out.
export const fixtures: Record<string, Success | Problem> = {
  "0.1+0.2": ok("0.3", "0.3", "exact"),
  "2+3*4": ok("14", "14", "exact"),
  "(2+3)*4": ok("20", "20", "exact"),
  "2^3^2": ok("512", "512", "exact"),
  "-2^2": ok("-4", "-4", "exact"),
  "2^-2": ok("0.25", "0.25", "exact"),
  "1/3*3": ok("1", "1", "exact"),
  "[1/7]*7": ok("1", "1", "exact"),
  "sqrt(9/4)": ok("1.5", "1.5", "exact"),
  "sqrt(2)*sqrt(2)": ok("~2", "2", "approximate"),
  "200+15%": ok("230", "230", "exact"),
  "200-15%": ok("170", "170", "exact"),
  "200*15%": ok("30", "30", "exact"),
  "200/15%": ok("4000/3", "1333." + "3".repeat(100), "repeating"),
  "15%": ok("0.15", "0.15", "exact"),
  "50+200+15%": ok("280", "280", "exact"),
  "50-200+15%": ok("-180", "-180", "exact"),
  "200+15%+10%": ok("253", "253", "exact"),
  "(50+200)+15%": ok("287.5", "287.5", "exact"),
  "15%+200": ok("200.15", "200.15", "exact"),
  "200+-15%": ok("170", "170", "exact"),
  "1/0": problem("division_by_zero", 1),
  "0^-1": problem("division_by_zero", 1),
  "sqrt(-4)": problem("negative_square_root", 0),
  "2^0.5": problem("exponent_not_integer", 1),
  "2^1001": problem("exponent_out_of_range", 1),
  "10^50/10^20": problem("result_too_large", 2),
  "2+": problem("invalid_expression", 2),
  "(2+3": problem("invalid_expression", 0),
  "(2+": problem("invalid_expression", 3),
  "2(3)": problem("invalid_expression", 1),
  "15%%": problem("invalid_expression", 3),
  "1.2.3": problem("invalid_number", 0),
  "2 1.2.3": problem("invalid_expression", 2),
  "sqrx(4)": problem("invalid_expression", 3),
  "sqrt (4)": problem("invalid_expression", 4),
  "((2+3": problem("invalid_expression", 1),
  ["1" + "+1".repeat(91) + "^1".repeat(11)]: problem(
    "expression_too_complex",
    201,
  ),
  ["1" + "^1".repeat(11) + "+1".repeat(90)]: problem(
    "expression_too_complex",
    21,
  ),
  "[1/3]/[~1.4142135623730950488]%": ok(
    "~23.57022603955158414672",
    "23.57022603955158414672",
    "approximate",
  ),
  "100*1e-999%": ok("0." + "0".repeat(998) + "1", "1e-999", "exact"),
  "200/0%": problem("division_by_zero", 3),
  "1e39*1000%": problem("result_too_large", 9),
  "1e39/0.1%": problem("result_too_large", 8),
  "1e-1000*1%": problem("result_too_precise", 9),
  "((1)/(7))": ok("1/7", "0." + "142857".repeat(16) + "1429", "repeating"),
  "((1)/(3))": ok("1/3", "0." + "3".repeat(100), "repeating"),
  "[1/3]*3": ok("1", "1", "exact"),
  "((1)/(2))^(2)": ok("0.25", "0.25", "exact"),
  "(([1/3])/[1/3]%)": ok("100", "100", "exact"),
  "(([1/3])/-[1/3]%)": ok("-100", "-100", "exact"),
  "(([1/3])/[~1.4142135623730950488]%)": ok(
    "~23.57022603955158414672",
    "23.57022603955158414672",
    "approximate",
  ),
  "(([1/3])/-[~1.4142135623730950488]%)": ok(
    "~-23.57022603955158414672",
    "-23.57022603955158414672",
    "approximate",
  ),
  "(([1/3])/--[~1.4142135623730950488]%)": ok(
    "~23.57022603955158414672",
    "23.57022603955158414672",
    "approximate",
  ),
  "(([1/3])/(([~1.4142135623730950488]%)))": ok(
    "~23.57022603955158414339",
    "23.57022603955158414339",
    "approximate",
  ),
  "(([1/3])/(-([~1.4142135623730950488]%)))": ok(
    "~-23.57022603955158414339",
    "-23.57022603955158414339",
    "approximate",
  ),
  "(([1/3])/([~1.4142135623730950488])%)": ok(
    "~23.57022603955158414672",
    "23.57022603955158414672",
    "approximate",
  ),
};

// Request failures have no expression position, even for invalid_expression.
export const requestFixtures: Record<string, Problem> = {
  "{}": problem(
    "missing_field",
    undefined,
    "The field expression is required.",
    "expression",
  ),
  '{"expression":null}': problem(
    "missing_field",
    undefined,
    "The field expression is required.",
    "expression",
  ),
  '{"expression":2}': problem(
    "invalid_expression",
    undefined,
    "The field expression must be a JSON string.",
    "expression",
  ),
};

export const requests: RecordedRequest[] = [];

export const handlers = [
  http.post("/api/v1/evaluate", async ({ request }) => {
    const body = await request.text();
    requests.push({
      url: new URL(request.url).pathname,
      method: request.method,
      contentType: request.headers.get("Content-Type"),
      body,
    });
    const { expression } = JSON.parse(body) as { expression: string };
    // A missing fixture is a test bug, so it fails loudly as a server error.
    const answer =
      requestFixtures[body] ??
      fixtures[expression] ??
      problem("internal_error");
    if ("code" in answer) {
      return HttpResponse.json(answer, {
        status: answer.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }
    return HttpResponse.json(answer);
  }),
];
