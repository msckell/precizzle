# Precizzle

A full-stack calculator for financial and accounting use. Every result is exact or explicitly marked as approximate: 0.1 + 0.2 returns 0.3, and 1 ÷ 3 × 3 returns 1.

- Frontend: React with TypeScript.
- Backend: a REST microservice in Go, the single source of truth for results.

Operations: addition, subtraction, multiplication, division, exponentiation, square root and percentage.

## Status

Work in progress. Setup and run instructions and API examples are added as each block lands.

## Development

Requirements: Go 1.27 and Node.js 24 with npm.

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

## Backend API

The Go backend in `backend/` has no dependencies outside the standard library. The full contract, with every rule and error, is in [docs/SPEC.md](docs/SPEC.md#4-api-contract).

### Running it

From `backend/`, `go run .` starts the API on `http://127.0.0.1:8080`, reachable only from this machine. The environment variable `PRECIZZLE_ADDR` sets another address:

```sh
PRECIZZLE_ADDR=127.0.0.1:9090 go run .           # bash
$env:PRECIZZLE_ADDR = "127.0.0.1:9090"; go run .  # PowerShell
```

During development the frontend reaches the API through the Vite dev server proxy, so the backend needs no CORS headers. Block 2 configures the proxy.

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
| `GET /healthz` | none | `{"status": "ok"}` when the server is up |

- Requests use `Content-Type: application/json`, and every number is a JSON string: `"0.1"`, `"1/3"`, `"1e-3"` or an approximate `"~1.4142135623730950488"`.
- A success returns `{"value", "decimal", "kind"}`:
  - `value` is the exact result, to send back when chaining.
  - `decimal` is for display.
  - `kind` is `exact`, `repeating` or `approximate`.

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

### Errors

Errors use RFC 9457 Problem Details (`application/problem+json`) with a stable `code` and, when one field is at fault, its name in `field`:

```sh
curl -s http://127.0.0.1:8080/api/v1/divide -H 'Content-Type: application/json' -d '{"a": "1", "b": "0"}'
# HTTP 422
# {"type":"https://github.com/msckell/precizzle/blob/main/docs/SPEC.md#division_by_zero","title":"Division by zero","status":422,"detail":"The divisor b is zero.","code":"division_by_zero","field":"b"}
```

- `400` means the request must be fixed, for example `invalid_number` or `missing_field`.
- `422` means the result does not exist or cannot be represented, for example `division_by_zero` or `result_too_large`.
- `500 internal_error` means a bug in the server.

The [errors table](docs/SPEC.md#errors) lists every code.

### Tests and CI

- **Tests:** `go test ./...` runs the table-driven tests. They cover every example and error in the SPEC, a fuzz target for the operand parser, and the HTTP layer through `httptest`.
- **CI:** [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs on every pull request and on `main`. It checks formatting with gofmt, runs `go vet`, runs the tests with the race detector, and fails when coverage drops below 95% in `internal/calc` or 90% in `internal/api`.

## Design decisions

The full list, with the reason for each decision, is in the [decisions log](docs/SPEC.md#decisions-log). The main ones:

- **No binary floating point.** Numbers travel as JSON strings, and neither the backend nor the frontend ever stores an operand or a result in a floating-point number.
- **Exact rationals.** A result keeps its exact value: `1 ÷ 3` returns the value `1/3`, so chaining `× 3` returns exactly `1`. The API stays stateless, because the client sends that value back.
- **Every result states its kind.** A result is `exact` (a terminating decimal), `repeating` (an exact fraction) or `approximate` (an irrational root, or a result computed from an approximate value). Approximate values carry a `~` prefix, so the approximation follows them through every later calculation.
- **Display is separate from value.** Every response also has a `decimal` for display: 100 fractional digits for repeating results and 20 for approximate ones, rounded half even. The UI truncates it and expands it on demand, up to 100 digits. Tiny non-zero results switch to scientific notation and never become 0.
- **Pocket-calculator percentages.** 200 + 15% = 230, 200 − 15% = 170, 200 × 15% = 30, 200 ÷ 15% = 1333.33…
- **Standard errors.** Errors use RFC 9457 Problem Details with a stable machine-readable `code`.

## Project documents

- [docs/SPEC.md](docs/SPEC.md): product definition, number model, API contract, decisions log and out of scope list.
- [docs/plans/](docs/plans/): the approved plan of each block.
- [PROMPTS.md](PROMPTS.md): the prompts that drove the project.
