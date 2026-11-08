# Precizzle

A full-stack calculator for financial and accounting use. Every result is exact or explicitly marked as approximate: 0.1 + 0.2 returns 0.3, and 1 ÷ 3 × 3 returns 1.

- Frontend: React with TypeScript.
- Backend: a REST microservice in Go, the single source of truth for results.

Operations: addition, subtraction, multiplication, division, exponentiation, square root and percentage.

## Quick start

With Docker running, from the repository root:

```sh
docker build -t precizzle .
docker run --rm -d -p 127.0.0.1:8080:8080 --name precizzle precizzle
```

Open http://127.0.0.1:8080 for the calculator, or call the API with `curl` in a POSIX shell such as Git Bash:

```sh
curl -s http://127.0.0.1:8080/api/v1/evaluate -H 'Content-Type: application/json' -d '{"expression": "0.1+0.2"}'
# {"value":"0.3","decimal":"0.3","kind":"exact"}
```

`docker stop precizzle` stops and removes the container. [Docker](#docker) has the details. To run the tests, see [Development](#development) and [Tests and coverage](#tests-and-coverage).

## Development

Requirements: Go 1.27 and Node.js 24.15 or later with npm. Docker is needed only for the image.

Backend, from `backend/`:

```sh
go test ./...         # run the tests
go test -cover ./...  # run the tests with coverage per package
go run .              # start the API on http://127.0.0.1:8080
```

Frontend, from `frontend/`:

```sh
npm ci              # install dependencies
npm test            # run the tests
npm run coverage    # run the tests with coverage
npm run lint        # ESLint
npm run typecheck   # TypeScript
npm run build       # production build in dist/
npm run dev         # start the dev server
```

### Known issues

- `npm audit` in `frontend/` reports a moderate advisory in `minimatch` ([GHSA-g3hv-75wg-m249](https://github.com/advisories/GHSA-g3hv-75wg-m249), a denial of service through crafted glob patterns, fixed in 11.0.1).
  - It reaches the project only through ESLint and typescript-eslint, which are development tools: it is not part of the built app, and the only glob patterns it processes are the project's own lint configuration.
  - ESLint 10.14 and typescript-eslint 8.72 require `minimatch` 10, and no patched 10.x release exists.
  - Forcing version 11 with npm `overrides` works in a local test, but it goes outside the versions ESLint supports, so it is not applied.
  - The advisory goes away when ESLint moves to `minimatch` 11.

## Tests and coverage

Statement coverage, measured with `go test -cover ./...` in `backend/` and `npm run coverage` in `frontend/` (D92):

| Layer | Scope | Coverage | CI minimum |
|---|---|---|---|
| Backend | `internal/calc`: number model and arithmetic | 100.0% | 95% |
| Backend | `internal/expr`: expression parser and evaluator | 99.5% | 95% |
| Backend | `internal/api`: HTTP handlers | 100.0% | 90% |
| Backend | `main`: server startup and static files | 89.5% | — |
| Frontend | `frontend/src` without `main.tsx`: 408 tests | 99.01% | 90% |

CI runs both on every pull request and on `main`. The Backend job writes the per-package table to the run summary, and the Frontend job prints the Vitest coverage report; each fails when a minimum is not met. The [frontend](#tests-and-ci) and [backend](#tests-and-ci-1) sections describe what the tests cover.

## Docker

The [`Dockerfile`](Dockerfile) builds one image that runs the whole app: the Go server answers the API and also serves the built frontend. Only Docker is needed. From the repository root, in bash or PowerShell:

```sh
docker build -t precizzle .
docker run --rm -d -p 127.0.0.1:8080:8080 --name precizzle precizzle
```

Open http://127.0.0.1:8080 for the app; the API answers on the same address, so the [API examples](#examples) work unchanged. The port is published only on the loopback address, so other machines on the network cannot reach it.

```sh
docker ps --filter name=precizzle   # STATUS shows (healthy) once GET /healthz answers
docker logs precizzle               # server log, empty unless an error occurs
docker stop precizzle               # stop and remove the container
```

- The build has three stages: Node builds the frontend, Go builds a static binary, and the final `alpine` image (about 23 MB) holds only the binary and the frontend build.
- The server runs as the non-root user `precizzle`.
- Inside the container it listens on `:8080` (`PRECIZZLE_ADDR`) and serves the frontend from the directory in `PRECIZZLE_STATIC_DIR`. Without that variable, as in development, the server answers only the API.
- The `docker` job in CI builds the image, starts it, waits until it is healthy and checks that `0.1+0.2` returns `0.3` and that `/` serves the app.

## Frontend

The React and TypeScript app in `frontend/` is a calculator built around an editable equation. It does no arithmetic: every result comes from `POST /api/v1/evaluate`.

### Running it

The backend and the dev server each keep running in their own terminal. Open two terminals at the repository root.

In the first one, start the backend and leave it running:

```sh
cd backend
go run .              # API on http://127.0.0.1:8080
```

In the second one, start the frontend:

```sh
cd frontend
npm ci                # first time, or after the dependencies change
npm run dev           # app on http://localhost:5173
```

The Vite dev server proxies `/api` to `127.0.0.1:8080`, so the backend must be running before the first calculation. `npm run build` writes the production app to `frontend/dist/`.

### Using it

- The calculation is an expression with a cursor. Click or tap between characters to place it, and type anywhere in the expression.
- ÷ draws a stacked fraction and takes the number before the cursor as its numerator; xʸ takes it as the base; √ inserts an empty root. A visible × is added where a √, a `(` or a reused result would otherwise touch a number.
- Backspace or Delete next to a fraction, root or exponent first selects all of it, and a second press deletes it.
- After a result, an operator continues from it, and a digit starts a new calculation.
- The history keeps the last 100 calculations in the browser. "Load" restores an expression and its result, "Use" inserts a result at the cursor, and "Clear history" asks for confirmation first.

The physical keyboard works anywhere on the page, except while the history is open:

| Key | Action |
|---|---|
| `0`–`9`, `.` | Digits and the decimal point |
| `+`, `-`, `*`, `%`, `(`, `)` | +, −, ×, %, `(` and `)` |
| `/` | ÷ (fraction) |
| `^` | xʸ |
| `r` | √ |
| Enter, `=` | Evaluate |
| Backspace, Delete | Delete before or after the cursor |
| Arrow keys, Home, End | Move the cursor; Up and Down move between numerator and denominator, and into and out of an exponent |
| Escape | Clear the expression, or close the history |

Keys pressed with Ctrl, Meta or Alt are left to the browser.

### Design notes

- **Equation.** A white page, the expression drawn as math (stacked fractions, roots, raised exponents) and the result in its own mint card. A purple box follows the part being edited. Long expressions stay on one line and scroll sideways.
- **Exact values in chips.** Chaining or reusing a result inserts a chip that shows the short decimal but keeps the exact `value`, such as `1/3`, so 1 ÷ 3 × 3 is exactly 1. Approximate chips are orange and marked with ≈.
- **Display rules.** Results are truncated, never rounded. Repeating results show 4 decimals and other long results 5, followed by "…"; each click on "…" shows 20, 40, 60, 80 and then 100 digits. Copy puts the full `decimal` on the clipboard. Tiny results use scientific notation.
- **Errors in place.** An error marks the part of the expression at fault, for example the fraction bar of a division by zero, and explains how to fix it.
- **Accessibility.** Every key has an accessible name, focus is visible, the expression is read as a whole, and results and errors are announced to screen readers.

### Tests and CI

From `frontend/`, `npm test` runs Vitest. Unit tests cover the editor, the serializer and its position map, the result formatting, the API client and the history storage. Integration tests render the app with React Testing Library and user-event, and Mock Service Worker answers `evaluate` from fixtures that match the real API. `npm run coverage` fails below 90% of statements.

The `frontend` job in [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs `npm ci`, ESLint, the Prettier check, the type check, the tests with coverage and the build on every pull request and on `main`.

## Backend API

The Go backend in `backend/` has no dependencies outside the standard library. The full contract, with every rule and error, is in [docs/SPEC.md](docs/SPEC.md#4-api-contract), and the expression rules are in [section 5](docs/SPEC.md#5-expressions).

### Running it

From `backend/`, `go run .` starts the API on `http://127.0.0.1:8080`, reachable only from this machine. The environment variable `PRECIZZLE_ADDR` sets another address:

```sh
PRECIZZLE_ADDR=127.0.0.1:9090 go run .           # bash
$env:PRECIZZLE_ADDR = "127.0.0.1:9090"; go run .  # PowerShell
```

During development the frontend reaches the API through the Vite dev server proxy, so the backend needs no CORS headers. The proxy is set in [`frontend/vite.config.ts`](frontend/vite.config.ts).

### Endpoints

| Endpoint | Request body | Computes |
|---|---|---|
| `POST /api/v1/add` | `{"a", "b"}` | a + b |
| `POST /api/v1/subtract` | `{"a", "b"}` | a − b |
| `POST /api/v1/multiply` | `{"a", "b"}` | a × b |
| `POST /api/v1/divide` | `{"a", "b"}` | a ÷ b |
| `POST /api/v1/power` | `{"base", "exponent"}` | base ^ exponent, with an integer exponent from −1000 to 1000 |
| `POST /api/v1/sqrt` | `{"value"}` | √value |
| `POST /api/v1/percentage` | `{"percent", "base"?, "operator"?}` | percent ÷ 100, or base `add`, `subtract`, `multiply` or `divide` percent% |
| `POST /api/v1/evaluate` | `{"expression"}` | an expression with several operations, see [Expressions](#expressions) |
| `GET /healthz` | none | `{"status": "ok"}` when the server is up |

- Requests use `Content-Type: application/json`, and every number is a JSON string: `"0.1"`, `"1/3"`, `"1e-3"` or an approximate `"~1.4142135623730950488"`.
- A success returns `{"value", "decimal", "kind"}`:
  - `value` is the exact result, to send back when chaining.
  - `decimal` is for display.
  - `kind` is `exact`, `repeating` or `approximate`.

### Expressions

`evaluate` computes an ASCII expression with the usual precedence. The frontend sends every calculation through it.

- **Tokens:**
  - Numbers without a sign (`12`, `0.5`, `1e-3`).
  - Values in brackets, with any operand form (`[1/7]`, `[-2.5]`, `[~1.4142135623730950488]`). This is how a result goes back into an expression.
  - The operators `+ - * / ^ %`, `sqrt(…)` and parentheses. Spaces may separate tokens.
- **Precedence:**
  - `^` groups from the right: `2^3^2` = 512.
  - Unary minus binds below it: `-2^2` = −4.
  - Then `*` and `/`, and then `+` and `-`, both from the left.
  - There is no implicit multiplication: `2(3)` is an error.
- **Percentages:** they follow the operand they belong to.
  - `50+200+15%` = 50 + 230 = 280.
  - `200+15%+10%` = 253.
  - `(50+200)+15%` = 287.5.
  - `200*15%` = 30, the same as the `percentage` endpoint.
- **Every step follows the rules of the single operations**, so `10^50/10^20` is `result_too_large` and `sqrt(2)*sqrt(2)` is `~2`.
- **Limits:** at most 100 operations, 10 of them powers, and 100 nested levels. Breaking a limit is `expression_too_complex`.

### Examples

The examples use `curl` in a POSIX shell such as Git Bash. In the responses, `…` stands for digits removed here; the real `decimal` of a repeating result has 100 fractional digits.

```sh
curl -s http://127.0.0.1:8080/api/v1/add -H 'Content-Type: application/json' -d '{"a": "0.1", "b": "0.2"}'
# {"value":"0.3","decimal":"0.3","kind":"exact"}

curl -s http://127.0.0.1:8080/api/v1/divide -H 'Content-Type: application/json' -d '{"a": "1", "b": "3"}'
# {"value":"1/3","decimal":"0.3333…3333","kind":"repeating"}

curl -s http://127.0.0.1:8080/api/v1/multiply -H 'Content-Type: application/json' -d '{"a": "1/3", "b": "3"}'
# {"value":"1","decimal":"1","kind":"exact"}

curl -s http://127.0.0.1:8080/api/v1/sqrt -H 'Content-Type: application/json' -d '{"value": "2"}'
# {"value":"~1.4142135623730950488","decimal":"1.4142135623730950488","kind":"approximate"}

curl -s http://127.0.0.1:8080/api/v1/percentage -H 'Content-Type: application/json' -d '{"percent": "15", "base": "200", "operator": "add"}'
# {"value":"230","decimal":"230","kind":"exact"}

curl -s http://127.0.0.1:8080/healthz
# {"status":"ok"}
```

Expressions:

```sh
curl -s http://127.0.0.1:8080/api/v1/evaluate -H 'Content-Type: application/json' -d '{"expression": "50+200+15%"}'
# {"value":"280","decimal":"280","kind":"exact"}

curl -s http://127.0.0.1:8080/api/v1/evaluate -H 'Content-Type: application/json' -d '{"expression": "1/3*3"}'
# {"value":"1","decimal":"1","kind":"exact"}

curl -s http://127.0.0.1:8080/api/v1/evaluate -H 'Content-Type: application/json' -d '{"expression": "1/7"}'
# {"value":"1/7","decimal":"0.1428…1429","kind":"repeating"}

curl -s http://127.0.0.1:8080/api/v1/evaluate -H 'Content-Type: application/json' -d '{"expression": "[1/7]*7"}'
# {"value":"1","decimal":"1","kind":"exact"}

curl -s http://127.0.0.1:8080/api/v1/evaluate -H 'Content-Type: application/json' -d '{"expression": "sqrt(2)*sqrt(2)"}'
# {"value":"~2","decimal":"2","kind":"approximate"}
```

### Errors

Errors use RFC 9457 Problem Details (`application/problem+json`) with a stable `code` and, when one field is at fault, its name in `field`:

```sh
curl -s http://127.0.0.1:8080/api/v1/divide -H 'Content-Type: application/json' -d '{"a": "1", "b": "0"}'
# HTTP 422
# {"type":"https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#division_by_zero","title":"Division by zero","status":422,"detail":"The divisor b is zero.","code":"division_by_zero","field":"b"}
```

An error about an expression, or about one of its operations, has `position` instead of `field`: the 0-based index of the character at fault.

```sh
curl -s http://127.0.0.1:8080/api/v1/evaluate -H 'Content-Type: application/json' -d '{"expression": "(2+"}'
# HTTP 400
# {"type":"https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#invalid_expression","title":"Invalid expression","status":400,"detail":"The expression cannot continue at position 3.","code":"invalid_expression","position":3}
```

- `400` means the request must be fixed, for example `invalid_number`, `missing_field` or `invalid_expression`.
- `422` means the result does not exist or cannot be represented, for example `division_by_zero` or `result_too_large`.
- `500 internal_error` means a bug in the server.

The [errors table](docs/SPEC.md#errors) lists every code.

### Tests and CI

- **Tests:** `go test ./...` runs the table-driven tests. They cover every example and error in the SPEC, the HTTP layer through `httptest`, and two fuzz targets: one for the operand parser and one for the expression evaluator.
- **Benchmark:** the costliest expression that fits in a request takes about 0.85 s on an Intel i5-8400 (`go test -run='^$' -bench=EvaluateWorstCase ./internal/expr`).
- **CI:** [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every pull request and on `main`. It checks formatting with gofmt, runs `go vet`, runs the tests with the race detector, and fails when coverage drops below 95% in `internal/calc` or `internal/expr`, or 90% in `internal/api`.

## Design decisions

The full list, with the reason for each decision, is in the [decisions log](docs/SPEC.md#decisions-log). The main ones:

- **No binary floating point.** Numbers travel as JSON strings, and neither the backend nor the frontend ever stores an operand or a result in a floating-point number.
- **Exact rationals.** A result keeps its exact value: `1 ÷ 3` returns the value `1/3`, so chaining `× 3` returns exactly `1`. The API stays stateless, because the client sends that value back.
- **Every result states its kind.** A result is `exact` (a terminating decimal), `repeating` (an exact fraction) or `approximate` (an irrational root, or a result computed from an approximate value). Approximate values carry a `~` prefix, so the approximation follows them through every later calculation.
- **Display is separate from value.** Every response also has a `decimal` for display: 100 fractional digits for repeating results and 20 for approximate ones, rounded half even. The UI truncates it and expands it on demand, up to 100 digits. Tiny non-zero results switch to scientific notation and never become 0.
- **Pocket-calculator percentages.** 200 + 15% = 230, 200 − 15% = 170, 200 × 15% = 30, 200 ÷ 15% = 1333.33…
- **Standard errors.** Errors use RFC 9457 Problem Details with a stable machine-readable `code`.

## Project documents

- [docs/prompts/00-master-prompt.md](docs/prompts/00-master-prompt.md): the master prompt that defined the product, the workflow and the plan in blocks.
- [docs/SPEC.md](docs/SPEC.md): product definition, number model, API contract, decisions log and out of scope list.
- [docs/plans/](docs/plans/): the approved plan of each block.
- [PROMPTS.md](PROMPTS.md): the prompts that drove the project.
- [Dockerfile](Dockerfile) and [`.github/workflows/ci.yml`](.github/workflows/ci.yml): the image and the CI checks.
