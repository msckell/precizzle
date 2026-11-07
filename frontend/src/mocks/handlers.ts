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
}

export interface RecordedRequest {
  url: string;
  method: string;
  contentType: string | null;
  body: string;
}

const titles: Record<string, [string, number]> = {
  invalid_expression: ["Invalid expression", 400],
  invalid_number: ["Invalid number", 400],
  exponent_not_integer: ["Exponent not integer", 400],
  exponent_out_of_range: ["Exponent out of range", 400],
  division_by_zero: ["Division by zero", 422],
  negative_square_root: ["Negative square root", 422],
  result_too_large: ["Result too large", 422],
  internal_error: ["Internal error", 500],
};

function problem(code: string, position?: number): Problem {
  const [title, status] = titles[code];
  return {
    type: `https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#${code}`,
    title,
    status,
    detail: `${title}.`,
    code,
    ...(position === undefined ? {} : { position }),
  };
}

function ok(value: string, decimal: string, kind: Kind): Success {
  return { value, decimal, kind };
}

// Every row of SPEC §5 "Expression examples", answered as the real API answers it.
// The long decimals are built from their digits instead of being written out.
const fixtures: Record<string, Success | Problem> = {
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
  "2(3)": problem("invalid_expression", 1),
  "15%%": problem("invalid_expression", 3),
  "1.2.3": problem("invalid_number", 0),
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
    const answer = fixtures[expression] ?? problem("internal_error");
    if ("code" in answer) {
      return HttpResponse.json(answer, {
        status: answer.status,
        headers: { "Content-Type": "application/problem+json" },
      });
    }
    return HttpResponse.json(answer);
  }),
];
